# Build stage using Go 1.24
FROM golang:1.24 AS builder
WORKDIR /app

# Cache dependency downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY ./*.go ./
COPY ./cmd ./cmd

# Build the binary with optimizations and dead code elimination flags
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /openapi-spec-converter ./cmd/openapi-spec-converter

# Final stage using a distroless image
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /openapi-spec-converter /openapi-spec-converter

USER nonroot:nonroot

ENTRYPOINT ["/openapi-spec-converter"]
