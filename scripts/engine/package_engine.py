#!/usr/bin/env python3
"""Create and benchmark lossless llama.cpp TAR+Zstandard bundles.

The selected level is the smallest archive whose measured extraction time is
within 15% of the fastest measured extraction. No archive is published when
it exceeds the configured size ceiling.
"""
from __future__ import annotations

import argparse
import hashlib
import json
import os
import shutil
import subprocess
import sys
import tarfile
import tempfile
import time
from pathlib import Path


def sha256(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        for block in iter(lambda: f.read(1024 * 1024), b""):
            h.update(block)
    return h.hexdigest()


def tree_digest(root: Path) -> dict[str, str]:
    result: dict[str, str] = {}
    for path in sorted(p for p in root.rglob("*") if p.is_file()):
        result[path.relative_to(root).as_posix()] = sha256(path)
    return result


def require_tool(name: str) -> str:
    found = shutil.which(name)
    if not found:
        raise RuntimeError(f"Required tool not found in PATH: {name}")
    return found


def run_pipeline(source: Path, archive: Path, level: int) -> float:
    started = time.perf_counter()
    with archive.open("wb") as output:
        tar = subprocess.Popen(
            ["tar", "--sort=name", "--mtime=@0", "--owner=0", "--group=0",
             "--numeric-owner", "-cf", "-", "-C", str(source.parent), source.name],
            stdout=subprocess.PIPE,
        )
        assert tar.stdout is not None
        zstd = subprocess.Popen(
            ["zstd", "-q", "-T0", f"-{level}", "-c"],
            stdin=tar.stdout, stdout=output,
        )
        tar.stdout.close()
        zstd_rc = zstd.wait()
        tar_rc = tar.wait()
    if tar_rc or zstd_rc:
        raise RuntimeError(f"Packaging failed (tar={tar_rc}, zstd={zstd_rc})")
    return time.perf_counter() - started


def extraction_time(archive: Path, destination: Path) -> float:
    started = time.perf_counter()
    with archive.open("rb") as source:
        zstd = subprocess.Popen(
            ["zstd", "-q", "-d", "-c"], stdin=source, stdout=subprocess.PIPE
        )
        assert zstd.stdout is not None
        tar = subprocess.Popen(
            ["tar", "-xf", "-", "-C", str(destination),
             "--no-same-owner", "--no-same-permissions"],
            stdin=zstd.stdout,
        )
        zstd.stdout.close()
        tar_rc = tar.wait()
        zstd_rc = zstd.wait()
    if tar_rc or zstd_rc:
        raise RuntimeError(f"Extraction failed (tar={tar_rc}, zstd={zstd_rc})")
    return time.perf_counter() - started


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--source", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path, required=True)
    parser.add_argument("--name", default="llama-engine")
    parser.add_argument("--levels", type=int, nargs="+", default=[3, 6, 9, 15, 19])
    parser.add_argument("--max-size-mib", type=float, default=90.0)
    args = parser.parse_args()

    source = args.source.resolve()
    if not source.is_dir():
        parser.error(f"Source directory does not exist: {source}")
    require_tool("tar")
    require_tool("zstd")
    args.output_dir.mkdir(parents=True, exist_ok=True)

    expected = tree_digest(source)
    if not expected:
        raise RuntimeError("Refusing to package an empty engine directory")

    results = []
    with tempfile.TemporaryDirectory(prefix="llama-zstd-bench-") as tmp:
        temp = Path(tmp)
        for level in args.levels:
            if level < 1 or level > 19:
                raise ValueError(f"Unsupported benchmark level: {level}")
            candidate = temp / f"{args.name}-level-{level}.tar.zst"
            compression_seconds = run_pipeline(source, candidate, level)
            extract_dir = temp / f"extract-{level}"
            extract_dir.mkdir()
            extraction_seconds = extraction_time(candidate, extract_dir)
            extracted_root = extract_dir / source.name
            actual = tree_digest(extracted_root)
            if actual != expected:
                raise RuntimeError(f"Lossless verification failed at Zstd level {level}")
            results.append({
                "level": level,
                "archive_bytes": candidate.stat().st_size,
                "compression_seconds": round(compression_seconds, 4),
                "extraction_seconds": round(extraction_seconds, 4),
                "sha256": sha256(candidate),
            })

        fastest = min(item["extraction_seconds"] for item in results)
        eligible = [
            item for item in results
            if item["extraction_seconds"] <= fastest * 1.15
            and item["archive_bytes"] <= args.max_size_mib * 1024 * 1024
        ]
        if not eligible:
            smallest = min(results, key=lambda x: x["archive_bytes"])
            raise RuntimeError(
                f"No tested archive satisfies both extraction and {args.max_size_mib} MiB "
                f"size gates. Smallest measured archive: {smallest['archive_bytes']} bytes."
            )
        selected = min(eligible, key=lambda x: x["archive_bytes"])
        final_archive = args.output_dir / f"{args.name}.tar.zst"
        shutil.copy2(temp / f"{args.name}-level-{selected['level']}.tar.zst", final_archive)

    report = {
        "format": "tar+zstd",
        "source_directory": source.name,
        "source_tree_sha256": hashlib.sha256(
            json.dumps(expected, sort_keys=True).encode()
        ).hexdigest(),
        "selection_rule": "smallest archive within 15% of fastest measured extraction",
        "max_size_mib": args.max_size_mib,
        "selected_level": selected["level"],
        "selected_archive": final_archive.name,
        "selected_archive_bytes": final_archive.stat().st_size,
        "selected_archive_sha256": sha256(final_archive),
        "benchmarks": results,
    }
    report_path = args.output_dir / "compression-report.json"
    report_path.write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
    (args.output_dir / f"{final_archive.name}.sha256").write_text(
        f"{sha256(final_archive)}  {final_archive.name}\n", encoding="utf-8"
    )
    print(json.dumps(report, indent=2))
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as exc:
        print(f"ERROR: {exc}", file=sys.stderr)
        raise SystemExit(1)
