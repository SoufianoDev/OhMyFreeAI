#!/usr/bin/env python3
"""Bundle CUDA toolkit shared libraries required by a llama.cpp CUDA build.

The NVIDIA driver library (libcuda.so.1) is supplied by the host and is never bundled.
"""
from __future__ import annotations
import argparse
import os
import re
import shutil
import subprocess
from pathlib import Path

CUDA_PREFIXES = (
    "libcudart.so", "libcublas.so", "libcublasLt.so", "libculibos.so",
    "libnvrtc.so", "libnvrtc-builtins.so", "libnvJitLink.so",
    "libcurand.so", "libcusparse.so", "libcusolver.so", "libcupti.so",
    "libnvToolsExt.so", "libnvfatbin.so", "libnvptxcompiler.so",
)

def is_cuda_runtime(name: str) -> bool:
    return name.startswith(CUDA_PREFIXES)

def dependencies(target: Path, env: dict[str, str]) -> list[tuple[str, Path]]:
    result = subprocess.run(["ldd", str(target)], text=True, stdout=subprocess.PIPE,
                            stderr=subprocess.STDOUT, env=env, timeout=45)
    if result.returncode:
        raise RuntimeError(f"ldd failed for {target}:\n{result.stdout}")
    found = []
    for line in result.stdout.splitlines():
        match = re.match(r"\s*(\S+)\s+=>\s+(/\S+)", line)
        if match and is_cuda_runtime(match.group(1)):
            found.append((match.group(1), Path(match.group(2))))
    return found

def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", type=Path, required=True)
    parser.add_argument("--cuda-lib-dir", type=Path, required=True)
    args = parser.parse_args()
    root = args.root.resolve()
    bindir = root / "bin"
    cuda_lib = args.cuda_lib_dir.resolve()
    if not bindir.is_dir() or not cuda_lib.is_dir():
        raise RuntimeError("Engine bin directory or CUDA library directory is missing.")
    env = os.environ.copy()
    env["LD_LIBRARY_PATH"] = os.pathsep.join(
        [str(bindir), str(cuda_lib), env.get("LD_LIBRARY_PATH", "")]
    )
    pending = [p for p in bindir.iterdir()
               if p.is_file() and (p.name == "llama-server" or ".so" in p.name)]
    visited: set[Path] = set()
    copied: dict[str, str] = {}
    while pending:
        target = pending.pop().resolve()
        if target in visited:
            continue
        visited.add(target)
        for soname, source in dependencies(target, env):
            if not source.is_file():
                raise RuntimeError(f"CUDA dependency {soname} missing at {source}")
            real = source.resolve(strict=True)
            destination = bindir / real.name
            if not destination.exists():
                shutil.copy2(real, destination)
                pending.append(destination)
            alias = bindir / soname
            if alias.name != real.name and not alias.exists():
                alias.symlink_to(real.name)
            copied[soname] = real.name
    if not any(n.startswith("libcudart.so") for n in copied):
        raise RuntimeError("No CUDA runtime library found; refusing partial CUDA bundle.")
    if not any(n.startswith("libcublas.so") for n in copied):
        raise RuntimeError("No cuBLAS library found; refusing partial CUDA bundle.")
    print("CUDA_RUNTIME_BUNDLING=PASS")
    for name, target in sorted(copied.items()):
        print(f"{name} -> {target}")
    print(f"CUDA_RUNTIME_LIBRARY_COUNT={len(copied)}")
    return 0

if __name__ == "__main__":
    raise SystemExit(main())
