
import yaml

DOCKER_COMPOSE_PATH = "docker-compose.yml"

def fix_build_context():
    with open(DOCKER_COMPOSE_PATH, "r") as f:
        content = f.read()

    # We will do a line-by-line replace to preserve comments and structure as much as possible,
    # since pyyaml isn't great at preserving comments.
    
    new_lines = []
    lines = content.splitlines()
    
    for line in lines:
        if "context: ./services" in line:
            new_lines.append(line.replace("context: ./services", "context: ."))
        elif "dockerfile: ./" in line and "dockerfile: ./services/" not in line:
            # Assumes format: dockerfile: ./service-name/Dockerfile
            # We want: dockerfile: ./services/service-name/Dockerfile
            new_lines.append(line.replace("dockerfile: ./", "dockerfile: ./services/"))
        else:
            new_lines.append(line)
            
    new_content = "\n".join(new_lines) + "\n"
    
    with open(DOCKER_COMPOSE_PATH, "w") as f:
        f.write(new_content)
    
    print("Updated docker-compose.yml build contexts.")

if __name__ == "__main__":
    fix_build_context()
