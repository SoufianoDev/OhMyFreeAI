# State Machines

## Runtime session

`unknown -> discovering -> available -> connecting -> connected -> starting -> ready`

Failure paths: `degraded`, `recovering`, `paused`, `terminated`, `unavailable`, `error`.

Only verified transitions are allowed. A reconnect first targets the existing session. A replacement session is created only after termination is confirmed. Each replacement gets a new session ID/generation.

## Engine

`not_installed -> installing -> installed -> starting -> loading_model -> ready -> stopping -> stopped`

Failure states include `incompatible`, `crashed`, `unhealthy`, `error`. A model-loaded marker alone does not equal ready; health and inference verification are required.

## Model assessment

`unknown -> metadata_pending -> artifact_pending -> format_check -> backend_check -> resource_check -> eligible | unsupported | unverified`.

A successful real run records a separate `verified` evidence state with exact environment/build metadata and timestamp. Evidence expires or becomes stale when the model revision, engine build, backend or relevant runtime profile changes.

## Download

`queued -> preparing -> downloading -> verifying -> completed`; alternate states: `paused`, `cancelling`, `cancelled`, `failed`, `integrity_error`.

Never report complete before checksum/size verification.

## Gateway

`stopped -> starting -> listening -> degraded -> stopping -> stopped`, with `error` for startup/runtime faults. A stale upstream generation must not receive new requests.

## Recovery policy

Classify before retrying. Auth, policy, permission, unsupported hardware and integrity errors are terminal until user action. Transient network/process failures use bounded exponential backoff with jitter. Never retry indefinitely or replay an inference request whose completion is uncertain.
