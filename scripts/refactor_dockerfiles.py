
import os

SERVICES_DIR = "services"
TEMPLATE = """# Builder
FROM golang:1.23 AS builder

WORKDIR /app

# 1. Copy workspace definition
COPY go.work go.work.sum ./

# 2. Copy shared lib
COPY services/shared-lib ./services/shared-lib

# 3. Copy service code
COPY services/{service_name} ./services/{service_name}

# 4. Fix go.work
RUN grep -v "./services/" go.work > go.work.new && \\
    mv go.work.new go.work && \\
    go work use ./services/shared-lib ./services/{service_name}

# 5. Build
RUN cd services/{service_name} && \\
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server cmd/main.go

# Runnable
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /app/server .

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/app/server"]
"""

def main():
    if not os.path.exists(SERVICES_DIR):
        print(f"Directory {SERVICES_DIR} not found.")
        return

    count = 0
    for service_name in os.listdir(SERVICES_DIR):
        service_path = os.path.join(SERVICES_DIR, service_name)
        if not os.path.isdir(service_path):
            continue
        
        dockerfile_path = os.path.join(service_path, "Dockerfile")
        if os.path.exists(dockerfile_path):
            print(f"Updating {dockerfile_path}...")
            content = TEMPLATE.format(service_name=service_name)
            with open(dockerfile_path, "w") as f:
                f.write(content)
            count += 1
            
    print(f"Successfully updated {count} Dockerfiles.")

if __name__ == "__main__":
    main()
