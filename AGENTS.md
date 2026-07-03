# Project agent memory

This file is the project's committed home for project-intrinsic agent knowledge: build, test, release, architecture, and sharp-edge notes that should travel with the code.

- Add durable project-specific notes here as they are discovered through real work.

## MCP transports

The same `*mcp.Server` (built by `mcpserver.New`) is served over two transports, chosen by deployment:

- **stdio** - `cmd/mcp` binary, for local editor/desktop clients that launch a subprocess. Contract covered by `internal/mcpserver/stdio_test.go`.
- **HTTP** - mounted at `/mcp` in the `cmd/api` chi router (streamable HTTP, stateless, JSON responses), built from the *same* service graph and DB handle as the REST API. Gated by `MCP_HTTP_ENABLED` (default `true`). Covered by `internal/server/mcp_http_test.go`.

Both transports share one sqlite store. `cmd/api` runs an `*http.Server` with signal-driven graceful `Shutdown` so REST + MCP stop cleanly.

Sharp edge: the go-sdk `StreamableHTTPHandler` enables **DNS-rebinding protection by default** - requests reaching a loopback listener with a non-loopback `Host` header get a 403. Reach `/mcp` via a loopback Host in local dev; for remote/proxied deploys set `DisableLocalhostProtection: true` only behind a trusted proxy. `/mcp` is currently **unauthenticated** (inherits the REST API's no-auth posture on the same port); add auth at the proxy or via middleware before exposing it.

## Container

Two images, orchestrated by `docker-compose.yml`:

- **Backend** (`Dockerfile`, repo root) - multi-stage `golang:1.25` build of `./cmd/api` to `gcr.io/distroless/static:nonroot` (uid 65532). Built with `CGO_ENABLED=0`: the SQLite driver is pure-Go `modernc.org/sqlite`, so the binary is fully static and needs no libc. The db + WAL sidecars live on a named volume at `/data`; the image bakes a nonroot-owned `/data` (via an alpine stage + `COPY --chown`) so a fresh volume is writable by uid 65532. Single-writer sqlite (`SetMaxOpenConns(1)`) -> single instance only, never scale replicas against one volume.
- **Frontend** (`web/Dockerfile` + `web/nginx.conf`) - `node:22` build of the Vite SPA -> `nginx:alpine` serving `dist` with SPA fallback, reverse-proxying `/api`, `/health`, `/mcp` to the `api` service. Build context is `./web` (the root `.dockerignore` excludes `web/` from the backend context).

Sharp edge (the `/mcp` DNS-rebinding caveat, in container form): the go-sdk protection above **can** 403 `/mcp` when the `Host` is non-loopback (as nginx forwards it behind the proxy). Left as a documented limitation - the Go code is deliberately not changed to opt out. Empirically, with the current go-sdk version, `/mcp` through the web proxy was verified to work (valid JSON-RPC), so it isn't broken today; treat it as a version-dependent risk and, if it 403s, reach `api` directly via a loopback `Host`. `/api` and `/health` proxy fine.

Healthcheck is orchestrator-level only: a compose HTTP probe on `/health`, placed on the `web` service (the distroless backend has no shell/curl, so a healthcheck command cannot run in it, and nothing is baked into the image). There is no Dockerfile `HEALTHCHECK`. The probe must hit `http://127.0.0.1/health`, **not** `localhost` - alpine resolves `localhost` to `::1` first but nginx `listen 80;` binds IPv4 only, so a `localhost` probe is refused.
