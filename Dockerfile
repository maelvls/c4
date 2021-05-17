# Build the manager binary
FROM golang:1.14-alpine as builder
WORKDIR /build

COPY . .

# Disable Cgo so that the binary doesn't rely on glibc and works with the
# scratch, alpine or distroless image.
RUN CGO_ENABLED=0 go build .

FROM scratch
COPY --from=builder /build/c4 /c4
ENTRYPOINT ["/c4"]
