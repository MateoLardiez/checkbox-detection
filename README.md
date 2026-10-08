# Checkbox Detection API

HTTP service that detects checkboxes in a document image and classifies each one as checked or unchecked.

## API

### `POST /detect`

**Input:** a document image, sent as a file upload (`multipart/form-data`).

**Output:**

```json
{
  "boxes": [
    { "bbox": [x1, y1, x2, y2], "is_checked": true },
    { "bbox": [x1, y1, x2, y2], "is_checked": false }
  ]
}
```

- `bbox` is in pixels of the uploaded image: top-left `(x1, y1)` and bottom-right `(x2, y2)`.
- An image with no checkboxes returns `{"boxes": []}`.
- Invalid input (empty file, not an image) returns a 4xx with an error message.

**Example:**

```bash
# TODO: fill in once the service exists
curl -X POST http://localhost:8000/detect -F "file=@sample.png"
```

## Requirements

### Functional

- The user can call `POST /detect` with an image file.
- The user receives the bounding box and checked state of every detected checkbox.
- Invalid input (empty file or not an image) is rejected with a clear error.

### Non-functional

- **Accuracy:** best effort, perfect accuracy is not expected.
- **Robustness:** bad input never crashes the server.
- **Determinism:** the same image always produces the same response (stable box ordering, fixed seeds, pinned dependencies).
- **Latency:** TODO: target time per image.

## Build and run

TODO

## System Structure

```
checkbox-detection/
├── go.mod
├── cmd/server/main.go        # starts the HTTP server
├── internal/
│   ├── api/handler.go        # POST /detect: parse upload, call detector, write JSON
│   └── detector/
│       ├── detector.go       # Detect(img) []Box, runs the pipeline
│       └── detector_test.go
├── testdata/                 # sample document images
├── Dockerfile
└── README.md
```

- `detector` has no HTTP knowledge: it takes image bytes and returns boxes, so it is testable with plain image files.
- `api` has no OpenCV knowledge: it validates input, calls the detector, and returns JSON or a 4xx error.
- Standard library `net/http` only. The single external dependency is [gocv](https://gocv.io) (OpenCV bindings for Go).

## Approach

Flow of a `POST /detect` request, from upload to response:

1. **Receive.** The client uploads an image as `multipart/form-data` (field `file`).
2. **Decode.** The handler reads the bytes and decodes them into an OpenCV image. If the file is empty or not an image, the request ends here with a 4xx.
3. **Preprocess.** Convert to grayscale and apply an adaptive threshold, so ink is white on a black background regardless of scan lighting.
4. **Detect.** Find contours and keep only those that look like checkboxes (small, roughly square, four corners, hollow). Each one yields a bounding box in the pixels of the uploaded image.
5. **Classify.** For each box, measure how much ink is inside it (ignoring the border). Enough ink means checked, otherwise unchecked.
6. **Respond.** Sort the boxes top to bottom, then left to right, and return them as JSON.

Each step is detailed in its own subsection below as the implementation lands.

## Tradeoffs

TODO

## Known limitations

TODO
