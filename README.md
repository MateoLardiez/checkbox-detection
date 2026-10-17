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
curl -X POST http://localhost:8000/detect -F "file=@testdata/sample-000.png"
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
- **Latency:** measured with `curl` against the Docker container on a development machine, including the upload: 0.08 s for the small sample (1586x846) and 0.21 to 0.25 s for the full-page samples (2550x4200 and 2550x3301). It grows with the number of pixels.

## Build and run

### Local

Requires Go 1.25+ and OpenCV 4.13 (the version gocv v0.43.0 expects). See [gocv's Linux install guide](https://gocv.io/getting-started/linux/).

```bash
go run ./cmd/server
```

The server listens on port 8000 by default. Set `PORT` to change it.

### Docker

No local Go or OpenCV needed, the build stage bundles both.

```bash
docker build -t checkbox-detection .
docker run --rm -p 8000:8000 checkbox-detection
```

The first build downloads the build base image (several GB); the resulting image is about 430 MB. The service is then available at `http://localhost:8000`.

### Try it

With the server running (locally or in Docker), in another terminal:

```bash
# valid image: returns the detected boxes
curl -X POST http://localhost:8000/detect -F "file=@testdata/sample-001.png"

# not an image: returns 400 with an error message
curl -X POST http://localhost:8000/detect -F "file=@README.md"
```

### Run the tests

Needs the local setup above (Go and OpenCV).

```bash
go test ./...
```

To save each sample image with the detections drawn on it (red = checked, green = unchecked):

```bash
cp .env.example .env
set -a && source .env && set +a
go test ./...
```

The images are written to `output/`.

## System Structure

```
checkbox-detection/
├── cmd/server/main.go          # starts the HTTP server
├── internal/
│   ├── api/
│   │   └── handler.go          # POST /detect: parse upload, call detector, write JSON
│   └── detector/
│       ├── detector.go         # Detect(data) []Box: runs the pipeline below
│       ├── preprocess.go       # grayscale + adaptive threshold
│       ├── candidates.go       # contours + filters: which shapes look like checkboxes
│       ├── classify.go         # checked / unchecked from the ink inside a box
│       └── visualize_test.go   # draws detections on the samples (opt-in, see "Run the tests")
├── testdata/                   # sample document images
├── Dockerfile
├── .env.example                # VISUALIZE_DIR for the visualizer
├── go.mod
└── README.md
```

Each source file has a `_test.go` next to it, except `main.go`.

- `detector` has no HTTP knowledge: it takes image bytes and returns boxes, so it is testable with plain image files.
- `api` has no OpenCV knowledge: it validates input, calls the detector, and returns JSON or a 4xx error.
- Standard library `net/http` only. The single external dependency is [gocv](https://gocv.io) (OpenCV bindings for Go).

## Approach

Classic computer vision with OpenCV (through gocv), no machine learning. Flow of a `POST /detect` request, from upload to response:

1. **Receive.** The client uploads an image as `multipart/form-data` (field `file`).
2. **Decode.** The handler reads the bytes and decodes them into an OpenCV image. If the file is empty or not an image, the request ends here with a 400.
3. **Preprocess** (`preprocess.go`). Convert to grayscale and apply an inverted adaptive threshold, so ink is white and paper black. "Adaptive" means each pixel is compared with the average of its 25x25 neighborhood (it counts as ink if it is 5 levels darker), which works with uneven scan lighting where a single global threshold would not.
4. **Find candidates** (`candidates.go`). Extract contours (the outlines of white regions) and keep those that look like a checkbox:
   - **Size:** each side between 0.5% and 3% of the image width. It is relative, so it does not depend on scan resolution.
   - **Proportion:** width/height between 0.67 and 1.5. The checkboxes in the samples are slightly wider than tall.
   - **Shape:** the contour simplifies to exactly four corners, and each one is close to a corner of its minimum rotated rectangle. This rejects slanted or irregular shapes (for example a hand-drawn scribble) while still accepting tilted squares.

   A checkbox whose border touches a table line fuses with it, so its outer contour is not a square. In that case its inner border (the hole) is used instead, and its box is grown by 8% per side to recover the border thickness.
5. **Classify** (`classify.go`). Ignore 25% of each side to skip the border, and measure the share of ink in the rest. 10% or more means checked. A ratio is used instead of "any ink" so scan specks do not count as marks. On the samples, empty boxes measured under 5% and marked ones over 15%, so 10% sits in the gap.
6. **Respond.** Sort the boxes by `y1`, `x1`, `x2`, `y2` (a total order, so the output is deterministic) and return them as JSON. Images are never resized, so `bbox` is read directly in pixels of the uploaded image.

The numbers above were tuned by eye on the four sample documents, using the visualizer (`visualize_test.go`) that draws the detections over each image.

## Tradeoffs

- **Classic computer vision instead of a trained model.** There are only four sample images and no labels to train or evaluate a model. Contours and filters need no training data, are fast, and always give the same output for the same input. The cost is hand-picked thresholds, tuned on those four samples, which may not transfer to other forms.
- **gocv (OpenCV bindings) instead of pure Go.** It gives well-tested image primitives (threshold, contours) instead of reimplementing them. The cost is cgo and a heavy setup: OpenCV 4.13 must be installed to build locally, and the Docker build image is about 3.9 GB (the final image is about 430 MB).
- **Multi-stage Dockerfile.** The build stage has Go and OpenCV (about 3.9 GB). The final stage is `debian:bookworm-slim` with only the binary and the shared libraries `ldd` lists for it (about 430 MB). The cost is a longer Dockerfile, and the runtime stage must be the same distro as the build stage (glibc), so Alpine does not work.
- **Relative sizes instead of fixed pixels.** Filters use fractions of the image width, so the same form scanned at another resolution still works. This assumes the image is a whole page.
- **Inner borders accepted as candidates.** It recovers checkboxes fused with table lines (the whole second sample page depends on it), but a square-looking hole in some other shape could become a false positive.

## Known limitations

- **Damaged or unusual checkboxes are missed.** In the samples: one whose border is broken by a thick X, and one struck through by a long stroke. Hand-drawn shapes are rejected on purpose.
- **Light marks.** A checkbox is "checked" only if at least 10% of its inside is ink, so a very thin or small mark can be reported as unchecked.
- **Tilted pages.** Tilted checkboxes are detected, but the returned `bbox` is always axis-aligned and the ink is measured in an axis-aligned region, so classification is expected to degrade on strongly tilted pages. This was not tested on real tilted scans.
- **Cropped or zoomed images.** Sizes are relative to the image width, so a crop that magnifies the checkboxes much more than a full page may fall outside the accepted range.
- **Only rectangular checkboxes.** Circles (radio buttons) and other shapes are not detected.
- **Possible false positives.** Square-looking shapes other than checkboxes can pass the filters. Letters of a diagonal watermark did in one case before a threshold change, and other cases may exist.
- **Large uploads.** There is no size limit, and the whole image is decoded in memory.

## Future Improvements

- Add another variant besides "checked" and "not checked", which can be "unknown". In this case, a person needs to validate it. This way we can avoid false positives.
- Classify tilted checkboxes using their rotated rectangle (or straighten the page first), and return the rotated corners.
- Label the sample documents by hand and measure precision and recall, to tune the thresholds with numbers.
- Reject uploads above a size limit with a 413.
- Order boxes by visual row (group nearby `y1` values) instead of strict `y1`, so a row reads left to right even when its boxes differ by a pixel or two.
