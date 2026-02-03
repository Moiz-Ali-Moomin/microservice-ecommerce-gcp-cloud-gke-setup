
import os

SERVICES_DIR = "services"

# Exact template requested by user, with placeholders for service_name and build_path
TEMPLATE = """# ---------- Builder ----------
FROM golang:1.24 AS builder

WORKDIR /app

# 1. Copy workspace definition
COPY go.work go.work.sum ./

# 2. Copy shared library (dependency layer)
COPY services/shared-lib ./services/shared-lib

# 3. Sync workspace (read-only, deterministic)
RUN go work sync

# 4. Copy service source
COPY services/{service_name} ./services/{service_name}

# 5. Build binary
WORKDIR /app/services/{service_name}
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \\
    go build -ldflags="-w -s" -o /app/server {build_path}


# ---------- Runtime ----------
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=builder /app/server /app/server

USER nonroot
EXPOSE 8080
ENTRYPOINT ["/app/server"]
"""

def standardize_dockerfiles():
    if not os.path.exists(SERVICES_DIR):
        print(f"Directory {SERVICES_DIR} not found.")
        return

    count = 0
    for service_name in os.listdir(SERVICES_DIR):
        service_path = os.path.join(SERVICES_DIR, service_name)
        if not os.path.isdir(service_path) or service_name == "shared-lib":
            continue
        
        dockerfile_path = os.path.join(service_path, "Dockerfile")
        # Ensure we only target folders that look like services (have a Dockerfile or main.go)
        if not os.path.exists(dockerfile_path) and not os.path.exists(os.path.join(service_path, "main.go")) and not os.path.exists(os.path.join(service_path, "cmd", "main.go")):
             continue

        # Auto-detect build path logic
        build_path = "cmd/main.go" # default from template
        if os.path.exists(os.path.join(service_path, "main.go")) and not os.path.exists(os.path.join(service_path, "cmd", "main.go")):
            build_path = "main.go"
        
        # Write new content
        new_content = TEMPLATE.format(service_name=service_name, build_path=build_path)
        with open(dockerfile_path, "w") as f:
            f.write(new_content)
        
        print(f"Standardized Dockerfile for {service_name} (Path: {build_path})")
        count += 1

    print(f"Successfully updated {count} Dockerfiles.")

if __name__ == "__main__":
    standardize_dockerfiles()
