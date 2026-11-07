# Third-Party Services and Components

| Service/component | Intended role | Data / risk | Status |
|---|---|---|---|
| Google Colab Free | Proposed interactive compute runtime | Account-linked sessions, variable resources and strict managed-runtime policy | Policy-gated; no remote inference activation yet |
| Google Drive | Persistent model/engine artifact storage where permitted | User files, Drive mount availability and I/O limits | Planned; stage to runtime-local disk |
| llama.cpp | Local inference engine | Downloaded binary and model artifacts | Primary engine; pin tested stable release |
| Hugging Face or other model host | User-selected model artifact source | Model downloads, tokens for gated repositories, model license terms | Per-artifact; no credential logging |
| OpenCode | First client integration | Local configuration changes and inference requests | Planned; backup and rollback required |
| Wails | Desktop shell | Native WebView and OS dependencies | Candidate; PoC required |
| aria2 | Optional download accelerator | External executable and download sources | Optional only; Go fallback required |

The final release must list exact versions, licenses, binary provenance, network destinations, data categories and whether each dependency is bundled or user-installed. This table is architectural intent, not a statement that all integrations are already implemented.
