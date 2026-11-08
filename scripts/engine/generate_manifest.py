#!/usr/bin/env python3
"""Generate provenance including exact bundled files and runtime dependencies."""
from __future__ import annotations

import argparse
import hashlib
import json
import os
import platform
import shutil
import subprocess
from datetime import datetime, timezone
from pathlib import Path


def command_output(command: list[str], timeout: int = 20) -> str | None:
    if not shutil.which(command[0]):
        return None
    result = subprocess.run(command, text=True, stdout=subprocess.PIPE,
                            stderr=subprocess.STDOUT, timeout=timeout)
    return result.stdout.strip() or None


def sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for block in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--source-ref", required=True)
    parser.add_argument("--source-commit", default=os.environ.get("SOURCE_COMMIT", ""))
    parser.add_argument("--source-mode", choices=["official-asset", "source-build"], required=True)
    parser.add_argument("--artifact-url", default="")
    parser.add_argument("--build-flags", default="")
    parser.add_argument("--backend", default="cpu")
    args = parser.parse_args()

    root = args.output.parent.resolve()
    files = []
    for path in sorted(p for p in root.rglob("*") if p.is_file() and p != args.output):
        files.append({
            "path": path.relative_to(root).as_posix(),
            "size_bytes": path.stat().st_size,
            "sha256": sha256(path),
        })

    server = root / "bin/llama-server"
    if not server.is_file():
        raise RuntimeError(f"Cannot create provenance: missing {server}")

    ldd = shutil.which("ldd")
    dependency_map = {}
    if ldd:
        targets = [server, *sorted(p for p in root.rglob("*.so*") if p.is_file())]
        for target in targets:
            result = subprocess.run([ldd, str(target)], cwd=root, text=True,
                                    stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
                                    timeout=30)
            if result.returncode != 0 or "not found" in result.stdout:
                raise RuntimeError(
                    f"Cannot record incomplete dependencies for {target.name}:\n{result.stdout}"
                )
            deps = []
            for line in result.stdout.splitlines():
                if "=>" in line:
                    deps.append(line.strip())
            dependency_map[target.relative_to(root).as_posix()] = deps

    data = {
        "schema_version": 2,
        "engine": "llama.cpp",
        "created_utc": datetime.now(timezone.utc).isoformat(),
        "source_repository": "https://github.com/ggml-org/llama.cpp",
        "source_ref": args.source_ref,
        "source_commit": args.source_commit or None,
        "source_mode": args.source_mode,
        "official_artifact_url": args.artifact_url or None,
        "platform": platform.platform(),
        "system": platform.system(),
        "architecture": platform.machine(),
        "python": platform.python_version(),
        "compiler": command_output(["c++", "--version"]),
        "cmake": command_output(["cmake", "--version"]),
        "ninja": command_output(["ninja", "--version"]),
        "zstd": command_output(["zstd", "--version"]),
        "glibc": command_output(["ldd", "--version"]),
        "backend_requested": args.backend,
        "build_flags": args.build_flags,
        "cuda": {
            "nvidia_smi": command_output(["nvidia-smi", "--query-gpu=name,driver_version",
                                          "--format=csv,noheader"]),
            "nvcc": command_output(["nvcc", "--version"]),
        },
        "package_filename": "llama-engine.tar.zst",
        "bundled_files": files,
        "runtime_dependencies": dependency_map,
        "validation_policy": {
            "binary_version": "required",
            "all_bundled_shared_objects": "ldd checked before publication",
            "real_inference": "required before runtime adoption",
            "cuda_inference": "must be tested on an actual compatible Colab GPU",
        },
    }
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
