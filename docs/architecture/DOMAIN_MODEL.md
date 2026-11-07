# Domain Model

## Core entities

- **Model**: stable logical identity and family metadata.
- **ModelRevision**: exact upstream repository, checkpoint/revision, tokenizer/config revision and license metadata.
- **Artifact**: file identity, format, quantization, size, checksum, source URI and local/persistent locations.
- **CompatibilityAssessment**: exact model revision + artifact + engine build + backend + runtime profile + result.
- **RuntimeProfile**: actual accelerator, VRAM, system RAM, disk, driver/runtime versions and policy/capability facts.
- **RuntimeSession**: unique session ID, generation, provider, lifecycle state, timestamps and capabilities.
- **EngineInstallation**: pinned version, build variant, target OS/architecture, backend and checksum.
- **EngineInstance**: owned process identity, model load, health and active request count.
- **GatewayInstance**: bind address, port, auth mode, upstream identity and state.
- **ClientInstallation**: discovered executable/config path and supported integration capabilities.
- **DownloadJob**: source, destination, expected size/hash, progress, retry state and cancellation.
- **BackupManifest**: format version, app version, creation time, included records and checksums.

## Identity rules

Model family is descriptive, not a compatibility key. Compatibility is keyed by exact checkpoint/revision, artifact conversion, quantization, tokenizer and required features, engine build/backend, runtime capabilities and test evidence.

Runtime sessions use immutable IDs and monotonically increasing generations. A response from an older generation is stale and cannot mutate the current session state.

## Readiness

Compatibility assessment and runtime readiness are separate. A model may be theoretically compatible but not fit the current Colab GPU. A model may fit memory but lack a supported feature path. A model may load but fail real inference. These states remain distinct in domain and UI.
