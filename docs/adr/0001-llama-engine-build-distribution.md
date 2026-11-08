# ADR-0001: llama.cpp Engine Build and Distribution

- **Status:** Accepted for implementation on `feature/llama-engine-build-factory`
- **Date:** 2026-11-08
- **Scope:** Linux x86_64 engine distribution for the initial Google Colab Free runtime
- **Related baseline:** OhMyFreeAI Architecture Foundation v3.3, sections 4.3–4.5

## Context

Colab sessions are ephemeral. Rebuilding llama.cpp from source during every session is too slow and wastes the available session time. The runtime must reuse a verified engine bundle from persistent storage and must not silently select an untested nightly build.

The architecture foundation explicitly prefers compatible official prebuilt llama.cpp binaries. A custom source build is a fallback only when no compatible official binary exists, a required feature is missing, or a specific tested patch/customization is necessary.

## Decision

1. **Prefer upstream official prebuilt releases.** Pin the exact release tag and artifact filename. Do not follow `latest` or nightly releases automatically in production.
2. **Use a reproducible build factory only for a demonstrated source-build requirement.** Build from a pinned upstream tag/commit using CMake + Ninja + ccache. Do not migrate llama.cpp to Meson.
3. **Keep the first target narrow:** Linux x86_64 CUDA build for the verified Colab Tesla T4 (Compute Capability 7.5), using CUDA Toolkit 13.0 to match the observed Colab driver/runtime generation. The CUDA build also contains the CPU backend for fallback. Do not generalize this artifact to other GPUs without testing.
4. **Build the T4 CUDA backend plus portable CPU variants with dynamic backend loading.** Explicitly enable `GGML_CUDA=ON`, `CMAKE_CUDA_ARCHITECTURES=75`, `BUILD_SHARED_LIBS=ON`, and `GGML_BACKEND_DL=ON`; package and dependency-check every generated backend shared object and required CUDA runtime libraries.\n5. **Separate build and runtime responsibilities.** GitHub Actions builds and validates packages; Colab detects its environment, selects a compatible pinned package, verifies it, and reuses it from Google Drive.
6. **Record provenance.** Every package must have a manifest recording source tag/commit, build configuration, compiler/toolchain versions, target architecture, CUDA/toolkit details where applicable, required shared libraries, package filename, and SHA-256.
7. **Validate before publication.** Check archive integrity, expected files, executable version output, dynamic dependencies, and backend/device visibility. A GitHub-hosted runner cannot prove that CUDA inference works on Colab; CUDA runtime acceptance must run on the actual Colab T4.
8. **Preserve rollback.** Keep the previous known-good engine bundle and do not replace it until the new package passes acceptance tests.
9. **Do not change the proven Colab launcher/Gateway/OpenCode path in this workstream.** Runtime integration is a separate step after the engine package is verified.

## Branch artifact policy

The requested final built package must be committed to the working branch, not left only as a temporary Actions artifact.

- The branch will contain build definitions, validation scripts, manifest, checksums, and the final compressed engine bundle.
- Before committing the bundle, the workflow must measure its compressed size and reject it if it approaches GitHub's per-file 100 MiB hard limit. The target ceiling is 90 MiB to leave safety margin.
- Do not commit intermediate object files, build directories, compiler caches, or uncompressed staging trees.
- If the final package exceeds the safe branch-file limit, stop publication and report the measured size. Do not silently substitute a Release asset or Git LFS; that requires a separate explicit decision.
- Because binary files remain in Git history after later updates, keep the number of committed engine versions low. Use versioned package names and remove obsolete tracked package files in a later commit only when the replacement is accepted; deletion does not erase historical Git objects.

## Build system

Use the upstream-supported CMake project with Ninja as the generator and ccache for compiler-result reuse. Meson is not selected: configure-time advantages do not establish faster C++/CUDA compilation, and maintaining a second build definition would diverge from upstream without a demonstrated benefit.

## Release and acceptance stages

1. **Source pin:** select a stable upstream tag or exact commit and record why it is selected.
2. **Compatibility discovery:** collect actual Colab GPU, driver, CUDA runtime/toolkit, architecture, and shared-library data.
3. **Build:** build only the required targets and variants.
4. **Static validation:** inspect binaries, dependencies, package contents, version metadata, and SHA-256.
5. **Runtime acceptance:** on Colab, verify backend/device detection and execute a minimal real inference/API smoke test.
6. **Branch publication:** commit only the validated compressed bundle, manifest, and checksum to this working branch.
7. **Runtime adoption:** update Colab bootstrap only after acceptance passes; keep the current working startup path intact until then.

## Failure policy

- No guessed CUDA target or architecture list; the initial target is evidence-based `sm_75` for the observed T4.
- No automatic nightly promotion.
- No repeated full source rebuild on ordinary Colab startup.
- No silent fallback from CUDA to CPU when CUDA was requested; report the incompatibility explicitly.
- No claim of CUDA readiness based solely on successful compilation.
- A failed acceptance test blocks publication and runtime adoption.

## Consequences

This approach minimizes recurring Colab startup cost and keeps the build reproducible. It also keeps the initial target matrix small. Committing a binary bundle satisfies the branch requirement but increases repository history; the size gate and single accepted package policy limit that cost. CUDA compatibility remains unproven until tested in the actual Colab environment.
