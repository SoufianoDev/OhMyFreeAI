#!/usr/bin/env python3
"""Run an isolated CPU inference acceptance test for the pinned Colab bundle.

This does not modify the existing OhMyFreeAI launcher or install the engine
into Google Drive. It downloads the immutable CI-accepted bundle into a
temporary directory, verifies SHA-256, extracts it, starts llama-server against
an explicitly supplied GGUF model, and sends one real completion request.
"""
from __future__ import annotations

import argparse
import hashlib
import json
import os
import platform
import shutil
import socket
import subprocess
import sys
import tarfile
import tempfile
import time
import urllib.error
import urllib.request
from pathlib import Path

REPO = "SoufianoDev/OhMyFreeAI"
ACCEPTED_COMMIT = "b60005b78db6320af023910951898693a14ca1aa"
BASE_URL = f"https://raw.githubusercontent.com/{REPO}/{ACCEPTED_COMMIT}/engine"
ARCHIVE_URL = f"{BASE_URL}/llama-engine.tar.zst"
CHECKSUM_URL = f"{BASE_URL}/llama-engine.tar.zst.sha256"


def download(url: str, destination: Path) -> None:
    request = urllib.request.Request(url, headers={"User-Agent": "OhMyFreeAI-Colab-SmokeTest/1.0"})
    with urllib.request.urlopen(request, timeout=60) as response, destination.open("wb") as out:
        shutil.copyfileobj(response, out)


def sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for block in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def free_port() -> int:
    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        return int(sock.getsockname()[1])


def request_json(url: str, payload: dict | None = None, timeout: float = 10) -> dict:
    data = None if payload is None else json.dumps(payload).encode("utf-8")
    headers = {"Content-Type": "application/json"} if data is not None else {}
    request = urllib.request.Request(url, data=data, headers=headers)
    with urllib.request.urlopen(request, timeout=timeout) as response:
        return json.loads(response.read().decode("utf-8"))


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--model", type=Path, required=True, help="Path to an existing GGUF model; it is not downloaded.")
    parser.add_argument("--threads", type=int, default=2)
    parser.add_argument("--ctx-size", type=int, default=512)
    parser.add_argument("--startup-timeout", type=int, default=240)
    args = parser.parse_args()

    model = args.model.expanduser().resolve()
    if not model.is_file() or model.stat().st_size == 0:
        raise RuntimeError(f"GGUF model does not exist or is empty: {model}")
    if platform.system() != "Linux" or platform.machine().lower() not in {"x86_64", "amd64"}:
        raise RuntimeError(f"This bundle targets Linux x86_64, got {platform.system()} {platform.machine()}")
    zstd = shutil.which("zstd")
    tar = shutil.which("tar")
    if not zstd or not tar:
        raise RuntimeError("Missing required extraction tools. Install zstd and tar in this Colab runtime.")

    with tempfile.TemporaryDirectory(prefix="ohmyfreeai-engine-smoke-") as temp_name:
        temp = Path(temp_name)
        archive = temp / "llama-engine.tar.zst"
        checksum_file = temp / "llama-engine.tar.zst.sha256"
        download(ARCHIVE_URL, archive)
        download(CHECKSUM_URL, checksum_file)

        checksum_fields = checksum_file.read_text(encoding="utf-8").strip().split()
        if not checksum_fields:
            raise RuntimeError("Downloaded checksum file is empty.")
        expected = checksum_fields[0].lower()
        actual = sha256(archive)
        if actual != expected:
            raise RuntimeError(f"SHA-256 mismatch: expected {expected}, got {actual}")
        print(f"SHA256 verified: {actual}")

        subprocess.run([zstd, "-t", str(archive)], check=True, timeout=60)
        extract_root = temp / "extract"
        extract_root.mkdir()
        with archive.open("rb") as source:
            decoder = subprocess.Popen([zstd, "-q", "-d", "-c"], stdin=source, stdout=subprocess.PIPE)
            assert decoder.stdout is not None
            extractor = subprocess.Popen(
                [tar, "-xf", "-", "-C", str(extract_root), "--no-same-owner", "--no-same-permissions"],
                stdin=decoder.stdout,
            )
            decoder.stdout.close()
            tar_rc = extractor.wait()
            zstd_rc = decoder.wait()
        if tar_rc or zstd_rc:
            raise RuntimeError(f"Archive extraction failed (tar={tar_rc}, zstd={zstd_rc})")

        engine_root = extract_root / "llama-engine"
        server = engine_root / "bin" / "llama-server"
        if not server.is_file() or not os.access(server, os.X_OK):
            raise RuntimeError("Bundle does not contain an executable bin/llama-server.")
        version = subprocess.run(
            [str(server), "--version"], cwd=server.parent, text=True,
            stdout=subprocess.PIPE, stderr=subprocess.STDOUT, check=True, timeout=30,
        ).stdout.strip()
        if "version: 0.9.0" not in version or "-dev" in version:
            raise RuntimeError(f"Unexpected engine version: {version}")
        print(f"Engine version: {version}")

        port = free_port()
        env = os.environ.copy()
        env["LD_LIBRARY_PATH"] = str(server.parent) + os.pathsep + env.get("LD_LIBRARY_PATH", "")
        log_path = temp / "llama-server.log"
        with log_path.open("w", encoding="utf-8") as log:
            process = subprocess.Popen(
                [str(server), "--model", str(model), "--host", "127.0.0.1",
                 "--port", str(port), "--ctx-size", str(args.ctx_size),
                 "--threads", str(args.threads)],
                cwd=server.parent, env=env, stdout=log, stderr=subprocess.STDOUT,
                text=True,
            )
            try:
                base = f"http://127.0.0.1:{port}"
                deadline = time.monotonic() + args.startup_timeout
                ready = False
                last_error = None
                while time.monotonic() < deadline:
                    if process.poll() is not None:
                        raise RuntimeError(f"llama-server exited early with code {process.returncode}. Log: {log_path}")
                    try:
                        request_json(base + "/health", timeout=3)
                        ready = True
                        break
                    except (urllib.error.URLError, TimeoutError, json.JSONDecodeError) as exc:
                        last_error = exc
                        time.sleep(1)
                if not ready:
                    raise RuntimeError(f"llama-server did not become healthy in time: {last_error}. Log: {log_path}")

                result = request_json(
                    base + "/completion",
                    {"prompt": "Reply with exactly: ENGINE_OK", "n_predict": 12, "temperature": 0},
                    timeout=120,
                )
                output = str(result.get("content", "")).strip()
                if not output:
                    raise RuntimeError(f"Completion endpoint returned no content: {result}")
                print("Inference response:", output)
                print("CPU model-backed inference: PASSED")
                return 0
            finally:
                if process.poll() is None:
                    process.terminate()
                    try:
                        process.wait(timeout=10)
                    except subprocess.TimeoutExpired:
                        process.kill()
                        process.wait(timeout=5)
                log.flush()
                if process.returncode not in (0, -15, 143):
                    print(f"llama-server exit code: {process.returncode}; log: {log_path}", file=sys.stderr)


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as exc:
        print(f"SMOKE TEST FAILED: {exc}", file=sys.stderr)
        raise SystemExit(1)
