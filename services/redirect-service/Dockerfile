FROM golang:1.24 as builder
WORKDIR /app
COPY shared-lib shared-lib
COPY redirect-service redirect-service
RUN go work init ./shared-lib ./redirect-service
RUN cd redirect-service && CGO_ENABLED=0 GOOS=linux go build -o redirect-service cmd/main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/redirect-service/redirect-service .
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/redirect-service"]
