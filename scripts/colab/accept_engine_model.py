#!/usr/bin/env python3
"""Tesla T4 CUDA acceptance test for the pinned SoufianoDev llama.cpp bundle.

All downloads and extraction stay under /content/ohmyfreeai-acceptance.
This test does not touch Google Drive, existing llama-server processes, or the
qwen-opencode session. Only the server process started by this script is stopped.
"""
from __future__ import annotations

import ctypes.util
import hashlib
import json
import os
import platform
import re
import shutil
import socket
import subprocess
import sys
import sysconfig
import time
import urllib.request
from pathlib import Path

ENGINE_COMMIT = "6e6075d3a449f1fff013137e82c02488d1ceedaa"
ENGINE_URL = (
    f"https://raw.githubusercontent.com/SoufianoDev/OhMyFreeAI/"
    f"{ENGINE_COMMIT}/engine/llama-engine.tar.zst"
)
ENGINE_SHA256 = "d97736fa7d6f192972e6bce4d77302b8c6206dd6e1d08cd0ad8f006f9f0e864f"

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
STARTUP_TIMEOUT = 1200
BASE_URL_HOST = "127.0.0.1"


def say(message: str) -> None:
    print(f"\n[OhMyFreeAI T4 CUDA] {message}", flush=True)


def sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for block in iter(lambda: handle.read(8 * 1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def run_checked(command: list[str], *, timeout: int = 300, env=None) -> str:
    result = subprocess.run(
        command, check=True, timeout=timeout, env=env, text=True,
        stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
    )
    return result.stdout.strip()


def download_verified(url: str, destination: Path, expected_sha256: str) -> None:
    destination.parent.mkdir(parents=True, exist_ok=True)
    if destination.is_file() and sha256(destination) == expected_sha256:
        say(f"Already downloaded and checksum verified: {destination.name}")
        return
    if destination.exists():
        destination.unlink()

    partial = destination.with_name(destination.name + ".part")
    curl = shutil.which("curl")
    if not curl:
        raise RuntimeError("curl is required in the Colab runtime.")

    def fetch(resume: bool) -> subprocess.CompletedProcess:
        command = [
            curl, "--location", "--fail", "--show-error",
            "--retry", "5", "--retry-all-errors", "--connect-timeout", "30",
            "--output", str(partial),
        ]
        if resume:
            command += ["--continue-at", "-"]
        command.append(url)
        return subprocess.run(command, text=True, stdout=subprocess.PIPE,
                              stderr=subprocess.STDOUT, timeout=7200)

    say(f"Downloading {destination.name}.")
    result = fetch(partial.exists() and partial.stat().st_size > 0)
    # A stale/complete partial can make curl resume fail with HTTP 416.
    # Discard only that partial and retry once from byte zero.
    if result.returncode != 0 and partial.exists():
        say("Resume failed; discarding the partial file and retrying cleanly.")
        partial.unlink()
        result = fetch(False)
    if result.returncode != 0:
        raise RuntimeError(f"Download failed for {destination.name}:\n{result.stdout[-4000:]}")

    actual = sha256(partial)
    if actual != expected_sha256:
        raise RuntimeError(
            f"SHA-256 mismatch for {destination.name}: expected {expected_sha256}, "
            f"got {actual}. Partial file retained for diagnosis."
        )
    partial.replace(destination)
    say(f"SHA-256 verified: {actual}")


def gpu_memory_used_mib() -> int:
    tool = shutil.which("nvidia-smi")
    if not tool:
        raise RuntimeError("nvidia-smi is unavailable; cannot verify actual GPU use.")
    output = run_checked([
        tool, "--query-gpu=memory.used", "--format=csv,noheader,nounits"
    ], timeout=20)
    return int(output.splitlines()[0].strip())


def gpu_report() -> str:
    tool = shutil.which("nvidia-smi")
    if not tool:
        return "nvidia-smi unavailable"
    result = subprocess.run(
        [tool, "--query-gpu=name,memory.total,memory.used,memory.free,driver_version",
         "--format=csv,noheader"],
        text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=20,
    )
    return result.stdout.strip()


def configure_cuda_runtime() -> tuple[dict, object]:
    try:
        import torch
    except Exception as exc:
        raise RuntimeError(f"PyTorch CUDA runtime is required in Colab: {exc}") from exc

    if not torch.cuda.is_available():
        raise RuntimeError("torch.cuda.is_available() is False.")
    name = torch.cuda.get_device_name(0)
    capability = torch.cuda.get_device_capability(0)
    cuda_version = torch.version.cuda or ""
    if "T4" not in name:
        raise RuntimeError(f"Expected Tesla T4, detected {name!r}.")
    if capability != (7, 5):
        raise RuntimeError(f"Expected compute capability 7.5, detected {capability}.")
    if not cuda_version.startswith("13.0"):
        raise RuntimeError(
            f"Expected PyTorch CUDA 13.0 runtime to match this bundle; got {cuda_version!r}."
        )

    paths: list[str] = []
    torch_lib = Path(torch.__file__).resolve().parent / "lib"
    if torch_lib.is_dir():
        paths.append(str(torch_lib))
    purelib = Path(sysconfig.get_paths()["purelib"])
    for libdir in sorted(purelib.glob("nvidia/*/lib")):
        if libdir.is_dir():
            paths.append(str(libdir))
    # Some Colab images place packages in a user site directory.
    for site in sorted(purelib.parent.glob("*/site-packages")):
        for libdir in sorted(site.glob("nvidia/*/lib")):
            if libdir.is_dir():
                paths.append(str(libdir))

    paths = list(dict.fromkeys(paths))
    if not any(Path(p).glob("libcudart.so*") for p in paths):
        raise RuntimeError("Could not locate libcudart in PyTorch/NVIDIA runtime libraries.")
    env = os.environ.copy()
    env["LD_LIBRARY_PATH"] = os.pathsep.join(
        paths + [env.get("LD_LIBRARY_PATH", "")]
    ).rstrip(os.pathsep)
    facts = {
        "gpu_name": name,
        "compute_capability": capability,
        "torch_version": torch.__version__,
        "cuda_version": cuda_version,
        "runtime_library_paths": paths,
    }
    return env, facts


def audit_shared_libraries(targets: list[Path], env: dict) -> None:
    for target in targets:
        output = run_checked(["ldd", str(target)], timeout=30, env=env)
        missing = [
            line.strip() for line in output.splitlines()
            if "not found" in line and "libcuda.so.1" not in line
        ]
        if missing:
            raise RuntimeError(
                f"Unresolved runtime dependencies for {target.name}:\n" +
                "\n".join(missing) + "\nFull ldd output:\n" + output
            )
        # libcuda is intentionally supplied by the Colab host driver, not PyTorch.
        if "libcuda.so.1 => not found" in output and not ctypes.util.find_library("cuda"):
            raise RuntimeError(
                f"{target.name} needs libcuda.so.1 but the host driver library is not discoverable."
            )
        say(f"ldd dependency audit passed: {target.name}")


def ensure_zstd() -> str:
    zstd = shutil.which("zstd")
    if zstd:
        return zstd
    say("Installing the small Ubuntu zstd package.")
    apt = shutil.which("apt-get")
    if not apt:
        raise RuntimeError("zstd is missing and apt-get is unavailable.")
    run_checked([apt, "update", "-qq"], timeout=600)
    run_checked([apt, "install", "-y", "-qq", "zstd"], timeout=600)
    zstd = shutil.which("zstd")
    if not zstd:
        raise RuntimeError("zstd installation did not provide the zstd executable.")
    return zstd


def extract_engine(zstd: str) -> Path:
    destination = ROOT / "engine"
    destination.mkdir(parents=True, exist_ok=True)
    say("Checking archive integrity and extracting the pinned CUDA engine.")
    run_checked([zstd, "-t", str(ENGINE_ARCHIVE)], timeout=180)
    run_checked([
        "tar", "--use-compress-program=" + zstd,
        "-xf", str(ENGINE_ARCHIVE), "-C", str(destination),
    ], timeout=300)
    bindir = destination / "llama-engine" / "bin"
    server = bindir / "llama-server"
    cuda_plugin = bindir / "libggml-cuda.so"
    if not server.is_file() or not os.access(server, os.X_OK):
        raise RuntimeError(f"llama-server is missing or not executable: {server}")
    if not cuda_plugin.is_file():
        raise RuntimeError(f"CUDA backend plugin is missing: {cuda_plugin}")
    return server


def http_json(url: str, payload: dict | None = None, timeout: int = 60) -> dict:
    data = None if payload is None else json.dumps(payload).encode("utf-8")
    headers = {"Content-Type": "application/json"} if data is not None else {}
    request = urllib.request.Request(url, data=data, headers=headers)
    with urllib.request.urlopen(request, timeout=timeout) as response:
        return json.loads(response.read().decode("utf-8"))


def free_port() -> int:
    with socket.socket() as sock:
        sock.bind((BASE_URL_HOST, 0))
        return int(sock.getsockname()[1])


def main() -> int:
    if platform.system() != "Linux" or platform.machine().lower() not in {"x86_64", "amd64"}:
        raise RuntimeError(f"Artifact targets Linux x86_64, got {platform.system()} {platform.machine()}.")
    if not shutil.which("tar"):
        raise RuntimeError("tar is missing from this Colab runtime.")

    ROOT.mkdir(parents=True, exist_ok=True)
    say("Validating the actual Colab GPU and CUDA runtime.")
    env, gpu = configure_cuda_runtime()
    for key, value in gpu.items():
        print(f"{key.upper()}:", value)
    print("GPU BEFORE:", gpu_report(), flush=True)
    baseline_vram = gpu_memory_used_mib()

    free_disk = shutil.disk_usage(ROOT).free / (1024 ** 3)
    if free_disk < 9.0:
        raise RuntimeError(f"Only {free_disk:.2f} GiB disk space is free; at least 9 GiB is required.")

    download_verified(ENGINE_URL, ENGINE_ARCHIVE, ENGINE_SHA256)
    model_url = (
        f"https://huggingface.co/{HF_REPO}/resolve/{HF_REVISION}/"
        f"{MODEL_FILE}?download=true"
    )
    download_verified(model_url, MODEL_PATH, MODEL_SHA256)
    if MODEL_PATH.stat().st_size != MODEL_BYTES:
        raise RuntimeError(
            f"Unexpected GGUF size {MODEL_PATH.stat().st_size}; expected {MODEL_BYTES}."
        )

    zstd = ensure_zstd()
    server = extract_engine(zstd)
    bindir = server.parent
    env["LD_LIBRARY_PATH"] = str(bindir) + os.pathsep + env.get("LD_LIBRARY_PATH", "")

    version = run_checked([str(server), "--version"], timeout=30, env=env)
    print("ENGINE_VERSION:", version)
    if "version: 0.9.0" not in version or "-dev" in version:
        raise RuntimeError(f"Unexpected llama-server version: {version}")

    cuda_plugin = bindir / "libggml-cuda.so"
    audit_shared_libraries([server, cuda_plugin], env)

    devices = run_checked([str(server), "--list-devices"], timeout=60, env=env)
    print("\nLLAMA_DEVICES:\n" + devices)
    if not re.search(r"CUDA\d*:", devices, flags=re.IGNORECASE):
        raise RuntimeError("llama-server did not enumerate a CUDA device.")

    port = free_port()
    threads = max(1, min(os.cpu_count() or 2, 4))
    command = [
        str(server), "--model", str(MODEL_PATH),
        "--host", BASE_URL_HOST, "--port", str(port),
        "--ctx-size", str(CTX_SIZE), "--threads", str(threads),
        "--n-gpu-layers", "99", "--parallel", "1",
    ]
    say("Starting llama-server with full GPU layer offload on Tesla T4.")
    print("COMMAND:", " ".join(command), flush=True)
    print("SERVER_LOG:", LOG_PATH, flush=True)

    with LOG_PATH.open("w", encoding="utf-8") as log:
        process = subprocess.Popen(
            command, cwd=bindir, env=env, stdin=subprocess.DEVNULL,
            stdout=log, stderr=subprocess.STDOUT, start_new_session=True,
        )
        try:
            base = f"http://{BASE_URL_HOST}:{port}"
            deadline = time.monotonic() + STARTUP_TIMEOUT
            ready = False
            last_error = None
            while time.monotonic() < deadline:
                if process.poll() is not None:
                    log.flush()
                    tail = "\n".join(LOG_PATH.read_text(errors="replace").splitlines()[-100:])
                    raise RuntimeError(
                        f"llama-server exited during model loading (code {process.returncode}).\n{tail}"
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
                tail = "\n".join(LOG_PATH.read_text(errors="replace").splitlines()[-100:])
                raise RuntimeError(
                    f"Model not ready after {STARTUP_TIMEOUT}s; last health error: {last_error}\n{tail}"
                )

            log.flush()
            server_log = LOG_PATH.read_text(errors="replace")
            offload = re.search(
                r"offloaded\s+\d+[/ ]\d+\s+layers?\s+to\s+GPU",
                server_log, flags=re.IGNORECASE,
            )
            if not offload:
                raise RuntimeError(
                    "Server became healthy but its log does not confirm GPU layer offload.\n" +
                    "\n".join(server_log.splitlines()[-100:])
                )

            used_vram = gpu_memory_used_mib()
            increase = used_vram - baseline_vram
            print("GPU AFTER MODEL LOAD:", gpu_report())
            print("VRAM_BASELINE_MIB:", baseline_vram)
            print("VRAM_USED_MIB:", used_vram)
            print("VRAM_INCREASE_MIB:", increase)
            if increase < 1024:
                raise RuntimeError(
                    f"GPU memory increased by only {increase} MiB; refusing CPU-fallback false positive."
                )

            say("Model is loaded on the T4. Sending a real chat completion.")
            result = http_json(
                base + "/v1/chat/completions",
                {
                    "model": "qwen35-9b-q4km",
                    "messages": [
                        {"role": "system", "content": "You are a concise assistant."},
                        {"role": "user", "content": "Reply with exactly this token: SOUFIANODEV_T4_CUDA_OK"},
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
            if "SOUFIANODEV_T4_CUDA_OK" not in answer:
                raise RuntimeError(
                    "Completion did not contain the required marker: " +
                    json.dumps(result, ensure_ascii=False)[:3000]
                )

            timings = result.get("timings") or {}
            print("\nMODEL_RESPONSE:", answer.strip())
            print("PROMPT_TOKENS:", result.get("usage", {}).get("prompt_tokens"))
            print("COMPLETION_TOKENS:", result.get("usage", {}).get("completion_tokens"))
            print("TOKENS_PER_SECOND:", timings.get("predicted_per_second"))
            print("GPU_NAME:", gpu["gpu_name"])
            print("CUDA_VERSION:", gpu["cuda_version"])
            print("VRAM_USED_MIB:", used_vram)
            print("\nACCEPTANCE_RESULT=PASS")
            print("Engine commit:", ENGINE_COMMIT)
            print("Engine SHA256:", ENGINE_SHA256)
            print("Model revision:", HF_REVISION)
            print("Inference: real completion with confirmed CUDA layer offload")
            return 0
        finally:
            if process.poll() is None:
                process.terminate()
                try:
                    process.wait(timeout=20)
                except subprocess.TimeoutExpired:
                    process.kill()
                    process.wait(timeout=10)
            log.flush()
            print("Acceptance server stopped. Files remain in:", ROOT)


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as exc:
        print("\nACCEPTANCE_RESULT=FAIL", file=sys.stderr)
        print(type(exc).__name__ + ": " + str(exc), file=sys.stderr)
        print("If the server started, inspect /content/ohmyfreeai-acceptance/llama-server.log", file=sys.stderr)
        raise SystemExit(1)
