# Download Architecture

## Job lifecycle

queued -> preparing -> downloading -> verifying -> completed; with paused, cancelling, cancelled, failed and integrity_error terminal/alternate states.

## Requirements

- Per-job identity, source, target, expected size/hash, progress, speed, resume support and retry policy.
- Resume only when the source supports safe range/resume semantics.
- Cancellation must stop owned transfer workers and close files.
- Write to a temporary path and atomically finalize only after size/hash verification.
- Never overwrite a verified existing artifact without an explicit replacement operation.
- aria2 is optional and must be detected, not silently installed. Go-native transfer is the fallback.
- Large downloads show clear progress and preserve verified partial metadata where safe.
