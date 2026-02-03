
import os

SERVICES_DIR = "services"
TEMPLATE = """FROM golang:1.24 AS builder
WORKDIR /app
COPY go.work go.work.sum ./
COPY services/shared-lib ./services/shared-lib
RUN go work sync
COPY services/{service_name} ./services/{service_name}
RUN cd services/{service_name} && \\
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/server {build_path}

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /app/server .
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/server"]
"""

def optimize_dockerfiles():
    if not os.path.exists(SERVICES_DIR):
        print(f"Directory {SERVICES_DIR} not found.")
        return

    count = 0
    for service_name in os.listdir(SERVICES_DIR):
        service_path = os.path.join(SERVICES_DIR, service_name)
        if not os.path.isdir(service_path) or service_name == "shared-lib":
            continue
        
        # Check for Dockerfile presence to confirm it's a service we want to modify
        dockerfile_path = os.path.join(service_path, "Dockerfile")
        if not os.path.exists(dockerfile_path):
            continue

        # Determine build path logic
        # 1. explicit override for admin-backoffice as known from previous context
        # 2. auto-detection logic: if cmd/main.go exists, use it. Else main.go
        
        main_in_root = os.path.exists(os.path.join(service_path, "main.go"))
        main_in_cmd = os.path.exists(os.path.join(service_path, "cmd", "main.go"))

        build_path = "cmd/main.go" # default convention
        if main_in_root and not main_in_cmd:
            build_path = "main.go"
        
        # Rewrite Dockerfile
        new_content = TEMPLATE.format(service_name=service_name, build_path=build_path)
        with open(dockerfile_path, "w") as f:
            f.write(new_content)
        
        print(f"Optimized Dockerfile for {service_name} (Build path: {build_path})")
        count += 1

    print(f"Successfully updated {count} Dockerfiles.")

if __name__ == "__main__":
    optimize_dockerfiles()
