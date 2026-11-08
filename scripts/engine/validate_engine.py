#!/usr/bin/env python3
"""Validate an extracted llama.cpp engine bundle without claiming GPU readiness."""
from __future__ import annotations

import argparse
import json
import shutil
import subprocess
import sys
from pathlib import Path


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", type=Path, required=True)
    parser.add_argument("--server", default="bin/llama-server")
    parser.add_argument("--require-ldd-clean", action="store_true")
    args = parser.parse_args()

    root = args.root.resolve()
    server = root / args.server
    if not server.is_file():
        raise RuntimeError(f"Missing server executable: {server}")
    if not server.stat().st_mode & 0o111:
        raise RuntimeError(f"Server is not executable: {server}")

    version = subprocess.run(
        [str(server), "--version"], cwd=root, text=True,
        stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=30,
    )
    if version.returncode != 0:
        raise RuntimeError(f"llama-server --version failed:\n{version.stdout}")

    ldd_result = None
    ldd = shutil.which("ldd")
    if args.require_ldd_clean and not ldd:
        raise RuntimeError("ldd is required for dependency validation but was not found in PATH")
    if ldd:
        ldd_result = subprocess.run(
            [ldd, str(server)], cwd=root, text=True,
            stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=30,
        )
        if args.require_ldd_clean and ldd_result.returncode != 0:
            raise RuntimeError(
                f"ldd failed with exit code {ldd_result.returncode}:\\n{ldd_result.stdout}"
            )
        missing = [line.strip() for line in ldd_result.stdout.splitlines()
                   if "not found" in line]
        if missing and args.require_ldd_clean:
            raise RuntimeError("Missing shared libraries:\n" + "\n".join(missing))

    report = {
        "status": "passed",
        "server": str(server.relative_to(root)),
        "version_output": version.stdout.strip(),
        "ldd_available": bool(ldd),
        "ldd_exit_code": ldd_result.returncode if ldd_result else None,
        "missing_shared_libraries": (
            [line.strip() for line in ldd_result.stdout.splitlines()
             if "not found" in line] if ldd_result else []
        ),
        "gpu_readiness": "NOT_TESTED — requires a compatible real Colab GPU",
        "inference_readiness": "NOT_TESTED — requires a model-backed smoke test",
    }
    print(json.dumps(report, indent=2))
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as exc:
        print(f"ERROR: {exc}", file=sys.stderr)
        raise SystemExit(1)
