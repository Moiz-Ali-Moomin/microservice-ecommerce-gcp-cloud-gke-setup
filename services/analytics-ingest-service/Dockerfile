FROM golang:1.24 as builder
WORKDIR /app
COPY shared-lib shared-lib
COPY analytics-ingest-service analytics-ingest-service
RUN go work init ./shared-lib ./analytics-ingest-service
RUN cd analytics-ingest-service && CGO_ENABLED=0 GOOS=linux go build -o analytics-ingest-service cmd/main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/analytics-ingest-service/analytics-ingest-service .
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/analytics-ingest-service"]
