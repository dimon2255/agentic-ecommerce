# --- Build stage ---
FROM golang:1.26-bookworm AS builder

WORKDIR /build/api
COPY api/go.mod api/go.sum ./
RUN go mod download

COPY api/ .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /build/server ./cmd/server/

# --- Runtime stage ---
FROM gcr.io/distroless/static-debian12

COPY --from=builder /build/server /app/server
COPY --from=builder /build/api/config.yaml /app/config.yaml

WORKDIR /app
EXPOSE 9090
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]
