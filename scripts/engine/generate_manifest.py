#!/usr/bin/env python3
"""Generate a provenance manifest from the actual build environment."""
from __future__ import annotations

import argparse
import json
import os
import platform
import shutil
import subprocess
from datetime import datetime, timezone
from pathlib import Path


def command_version(command: list[str]) -> str | None:
    if not shutil.which(command[0]):
        return None
    result = subprocess.run(command, text=True, stdout=subprocess.PIPE,
                            stderr=subprocess.STDOUT, timeout=20)
    return result.stdout.strip().splitlines()[0] if result.stdout.strip() else None


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--source-ref", required=True)
    parser.add_argument("--source-mode", choices=["official-asset", "source-build"], required=True)
    parser.add_argument("--artifact-url", default="")
    parser.add_argument("--build-flags", default="")
    parser.add_argument("--backend", default="cpu")
    args = parser.parse_args()

    data = {
        "schema_version": 1,
        "engine": "llama.cpp",
        "created_utc": datetime.now(timezone.utc).isoformat(),
        "source_repository": "https://github.com/ggml-org/llama.cpp",
        "source_ref": args.source_ref,
        "source_mode": args.source_mode,
        "official_artifact_url": args.artifact_url or None,
        "platform": platform.platform(),
        "system": platform.system(),
        "architecture": platform.machine(),
        "python": platform.python_version(),
        "compiler": command_version(["c++", "--version"]),
        "cmake": command_version(["cmake", "--version"]),
        "ninja": command_version(["ninja", "--version"]),
        "zstd": command_version(["zstd", "--version"]),
        "backend_requested": args.backend,
        "build_flags": args.build_flags,
        "cuda": {
            "nvidia_smi": command_version(["nvidia-smi", "--query-gpu=name,driver_version",
                                           "--format=csv,noheader"]),
            "nvcc": command_version(["nvcc", "--version"]),
        },
        "validation_policy": {
            "binary_version": "required",
            "shared_libraries": "required",
            "real_inference": "required before runtime adoption",
            "cuda_inference": "must be tested on an actual compatible Colab GPU",
        },
    }
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
