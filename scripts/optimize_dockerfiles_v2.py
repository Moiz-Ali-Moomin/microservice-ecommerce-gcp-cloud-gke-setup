
import os

SERVICES_DIR = "services"
# Template that solves the go.work missing module issue by creating a build-local go.work
TEMPLATE = """FROM golang:1.24 AS builder
WORKDIR /app

# 1. Copy global workspace files (User requirement, though we will override usage)
COPY go.work go.work.sum ./

# 2. Copy shared library (Dependency)
COPY services/shared-lib ./services/shared-lib

# 3. Copy target service go.mod (Required for workspace init before full source copy)
COPY services/{service_name}/go.mod ./services/{service_name}/go.mod

# 4. Create a valid go.work for this build context (Prunes missing modules)
#    We overwrite the global go.work because it references modules we haven't copied.
RUN go work init ./services/shared-lib ./services/{service_name} && \\
    go work edit -go=1.24.0

# 5. Download dependencies (Cached layer - invalidates only on go.mod changes)
RUN go work sync

# 6. Copy target service source code (Invalidates on code changes)
COPY services/{service_name} ./services/{service_name}

# 7. Build
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
        
        dockerfile_path = os.path.join(service_path, "Dockerfile")
        if not os.path.exists(dockerfile_path):
            continue

        # Auto-detect build path
        main_in_root = os.path.exists(os.path.join(service_path, "main.go"))
        main_in_cmd = os.path.exists(os.path.join(service_path, "cmd", "main.go"))

        build_path = "cmd/main.go" # default
        if main_in_root and not main_in_cmd:
            build_path = "main.go"
        
        new_content = TEMPLATE.format(service_name=service_name, build_path=build_path)
        with open(dockerfile_path, "w") as f:
            f.write(new_content)
        
        print(f"Fixed Dockerfile for {service_name}")
        count += 1

    print(f"Successfully updated {count} Dockerfiles.")

if __name__ == "__main__":
    optimize_dockerfiles()
