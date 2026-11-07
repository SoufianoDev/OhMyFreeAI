# Security and Recovery

## Process ownership

A PID alone is not process identity. Record executable identity, start time, command fingerprint and ownership token. Linux adapters should use /proc and pidfd where available; Windows adapters should use process handles and Job Objects. Never kill a process solely because its PID matches a stale file.

## Network boundary

Gateway binds to loopback by default. Control IPC is separate from inference endpoints and requires local authentication. Do not expose arbitrary remote proxy endpoints.

## Secrets and privacy

Use OS-backed secret storage where available. Store references, not secret values, in SQLite or backups. Redact headers, tokens, prompts and user paths from diagnostics unless explicitly opted in.

## Recovery

1. Classify the error.
2. Reconnect to the existing runtime/session first.
3. Create a replacement only after termination is confirmed.
4. Use bounded exponential backoff with jitter for transient failures.
5. Stop retries on auth, policy, permission, unsupported hardware or integrity failures.
6. Use session ID + generation to reject stale events and responses.
7. Never replay an inference request with uncertain completion.
8. Report the exact recovery action and final verified state.

## Shutdown

Cancel owned work, stop only owned processes, flush durable state, close DB handles and release OS resources. No orphaned goroutines/processes.
