# ---------- Builder ----------
FROM golang:1.24 AS builder

WORKDIR /app

# 1. Copy workspace definition
COPY go.work go.work.sum ./

# 2. Copy shared library (dependency layer)
COPY services/shared-lib ./services/shared-lib

# 3. Sync workspace (read-only, deterministic)
RUN go work sync

# 4. Copy service source
COPY services/api-gateway ./services/api-gateway

# 5. Build binary
WORKDIR /app/services/api-gateway
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/server cmd/main.go


# ---------- Runtime ----------
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=builder /app/server /app/server

USER nonroot
EXPOSE 8080
ENTRYPOINT ["/app/server"]
