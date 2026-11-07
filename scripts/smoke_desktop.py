#!/usr/bin/env python3
"""Exercise the real Wails frontend-to-Go bridge twice under a virtual X server."""

from __future__ import annotations

import os
import shutil
import subprocess
import sys
import tempfile
import time
from pathlib import Path


def smoke_once(binary: Path, marker: Path, log_path: Path, attempt: int) -> None:
    marker.unlink(missing_ok=True)
    env = os.environ.copy()
    env.update(
        {
            "OMFAI_WEBVIEW_SMOKE_MARKER": str(marker),
            "GDK_BACKEND": "x11",
            "GSK_RENDERER": "cairo",
            "LIBGL_ALWAYS_SOFTWARE": "1",
            "WEBKIT_DISABLE_COMPOSITING_MODE": "1",
        }
    )

    xvfb = shutil.which("Xvfb")
    if not xvfb:
        raise RuntimeError("Xvfb is required for the Linux WebView smoke test.")
    display = ":97"
    with log_path.open("wb") as output:
        display_process = subprocess.Popen(
            [xvfb, display, "-screen", "0", "1280x820x24", "-nolisten", "tcp", "-ac"],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )
        process = None
        try:
            time.sleep(0.5)
            if display_process.poll() is not None:
                raise RuntimeError("Xvfb exited before the desktop application could start.")
            env["DISPLAY"] = display
            process = subprocess.Popen(
                [str(binary)],
                cwd=binary.parent,
                env=env,
                stdout=output,
                stderr=subprocess.STDOUT,
            )
            deadline = time.monotonic() + 30
            while time.monotonic() < deadline:
                if marker.is_file():
                    value = marker.read_text(encoding="utf-8")
                    if value != "GetSnapshot invoked\n":
                        raise RuntimeError(f"Unexpected WebView marker: {value!r}")
                    print(f"WEBVIEW_GO_BRIDGE_PASS attempt={attempt}", flush=True)
                    return
                if process.poll() is not None:
                    raise RuntimeError(
                        f"Desktop process exited before the frontend called GetSnapshot "
                        f"(exit={process.returncode})."
                    )
                time.sleep(0.2)
            raise RuntimeError("Timed out waiting for the frontend-to-Go GetSnapshot call.")
        finally:
            if process is not None and process.poll() is None:
                process.terminate()
                try:
                    process.wait(timeout=8)
                except subprocess.TimeoutExpired:
                    process.kill()
                    process.wait(timeout=5)
            if display_process.poll() is None:
                display_process.terminate()
                try:
                    display_process.wait(timeout=5)
                except subprocess.TimeoutExpired:
                    display_process.kill()
                    display_process.wait(timeout=5)

    if not marker.is_file():
        details = log_path.read_text(encoding="utf-8", errors="replace")[-5000:]
        raise RuntimeError("WebView smoke marker missing. Last application output:\n" + details)


def main() -> int:
    if len(sys.argv) != 2:
        print("Usage: smoke_desktop.py PATH_TO_EXECUTABLE", file=sys.stderr)
        return 2

    binary = Path(sys.argv[1]).resolve()
    if not binary.is_file():
        print(f"Desktop executable not found: {binary}", file=sys.stderr)
        return 2
    if os.name != "posix":
        print("This smoke test requires Linux and Xvfb.", file=sys.stderr)
        return 2

    try:
        with tempfile.TemporaryDirectory(prefix="omfai-webview-smoke-") as temporary:
            root = Path(temporary)
            marker = root / "bridge-called"
            for attempt in (1, 2):
                smoke_once(binary, marker, root / f"attempt-{attempt}.log", attempt)
        print("WEBVIEW_REOPEN_PASS")
        return 0
    except (OSError, RuntimeError, subprocess.SubprocessError) as error:
        print(f"WEBVIEW_SMOKE_FAILED: {error}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
