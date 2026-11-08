#!/usr/bin/env python3
"""Validate llama.cpp server and every bundled dynamic backend dependency."""
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

    # --list-devices intentionally lists accelerator devices only in llama.cpp
    # v0.9.0; a CPU-only build correctly prints "(none)". Do not use this
    # output as a CPU-backend health check.
    devices = subprocess.run(
        [str(server), "--list-devices"], cwd=root, text=True,
        stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=30,
    )
    if devices.returncode != 0:
        raise RuntimeError(
            "llama-server --list-devices failed:\n" + devices.stdout
        )

    # With GGML_BACKEND_DL=ON, the CPU implementation is shipped as a
    # dynamically loaded plugin. Verify the generic x86-64 variant is present;
    # actual backend initialization still requires a model-backed smoke test.
    cpu_plugins = sorted(root.glob("bin/libggml-cpu-*.so"))
    if not cpu_plugins:
        raise RuntimeError(
            "No dynamically loaded CPU backend plugin found in bin/ "
            "(expected libggml-cpu-*.so)."
        )
    generic_cpu = root / "bin/libggml-cpu-x64.so"
    if not generic_cpu.is_file():
        raise RuntimeError(
            "Missing portable x86-64 CPU backend plugin: bin/libggml-cpu-x64.so"
        )

    ldd = shutil.which("ldd")
    if args.require_ldd_clean and not ldd:
        raise RuntimeError("ldd is required for dependency validation but was not found in PATH")

    # Dynamic backend loading means the server executable is not the only ELF
    # runtime component. Check every bundled shared object as well.
    targets = [server, *sorted(p for p in root.rglob("*.so*") if p.is_file())]
    dependency_checks = []
    if ldd:
        for target in targets:
            result = subprocess.run(
                [ldd, str(target)], cwd=root, text=True,
                stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=30,
            )
            missing = [line.strip() for line in result.stdout.splitlines()
                       if "not found" in line]
            item = {
                "file": str(target.relative_to(root)),
                "exit_code": result.returncode,
                "missing_shared_libraries": missing,
                "output": result.stdout.strip(),
            }
            dependency_checks.append(item)
            if args.require_ldd_clean and (result.returncode != 0 or missing):
                raise RuntimeError(
                    f"Dependency validation failed for {item['file']}:\n{result.stdout}"
                )

    report = {
        "status": "passed",
        "server": str(server.relative_to(root)),
        "version_output": version.stdout.strip(),
        "device_enumeration_output": devices.stdout.strip(),
        "cpu_backend_plugin_count": len(cpu_plugins),
        "cpu_backend_plugins": [p.name for p in cpu_plugins],
        "cpu_backend_validation": "portable x86-64 plugin presence checked; actual initialization requires model-backed smoke test",
        "dependency_targets_checked": len(dependency_checks),
        "dependency_checks": dependency_checks,
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
