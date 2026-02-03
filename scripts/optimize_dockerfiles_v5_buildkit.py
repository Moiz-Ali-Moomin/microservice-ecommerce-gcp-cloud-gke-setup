
import os

SERVICES_DIR = "services"

# Template with BuildKit cache mounts
TEMPLATE = """# syntax=docker/dockerfile:1

FROM golang:1.24 AS builder
WORKDIR /app

# 1. Copy shared library (Dependency)
COPY services/shared-lib ./services/shared-lib

# 2. Copy go.mod and go.sum (Dependency definitions)
COPY services/{service_name}/go.mod services/{service_name}/go.sum ./services/{service_name}/

# 3. Download dependencies with Cache Mounts
#    Persists /go/pkg/mod across builds, even if go.mod changes.
WORKDIR /app/services/{service_name}
RUN --mount=type=cache,target=/go/pkg/mod \\
    go mod download

# 4. Copy source code
COPY services/{service_name} /app/services/{service_name}

# 5. Build binary with Cache Mounts
#    Persists build cache (/root/.cache/go-build) for faster compilation.
RUN --mount=type=cache,target=/go/pkg/mod \\
    --mount=type=cache,target=/root/.cache/go-build \\
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \\
    go build -ldflags="-w -s" -o /app/server {build_path}

# ---------- Runtime ----------
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /app/server /app/server
USER nonroot
EXPOSE 8080
ENTRYPOINT ["/app/server"]
"""

def optimize_buildkit():
    if not os.path.exists(SERVICES_DIR):
        print(f"Directory {SERVICES_DIR} not found.")
        return

    count = 0
    for service_name in os.listdir(SERVICES_DIR):
        service_path = os.path.join(SERVICES_DIR, service_name)
        if not os.path.isdir(service_path) or service_name == "shared-lib":
            continue
        
        dockerfile_path = os.path.join(service_path, "Dockerfile")
        # Ensure it's a valid service directory
        if not os.path.exists(dockerfile_path) and not os.path.exists(os.path.join(service_path, "main.go")) and not os.path.exists(os.path.join(service_path, "cmd", "main.go")):
             continue

        # Auto-detect build path logic
        build_path = "cmd/main.go" # default
        if os.path.exists(os.path.join(service_path, "main.go")) and not os.path.exists(os.path.join(service_path, "cmd", "main.go")):
            build_path = "main.go"
        
        # Write new content
        new_content = TEMPLATE.format(service_name=service_name, build_path=build_path)
        with open(dockerfile_path, "w") as f:
            f.write(new_content)
        
        print(f"Applied BuildKit Cache Mounts to {service_name}")
        count += 1

    print(f"Successfully updated {count} Dockerfiles.")

if __name__ == "__main__":
    optimize_buildkit()
