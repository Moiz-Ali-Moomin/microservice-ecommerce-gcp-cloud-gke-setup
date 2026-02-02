# Builder
FROM golang:1.23 as builder
WORKDIR /app


COPY shared-lib shared-lib
COPY feature-flag-service feature-flag-service
RUN go work init ./shared-lib ./feature-flag-service
RUN cd feature-flag-service && CGO_ENABLED=0 GOOS=linux go build -o feature-flag-service cmd/main.go

# Runnable
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/feature-flag-service/feature-flag-service .
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/feature-flag-service"]

