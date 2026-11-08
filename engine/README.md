# llama.cpp Engine Bundle

This directory is the publication location for the single accepted Linux x86_64
engine bundle on `feature/llama-engine-build-factory`.

The binary archive is intentionally absent until a pinned upstream asset or
source revision has been selected and the resulting package passes validation.
Do not commit a placeholder archive, an untested binary, build output, or
intermediate compression candidates here.

## Expected accepted outputs

- `llama-engine.tar.zst` — final lossless TAR+Zstandard bundle.
- `llama-engine.tar.zst.sha256` — SHA-256 checksum for the archive.
- `manifest.json` — source and build provenance.
- `compression-report.json` — measured compression/extraction results for
  Zstandard levels 3, 6, 9, 15, and 19.

The manual workflow in
`.github/workflows/build-llama-engine.yml` publishes these files only after
binary validation, archive integrity checks, lossless extraction comparison,
and the 90 MiB size gate succeed.

## Acceptance boundary

A passing CPU build does not establish CUDA readiness. CUDA must be validated
with a compatible real GPU in Google Colab, including an actual model-backed
inference test, before the runtime may select a CUDA package.
