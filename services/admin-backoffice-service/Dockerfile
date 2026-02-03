FROM golang:1.24 AS builder
WORKDIR /app

# 1. Copy shared library (Dependency)
COPY services/shared-lib ./services/shared-lib

# 2. Copy go.mod and go.sum (Dependency definitions)
#    This layer is cached unless dependencies change.
COPY services/admin-backoffice-service/go.mod services/admin-backoffice-service/go.sum ./services/admin-backoffice-service/

# 3. Download dependencies
#    This runs ONLY if go.mod/go.sum (or shared-lib) changed.
WORKDIR /app/services/admin-backoffice-service
RUN go mod download

# 4. Copy source code
#    This layer changes frequently but won't trigger re-download.
COPY services/admin-backoffice-service /app/services/admin-backoffice-service

# 5. Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/server main.go

# ---------- Runtime ----------
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /app/server /app/server
USER nonroot
EXPOSE 8080
ENTRYPOINT ["/app/server"]
