# Runtime and Engine Architecture

Runtime and Engine are independent modules.

## Runtime responsibilities

Discover provider/session, obtain capability and resource snapshots, enforce policy status, connect/reconnect, report session lifecycle and expose a typed execution channel only when permitted.

## Engine responsibilities

Resolve a pinned distribution, verify checksum/signature where available, install or reuse, configure backend, launch and supervise llama.cpp, load exact artifacts, expose health and stop only owned processes.

## llama.cpp distribution

- Primary engine remains upstream llama.cpp.
- Prefer official stable prebuilt release artifacts matching OS/architecture/backend.
- Pin exact version, asset URL, checksum, target and build options in a manifest.
- Do not rebuild each runtime session.
- Build from source only when no compatible prebuilt exists or a required supported feature/patch makes it necessary; explain the reason and preserve reproducibility.
- Do not use nightly/prerelease builds as production defaults without explicit approval.
- CUDA variant must match actual Colab GPU, driver and runtime. A CUDA-enabled binary is not proof of successful execution.

## Persistent and ephemeral storage

Persistent Drive storage may hold verified model artifacts, engine distributions, manifests, checksums and download metadata when mounting/use is allowed. Never execute an engine or load model weights directly from mounted Drive. Stage to local runtime disk, verify, then execute.

## Runtime policy

Provider policy is checked before session activation. A runtime that is technically reachable but policy-blocked is not available to application use cases. No tunnel, proxy, SSH, remote desktop, account rotation or notebook UI bypass is permitted.
