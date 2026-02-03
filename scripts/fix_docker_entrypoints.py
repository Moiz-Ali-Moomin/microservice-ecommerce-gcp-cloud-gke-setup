
import os

SERVICES_DIR = "services"

def fix_dockerfiles_entrypoint():
    if not os.path.exists(SERVICES_DIR):
        print(f"Directory {SERVICES_DIR} not found.")
        return

    fixed_count = 0
    for service_name in os.listdir(SERVICES_DIR):
        service_path = os.path.join(SERVICES_DIR, service_name)
        if not os.path.isdir(service_path):
            continue
        
        main_in_root = os.path.exists(os.path.join(service_path, "main.go"))
        main_in_cmd = os.path.exists(os.path.join(service_path, "cmd", "main.go"))
        
        dockerfile_path = os.path.join(service_path, "Dockerfile")
        if not os.path.exists(dockerfile_path):
            continue

        target_build_path = None
        if main_in_root and not main_in_cmd:
            print(f"[{service_name}]: main.go found in ROOT. Fixing Dockerfile...")
            target_build_path = "main.go"
        elif main_in_cmd:
            # Assumed correct by default template, but good to check if we need to revert?
            # actually the template uses cmd/main.go, so if it is there, we are good.
            continue
        else:
            print(f"[{service_name}]: WARNING - Could not find main.go in root or cmd/!")
            continue

        if target_build_path:
            with open(dockerfile_path, "r") as f:
                content = f.read()
            
            # Look for the build command
            # Template: go build ... -o /app/server cmd/main.go
            if "cmd/main.go" in content:
                new_content = content.replace("cmd/main.go", target_build_path)
                with open(dockerfile_path, "w") as f:
                    f.write(new_content)
                fixed_count += 1
                print(f"[{service_name}]: Dockerfile updated.")

    print(f"Fixed {fixed_count} Dockerfiles.")

if __name__ == "__main__":
    fix_dockerfiles_entrypoint()
