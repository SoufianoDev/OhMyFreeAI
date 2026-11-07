# CLI Architecture

The GUI remains the default experience. CLI dispatch happens before any GUI, database or runtime initialization; every implemented command must return without creating a window.

## Implemented command contract

| Invocation | Behavior |
|---|---|
| no arguments | Launch the desktop GUI |
| `--help`, `-h` | Print usage and exit successfully |
| `--version`, `-v` | Print the application version and exit successfully |
| `doctor` | Print local platform, Go runtime, config path and database-file presence |
| `backup export <file.omfaibackup>` | Export the existing local SQLite database to a verified archive; refuses overwrite and does not create a first-run database |
| `backup inspect <file.omfaibackup>` | Fully validate an archive in temporary storage and print its manifest metadata without restoring it |
| unknown command | Print an error and help hint; exit with status 2; never launch GUI |
| extra arguments | Reject with status 2; never launch GUI |

The doctor command is intentionally read-only. A database file being present does not prove database integrity, runtime connectivity, GPU availability or model readiness. It must not create directories, initialize the database, install dependencies or change permissions.

## Planned commands

The following remain unimplemented and must not be advertised as working: `status`, `models`, `runtime`, `engine`, `gateway`, `client opencode`, `backup restore`, and `logs`.

## Future CLI requirements

- Commands call the same application services as the GUI; no duplicated domain logic.
- Support human-readable and machine-readable JSON output.
- Stable documented exit codes and typed errors. Current contract: 0 success, 1 command execution failure, 2 syntax/unknown-command failure.
- Backup inspect is read-only and validates the full archive before reporting metadata.
- Backup export opens only an existing database; it does not run migrations or create a first-run database.
- Restore must coordinate the application/database lifecycle and preserve rollback capability.
- CLI must not duplicate domain logic or silently elevate privileges.
