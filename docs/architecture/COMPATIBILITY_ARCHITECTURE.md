# Model and Runtime Compatibility

Compatibility is a multi-dimensional, evidence-backed assessment, not a model-family allowlist.

## Required dimensions

1. Model family and exact architecture.
2. Exact checkpoint/revision, configuration and tokenizer.
3. Conversion path and provenance.
4. Artifact format and quantization.
5. Required features: chat template, tool calling, vision/audio, MoE routing, speculative decoding, MTP, long context, multimodal projector and special attention requirements.
6. Engine version/build and selected backend.
7. Runtime hardware/software capabilities.
8. Resource fit under a named operating profile.
9. Real load, health and inference evidence.

## Runtime resource assessment

Capture actual accelerator model/count, free and total VRAM, system RAM, local disk, driver/CUDA runtime, CPU architecture/features, engine backend, available GPU layers, context length, KV-cache types, Flash Attention capability and expected compute/workspace overhead.

Do not estimate fit from parameter count alone. Include weight memory, KV cache, graph/compute buffers, runtime overhead and safety headroom. Report assumptions and confidence.

## Result vocabulary

- **Supported** — documented path and verified on the exact compatibility tuple.
- **Supported with limits** — verified with explicit restrictions (context, modalities, backend, quantization or concurrency).
- **Unsupported** — a concrete incompatibility is known; show why and the minimum requirement or missing feature.
- **Unverified** — plausible but no successful test evidence for this exact tuple.
- **Resource-ineligible** — technically supported but current hardware cannot meet the selected profile.
- **Policy-blocked** — runtime usage is not approved under provider policy.

## Family handling

Qwen, Llama, Gemma, Mistral/Mixtral, DeepSeek, MiniMax, Phi, GLM, MiniCPM and Muse are catalog families only. Yi, InternLM, Falcon and Command remain version/checkpoint-specific. No family-level status may imply all versions are supported.

Examples of required specificity:
- MiniMax M3 needs explicit MSA/Flash Attention and KV-cache/context-shift constraints.
- Muse Glimmer or Muse Spark open implementation does not prove that all weights are downloadable or licensing permits their use.
- A GLM-5.3-Flash result does not generalize to every GLM checkpoint.
- MiniCPM-V is a multimodal path, not interchangeable with text-only MiniCPM.

## UI contract

Show supported and unsupported catalog entries. Every unsupported entry states the missing capability and minimum requirements when those can be grounded. Every unverified entry says it has not been tested. Never turn unknown into supported.
