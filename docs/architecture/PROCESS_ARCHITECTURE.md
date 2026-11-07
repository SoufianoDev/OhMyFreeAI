# Process Architecture

## Ownership

Every child process launched by OhMyFreeAI is registered with an owner ID, executable path, arguments fingerprint, working directory, start time, process identity and lifecycle state. Secrets are never copied into diagnostics.

## Start

Validate executable provenance and configuration, reserve the intended endpoint, create an owned process record, launch with explicit environment, capture bounded stdout/stderr and wait for a real health signal.

## Stop

Request graceful shutdown first, wait with a deadline, then use platform-specific termination only for a process whose identity still matches the owned record. Never terminate by PID alone.

## Reconciliation

On startup, reconcile stored process records against actual process identity. A reused PID or mismatched executable is stale metadata, not an owned process.

## Cancellation and shutdown

Every process wait, stream and monitor has a context/cancellation path. Shutdown stops only owned processes and never touches unrelated user services.
