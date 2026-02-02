# Builder
FROM golang:1.23 as builder
WORKDIR /app


COPY shared-lib shared-lib
COPY admin-backoffice-service admin-backoffice-service
RUN go work init ./shared-lib ./admin-backoffice-service
RUN cd admin-backoffice-service && CGO_ENABLED=0 GOOS=linux go build -o admin-backoffice-service main.go

# Runnable
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/admin-backoffice-service/admin-backoffice-service .
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/admin-backoffice-service"]

