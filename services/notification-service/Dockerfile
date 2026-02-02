# Builder
FROM golang:1.22 as builder
WORKDIR /app
COPY go.work .
COPY go.work.sum .
COPY services/common services/common
COPY services/notification-service services/notification-service
RUN cd services/notification-service && CGO_ENABLED=0 GOOS=linux go build -o notification-service cmd/main.go

# Runnable
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/services/notification-service/notification-service .
USER 65532:65532
ENTRYPOINT ["/notification-service"]
