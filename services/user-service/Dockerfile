# Builder
FROM golang:1.23 AS builder

WORKDIR /app

# 1. Copy workspace definition
COPY go.work go.work.sum ./

# 2. Copy shared lib
COPY services/shared-lib ./services/shared-lib

# 3. Copy service code
COPY services/user-service ./services/user-service

# 4. Fix go.work
RUN grep -v "./services/" go.work > go.work.new && \
    mv go.work.new go.work && \
    go work use ./services/shared-lib ./services/user-service

# 5. Build
RUN cd services/user-service && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server cmd/main.go

# Runnable
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /app/server .

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/app/server"]
