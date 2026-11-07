# Gateway and Client Integration

## Local Gateway

Default bind: `127.0.0.1:18081`. Do not bind to all interfaces by default.

Initial OpenAI-compatible routes:
- `GET /health`
- `GET /v1/models`
- `POST /v1/chat/completions`

## Streaming contract

SSE parser/writer must preserve blank lines as event delimiters, support multiple data lines, flush incrementally, propagate upstream errors, honor cancellation, enforce backpressure and timeouts, and send a correct terminal event. Do not strip blank lines from upstream streams.

## Request lifecycle

Each request carries a request ID, deadline, cancellation context and runtime generation. If the upstream generation changes, reject or terminate the stale stream; do not silently replay uncertain inference.

## Security

Loopback-only by default, local authentication for control operations, strict request size/concurrency limits, explicit CORS policy, no secrets in logs, and no unrestricted proxy behavior.

## OpenCode

Discover CLI/GUI and its actual configuration path. Back up before editing. Apply atomic, minimal updates and preserve unrelated settings. Verify endpoint, model discovery, non-streaming, streaming, cancellation, context behavior and reconnect. Roll back on failure. OpenCode is the only initial client; other clients are deferred.
