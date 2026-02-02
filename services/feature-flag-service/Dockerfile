# Builder
FROM golang:1.22 as builder
WORKDIR /app
COPY go.work .
COPY go.work.sum .
COPY services/common services/common
COPY services/feature-flag-service services/feature-flag-service
RUN cd services/feature-flag-service && CGO_ENABLED=0 GOOS=linux go build -o feature-flag-service cmd/main.go

# Runnable
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/services/feature-flag-service/feature-flag-service .
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/feature-flag-service"]
