FROM golang:1.23 as builder
WORKDIR /app
# 
COPY shared-lib shared-lib
COPY cart-service cart-service
RUN go work init ./shared-lib ./cart-service
RUN cd cart-service && CGO_ENABLED=0 GOOS=linux go build -o cart-service cmd/main.go

# Runnable
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/cart-service/cart-service .
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/cart-service"]

