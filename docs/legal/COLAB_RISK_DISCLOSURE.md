# Google Colab Free — Policy and Risk Disclosure

**Status: feasibility gate; do not treat Colab Free as a guaranteed inference host.**

## Official policy facts

Google's Colab FAQ states:
- Free compute resources, including GPUs and TPUs, are provided at no charge, but resources are not guaranteed or unlimited and usage limits fluctuate.
- Managed virtual machines are deleted after idle periods and have a maximum lifetime.
- Colab managed runtimes disallow file hosting/media serving and other web-service offerings unrelated to interactive compute, remote proxies, cryptocurrency mining, denial-of-service, password cracking, account multiplication to work around restrictions, and containerization used to circumvent policy.
- Free-tier managed runtimes may be terminated without warning for remote control such as SSH shells or remote desktops, bypassing the notebook UI to interact primarily through a web UI, and distributed workers.

Official source: https://research.google.com/colaboratory/faq.html

## Product impact

The intended pattern of a local client calling a local Gateway that forwards inference requests to a remote llama-server may constitute remote control, a web-service offering, or a notebook-UI bypass depending on how it is operated. This is a serious policy risk. Prior successful experiments do not establish provider approval.

## Engineering rules

- Do not implement tunnels, proxy relays, SSH/remote desktop, hidden web endpoints, account rotation or other evasion.
- Do not promise GPU model, memory, uptime, session duration or persistence.
- Keep provider policy state separate from technical health.
- A technically reachable session may still be `policy_blocked`.
- Do not activate the Colab remote-inference adapter until the exact interaction has been reviewed against current official terms and found permitted.
- If no compliant free usage pattern supports the required product flow, state that the feature is infeasible under the current free-only constraint rather than disguising the limitation.

## User disclosure

Colab Free is best-effort interactive compute. Sessions and resources can change or terminate without warning. Users must remain within Google's current terms. OhMyFreeAI does not guarantee availability or override provider restrictions.


## Official Colab MCP API — scope and limits (reviewed 2026-11-08)

Google now documents an official Colab MCP API. Its runtime-management reference describes creating managed runtimes, returning authenticated runtime connection information, and using the standard Jupyter REST API and/or a client library to interact with the runtime.

This is an official runtime-management interface, but the API reference does **not** state that a managed runtime may be repurposed as a general-purpose model-serving host, nor does it grant an exception to the Colab FAQ's restrictions for free managed runtimes. In particular, an authenticated Jupyter connection is not evidence that a persistent `llama-server`, an externally consumed inference endpoint, or OpenCode-driven remote inference is an approved use.

For OhMyFreeAI, keep these cases distinct:

- **Runtime/session management through Google's documented MCP API:** an official API surface that may be evaluated for notebook-oriented lifecycle management.
- **User-driven interactive notebook computation:** remains subject to the FAQ's interactive-compute priority and all applicable restrictions.
- **Long-running model server or inference API consumed by OpenCode:** not cleared by the existence of the MCP API; keep remote-inference activation disabled until the exact workflow is explicitly confirmed as permitted.

Do not use the MCP API, Jupyter REST API, or any other interface as a means to evade Colab's restrictions.

Official references:
- Colab FAQ: https://research.google.com/colaboratory/intl/en-GB/faq.html
- Colab MCP `create_runtime` reference: https://developers.google.com/colab/api/reference/mcp/tools_list/create_runtime
