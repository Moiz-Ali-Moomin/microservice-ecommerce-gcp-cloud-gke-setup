FROM golang:1.23 as builder
WORKDIR /app
# 
COPY shared-lib shared-lib
COPY auth-service auth-service
RUN go work init ./shared-lib ./auth-service
RUN cd auth-service && CGO_ENABLED=0 GOOS=linux go build -o auth-service cmd/main.go

# Runnable
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/auth-service/auth-service .
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/auth-service"]

