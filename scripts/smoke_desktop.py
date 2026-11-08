#!/usr/bin/env python3
"""Exercise the real Wails frontend-to-Go bridge twice under a virtual X server."""

from __future__ import annotations

import os
import re
import shutil
import signal
import select
import subprocess
import sys
import tempfile
import time
from pathlib import Path


def parse_snapshot_marker(value: str) -> int:
    match = re.fullmatch(r"GetSnapshot success count=([1-9][0-9]*)\n?", value)
    if match is None:
        raise ValueError(f"invalid snapshot marker: {value!r}")
    return int(match.group(1))


def diagnostic_excerpt(path: Path) -> str:
    text = path.read_text(encoding="utf-8", errors="replace")
    if len(text) <= 10000:
        return text
    return text[:5000] + "\\n... middle of native log omitted ...\\n" + text[-5000:]


def process_diagnostics(process: subprocess.Popen[bytes], display_process: subprocess.Popen[bytes], display: str) -> str:
    """Capture a bounded Linux process/X11 snapshot without exposing environment variables."""
    pending = [process.pid, display_process.pid]
    seen: set[int] = set()
    rows: list[str] = []

    while pending and len(seen) < 24:
        current_pid = pending.pop(0)
        if current_pid in seen:
            continue
        seen.add(current_pid)
        proc = Path("/proc") / str(current_pid)
        try:
            status = (proc / "status").read_text(encoding="utf-8", errors="replace")
            fields = {}
            for line in status.splitlines():
                if line.startswith(("Name:", "State:", "PPid:")):
                    key, value = line.split(":", 1)
                    fields[key] = value.strip()
            wchan = (proc / "wchan").read_text(encoding="ascii", errors="replace").strip()
            rows.append(
                f"pid={current_pid} name={fields.get('Name', '?')} "
                f"state={fields.get('State', '?')} ppid={fields.get('PPid', '?')} "
                f"wchan={wchan or '?'}"
            )
            children = (proc / "task" / str(current_pid) / "children").read_text(
                encoding="ascii", errors="replace"
            ).split()
            pending.extend(int(child) for child in children if child.isdigit())
        except (OSError, ValueError):
            rows.append(f"pid={current_pid} unavailable")

    socket_path = Path("/tmp/.X11-unix") / f"X{display.removeprefix(':')}"
    try:
        socket_state = "present" if socket_path.exists() else "missing"
    except OSError as error:
        socket_state = f"unavailable ({error})"

    return (
        f"Process snapshot (launcher_pid={process.pid}, launcher_exit={process.poll()}, "
        f"xvfb_pid={display_process.pid}, xvfb_exit={display_process.poll()}, "
        f"display={display}, x_socket={socket_state}):\\n"
        + "\\n".join(rows)
    )


def smoke_once(binary: Path, marker: Path, log_path: Path, attempt: int) -> None:
    marker.unlink(missing_ok=True)
    env = os.environ.copy()
    # Xvfb is launched with -ac, so inherited host/container X11 cookies are
    # both unnecessary and potentially point to an inaccessible authority file.
    env.pop("XAUTHORITY", None)
    env.pop("XAUTHLOCALHOSTNAME", None)
    env.pop("AUTHFILE", None)
    env.update(
        {
            "OMFAI_WEBVIEW_SMOKE_MARKER": str(marker),
            "GDK_BACKEND": "x11",
            "GSK_RENDERER": "cairo",
            "LIBGL_ALWAYS_SOFTWARE": "1",
            "WEBKIT_DISABLE_COMPOSITING_MODE": "1",
            # GitHub-hosted/container runners cannot create the network namespace
            # required by WebKit's bubblewrap sandbox. This is test-only; never
            # propagate it into the installed application environment.
            "WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS": "1",
            "GTK_A11Y": "none",
        }
    )

    xvfb = shutil.which("Xvfb")
    if not xvfb:
        raise RuntimeError("Xvfb is required for the Linux WebView smoke test.")
    with log_path.open("wb") as output:
        # Ask Xvfb to allocate an unused display number. A fixed :97 can collide
        # with stale sockets or another process in shared container namespaces.
        display_process = subprocess.Popen(
            [xvfb, "-displayfd", "1", "-screen", "0", "1280x820x24", "-nolisten", "tcp", "-ac"],
            stdout=subprocess.PIPE,
            stderr=output,
        )
        process = None
        try:
            ready, _, _ = select.select([display_process.stdout], [], [], 10)
            if not ready:
                raise RuntimeError("Xvfb did not allocate a display within 10 seconds.")
            display_number = display_process.stdout.readline().decode("ascii", errors="replace").strip()
            if not display_number.isdigit():
                raise RuntimeError(f"Xvfb returned an invalid display number: {display_number!r}")
            display = f":{display_number}"
            if display_process.poll() is not None:
                raise RuntimeError("Xvfb exited before the desktop application could start.")
            env["DISPLAY"] = display
            command = [str(binary)]
            dbus_runner = shutil.which("dbus-run-session")
            if dbus_runner:
                # Model a real desktop session explicitly. Without a session bus,
                # GLib may try D-Bus autolaunch in a headless container and block
                # before WebKit creates its web process.
                command = [dbus_runner, "--", *command]
            process = subprocess.Popen(
                command,
                cwd=binary.parent,
                env=env,
                stdout=output,
                stderr=subprocess.STDOUT,
                start_new_session=True,
            )
            print(
                f"WEBVIEW_SMOKE_START attempt={attempt} display={display} "
                f"dbus_session={'enabled' if dbus_runner else 'unavailable'}",
                flush=True,
            )
            deadline = time.monotonic() + 60
            initial_call_logged = False
            while time.monotonic() < deadline:
                if marker.is_file():
                    value = marker.read_text(encoding="utf-8")
                    if value.startswith("GetSnapshot failed:"):
                        raise RuntimeError(value.strip())
                    try:
                        call_count = parse_snapshot_marker(value)
                    except ValueError as error:
                        raise RuntimeError(f"Unexpected WebView marker: {value!r}") from error
                    if call_count == 2:
                        print(f"WEBVIEW_GO_BRIDGE_PASS attempt={attempt} calls={call_count}", flush=True)
                        return
                    if call_count == 1 and not initial_call_logged:
                        print(f"WEBVIEW_GO_BRIDGE_PROGRESS attempt={attempt} initial_call=1", flush=True)
                        initial_call_logged = True
                    if call_count != 1:
                        raise RuntimeError(f"Expected one initial bridge call before UI import, got {call_count}.")
                if process.poll() is not None:
                    output.flush()
                    details = diagnostic_excerpt(log_path)
                    raise RuntimeError(
                        f"Desktop process exited before the frontend called GetSnapshot "
                        f"(exit={process.returncode}). Application output:\\n{details}"
                    )
                time.sleep(0.2)
            output.flush()
            details = log_path.read_text(encoding="utf-8", errors="replace")[-5000:]
            diagnostics = process_diagnostics(process, display_process, display)
            raise RuntimeError(
                "Timed out waiting for the frontend-to-Go GetSnapshot call.\\n"
                + diagnostics
                + "\\nApplication output:\\n"
                + details
            )
        finally:
            if process is not None and process.poll() is None:
                try:
                    os.killpg(process.pid, signal.SIGTERM)
                except ProcessLookupError:
                    pass
                try:
                    process.wait(timeout=8)
                except subprocess.TimeoutExpired:
                    try:
                        os.killpg(process.pid, signal.SIGKILL)
                    except ProcessLookupError:
                        pass
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
