#!/usr/bin/env python3
"""Exercise the built desktop executable's CLI without opening its GUI."""

from __future__ import annotations

import json
import os
import sqlite3
import zipfile
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


def create_migrated_database(database_path: Path) -> None:
    migrations = Path(__file__).resolve().parent.parent / "migrations"
    database_path.parent.mkdir(parents=True, exist_ok=True)
    connection = sqlite3.connect(database_path)
    try:
        connection.execute("PRAGMA journal_mode = WAL")
        connection.execute("PRAGMA foreign_keys = ON")
        for migration in sorted(migrations.glob("[0-9][0-9][0-9]_*.sql")):
            if migration.name.endswith(".down.sql"):
                continue
            prefix, description = migration.name.split("_", 1)
            connection.executescript(migration.read_text(encoding="utf-8"))
            connection.execute(
                "INSERT INTO schema_migrations(version, description, applied_at) VALUES (?, ?, ?)",
                (int(prefix), description.removesuffix(".sql"), "2026-11-07T00:00:00Z"),
            )
        connection.commit()
    finally:
        connection.close()

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
                    f"{arguments!r}: exit={result.returncode}, expected=1 and {expected_error!r}\n{combined}"
                )
            if "Usage:" in combined:
                raise SystemExit(f"{arguments!r}: backup command unexpectedly fell back to help/GUI\n{combined}")
            print(f"CLI_SMOKE_PASS command={' '.join(arguments[:2])} exit=1")

        database_path = isolated_config / "OhMyFreeAI" / "state.db"
        create_migrated_database(database_path)
        archive_path = Path(temporary) / "native CLI backup.omfaibackup"
        export_result = run(binary, ["backup", "export", str(archive_path)], environment)
        if export_result.returncode != 0 or not archive_path.is_file():
            raise SystemExit(f"backup export failed: {export_result.stdout}{export_result.stderr}")
        print("CLI_SMOKE_PASS command=backup export exit=0")

        with zipfile.ZipFile(archive_path) as archive:
            names = set(archive.namelist())
            if names != {"manifest.json", "state.sqlite"}:
                raise SystemExit(f"backup archive contains unexpected entries: {sorted(names)}")
            manifest = json.loads(archive.read("manifest.json"))
            if manifest.get("format") != "OhMyFreeAI" or manifest.get("format_version") != 1:
                raise SystemExit(f"backup manifest is invalid: {manifest!r}")

        inspect_result = run(binary, ["backup", "inspect", str(archive_path)], environment)
        if inspect_result.returncode != 0 or "Backup valid" not in inspect_result.stdout:
            raise SystemExit(f"backup inspect failed: {inspect_result.stdout}{inspect_result.stderr}")
        print("CLI_SMOKE_PASS command=backup inspect exit=0")

        original_archive = archive_path.read_bytes()
        overwrite_result = run(binary, ["backup", "export", str(archive_path)], environment)
        if overwrite_result.returncode != 1 or "already exists" not in overwrite_result.stderr:
            raise SystemExit(f"backup export must reject overwrite: {overwrite_result.stdout}{overwrite_result.stderr}")
        if archive_path.read_bytes() != original_archive:
            raise SystemExit("backup export changed an existing archive")
        print("CLI_SMOKE_PASS backup_no_overwrite=true")

        if not database_path.is_file():
            raise SystemExit("backup CLI unexpectedly removed the live database")
        print("CLI_SMOKE_PASS existing_database_preserved=true")


    print("CLI_SMOKE_PASS no_gui_fallback=true doctor_read_only=true")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
