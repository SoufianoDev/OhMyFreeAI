#!/usr/bin/env python3
"""Exercise the built desktop executable's CLI without opening its GUI."""

from __future__ import annotations

import os
import subprocess
import sys
import tempfile
from pathlib import Path


def run(binary: Path, arguments: list[str], environment: dict[str, str]) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        [str(binary), *arguments],
        capture_output=True,
        text=True,
        env=environment,
        timeout=10,
        check=False,
    )


def main() -> int:
    if len(sys.argv) != 2:
        print("usage: smoke_cli.py PATH_TO_EXECUTABLE", file=sys.stderr)
        return 2

    binary = Path(sys.argv[1]).resolve()
    if not binary.is_file():
        print(f"application executable not found: {binary}", file=sys.stderr)
        return 2

    with tempfile.TemporaryDirectory(prefix="ohmyfreeai-cli-smoke-") as temporary:
        isolated_config = Path(temporary) / "config"
        environment = os.environ.copy()
        if os.name == "nt":
            environment["APPDATA"] = str(isolated_config)
        else:
            environment["XDG_CONFIG_HOME"] = str(isolated_config)

        cases = (
            (["--help"], 0, "Usage:"),
            (["--version"], 0, "OhMyFreeAI "),
            (["doctor"], 0, "Database file:"),
            (["definitely-not-a-command"], 2, "unknown command"),
        )
        for arguments, expected_code, expected_output in cases:
            result = run(binary, arguments, environment)
            combined = result.stdout + result.stderr
            if result.returncode != expected_code:
                raise SystemExit(
                    f"{arguments!r}: exit={result.returncode}, expected={expected_code}\n{combined}"
                )
            if expected_output not in combined:
                raise SystemExit(f"{arguments!r}: missing {expected_output!r}\n{combined}")
            print(f"CLI_SMOKE_PASS command={arguments[0]} exit={result.returncode}")

        missing_archive = Path(temporary) / "missing backup.omfaibackup"
        backup_cases = (
            (["backup", "export", str(missing_archive)], "application database does not exist"),
            (["backup", "inspect", str(missing_archive)], "invalid OhMyFreeAI backup archive"),
        )
        for arguments, expected_error in backup_cases:
            result = run(binary, arguments, environment)
            combined = result.stdout + result.stderr
            if result.returncode != 1 or expected_error not in combined:
                raise SystemExit(
                    f"{arguments!r}: exit={result.returncode}, expected=1 and {expected_error!r}\\n{combined}"
                )
            if "Usage:" in combined:
                raise SystemExit(f"{arguments!r}: backup command unexpectedly fell back to help/GUI\\n{combined}")
            print(f"CLI_SMOKE_PASS command={' '.join(arguments[:2])} exit=1")

        if (isolated_config / "OhMyFreeAI").exists():
            raise SystemExit("read-only CLI commands unexpectedly created application state")

    print("CLI_SMOKE_PASS no_gui_fallback=true doctor_read_only=true")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
