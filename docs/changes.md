### Changes in this update

Date: 2025-11-17

#### What I improved
- Added explicit Not Implemented flow for unready providers (`gemini`, `claude`) to return HTTP 501 instead of a generic 502.
- Improved upstream error propagation: when providers return non-2xx, the gateway now surfaces the upstream HTTP status (e.g. 400/401/429) with a concise message in the uniform error envelope.
- Added graceful shutdown on SIGINT/SIGTERM with time-bounded server shutdown.
- Added sensible defaults: if `PROVIDER` is empty it defaults to `openai`; for `openai` with empty `PROVIDER_URL`, it defaults to `https://api.openai.com/v1`.
- Hardened HTTP server timeouts (read/write/idle) for robustness.
- Optimized Docker image with a multi-stage build and non-root user.
- Updated README and specification to reflect the new behavior.

#### Files changed
- `handlers/provider.go`
  - Added `NotImplementedError` to represent providers that are not yet implemented.
- `handlers/gemini.go`, `handlers/claude.go`
  - Return `NotImplementedError` instead of a generic error.
- `router.go`
  - Added error branching in `handleChat` to:
    - Return 501 for not-implemented providers
    - Surface upstream status/message from provider errors (`HTTPStatusError`)
  - Added helper `extractProviderMessage` to derive a concise message from upstream bodies.
- `main.go`
  - Added defaults for `PROVIDER` and `PROVIDER_URL` (for `openai`)
  - Added server read/write/idle timeouts
  - Implemented graceful shutdown via `signal.NotifyContext` and `Server.Shutdown`
- `README.md`
  - Documented 501 behavior and default `PROVIDER_URL` for `openai`
  - Clarified upstream error surfacing
- `docs/specification.md`
  - Added HTTP 501 to the error table and examples
- `example.env`
  - Clarified default behavior for `PROVIDER_URL` when using `openai`
- `Dockerfile`
  - Switched to multi-stage build (alpine) with non-root runtime user and certs

#### Impact
- Clearer ergonomics for users selecting unimplemented providers.
- Better debuggability when providers return errors (status and message preserved).
- Safer production behavior with graceful shutdown and timeouts.
- Smaller, more secure container image.


