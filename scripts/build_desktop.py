#!/usr/bin/env python3
"""Build the native OhMyFreeAI desktop application from the official Wails v3 template."""

from __future__ import annotations

import argparse
import os
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
MODULE_PATH = "github.com/SoufianoDev/OhMyFreeAI/desktop"
ROOT_MODULE = "github.com/SoufianoDev/OhMyFreeAI"


def run(command: list[str], cwd: Path) -> None:
    print("+", " ".join(command), flush=True)
    subprocess.run(command, cwd=cwd, check=True)


def require_tools() -> None:
    missing = [name for name in ("wails3", "go", "node", "npm") if shutil.which(name) is None]
    if missing:
        raise SystemExit("Missing required build tools: " + ", ".join(missing))


def copy_frontend(project: Path) -> None:
    source = ROOT / "frontend"
    destination = project / "frontend"

    for name in ("index.html", "app.js"):
        shutil.copy2(source / name, destination / name)

    shutil.copytree(source / "styles", destination / "styles", dirs_exist_ok=True)

    index_path = destination / "index.html"
    html = index_path.read_text(encoding="utf-8")
    original = '<script type="module" src="app.js"></script>'
    replacement = '<script type="module" src="/src/main.js"></script>'
    if original not in html:
        raise RuntimeError("The root frontend index.html no longer has the expected app.js entrypoint.")
    index_path.write_text(html.replace(original, replacement, 1), encoding="utf-8")


def attach_generated_binding(project: Path) -> None:
    bindings = project / "frontend" / "bindings"
    source_dir = project / "frontend" / "src"
    source_dir.mkdir(parents=True, exist_ok=True)

    candidates: list[Path] = []
    if bindings.exists():
        for path in bindings.rglob("*.js"):
            source = path.read_text(encoding="utf-8", errors="replace")
            if "GetSnapshot" in source and "SaveConfig" in source:
                candidates.append(path)

    if not candidates:
        raise RuntimeError("Wails did not generate the AppService GetSnapshot/SaveConfig binding.")

    binding = sorted(candidates)[0]
    relative = Path(os.path.relpath(binding, source_dir)).as_posix()
    if not relative.startswith("."):
        relative = "./" + relative

    bridge = (
        f'import {{ GetSnapshot, SaveConfig }} from "{relative}";\n'
        'import "../app.js";\n'
        'window.OhMyFreeAI = Object.freeze({ GetSnapshot, SaveConfig });\n'
        'window.dispatchEvent(new Event("ohmyfreeai:bridge-ready"));\n'
    )
    (source_dir / "main.js").write_text(bridge, encoding="utf-8")


def build(output_directory: Path) -> Path:
    require_tools()
    build_root = ROOT / ".build"
    build_root.mkdir(parents=True, exist_ok=True)
    output_directory.mkdir(parents=True, exist_ok=True)

    with tempfile.TemporaryDirectory(prefix="wails-", dir=build_root) as temporary:
        temporary_root = Path(temporary)
        project = temporary_root / "OhMyFreeAI"

        run(
            [
                "wails3",
                "init",
                "-n",
                "OhMyFreeAI",
                "-t",
                "vanilla-js",
                "-d",
                str(project),
                "-mod",
                MODULE_PATH,
                "-skipgomodtidy",
            ],
            cwd=temporary_root,
        )

        shutil.copy2(ROOT / "desktop-src" / "main.go.tmpl", project / "main.go")
        shutil.copy2(ROOT / "desktop-src" / "app_service.go.tmpl", project / "app_service.go")
        demo_service = project / "greetservice.go"
        if demo_service.exists():
            demo_service.unlink()

        copy_frontend(project)

        relative_root = Path(os.path.relpath(ROOT, project)).as_posix()
        run(["go", "mod", "edit", "-go=1.27.3"], cwd=project)
        run(["go", "mod", "edit", f"-require={ROOT_MODULE}@v0.0.0"], cwd=project)
        run(["go", "mod", "edit", f"-replace={ROOT_MODULE}={relative_root}"], cwd=project)
        run(["go", "mod", "tidy"], cwd=project)

        run(["wails3", "generate", "bindings"], cwd=project)
        attach_generated_binding(project)
        run(["wails3", "build"], cwd=project)

        binary_directory = project / "bin"
        if not binary_directory.is_dir():
            raise RuntimeError("Wails build did not create the bin directory.")
        binaries = [
            path for path in binary_directory.iterdir()
            if path.is_file() and path.name.lower().startswith("ohmyfreeai")
            and path.suffix.lower() in ("", ".exe")
        ]
        if not binaries:
            raise RuntimeError("Wails build did not produce the OhMyFreeAI executable.")

        binary = sorted(binaries, key=lambda path: path.suffix.lower() == ".exe")[0]
        target_name = "OhMyFreeAI.exe" if os.name == "nt" else "OhMyFreeAI"
        destination = output_directory / target_name
        shutil.copy2(binary, destination)
        return destination


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--output-dir",
        type=Path,
        default=ROOT / "dist",
        help="directory for the resulting native executable",
    )
    args = parser.parse_args()

    try:
        output = build(args.output_dir.resolve())
    except subprocess.CalledProcessError as error:
        print(f"Desktop build failed with exit code {error.returncode}.", file=sys.stderr)
        return error.returncode or 1
    except (OSError, RuntimeError) as error:
        print(f"Desktop build failed: {error}", file=sys.stderr)
        return 1

    print(f"Desktop executable: {output}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
