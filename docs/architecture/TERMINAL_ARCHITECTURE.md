# Terminal Architecture

The local terminal is a native platform PTY. It supports resize, stdin/stdout/stderr, signals, exit status, environment policy and clean cancellation.

A remote command session is not automatically a PTY. If a provider adapter cannot supply PTY semantics, label it accurately as a remote command session and document missing capabilities.

Remote terminal support is not enabled for Colab Free because of the current policy restrictions on remote control. No SSH, reverse tunnel, proxy or notebook UI bypass is permitted.
