# Builder
FROM golang:1.24 AS builder

WORKDIR /app

# 1. Copy workspace definition
COPY go.work go.work.sum ./

# 2. Copy shared lib
COPY services/shared-lib ./services/shared-lib

# 3. Copy service code
COPY services/audit-service ./services/audit-service

# 4. Fix go.work
RUN grep -v "./services/" go.work > go.work.new && \
    mv go.work.new go.work && \
    go work use ./services/shared-lib ./services/audit-service

# 5. Build
RUN cd services/audit-service && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server cmd/main.go

# Runnable
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /app/server .

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/app/server"]
