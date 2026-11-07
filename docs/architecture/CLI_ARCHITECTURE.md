# CLI Architecture

The GUI remains the default experience. CLI dispatch happens before any GUI, database or runtime initialization; every implemented command must return without creating a window.

## Implemented command contract

| Invocation | Behavior |
|---|---|
| no arguments | Launch the desktop GUI |
| `--help`, `-h` | Print usage and exit successfully |
| `--version`, `-v` | Print the application version and exit successfully |
| `doctor` | Print local platform, Go runtime, config path and database-file presence |
| unknown command | Print an error and help hint; exit with status 2; never launch GUI |
| extra arguments | Reject with status 2; never launch GUI |

The doctor command is intentionally read-only. A database file being present does not prove database integrity, runtime connectivity, GPU availability or model readiness. It must not create directories, initialize the database, install dependencies or change permissions.

## Planned commands

The following remain unimplemented and must not be advertised as working: `status`, `models`, `runtime`, `engine`, `gateway`, `client opencode`, `backup create`, `backup inspect`, `backup restore`, and `logs`.

## Future CLI requirements

- Commands call the same application services as the GUI; no duplicated domain logic.
- Support human-readable and machine-readable JSON output.
- Stable documented exit codes and typed errors.
- Backup inspect must be read-only and validate the full archive before reporting metadata.
- Restore must coordinate the application/database lifecycle and preserve rollback capability.
- CLI must not duplicate domain logic or silently elevate privileges.
