# llama.cpp Engine Packaging Tools

These scripts prepare and validate a Linux engine bundle. They do not select a
moving upstream branch or claim that CUDA works merely because a binary
contains CUDA support.

The Colab target is a CUDA-enabled Linux x86_64 build for the verified Tesla T4
(Compute Capability 7.5), built from pinned stable `v0.9.0` source with CUDA
Toolkit 13.0. The bundle includes the portable CPU backend for fallback, but
CUDA is mandatory for this build target. Actual GPU readiness requires the
real Colab T4 acceptance test.

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
  --backend cuda-t4-sm75 --cuda-architectures 75
```

Never use a moving branch name or `latest` as the production source pin.


## CUDA T4 build target

The CI workflow installs CUDA Toolkit 13.0 on Ubuntu 22.04 without installing a host NVIDIA driver, then builds with `GGML_CUDA=ON`, `CMAKE_CUDA_ARCHITECTURES=75`, and `GGML_NATIVE=OFF`. CUDA toolkit runtime libraries are copied into `bin/` for origin-relative loading. `libcuda.so.1` is not bundled; it is supplied by the Colab host driver. CI validates the CUDA plugin and packaged toolkit dependencies. Only the real Colab T4 test validates device initialization and inference.
