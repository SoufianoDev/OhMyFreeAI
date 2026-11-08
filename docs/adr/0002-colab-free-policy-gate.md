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


## Update — official Colab MCP API reviewed 2026-11-08

Google documents an official MCP interface for Colab runtime management. The `create_runtime` reference returns authenticated runtime connection information and describes interaction through the standard Jupyter REST API and/or a client library.

This does not change the decision above. The MCP reference documents an API for managing and interacting with notebook runtimes; it does not expressly authorize turning a free managed runtime into a persistent general-purpose inference server or exposing model inference to OpenCode. The free-tier restrictions in the Colab FAQ remain applicable. Therefore the adapter remains disabled pending confirmation of the exact intended workflow.

References:
- https://developers.google.com/colab/api/reference/mcp/tools_list/create_runtime
- https://research.google.com/colaboratory/intl/en-GB/faq.html
