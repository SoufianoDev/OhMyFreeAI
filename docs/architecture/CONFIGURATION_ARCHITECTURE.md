# Configuration Architecture

## Sources and precedence

Defaults -> persisted SQLite settings -> environment overrides -> explicit command-line overrides. The effective configuration is validated before services start.

## Rules

- Keep schema version and migrations explicit.
- Separate secrets from ordinary settings; config stores secret references only.
- Reject invalid combinations with typed field-level errors.
- Show effective non-secret configuration in diagnostics.
- Changes to active runtime settings use validate -> stage -> apply -> health-check -> commit, with rollback on failure.
- Never silently rewrite unrelated client settings or environment files.
