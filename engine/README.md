# llama.cpp Engine Bundle

This directory contains the CI-accepted Linux x86_64 CPU engine bundle on
`feature/llama-engine-build-factory`.

## Accepted bundle

- `llama-engine.tar.zst` — lossless TAR+Zstandard bundle built from upstream
  `ggml-org/llama.cpp` tag `v0.9.0`.
- `llama-engine.tar.zst.sha256` — SHA-256 checksum for the archive.
- `manifest.json` — source pin, build provenance, bundled file hashes, and
  shared-library dependency report.
- `compression-report.json` — measured compression/extraction results for
  Zstandard levels 3, 6, 9, 15, and 19.

The accepted source commit is
`cd959bae99ff4bed22bef2604f45241c09bbb8c8`. The archive is approximately
12.71 MiB and passed the CI build, executable/version checks, shared-library
dependency checks, archive integrity, lossless extraction, and 90 MiB size gate.

## Colab runtime acceptance

A passing CI build does not prove model inference works on the actual Colab
runtime. Before adopting this engine in the existing launcher, run the isolated
smoke test against a GGUF model already present in the runtime:

```bash
python3 scripts/colab/engine_smoke_test.py --model /path/to/your/model.gguf
```

The script downloads the immutable bundle revision, verifies its SHA-256,
extracts it into a temporary directory, starts a local CPU server, and sends a
real completion request. It does not modify the existing launcher or install
the engine into Drive. A successful CPU smoke test is required before runtime
adoption. CUDA readiness remains untested and requires a separate test on an
actual compatible Colab GPU.

## Publication policy

The workflow publishes only after binary validation, archive integrity checks,
lossless extraction comparison, and the 90 MiB size gate succeed. Do not commit
intermediate object files, build directories, compiler caches, or uncompressed
staging trees. Keep the current working Colab launcher/Gateway/OpenCode path
intact until runtime acceptance is complete.
