# Builder
FROM golang:1.23 as builder
WORKDIR /app


COPY shared-lib shared-lib
COPY attribution-service attribution-service
RUN go work init ./shared-lib ./attribution-service
RUN cd attribution-service && CGO_ENABLED=0 GOOS=linux go build -o attribution-service cmd/main.go

# Runnable (Distroless)
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/attribution-service/attribution-service .
USER 65532:65532
ENTRYPOINT ["/attribution-service"]

