# CLI Architecture

The GUI is the default user experience. CLI commands are dispatched separately but call the same application services.

Planned commands: `doctor`, `status`, `models`, `runtime`, `engine`, `gateway`, `client opencode`, `backup create`, `backup inspect`, `backup restore`, `logs`.

CLI output supports human-readable and machine-readable JSON modes. Exit codes are stable and documented. CLI must not duplicate domain logic or silently elevate privileges.
