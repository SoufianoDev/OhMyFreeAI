# llama.cpp Engine Packaging Tools

These scripts prepare and validate a Linux engine bundle. They do not select a
moving upstream branch or claim that CUDA works merely because a binary
contains CUDA support.

The initial manual workflow defaults to the stable `v0.9.0` source tag and a
CPU-only source build. This is deliberate: the v0.9.0 release page identifies
a separate nightly build asset rather than a stable Linux binary bundle. The
official-asset path remains available only when an exact official release
asset URL is supplied and pinned to its matching tag.

## Discover the Colab runtime

Run this in the target Colab session before choosing a CUDA package:

```bash
python3 scripts/engine/detect_environment.py > colab-engine-environment.json
```

The report records Linux/CPU details, memory, disk, NVIDIA driver/GPU
information, and CUDA compiler/runtime evidence when available. Missing GPU
tools are reported as unavailable; they are not treated as proof of a CPU-only
runtime or as a reason to guess a CUDA target.

## Package and benchmark

```bash
python3 scripts/engine/package_engine.py \
  --source build/bin \
  --output-dir dist \
  --levels 3 6 9 15 19 \
  --max-size-mib 90
```

The source directory should contain the complete runtime bundle, including any
required shared libraries and the upstream license file. The script tests each
archive by extracting it and comparing SHA-256 for every file. It selects the
smallest archive whose extraction time is within 15% of the fastest measured
extraction, provided it stays under the size ceiling.

Outputs:

- `dist/llama-engine.tar.zst`
- `dist/llama-engine.tar.zst.sha256`
- `dist/compression-report.json`

## Validate extracted engine

```bash
python3 scripts/engine/validate_engine.py --root path/to/llama-engine \
  --require-ldd-clean
```

This checks the executable and dynamic dependencies. It does not replace a
model-backed inference test or a real Colab GPU acceptance test.

## Provenance manifest

```bash
python3 scripts/engine/generate_manifest.py \
  --output dist/manifest.json \
  --source-ref <pinned-tag-or-commit> \
  --source-mode source-build \
  --backend cpu
```

Never use a moving branch name or `latest` as the production source pin.
