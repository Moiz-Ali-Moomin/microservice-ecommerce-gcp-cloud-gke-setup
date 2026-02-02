FROM golang:1.23 as builder
WORKDIR /app
# 
COPY shared-lib shared-lib
COPY landing-service landing-service
RUN go work init ./shared-lib ./landing-service
RUN cd landing-service && CGO_ENABLED=0 GOOS=linux go build -o landing-service cmd/main.go

# Runnable
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/landing-service/landing-service .
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/landing-service"]

