# syntax=docker/dockerfile:1

FROM golang:1.22.0 AS builder
WORKDIR /app

# 1. Copy shared library (Dependency)
COPY services/shared-lib ./services/shared-lib

# 2. Copy go.mod (required) and go.sum (optional via glob pattern)
#    The go.su[m] pattern matches go.sum if present, but doesn't fail if missing.
COPY services/audit-service/go.mod services/audit-service/go.su[m] ./services/audit-service/

# 3. Download dependencies with Cache Mounts
WORKDIR /app/services/audit-service
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# 4. Copy source code
COPY services/audit-service /app/services/audit-service

# 5. Build binary with Cache Mounts
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
