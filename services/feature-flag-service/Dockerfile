# syntax=docker/dockerfile:1

FROM golang:1.24 AS builder
WORKDIR /app

# 1. Copy shared library (Dependency)
COPY services/shared-lib ./services/shared-lib

# 2. Copy go.mod and go.sum (Dependency definitions)
COPY services/feature-flag-service/go.mod services/feature-flag-service/go.sum ./services/feature-flag-service/

# 3. Download dependencies with Cache Mounts
#    Persists /go/pkg/mod across builds, even if go.mod changes.
WORKDIR /app/services/feature-flag-service
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# 4. Copy source code
COPY services/feature-flag-service /app/services/feature-flag-service

# 5. Build binary with Cache Mounts
#    Persists build cache (/root/.cache/go-build) for faster compilation.
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
ENTRYPOINT ["/app/server"]
