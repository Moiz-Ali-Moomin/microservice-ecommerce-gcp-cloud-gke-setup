FROM golang:1.24 as builder
WORKDIR /app
# COPY go.work .
COPY shared-lib shared-lib
COPY api-gateway api-gateway
RUN go work init ./shared-lib ./api-gateway
RUN cd api-gateway && CGO_ENABLED=0 GOOS=linux go build -o api-gateway cmd/main.go

# Runnable
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/api-gateway/api-gateway .
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/api-gateway"]
