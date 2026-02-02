FROM golang:1.24 as builder
WORKDIR /app
COPY shared-lib shared-lib
COPY audit-service audit-service
RUN go work init ./shared-lib ./audit-service
RUN cd audit-service && CGO_ENABLED=0 GOOS=linux go build -o audit-service cmd/main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/audit-service/audit-service .
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/audit-service"]
