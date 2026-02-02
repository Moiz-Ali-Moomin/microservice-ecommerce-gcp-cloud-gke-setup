# Builder
FROM golang:1.23 as builder
WORKDIR /app
COPY shared-lib shared-lib
COPY storefront-service storefront-service
RUN go work init ./shared-lib ./storefront-service
RUN cd storefront-service && CGO_ENABLED=0 GOOS=linux go build -o storefront-service cmd/main.go

# Runnable
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/storefront-service/storefront-service .
COPY --from=builder /app/storefront-service/templates ./templates
COPY --from=builder /app/storefront-service/static ./static
USER 65532:65532
EXPOSE 3000
ENTRYPOINT ["/storefront-service"]

