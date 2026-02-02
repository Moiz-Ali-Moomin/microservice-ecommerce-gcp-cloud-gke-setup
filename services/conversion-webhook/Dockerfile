# Builder
FROM golang:1.23 as builder
WORKDIR /app


COPY shared-lib shared-lib
COPY conversion-webhook conversion-webhook
RUN go work init ./shared-lib ./conversion-webhook
RUN cd conversion-webhook && CGO_ENABLED=0 GOOS=linux go build -o conversion-webhook cmd/main.go

# Runnable
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/conversion-webhook/conversion-webhook .
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/conversion-webhook"]

