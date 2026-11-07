#!/usr/bin/env python3
"""OhMyFreeAI developer entry point. Python standard library only."""
from __future__ import annotations

import argparse
import platform
import re
import shutil
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
REQUIRED_DOCS = (
    "docs/architecture/PROJECT_SPEC.md",
    "docs/architecture/SYSTEM_ARCHITECTURE.md",
    "docs/architecture/MODULE_BOUNDARIES.md",
    "docs/architecture/COMPATIBILITY_ARCHITECTURE.md",
    "docs/legal/COLAB_RISK_DISCLOSURE.md",
)


def run(command: list[str]) -> int:
    print("$", " ".join(command))
    return subprocess.run(command, cwd=ROOT, check=False).returncode


def version(name: str, args: list[str] | None = None) -> str:
    path = shutil.which(name)
    if not path:
        return "not installed"
    result = subprocess.run([path, *(args or ["--version"])], cwd=ROOT,
                            capture_output=True, text=True, check=False)
    output = (result.stdout or result.stderr).strip().splitlines()
    return output[0] if output else f"available (exit {result.returncode})"


def doctor(_: argparse.Namespace) -> int:
    print(f"OhMyFreeAI doctor — {platform.system()} {platform.machine()}")
    for name, args in (("python3", ["--version"]), ("go", ["version"]),
                       ("node", ["--version"]), ("npm", ["--version"]),
                       ("wails", ["version"]), ("wails3", ["version"])):
        print(f"{name:8} {version(name, args)}")
    go_output = version("go", ["version"])
    match = re.search(r"go(\d+)\.(\d+)\.(\d+)", go_output)
    if match:
        current = tuple(int(part) for part in match.groups())
        target = (1, 27, 1)
        state = "meets target" if current >= target else "below project target 1.27.3"
        print(f"Go target {target[0]}.{target[1]}.{target[2]}: {state}")
    print("No dependencies were installed or changed.")
    return 0


def go_command(subcommand: str, _: argparse.Namespace) -> int:
    if not shutil.which("go"):
        print("Go is required for this command.", file=sys.stderr)
        return 2
    return run(["go", subcommand, "./..."])


def test(_: argparse.Namespace) -> int:
    go_status = go_command("test", _)
    if go_status != 0:
        return go_status
    return run([sys.executable, "-m", "unittest", "discover", "-s", "tests", "-v"])


def fmt(_: argparse.Namespace) -> int:
    if not shutil.which("gofmt"):
        print("gofmt is required for this command.", file=sys.stderr)
        return 2
    files = sorted(str(p.relative_to(ROOT)) for p in ROOT.rglob("*.go")
                   if not any(part.startswith(".") for part in p.relative_to(ROOT).parts))
    if not files:
        print("No Go source files found.")
        return 0
    return run(["gofmt", "-w", *files])


def arch_check(_: argparse.Namespace) -> int:
    errors: list[str] = []
    for rel in REQUIRED_DOCS:
        if not (ROOT / rel).is_file():
            errors.append(f"missing required architecture document: {rel}")
    domain = ROOT / "internal" / "domain"
    if domain.exists():
        for path in domain.rglob("*.go"):
            source = path.read_text(encoding="utf-8")
            if re.search(r'"(github\.com/wailsapp/wails|database/sql|net/http)"', source):
                errors.append(f"framework or I/O import in domain layer: {path.relative_to(ROOT)}")
    if errors:
        print("Architecture check failed:", file=sys.stderr)
        for error in errors:
            print(f"- {error}", file=sys.stderr)
        return 1
    print("Architecture document and domain-boundary checks passed.")
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="command", required=True)
    sub.add_parser("doctor", help="inspect local tools without installing anything").set_defaults(func=doctor)
    sub.add_parser("test", help="run Go and Python standard-library tests").set_defaults(func=test)
    sub.add_parser("build", help="compile Go packages without producing a release package").set_defaults(func=lambda a: go_command("build", a))
    sub.add_parser("fmt", help="format Go source files with gofmt").set_defaults(func=fmt)
    sub.add_parser("arch-check", help="check required docs and domain import boundaries").set_defaults(func=arch_check)
    args = parser.parse_args()
    return args.func(args)


if __name__ == "__main__":
    raise SystemExit(main())
