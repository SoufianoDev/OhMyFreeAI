#!/usr/bin/env python3
"""Collect actual Linux/Colab compatibility data without assuming a GPU exists."""
from __future__ import annotations

import json
import os
import platform
import shutil
import subprocess
from datetime import datetime, timezone
from pathlib import Path


def command(args: list[str], timeout: int = 15) -> dict:
    executable = shutil.which(args[0])
    if not executable:
        return {"available": False, "command": args}
    try:
        result = subprocess.run(
            [executable, *args[1:]], text=True, stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT, timeout=timeout, check=False,
        )
        return {
            "available": True,
            "returncode": result.returncode,
            "output": result.stdout.strip(),
        }
    except Exception as exc:
        return {"available": True, "error": str(exc)}


def read_text(path: str) -> str | None:
    try:
        return Path(path).read_text(encoding="utf-8").strip()
    except OSError:
        return None


def main() -> None:
    meminfo = read_text("/proc/meminfo") or ""
    memory = {}
    for line in meminfo.splitlines():
        key, _, rest = line.partition(":")
        if key in {"MemTotal", "MemAvailable", "SwapTotal", "SwapFree"}:
            memory[key] = rest.strip()

    data = {
        "schema_version": 1,
        "collected_utc": datetime.now(timezone.utc).isoformat(),
        "system": platform.system(),
        "platform": platform.platform(),
        "kernel": platform.release(),
        "architecture": platform.machine(),
        "python": platform.python_version(),
        "cpu": {
            "processor": platform.processor(),
            "logical_cores": os.cpu_count(),
            "lscpu": command(["lscpu"]),
            "flags": next(
                (line.split(":", 1)[1].strip()
                 for line in (read_text("/proc/cpuinfo") or "").splitlines()
                 if line.lower().startswith("flags")),
                None,
            ),
        },
        "memory": memory,
        "disk": command(["df", "-h", "/content" if Path("/content").exists() else "/"]),
        "gpu": {
            "nvidia_smi": command([
                "nvidia-smi",
                "--query-gpu=name,uuid,compute_cap,driver_version,memory.total",
                "--format=csv,noheader",
            ]),
            "cuda_compiler": command(["nvcc", "--version"]),
            "cuda_runtime_files": [
                str(path) for path in (
                    list(Path("/usr/local").glob("cuda*/version.txt")) +
                    list(Path("/usr/local").glob("cuda*/lib64/libcudart.so*"))
                )[:20]
            ],
        },
        "notes": [
            "No GPU or CUDA availability is inferred from a missing command.",
            "This report is discovery data, not proof that llama.cpp GPU inference works.",
        ],
    }
    print(json.dumps(data, indent=2))


if __name__ == "__main__":
    main()
