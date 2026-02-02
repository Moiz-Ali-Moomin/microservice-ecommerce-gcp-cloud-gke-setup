# Builder
FROM golang:1.23 as builder
WORKDIR /app


COPY shared-lib shared-lib
COPY reporting-service reporting-service
RUN go work init ./shared-lib ./reporting-service
RUN cd reporting-service && CGO_ENABLED=0 GOOS=linux go build -o reporting-service cmd/main.go

# Runnable
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/reporting-service/reporting-service .
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/reporting-service"]

