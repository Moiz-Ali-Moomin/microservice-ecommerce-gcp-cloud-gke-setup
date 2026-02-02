FROM golang:1.24 as builder
WORKDIR /app
# COPY go.work .
COPY shared-lib shared-lib
COPY offer-service offer-service
RUN go work init ./shared-lib ./offer-service
RUN cd offer-service && CGO_ENABLED=0 GOOS=linux go build -o offer-service cmd/main.go

# Runnable (Distroless or Scratch for security/size)
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/offer-service/offer-service .
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/offer-service"]
