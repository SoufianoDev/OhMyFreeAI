# ADR 0002: Colab Free policy gate

- Status: Accepted
- Date: 2026-11-07

## Context

Google Colab Free provides best-effort compute, not guaranteed GPU/session capacity. Its official FAQ prohibits remote control such as SSH shells/remote desktops, bypassing notebook UI to interact primarily through a web UI, and unrelated web-service offerings on managed runtimes. The product's local client -> Gateway -> remote llama-server flow may fall within these categories.

## Decision

The Colab remote inference adapter is policy-gated and must remain disabled until the exact interaction pattern is reviewed and determined to be permitted under current official terms. The application must not implement tunnels, proxy relays, SSH, remote desktop, notebook UI bypass, account rotation or other evasion.

## Consequences

- Technical reachability does not imply policy eligibility.
- UI distinguishes policy-blocked from offline/unavailable.
- If no compliant free pattern supports the product flow, document the feature as infeasible under the free-only requirement rather than work around restrictions.

Source: https://research.google.com/colaboratory/faq.html
