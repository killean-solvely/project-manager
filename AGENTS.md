# Project agent memory

This file is the project's committed home for project-intrinsic agent knowledge: build, test, release, architecture, and sharp-edge notes that should travel with the code.

- Add durable project-specific notes here as they are discovered through real work.

## MCP transports

The same `*mcp.Server` (built by `mcpserver.New`) is served over two transports, chosen by deployment:

- **stdio** - `cmd/mcp` binary, for local editor/desktop clients that launch a subprocess. Contract covered by `internal/mcpserver/stdio_test.go`.
- **HTTP** - mounted at `/mcp` in the `cmd/api` chi router (streamable HTTP, stateless, JSON responses), built from the *same* service graph and DB handle as the REST API. Gated by `MCP_HTTP_ENABLED` (default `true`). Covered by `internal/server/mcp_http_test.go`.

Both transports share one sqlite store. `cmd/api` runs an `*http.Server` with signal-driven graceful `Shutdown` so REST + MCP stop cleanly.

Sharp edge: the go-sdk `StreamableHTTPHandler` enables **DNS-rebinding protection by default** - requests reaching a loopback listener with a non-loopback `Host` header get a 403. Reach `/mcp` via a loopback Host in local dev; for remote/proxied deploys set `DisableLocalhostProtection: true` only behind a trusted proxy. `/mcp` is currently **unauthenticated** (inherits the REST API's no-auth posture on the same port); add auth at the proxy or via middleware before exposing it.
