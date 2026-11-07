# Historical prototype audit

- Audit type: read-only static inspection
- Date: 2026-11-07
- Execution policy: none of the supplied patch/monitor scripts or archived applications were executed; no Colab command or local process was touched.

## Sources inspected

| Source | SHA-256 |
|---|---|
| SoufianoDev-Monitor-GUI-Remote-Complete.zip | f3f55e831502231faea4292deb7789bc305e1c273fae98d334a1f5c025f6c01f |
| SoufianoDev-Monitor-GUI.zip | 73336decf712690392c483eeaa3efa17c88e9497f130bc6b5bfc404c9acc3b40 |
| fix-soufianodev-free-ai.py | 6fae50b6375bd029fa4e3215f1932e3850bc7b0d443949c255d89a7248fa537e |
| fix-soufianodev-sse-framing.py | f589bbc1c4e26a20fc41abe82614a8a193d5ccf179c60c24fcef0a1555ad98c0 |
| monitor-soufianodev-free-ai.py | 1505dc10a9984d6c8614d5e3286e459b80ba7595de0e70af37353321c3684185 |

The archives contain a Wails v2.10.2 Linux monitor. The remote-complete archive extends the earlier local monitor with Colab telemetry, a busy-session guard, GPU/VRAM/RAM/server telemetry and JSONL history. The Python monitor is the clearest behavioral reference for the read-only monitoring loop.

## Recovered working path

- Local OpenCode client.
- Local OpenAI-compatible Gateway at 127.0.0.1:18081.
- Colab CLI session named qwen-opencode.
- Remote llama-server endpoint at 127.0.0.1:8081, queried through colab exec.
- Remote telemetry reads llama-server /props, /proc process arguments and RSS, nvidia-smi GPU metrics, and /proc/meminfo.
- The monitor reads ~/.config/soufianodev-free-ai/runtime.env for CONTEXT_SIZE and CACHE_TYPE and tails the local gateway log.
- Remote sampling is normally every 8 seconds. When a matching colab exec process is detected for qwen-opencode, remote polling is deferred and retried after approximately 3 seconds.
- The monitor does not stop or restart OpenCode, the Gateway or the Colab server.
- The SSE repair preserves blank upstream lines. This is essential because blank lines delimit SSE events; filtering them breaks event framing and can prevent OpenCode from seeing complete events and the terminal [DONE] marker.
- The context/cache patch targets CONTEXT_SIZE=262144 and CACHE_TYPE=q8_0, verifies the final context and /v1/models metadata, runs a real completion test, and checks minimum post-test free VRAM/RAM before printing its final readiness marker.

## Important differences from the new product

- The archived GUI is Wails v2.10.2, Go 1.24, CGO/GCC, GTK3 and WebKit2GTK 4.1, and its README targets Ubuntu 25.04. It is a Linux monitor, not a cross-platform product shell.
- Its main.go is a monolithic monitor with Linux /proc and syscall.Kill usage, hard-coded session/ports/paths, map[string]any telemetry and no production service/port boundaries.
- The archived UI uses a dark bespoke dashboard with decorative orbit/hero graphics and dense cards. It is not the Material 3 product UI and should not be copied into the new application.
- The historical JSC_SIGNAL_FOR_GC=12 workaround is documented for that Wails v2/WebKit2GTK build. Do not blindly carry it into Wails v3/GTK4/WebKitGTK 6.0; reproduce the actual runtime warning and validate the selected signal first.

## Static findings and risks

### fix-soufianodev-free-ai.py
- It performs exact text replacements and regex substitution against the user's launcher. It checks several expected blocks and compiles the embedded Colab Python before writing.
- It backs up the launcher and runtime.env, and writes the launcher through a temporary file before replacing it.
- **Data-loss risk:** it overwrites runtime.env with only CONTEXT_SIZE and CACHE_TYPE. Any other existing runtime settings, comments or future keys are discarded. It does not merge/preserve unrelated keys.
- The runtime.env write is not atomic, and an error after replacing the launcher but before the final bash -n/rollback path can leave a partially updated pair of files.
- The script validates shell syntax and embedded Python syntax, but does not execute the launcher or prove the remote model/runtime is compatible.
- The 1536 MiB free-VRAM and 1024 MiB available-RAM thresholds are heuristics, not a complete model-fit calculation.
- It assumes the current llama.cpp build exposes the expected /props and /v1/models metadata fields, and that the model/backend supports Q8_0 K/V at 262144 context.
- GPU parsing splits CSV on commas and selects only the first GPU; GPU names containing commas or multi-GPU layouts can invalidate the parse/interpretation.
- Its colab exec pipeline uses tee and then validates explicit readiness markers. This protects against a zero exit status with an incomplete remote run, but it is still a launcher patch, not a live verification performed during this audit.

### fix-soufianodev-sse-framing.py
- It backs up and patches both the persistent launcher and the installed Gateway using exact-match checks.
- It correctly preserves every upstream SSE line, including blank delimiters.
- The two-file patch is not transactional: if the second patch fails, the first may remain modified.
- It terminates matching Gateway processes using /proc command-line matching, waits 10 seconds, then escalates to SIGKILL. This is narrower than killing by port, but it still relies on command-line identity and has a PID-reuse race.
- It does not verify the listener's owner before stopping the matching Gateway. If an unrelated process owns the port after the managed Gateway is stopped, the restart can fail.
- If restart or the SSE test fails, the script does not restore both original files or automatically stop/roll back the newly started Gateway.
- It writes the PID file directly rather than atomically.
- The SSE validation itself is useful: it parses blank-line-delimited events, rejects malformed JSON, requires a finish_reason, requires [DONE], and checks a known response token.

### monitor-soufianodev-free-ai.py and Wails v2 monitor
- Both implement the useful busy-session guard, but detection scans /proc and can race with a new Colab command starting immediately after the scan.
- Process liveness is based on PID existence (os.kill(pid, 0) / syscall.Kill(pid, 0)); it does not prove that the PID still belongs to the intended process.
- The Wails app's OpenProjectFolder/OpenHistory use xdg-open and the app depends on Linux /proc/syscall behavior; it is not a Windows/macOS implementation.
- The Python monitor's JSONL file is opened with default permissions; it does not explicitly enforce mode 0600. The Go monitor uses 0600 on creation but does not tighten an already-existing file.
- Both monitor implementations read the entire gateway log before selecting the last few lines, and append history without rotation/retention. Large files can increase memory and I/O.
- Remote error text can include the tail of colab CLI stdout/stderr. It is truncated, but not explicitly redacted.
- The remote probe reports the first nvidia-smi row, not an aggregate across GPUs. It does not report CPU, and the code correctly avoids inventing a CPU metric that the original probe does not measure.
- The remote context value from /props and the server's --ctx-size argument are separate signals; neither alone proves actual model compatibility or available memory for a future context length.
- The monitor is read-only telemetry. It is not a runtime lifecycle manager, process supervisor or inference health certificate.

## What should be preserved

1. Never issue telemetry colab exec while another colab exec is active for the target session; use a serialized command broker rather than independent polling and control paths.
2. Keep the SSE blank-line framing invariant and add a Gateway regression test for complete event boundaries and [DONE].
3. Treat CONTEXT_SIZE=262144 and CACHE_TYPE=q8_0 as desired configuration values, not proof that every model/checkpoint/hardware combination can run them.
4. Require actual health, loaded model identity, verified configuration and a successful completion before reporting model readiness.
5. Keep the monitor read-only and ensure monitoring failures never block inference or control.
6. Use typed telemetry, explicit Live/Stale/Unavailable/Error states, bounded history, private file permissions, redacted diagnostics and identity-safe process ownership in the new architecture.
7. Keep the old monitor UI as historical reference only; continue with the separate Material 3 product design.

## Still unverified

- The current Colab session and current llama.cpp version still expose the exact /props and /v1/models fields assumed by the patches.
- Q8_0 K/V and 262144 context are compatible with the exact model checkpoint, tokenizer, llama.cpp build and available Colab hardware.
- The minimum memory reserve is sufficient for the selected model and workload; it is not a universal threshold.
- The busy-process detector covers every invocation form used by the installed Colab CLI.
- The historical SSE fix is still required by the current Gateway version; this audit did not execute a live streaming request.
- The Wails v2 signal workaround is needed or safe in the new Wails v3/GTK4/WebKitGTK 6.0 runtime.
- No scripts, launchers, Gateway processes, OpenCode sessions or Colab sessions were modified or executed during this audit.
