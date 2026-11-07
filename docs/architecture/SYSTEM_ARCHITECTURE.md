# System Architecture

## Architectural style

A modular monolith using Clean Architecture. Domain and application services are independent of Wails, Colab, llama.cpp, SQLite, OpenCode, and OS-specific APIs. Adapters implement ports at the boundaries.

## Layers

1. **Domain** — entities, value objects, state machines, compatibility rules and domain errors. No I/O or framework dependencies.
2. **Application** — use cases, orchestration, policies, transactions and lifecycle coordination. Depends on domain and ports.
3. **Ports** — interfaces for persistence, process control, runtime sessions, model storage, inference, client configuration, telemetry, terminal and platform services.
4. **Adapters** — SQLite, filesystem, llama.cpp process adapter, permitted runtime adapters, Gateway HTTP/SSE, OpenCode config, Linux/Windows process adapters, Wails UI adapter.
5. **Delivery** — desktop shell and CLI. Both call the same application services; neither owns business logic.

## Runtime topology

- Desktop process owns configuration, lifecycle, model metadata, local storage, monitoring, Gateway, client integration and local terminal.
- Model artifacts live in persistent storage and are staged to runtime-local storage before execution.
- The inference engine is a separately supervised process where the platform permits it.
- The local Gateway binds to loopback by default at `127.0.0.1:18081`.
- Remote runtime interaction is not presumed legal or available. Each adapter declares its policy status and capabilities before activation.

## Dependency direction

`delivery -> application -> domain`; adapters implement application ports. Domain never imports Wails, SQLite, HTTP, OS process packages, Colab SDKs or llama.cpp bindings.

## Lifecycle

Startup: load and validate config -> open/migrate DB -> initialize platform adapters -> inspect owned processes -> reconcile runtime state -> start local Gateway if configured -> expose UI.

Shutdown: reject new work -> cancel use cases -> stop only owned processes -> flush persistence/telemetry -> close DB -> release OS resources.

Every long-lived goroutine, process, session and stream has an owner and cancellation path.
