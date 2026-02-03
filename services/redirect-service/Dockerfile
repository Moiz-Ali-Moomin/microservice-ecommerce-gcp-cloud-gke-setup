FROM golang:1.24 AS builder
WORKDIR /app
COPY go.work go.work.sum ./
COPY services/shared-lib ./services/shared-lib
RUN go work sync
COPY services/redirect-service ./services/redirect-service
RUN cd services/redirect-service && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/server cmd/main.go

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /app/server .
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/server"]
