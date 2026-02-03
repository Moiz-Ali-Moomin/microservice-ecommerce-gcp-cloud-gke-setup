FROM golang:1.24 AS builder
WORKDIR /app

COPY services/shared-lib ./services/shared-lib
COPY services/storefront-service ./services/storefront-service

WORKDIR /app/services/storefront-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/server cmd/main.go

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /app/server /app/server
USER nonroot
EXPOSE 8080
ENTRYPOINT ["/app/server"]
