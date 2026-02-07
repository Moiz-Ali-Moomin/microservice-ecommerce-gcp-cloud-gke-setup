# syntax=docker/dockerfile:1

FROM golang:1.24 AS builder
WORKDIR /app

# 1. Copy shared library
COPY services/shared-lib ./services/shared-lib

# 2. Copy go.mod and optional go.sum
COPY services/auth-service/go.mod services/auth-service/go.su[m] ./services/auth-service/

# 3. Download dependencies with cache
WORKDIR /app/services/auth-service
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# 4. Copy source code
COPY services/auth-service /app/services/auth-service

# 5. Build binary
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/server cmd/main.go

# ---------- Runtime ----------
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=builder /app/server /app/server

USER nonroot
EXPOSE 8080

# ✅ CHANGED: CMD instead of ENTRYPOINT
CMD ["/app/server"]
