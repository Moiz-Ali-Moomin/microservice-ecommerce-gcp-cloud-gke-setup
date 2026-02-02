# Builder
FROM golang:1.23 as builder
WORKDIR /app


COPY shared-lib shared-lib
COPY notification-service notification-service
RUN go work init ./shared-lib ./notification-service
RUN cd notification-service && CGO_ENABLED=0 GOOS=linux go build -o notification-service cmd/main.go

# Runnable
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/notification-service/notification-service .
USER 65532:65532
ENTRYPOINT ["/notification-service"]

