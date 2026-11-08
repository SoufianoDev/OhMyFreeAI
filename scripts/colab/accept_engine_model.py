#!/usr/bin/env python3
"""Colab acceptance test: pinned SoufianoDev CPU engine + the project's Qwen3.5-9B GGUF.

Downloads only into /content/ohmyfreeai-acceptance. It never touches an existing
llama-server, the qwen-opencode session setup, Google Drive, or project files.
The published engine is CPU-only (GGML_CUDA=OFF); a visible Colab GPU is reported
but is not used by this artifact.
"""
from __future__ import annotations

import hashlib
import json
import os
import platform
import shutil
import socket
import subprocess
import sys
import time
import urllib.error
import urllib.request
from pathlib import Path

# Immutable, CI-accepted engine publication.
ENGINE_COMMIT = "b60005b78db6320af023910951898693a14ca1aa"
ENGINE_URL = (
    "https://raw.githubusercontent.com/SoufianoDev/OhMyFreeAI/"
    + ENGINE_COMMIT + "/engine/llama-engine.tar.zst"
)
ENGINE_SHA256 = "302acba151d40d983a19fbc2bbb59809890655173b65b7eef80aad5263393b9f"

# This is the Qwen3.5-9B Q4_K_M checkpoint referenced by the project's
# previously used Colab launcher. Both repository revision and LFS SHA256
# are pinned; no "main" branch or moving model alias is used.
HF_REVISION = "182be2fd6c7bc44887d88a91cb03ff009cc9f549"
HF_REPO = "bartowski/Qwen_Qwen3.5-9B-GGUF"
MODEL_FILE = "Qwen_Qwen3.5-9B-Q4_K_M.gguf"
MODEL_SHA256 = "d784ce9eda1a5a7b51e8f705a9e6310844bf4f173654d115823c775fdea56d43"
MODEL_BYTES = 6169341984

ROOT = Path("/content/ohmyfreeai-acceptance")
ENGINE_ARCHIVE = ROOT / "llama-engine.tar.zst"
MODEL_PATH = ROOT / "models" / MODEL_FILE
LOG_PATH = ROOT / "llama-server.log"
CTX_SIZE = 2048
STARTUP_TIMEOUT = 900


def say(message: str) -> None:
    print(f"\n[OhMyFreeAI test] {message}", flush=True)


def sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for block in iter(lambda: handle.read(8 * 1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def run_checked(command: list[str], *, timeout: int = 300) -> None:
    subprocess.run(command, check=True, timeout=timeout)


def download_verified(url: str, destination: Path, expected_sha256: str) -> None:
    destination.parent.mkdir(parents=True, exist_ok=True)
    if destination.is_file():
        actual = sha256(destination)
        if actual == expected_sha256:
            say(f"Already present and SHA-256 verified: {destination.name}")
            return
        bad = destination.with_name(destination.name + ".invalid")
        if bad.exists():
            bad.unlink()
        destination.replace(bad)
        say(f"Existing file had the wrong hash; preserved it as {bad.name}")

    partial = destination.with_name(destination.name + ".part")
    curl = shutil.which("curl")
    if not curl:
        raise RuntimeError("curl is required in the Colab runtime.")
    say(f"Downloading {destination.name}; existing .part downloads will resume.")
    run_checked([
        curl, "--location", "--fail", "--show-error",
        "--retry", "5", "--retry-all-errors", "--connect-timeout", "30",
        "--continue-at", "-", "--output", str(partial), url,
    ], timeout=7200)

    actual = sha256(partial)
    if actual != expected_sha256:
        raise RuntimeError(
            f"SHA-256 mismatch for {destination.name}: expected "
            f"{expected_sha256}, got {actual}. Partial file retained for diagnosis."
        )
    partial.replace(destination)
    say(f"SHA-256 verified: {actual}")


def read_mem_available_gib() -> float:
    values = {}
    for line in Path("/proc/meminfo").read_text().splitlines():
        if ":" in line:
            key, value = line.split(":", 1)
            if key in {"MemTotal", "MemAvailable"}:
                values[key] = int(value.strip().split()[0]) / 1024 / 1024
    return values.get("MemAvailable", 0.0)


def gpu_report() -> str:
    nvidia_smi = shutil.which("nvidia-smi")
    if not nvidia_smi:
        return "No nvidia-smi detected."
    result = subprocess.run(
        [nvidia_smi, "--query-gpu=name,memory.total,memory.free,driver_version",
         "--format=csv,noheader"],
        text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=15,
    )
    return result.stdout.strip() or "GPU details unavailable."


def ensure_zstd() -> str:
    zstd = shutil.which("zstd")
    if zstd:
        return zstd
    say("zstd is missing; installing the small Ubuntu zstd package.")
    apt = shutil.which("apt-get")
    if not apt:
        raise RuntimeError("zstd is missing and apt-get is unavailable.")
    run_checked([apt, "update", "-qq"], timeout=600)
    run_checked([apt, "install", "-y", "-qq", "zstd"], timeout=600)
    zstd = shutil.which("zstd")
    if not zstd:
        raise RuntimeError("zstd installation did not provide the zstd executable.")
    return zstd


def extract_engine(zstd: str, destination: Path) -> Path:
    destination.mkdir(parents=True, exist_ok=True)
    say("Checking Zstandard archive integrity.")
    run_checked([zstd, "-t", str(ENGINE_ARCHIVE)], timeout=120)
    say("Extracting the engine into the isolated acceptance directory.")
    run_checked([
        "tar", "--use-compress-program=" + zstd,
        "-xf", str(ENGINE_ARCHIVE), "-C", str(destination),
    ], timeout=300)
    server = destination / "llama-engine" / "bin" / "llama-server"
    if not server.is_file() or not os.access(server, os.X_OK):
        raise RuntimeError(f"llama-server is missing or not executable: {server}")
    return server


def http_json(url: str, payload: dict | None = None, timeout: int = 60) -> dict:
    data = None if payload is None else json.dumps(payload).encode("utf-8")
    headers = {"Content-Type": "application/json"} if data is not None else {}
    request = urllib.request.Request(url, data=data, headers=headers)
    with urllib.request.urlopen(request, timeout=timeout) as response:
        return json.loads(response.read().decode("utf-8"))


def free_port() -> int:
    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        return int(sock.getsockname()[1])


def main() -> int:
    if platform.system() != "Linux" or platform.machine().lower() not in {"x86_64", "amd64"}:
        raise RuntimeError(
            f"The published artifact targets Linux x86_64, got "
            f"{platform.system()} {platform.machine()}."
        )
    if not shutil.which("tar"):
        raise RuntimeError("tar is missing from this Colab runtime.")

    ROOT.mkdir(parents=True, exist_ok=True)
    say("Runtime facts")
    print("CPU threads:", os.cpu_count())
    print("RAM available (GiB):", round(read_mem_available_gib(), 2))
    print("GPU:", gpu_report())
    print("Engine target: Linux x86_64 CPU (CUDA is disabled in this published bundle).")

    # The GGUF is 6.17 GB. Stop early if the current runtime cannot reasonably
    # load it, rather than letting the kernel be OOM-killed during the test.
    available = read_mem_available_gib()
    if available < 7.0:
        raise RuntimeError(
            f"Only {available:.2f} GiB RAM is available. This test requires "
            "at least 7 GiB free for the 9B Q4_K_M model and a small context. "
            "No downloads have been started."
        )
    free_disk = shutil.disk_usage(ROOT).free / (1024 ** 3)
    if free_disk < 9.0:
        raise RuntimeError(f"Only {free_disk:.2f} GiB disk space is free; at least 9 GiB is required.")

    # Engine archive and model are both immutable and checksum-pinned.
    download_verified(ENGINE_URL, ENGINE_ARCHIVE, ENGINE_SHA256)
    model_url = (
        f"https://huggingface.co/{HF_REPO}/resolve/{HF_REVISION}/"
        f"{MODEL_FILE}?download=true"
    )
    download_verified(model_url, MODEL_PATH, MODEL_SHA256)
    if MODEL_PATH.stat().st_size != MODEL_BYTES:
        raise RuntimeError(
            f"Unexpected GGUF byte size: {MODEL_PATH.stat().st_size}; "
            f"expected {MODEL_BYTES}."
        )

    zstd = ensure_zstd()
    server = extract_engine(zstd, ROOT / "engine")
    server_dir = server.parent

    env = os.environ.copy()
    env["LD_LIBRARY_PATH"] = str(server_dir) + os.pathsep + env.get("LD_LIBRARY_PATH", "")

    version = subprocess.run(
        [str(server), "--version"], cwd=server_dir, env=env,
        text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
        check=True, timeout=30,
    ).stdout.strip()
    print("Engine version:", version)
    if "version: 0.9.0" not in version or "-dev" in version:
        raise RuntimeError(f"Unexpected llama-server version: {version}")

    ldd = subprocess.run(
        ["ldd", str(server)], cwd=server_dir, env=env,
        text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
        check=False, timeout=30,
    )
    if "not found" in ldd.stdout:
        raise RuntimeError("Missing shared libraries:\n" + ldd.stdout)

    port = free_port()
    threads = max(1, min(os.cpu_count() or 2, 4))
    command = [
        str(server),
        "--model", str(MODEL_PATH),
        "--host", "127.0.0.1",
        "--port", str(port),
        "--ctx-size", str(CTX_SIZE),
        "--threads", str(threads),
        "--n-gpu-layers", "0",
        "--parallel", "1",
    ]
    say("Starting the pinned CPU llama-server against Qwen3.5-9B Q4_K_M.")
    print("Command:", " ".join(command), flush=True)
    print("Server log:", LOG_PATH, flush=True)

    with LOG_PATH.open("w", encoding="utf-8") as log:
        process = subprocess.Popen(
            command, cwd=server_dir, env=env,
            stdin=subprocess.DEVNULL, stdout=log, stderr=subprocess.STDOUT,
            start_new_session=True,
        )
        try:
            base = f"http://127.0.0.1:{port}"
            deadline = time.monotonic() + STARTUP_TIMEOUT
            ready = False
            last_error = None
            while time.monotonic() < deadline:
                if process.poll() is not None:
                    log.flush()
                    tail = "\n".join(LOG_PATH.read_text(errors="replace").splitlines()[-70:])
                    raise RuntimeError(
                        f"llama-server exited during model loading "
                        f"(code {process.returncode}).\n{tail}"
                    )
                try:
                    with urllib.request.urlopen(base + "/health", timeout=5) as response:
                        if response.status == 200:
                            ready = True
                            break
                except Exception as exc:
                    last_error = repr(exc)
                time.sleep(2)

            if not ready:
                log.flush()
                tail = "\n".join(LOG_PATH.read_text(errors="replace").splitlines()[-70:])
                raise RuntimeError(
                    f"Model did not become ready within {STARTUP_TIMEOUT}s. "
                    f"Last health error: {last_error}\n{tail}"
                )

            say("Model loaded. Sending a real non-streaming chat completion.")
            result = http_json(
                base + "/v1/chat/completions",
                {
                    "model": "qwen35-9b-q4km",
                    "messages": [
                        {"role": "system", "content": "You are a concise assistant."},
                        {"role": "user", "content": "Reply with exactly this token: SOUFIANODEV_ENGINE_OK"},
                    ],
                    "temperature": 0,
                    "max_tokens": 32,
                    "stream": False,
                    "chat_template_kwargs": {"enable_thinking": False},
                },
                timeout=600,
            )
            message = result.get("choices", [{}])[0].get("message", {})
            answer = message.get("content") or message.get("reasoning_content") or ""
            if not answer.strip():
                raise RuntimeError("The model loaded, but the completion response contained no text: " +
                                   json.dumps(result, ensure_ascii=False)[:2000])
            print("\nMODEL_RESPONSE:")
            print(answer.strip())
            print("\nPROMPT_TOKENS:", result.get("usage", {}).get("prompt_tokens"))
            print("COMPLETION_TOKENS:", result.get("usage", {}).get("completion_tokens"))
            print("GPU_VISIBLE_BUT_NOT_USED_BY_THIS_CPU_BUNDLE:", gpu_report())
            print("\nACCEPTANCE_RESULT=PASS")
            print("Engine: pinned SoufianoDev bundle, llama.cpp v0.9.0")
            print("Model: Qwen3.5-9B Q4_K_M, immutable Hugging Face revision")
            print("Inference: real /v1/chat/completions request succeeded")
            return 0
        finally:
            if process.poll() is None:
                process.terminate()
                try:
                    process.wait(timeout=15)
                except subprocess.TimeoutExpired:
                    process.kill()
                    process.wait(timeout=10)
            log.flush()
            print("Test server stopped. Downloaded files remain in:", ROOT)


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as exc:
        print("\nACCEPTANCE_RESULT=FAIL", file=sys.stderr)
        print(type(exc).__name__ + ": " + str(exc), file=sys.stderr)
        print("If the server started, inspect /content/ohmyfreeai-acceptance/llama-server.log", file=sys.stderr)
        raise SystemExit(1)
