from pathlib import Path
from ruamel.yaml import YAML

yaml = YAML()
yaml.preserve_quotes = True
yaml.width = 4096

for f in Path(".").rglob("*.y*ml"):
    try:
        with open(f, encoding="utf-8") as fp:
            data = yaml.load(fp)
        with open(f, "w", encoding="utf-8") as fp:
            yaml.dump(data, fp)
        print(f"fixed: {f}")
    except Exception as e:
        print(f"FAILED: {f} -> {e}")
