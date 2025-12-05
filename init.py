#!/usr/bin/env python3
import subprocess
import pathlib
import sys
import re

def repo_root():
    try:
        p = subprocess.run(["git", "rev-parse", "--show-toplevel"], check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
        return pathlib.Path(p.stdout.strip())
    except Exception:
        return pathlib.Path(__file__).resolve().parent

def to_lower_snake(name: str) -> str:
    name = re.sub(r"[^A-Za-z0-9]+", "_", name)
    name = re.sub(r"([a-z0-9])([A-Z])", r"\1_\2", name)
    name = re.sub(r"_+", "_", name).strip("_")
    return name.lower()

def replace_in_file(p: pathlib.Path, placeholder: str, value: str) -> None:
    s = p.read_text(encoding="utf-8")
    if placeholder in s:
        p.write_text(s.replace(placeholder, value), encoding="utf-8")

def main():
    root = repo_root()
    repo_name = root.name
    app_name = to_lower_snake(repo_name)
    for rel in ["infra/vars/const.go", "Makefile"]:
        p = root / rel
        if p.exists():
            replace_in_file(p, "APP_NAME_PLACEHOLDER", app_name)
        else:
            print(f"missing {p}", file=sys.stderr)
    print(app_name)

if __name__ == "__main__":
    main()
