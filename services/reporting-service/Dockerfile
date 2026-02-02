# Builder
FROM golang:1.22 as builder
WORKDIR /app
COPY go.work .
COPY go.work.sum .
COPY services/common services/common
COPY services/reporting-service services/reporting-service
RUN cd services/reporting-service && CGO_ENABLED=0 GOOS=linux go build -o reporting-service cmd/main.go

# Runnable
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/services/reporting-service/reporting-service .
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/reporting-service"]
