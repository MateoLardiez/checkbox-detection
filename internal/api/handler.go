package api

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/MateoLardiez/checkbox-detection/internal/detector"
)

type detectResponse struct {
	Boxes []detector.Box `json:"boxes"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// maxUploadBytes caps the whole request body, so the uploaded image is slightly smaller.
const maxUploadBytes = 50 << 20

func DetectHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)

	file, _, err := r.FormFile("file")
	if err != nil {
		if isTooLarge(err) {
			writeTooLarge(w)
			return
		}
		writeError(w, http.StatusBadRequest, "expected a multipart file in field 'file'")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		if isTooLarge(err) {
			writeTooLarge(w)
			return
		}
		writeError(w, http.StatusBadRequest, "could not read the uploaded file")
		return
	}
	if len(data) == 0 {
		writeError(w, http.StatusBadRequest, "the uploaded file is empty")
		return
	}

	boxes, err := detector.Detect(data)
	if errors.Is(err, detector.ErrInvalidImage) {
		writeError(w, http.StatusBadRequest, "the uploaded file is not a valid image")
		return
	}
	if err != nil {
		log.Printf("detect failed: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, detectResponse{Boxes: boxes})
}

func isTooLarge(err error) bool {
	var tooLarge *http.MaxBytesError
	return errors.As(err, &tooLarge)
}

func writeTooLarge(w http.ResponseWriter) {
	writeError(w, http.StatusRequestEntityTooLarge, "the uploaded file is larger than 50 MB")
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("write response failed: %v", err)
	}
}
