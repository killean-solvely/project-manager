# projectmanager

A personal, local-first tool for capturing project ideas, curating them, and
driving active projects through documented, well-tracked work. Designed MCP-first
so Claude (or any AI service) can read and act on it autonomously.

See [`PROJECT_SPEC.md`](PROJECT_SPEC.md) for the full design and rationale.

## Model in one breath

- **Portfolio** is a *lifecycle*, not a kanban: `idea → active → archived`. Active
  projects carry a mode (`developing` or `maintaining`).
- **Each project** owns a *docs library* (typed docs, tracked for completeness) and
  a *kanban board* (columns + task cards that can link to docs).

## Architecture

All behavior lives in the service layer. Two thin adapters sit over it — an HTTP
API (`cmd/api`) and a stdio MCP server (`cmd/mcp`) — and neither holds logic.
`cmd/api` also mounts the same MCP server as an HTTP handler at `/mcp`, so MCP
ships over both stdio and HTTP from one service graph and one database handle.

```
cmd/api/                 HTTP server: wires repos + services + chi; also mounts
                         MCP over HTTP at /mcp (see MCP_HTTP_ENABLED below)
cmd/mcp/                 MCP server: wires repos + services + stdio
internal/
  config/                env + .env loading (godotenv + viper), defaults
  models/                pure domain types (uuid IDs, no tags)
  persistence/           repo interfaces + implementations
    memory/              in-memory repos (used in tests)
    sqlite/              SQLite repos (used by both binaries)
  projects/              lifecycle service + state-machine invariants
  docs/                  docs library + completeness checklist
  boards/                board + columns + tasks (one cohesive service)
  server/                chi router, DTOs, request types, handlers
  mcpserver/             MCP tools, resources, prompts
```

## Configuration

Settings are loaded once at startup by `internal/config` (godotenv + viper) and
passed down explicitly. Sources, highest precedence first: real environment
variables, then a `.env` file in the working directory (optional - see
`.env.example`), then built-in defaults.

| Variable | Default | Purpose |
|---|---|---|
| `PORT` | `4523` | HTTP listen port for `cmd/api` |
| `DB_PATH` | `~/.projectmanager/projectmanager.db` | SQLite file, shared by `cmd/api` and `cmd/mcp` |
| `MCP_HTTP_ENABLED` | `true` | Mount the MCP server at `/mcp` in `cmd/api`; does not affect the stdio `cmd/mcp` binary |

`PM_DB_PATH` is still honored as a legacy alias for `DB_PATH`; if both are set,
`DB_PATH` wins.

## Persistence

State lives in a single SQLite file (pure-Go `modernc.org/sqlite`, no cgo). Both
`cmd/api` and `cmd/mcp` open the same file, so they share one store.

- Default path: `~/.projectmanager/projectmanager.db`
- Override with `DB_PATH`, e.g. `DB_PATH=./dev.db make run` (or set it in `.env`)

The schema is applied automatically on open; there is no separate migration step yet.

## Run

```sh
make run          # starts the API on :4523 (set PORT to change)
# or: go run ./cmd/api

curl localhost:4523/health
```

Quick smoke test:

```sh
# create an idea
curl -s -XPOST localhost:4523/api/projects \
  -d '{"name":"My idea","summary":"a thing","tags":["x"]}'

# list ideas
curl -s 'localhost:4523/api/projects?status=idea'

# promote it (use the id from above)
curl -s -XPOST localhost:4523/api/projects/<id>/promote -d '{"mode":"developing"}'

# what docs are still missing?
curl -s localhost:4523/api/projects/<id>/documents/missing
```

## API

| Method | Path | Purpose |
|---|---|---|
| GET | `/api/projects?status=&` | list projects (optional status filter) |
| POST | `/api/projects` | create an idea |
| GET | `/api/projects/{id}` | get a project |
| PATCH | `/api/projects/{id}` | update name/summary/description/tags |
| POST | `/api/projects/{id}/promote` | idea → active (`{mode?}`) |
| POST | `/api/projects/{id}/mode` | switch active mode (`{mode}`) |
| POST | `/api/projects/{id}/archive` | → archived (`{reason?}`) |
| POST | `/api/projects/{id}/revive` | archived → active (`{mode?}`) |
| GET | `/api/projects/{id}/documents` | list a project's docs |
| GET | `/api/projects/{id}/documents/missing` | required docs not yet complete |
| PUT | `/api/projects/{id}/documents/{type}` | upsert a doc by type (`{title,content,status}`) |
| GET | `/api/projects/{id}/board` | board + columns (created on first access) |
| GET | `/api/projects/{id}/tasks` | list tasks |
| POST | `/api/projects/{id}/tasks` | create a task (`{column_id,title,...}`) |
| PATCH | `/api/tasks/{id}` | update a task |
| POST | `/api/tasks/{id}/move` | move a task (`{column_id,position}`) |
| POST | `/api/tasks/{id}/link` | link/unlink a document (`{document_id?}`) |
| DELETE | `/api/tasks/{id}` | delete a task |

## MCP server

The MCP server exposes the same services to Claude (or any MCP client) — it
imports `internal/projects`, `internal/docs`, and `internal/boards` directly and
holds no logic of its own. It is served over two transports from one
`*mcp.Server`, built once and shared:

- **stdio** — the `cmd/mcp` binary, for local editor/desktop clients that launch
  a subprocess.
- **HTTP** — mounted at `/mcp` in `cmd/api` (streamable HTTP, stateless, JSON
  responses), gated by `MCP_HTTP_ENABLED` (default `true`, see
  [Configuration](#configuration)). It shares `cmd/api`'s port, database handle,
  and no-auth posture — add auth at a proxy or middleware before exposing it
  beyond the local machine. The go-sdk handler also enables DNS-rebinding
  protection by default, so requests must reach it via a loopback `Host` header
  in local dev (or set `DisableLocalhostProtection` behind a trusted proxy for
  remote deploys).

```sh
make mcp           # run the stdio server directly (go run ./cmd/mcp)
make mcp-build     # build a binary at bin/pm-mcp (faster startup for clients)
```

Wire the stdio server into **Claude Code**:

```sh
make mcp-build
claude mcp add projectmanager -- "$(pwd)/bin/pm-mcp"
```

Or **Claude Desktop** (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "projectmanager": { "command": "/absolute/path/to/projectmanager/bin/pm-mcp" }
  }
}
```

What it exposes:

- **Tools** — `list_projects`, `get_project`, `create_idea`, `update_project`,
  `promote_project`, `set_project_mode`, `archive_project`, `revive_project`,
  `list_documents`, `upsert_document`, `get_missing_docs`, `get_board`, `list_tasks`,
  `create_task`, `update_task`, `move_task`, `link_task_document`, `delete_task`.
- **Resources** — `pm://project/{id}`, `pm://project/{id}/doc/{type}`,
  `pm://project/{id}/board`.
- **Prompts** — `flesh_out_idea`, `generate_missing_docs`, `draft_tasks_from_spec`.

> **Shared store:** `cmd/api` and `cmd/mcp` read and write the same SQLite file
> (see [Persistence](#persistence)), so anything created through Claude shows up in
> the API and vice versa — they default to the same path.

## Web frontend

A React + TypeScript + Tailwind SPA in `web/`, built with Vite. It talks to the API
through a dev proxy (`/api` → `:4523`), so run both together:

```sh
make run                  # terminal 1: API on :4523
make web-install          # first time only
make web                  # terminal 2: UI on :5173
```

Open http://localhost:5173 — the portfolio shows projects by lifecycle stage; open a
project for its docs library (markdown editor + completeness checklist) and its kanban
board (drag cards between columns).

## Development

```sh
make test     # go test ./...
make vet      # go vet ./...
make tidy     # go mod tidy
```

## Next steps

1. ~~`cmd/mcp` — stdio MCP server over the same services.~~ **Done.**
2. ~~SQLite persistence shared by both binaries.~~ **Done.**
3. ~~`web/` — the React frontend (portfolio + project dashboard).~~ **Done.**

Possible polish: search/filter across projects, intra-column drag reordering,
per-project doc templates, and auth if it ever leaves the local machine.
