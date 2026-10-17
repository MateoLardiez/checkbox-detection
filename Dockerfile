# Build stage: bundles Go 1.25.5 and OpenCV 4.13.0 (the version gocv v0.43.0 expects).
FROM ghcr.io/hybridgroup/opencv:4.13.0 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
RUN go build -o /server ./cmd/server

# Gather the shared libraries the binary needs, except the C library,
# which the runtime image already has.
RUN mkdir /deps && ldd /server | awk '/=>/ {print $3}' \
    | grep -Ev 'libc\.so|libm\.so|libpthread|libresolv' \
    | xargs -I{} cp -L {} /deps/

# Runtime stage: same distro as the build stage (Debian bookworm, glibc).
FROM debian:bookworm-slim

COPY --from=build /deps/ /usr/local/lib/
RUN ldconfig
COPY --from=build /server /server

ENV PORT=8000
EXPOSE 8000

CMD ["/server"]
