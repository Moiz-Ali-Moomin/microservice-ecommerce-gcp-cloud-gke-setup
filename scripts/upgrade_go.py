
import os
import re

SERVICES_DIR = "services"
TARGET_GO_VERSION = "1.24.0"
TARGET_DOCKER_IMAGE = "golang:1.24"
GO_MOD_VERSION = "1.24.0" # go.mod usually uses 1.24

def update_dockerfile(path):
    with open(path, "r") as f:
        content = f.read()
    
    # Replace FROM golang:1.xx... with FROM golang:1.24
    # This regex looks for "FROM golang:" followed by digits and dots, optionally as builder
    new_content = re.sub(r"FROM golang:[0-9.]+(.*)", f"FROM {TARGET_DOCKER_IMAGE}\\1", content)
    
    if content != new_content:
        print(f"Updating Dockerfile: {path}")
        with open(path, "w") as f:
            f.write(new_content)
        return True
    return False

def update_go_mod(path):
    with open(path, "r") as f:
        content = f.read()
    
    # Replace go 1.xx with go 1.24
    new_content = re.sub(r"^go [0-9.]+$", f"go {GO_MOD_VERSION}", content, flags=re.MULTILINE)
    
    if content != new_content:
        print(f"Updating go.mod: {path}")
        with open(path, "w") as f:
            f.write(new_content)
        return True
    return False

def main():
    if not os.path.exists(SERVICES_DIR):
        print(f"Directory {SERVICES_DIR} not found.")
        return

    docker_count = 0
    mod_count = 0

    # 1. Update services
    for service_name in os.listdir(SERVICES_DIR):
        service_path = os.path.join(SERVICES_DIR, service_name)
        if not os.path.isdir(service_path):
            continue
        
        # Update Dockerfile
        dockerfile_path = os.path.join(service_path, "Dockerfile")
        if os.path.exists(dockerfile_path):
            if update_dockerfile(dockerfile_path):
                docker_count += 1
        
        # Update go.mod
        go_mod_path = os.path.join(service_path, "go.mod")
        if os.path.exists(go_mod_path):
            if update_go_mod(go_mod_path):
                mod_count += 1

    print(f"Updated {docker_count} Dockerfiles and {mod_count} go.mod files.")

if __name__ == "__main__":
    main()
