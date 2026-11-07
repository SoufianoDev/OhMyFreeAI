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
