# Builder
FROM golang:1.22 as builder
WORKDIR /app
COPY go.work .
COPY go.work.sum .
COPY services/common services/common
COPY services/admin-backoffice-service services/admin-backoffice-service
RUN cd services/admin-backoffice-service && CGO_ENABLED=0 GOOS=linux go build -o admin-backoffice-service cmd/main.go

# Runnable
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/services/admin-backoffice-service/admin-backoffice-service .
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/admin-backoffice-service"]
