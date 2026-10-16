# Bundles Go 1.25.5 and OpenCV 4.13.0 (the version gocv v0.43.0 expects),
FROM ghcr.io/hybridgroup/opencv:4.13.0

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
RUN go build -o /server ./cmd/server

ENV PORT=8000
EXPOSE 8000

CMD ["/server"]
