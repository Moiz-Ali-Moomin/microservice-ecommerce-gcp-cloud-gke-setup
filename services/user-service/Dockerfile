FROM golang:1.24 AS builder
WORKDIR /app

# 1. Copy shared library (Dependency)
COPY services/shared-lib ./services/shared-lib

# 2. Copy go.mod and go.sum (Dependency definitions)
#    This layer is cached unless dependencies change.
COPY services/user-service/go.mod services/user-service/go.sum ./services/user-service/

# 3. Download dependencies
#    This runs ONLY if go.mod/go.sum (or shared-lib) changed.
WORKDIR /app/services/user-service
RUN go mod download

# 4. Copy source code
#    This layer changes frequently but won't trigger re-download.
COPY services/user-service /app/services/user-service

# 5. Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/server cmd/main.go

# ---------- Runtime ----------
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /app/server /app/server
USER nonroot
EXPOSE 8080
ENTRYPOINT ["/app/server"]
