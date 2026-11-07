# Monitoring Architecture

## Independent planes

- **Health**: whether the runtime/engine/Gateway can perform its intended operation.
- **Telemetry**: resource and process measurements.
- **Control**: start/stop/reconnect/configuration actions.

Telemetry must never block control or inference. Health cannot be inferred from a single telemetry sample.

## Statuses

Live, Stale, Unavailable, Error, Starting, Recovering and Paused are explicit states. Every sample carries timestamp, source, runtime generation and freshness. Missing data is not zero.

## Sampling

Use bounded sampling intervals and bounded history. Skip or defer remote probes when a conflicting runtime command is active. Never run concurrent probes that contend with an active inference operation.

## Privacy

Do not collect prompts, message bodies, credentials or arbitrary environment variables. Diagnostics are redacted by default.
