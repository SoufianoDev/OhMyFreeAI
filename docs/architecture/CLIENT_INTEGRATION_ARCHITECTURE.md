# Client Integration Architecture

OpenCode is the first supported client. Discover the executable and configuration path; inspect current configuration before editing; create a backup; apply the smallest atomic change; preserve unrelated settings; validate; and roll back if verification fails.

Verification covers model listing, non-streaming completion, streaming event framing, cancellation, context metadata and reconnect. Other clients are deferred until this integration is stable and a real user need is demonstrated.
