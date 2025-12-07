#!/usr/bin/env python3
import subprocess
import pathlib
import sys
import re
import shutil

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

def get_module_path() -> str | None:
    try:
        p = subprocess.run(["git", "remote", "get-url", "origin"], check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
        url = p.stdout.strip()
    except Exception:
        return None
    m = re.match(r"git@([^:]+):/*([^/]+/[^/]+?)(?:\.git)?$", url)
    if m:
        host, path = m.group(1), m.group(2)
        return f"{host}/{path}"
    m = re.match(r"https?://([^/]+)/(.+?)(?:\.git)?$", url)
    if m:
        host, path = m.group(1), m.group(2)
        return f"{host}/{path}"
    return None

def replace_go_imports(root: pathlib.Path, old: str, new: str) -> None:
    for p in root.rglob("*.go"):
        s = p.read_text(encoding="utf-8")
        if old in s:
            p.write_text(s.replace(old, new), encoding="utf-8")

def update_go_mod(root: pathlib.Path, module: str) -> None:
    p = root / "go.mod"
    if not p.exists():
        return
    s = p.read_text(encoding="utf-8")
    s = re.sub(r"^module\s+\S+", f"module {module}", s, count=1, flags=re.MULTILINE)
    p.write_text(s, encoding="utf-8")
    if shutil.which("go"):
        try:
            subprocess.run(["go", "mod", "tidy"], check=True, cwd=str(root))
        except Exception:
            pass

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
    try:
        gitignore = (root / ".gitignore")
        content = gitignore.read_text(encoding="utf-8") if gitignore.exists() else ""
    except Exception:
        content = ""
    patterns = [app_name, "webui/dist"]
    missing = []
    for pat in patterns:
        if not re.search(rf"(?m)^\s*{re.escape(pat)}\s*$", content):
            missing.append(pat)
    if missing:
        if content and not content.endswith("\n"):
            content += "\n"
        content += "\n".join(missing) + "\n"
        (root / ".gitignore").write_text(content, encoding="utf-8")
    mod = get_module_path()
    if mod:
        replace_go_imports(root, "example.com/template", mod)
        update_go_mod(root, mod)
        print(mod)
    print(app_name)

if __name__ == "__main__":
    main()
