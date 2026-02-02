# Builder
FROM golang:1.22 as builder
WORKDIR /app
COPY go.work .
COPY go.work.sum .
COPY services/common services/common
COPY services/attribution-service services/attribution-service
RUN cd services/attribution-service && CGO_ENABLED=0 GOOS=linux go build -o attribution-service cmd/main.go

# Runnable (Distroless)
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/services/attribution-service/attribution-service .
USER 65532:65532
ENTRYPOINT ["/attribution-service"]
