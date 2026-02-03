FROM golang:1.24 AS builder
WORKDIR /app

# 1. Copy global workspace files (User requirement, though we will override usage)
COPY go.work go.work.sum ./

# 2. Copy shared library (Dependency)
COPY services/shared-lib ./services/shared-lib

# 3. Copy target service go.mod (Required for workspace init before full source copy)
COPY services/redirect-service/go.mod ./services/redirect-service/go.mod

# 4. Create a valid go.work for this build context (Prunes missing modules)
#    We overwrite the global go.work because it references modules we haven't copied.
RUN go work init ./services/shared-lib ./services/redirect-service && \
    go work edit -go=1.24.0

# 5. Download dependencies (Cached layer - invalidates only on go.mod changes)
RUN go work sync

# 6. Copy target service source code (Invalidates on code changes)
COPY services/redirect-service ./services/redirect-service

# 7. Build
RUN cd services/redirect-service && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/server cmd/main.go

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /app/server .
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/server"]
