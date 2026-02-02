# Builder
FROM golang:1.22 as builder
WORKDIR /app
COPY go.work .
COPY go.work.sum .
COPY services/common services/common
COPY services/conversion-webhook services/conversion-webhook
RUN cd services/conversion-webhook && CGO_ENABLED=0 GOOS=linux go build -o conversion-webhook cmd/main.go

# Runnable
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/services/conversion-webhook/conversion-webhook .
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/conversion-webhook"]
