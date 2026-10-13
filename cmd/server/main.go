package main

import (
	"log"
	"net/http"
	"os"

	"github.com/MateoLardiez/checkbox-detection/internal/api"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /detect", api.DetectHandler)

	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
