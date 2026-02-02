FROM golang:1.23 as builder
WORKDIR /app
# 
COPY shared-lib shared-lib
COPY user-service user-service
RUN go work init ./shared-lib ./user-service
RUN cd user-service && CGO_ENABLED=0 GOOS=linux go build -o user-service cmd/main.go

# Runnable
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/user-service/user-service .
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/user-service"]

