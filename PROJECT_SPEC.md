# Project Spec — Idea & Project Manager

A personal tool for capturing project ideas, curating them, and driving the ones
that go active through to documented, well-tracked work. Built MCP-first so Claude
(or any AI service) can read and act on it autonomously.

> Status: **planning**. This document is the contract. When we build, the backend
> is scaffolded from this with the `go-project-init` skill, which produces exactly
> the layered structure described under Architecture.

---

## 1. Goals & non-goals

**Goals**
- Capture rough ideas fast, then curate them into a real backlog.
- Move an idea into an *active* project that has living documentation + a kanban board.
- Support *continuous work* (maintenance), not just build-to-done.
- Expose the whole thing over MCP as a first-class interface, not an afterthought —
  an AI can flesh out an idea, fill in missing docs, and manage the board on its own.

**Non-goals (v1)**
- Multi-user / auth / sharing. This is a single-user, local-first tool.
- Hosting. It runs on the local machine.
- Real-time collaboration, notifications, integrations.

**Decisions locked**
- Scope: single-user, local-first.
- Persistence: in-memory first, then **SQLite** (file-based, zero-ops). No Postgres unless it ever goes hosted.
- Docs: a **separate first-class library**; task cards *link* to docs. Board = work, docs = deliverables.
- MCP transport: **local stdio**, in-process (the MCP binary imports the service layer directly — no network hop).

---

## 2. The core mental model

Two levels with different semantics:

- **Level 1 — Portfolio = a lifecycle, not a kanban.** Projects don't burn down to
  "done"; they *mature*: `idea → active → archived`. Some sit in `active/maintaining`
  indefinitely. Model it as a small state machine; render it as columns if desired.
- **Level 2 — Project dashboard = where kanban actually lives.** Each project owns a
  docs library (typed, tracked for completeness) and a kanban task board.

---

## 3. Architecture

MCP is a **peer interface**, not a bolt-on. All behavior lives in a service layer;
REST and MCP are two thin adapters that both call it. Design the service interface
first — REST DTOs and MCP tool schemas are just two serializations of it.

```
cmd/
  api/        chi REST server        ─┐ both import internal/services;
  mcp/        stdio MCP server        ─┘ neither holds business logic
internal/
  models/         project, document, board, column, task
  persistence/    repo interfaces + in-memory impls  (→ sqlite later)
  services/       ALL business logic + the lifecycle state machine
  transport/
    http/         chi routes + request/response DTOs   (for React)
    mcp/          tool, resource, and prompt handlers  (for Claude / any AI)
web/              React + Vite app
```

- **Backend:** Go, chi router. Official `modelcontextprotocol/go-sdk` for MCP
  (confirm current state at build time). SQLite via a single embedded file.
- **Frontend:** React + Vite, React Query against the REST API, dnd-kit for the
  board, a markdown editor for docs.

---

## 4. Domain model

### Project (the lifecycle entity — an "idea" is just a project with `status=idea`)
| field | type | notes |
|---|---|---|
| id | uuid | |
| name | string | |
| summary | string | one-liner |
| description | text (markdown) | the spitball / braindump; grows over time |
| status | enum | `idea` \| `active` \| `archived` |
| mode | enum, nullable | `developing` \| `maintaining`; only set when `active` |
| tags | []string | |
| archived_reason | string, nullable | only set when `archived` |
| created_at / updated_at | timestamp | |
| promoted_at / archived_at | timestamp, nullable | lifecycle audit trail |

A project owns many Documents and (v1) one Board. Schema permits multiple boards later.

### Document (first-class artifact in the project's docs library)
| field | type | notes |
|---|---|---|
| id | uuid | |
| project_id | uuid | |
| type | enum | `overview` \| `technical` \| `spec` \| `api` \| `runbook` \| `other` |
| title | string | |
| content | text (markdown) | |
| status | enum | `draft` \| `in_review` \| `complete` |
| created_at / updated_at | timestamp | |

> "Missing" is **not** a stored status — a missing doc is simply the absence of a
> Document row for an expected type. See Doc template below.

### Doc template (drives the completeness checklist)
A v1 global default list of expected doc types with a `required` flag (e.g. overview +
technical + spec required, api + runbook optional). `get_missing_docs(project)` =
required template types that have no Document, or whose Document is still `draft`.
Per-project template overrides are a later addition.

### Board / Column
- **Board**: `id`, `project_id`, `name`. One per project in v1.
- **Column**: `id`, `board_id`, `name`, `position`. Default columns:
  `Backlog`, `Todo`, `In progress`, `Done` — renamable/reorderable.

### Task (card / ticket)
| field | type | notes |
|---|---|---|
| id | uuid | |
| board_id | uuid | |
| column_id | uuid | column = the task's status; source of truth for "where it is" |
| title | string | |
| description | text (markdown) | |
| priority | enum | `none` \| `low` \| `medium` \| `high` |
| labels | []string | |
| document_id | uuid, nullable | links a task to a doc ("write the API spec" → api doc) |
| position | int | order within column |
| created_at / updated_at | timestamp | |
| completed_at | timestamp, nullable | set when moved to a done column |

---

## 5. Lifecycle state machine

States: `idea`, `active`, `archived`.

| transition | from → to | rules |
|---|---|---|
| promote | idea → active | sets `mode` (default `developing`), stamps `promoted_at` |
| drop | idea → archived | kill an idea that won't be built |
| archive | active → archived | requires `reason`, stamps `archived_at`, clears `mode` |
| revive | archived → active | restores; sets `mode` (default `maintaining`) |
| set mode | active → active | `developing ↔ maintaining` (a field update, not a status change) |

Invariants enforced in the service layer: `mode` is non-null **iff** `status=active`;
`archived_reason` is set **iff** `status=archived`.

---

## 6. Persistence interfaces

Repo interfaces in `internal/persistence`, in-memory impls first, SQLite later.

- `ProjectRepo`: Create, Get, List(filter{status, tags}), Update, Delete
- `DocumentRepo`: Create, Get, GetByProjectAndType, ListByProject, Update, Delete
- `BoardRepo`: Create, GetByProject, Update
- `ColumnRepo`: Create, ListByBoard, Update, Reorder, Delete
- `TaskRepo`: Create, Get, ListByBoard(filters), Update, Move(column, position), Delete

---

## 7. Services (the contract)

- **ProjectService**: CreateIdea, Get, List, Update, PromoteToActive(id, mode),
  SetMode(id, mode), Archive(id, reason), Revive(id), DropIdea(id)
- **DocumentService**: List(projectID), Get(id), Upsert(projectID, type, title, content, status),
  Delete(id), GetMissingDocs(projectID), GetDocTemplate()
- **BoardService**: GetForProject(projectID), AddColumn, RenameColumn, ReorderColumns, RemoveColumn
- **TaskService**: List(boardID, filters), Get(id), Create, Update, Move(taskID, columnID, position),
  LinkDocument(taskID, docID), Delete(id)

---

## 8. REST API (for the React frontend)

```
GET    /api/projects?status=&tag=
POST   /api/projects                      # create idea
GET    /api/projects/{id}
PATCH  /api/projects/{id}
POST   /api/projects/{id}/promote         # { mode }
POST   /api/projects/{id}/archive         # { reason }
POST   /api/projects/{id}/revive
GET    /api/projects/{id}/documents
GET    /api/projects/{id}/documents/missing
PUT    /api/projects/{id}/documents/{type}   # upsert by type
GET    /api/projects/{id}/board
GET    /api/projects/{id}/tasks
POST   /api/projects/{id}/tasks
PATCH  /api/tasks/{id}
POST   /api/tasks/{id}/move               # { column_id, position }
DELETE /api/tasks/{id}
```

---

## 9. MCP surface (the reason this whole thing exists)

Served over stdio, in-process, importing `internal/services` directly. Plan for all
three MCP primitives:

**Tools (actions)**
- `list_projects(status?, tag?)`, `get_project(id)`
- `create_idea(name, summary?, description?, tags?)`, `update_project(id, …)`
- `promote_project(id, mode?)`, `set_project_mode(id, mode)`, `archive_project(id, reason)`, `revive_project(id)`
- `list_documents(project_id)`, `get_document(id)`, `upsert_document(project_id, type, title, content, status)`
- `get_missing_docs(project_id)`  ← powers "help fill out missing details"
- `list_tasks(project_id, filters?)`, `create_task(...)`, `update_task(id, …)`, `move_task(id, column_id, position)`, `link_task_document(task_id, document_id)`

**Resources (read-only context)** — let an AI *read* without spending a tool call
- `project://{id}` — project summary + lifecycle state
- `project://{id}/docs/{type}` — a document's markdown
- `board://{project_id}` — current board snapshot

**Prompts (canned workflows)**
- `flesh_out_idea(project_id)` — expand a raw idea into summary + initial docs
- `generate_missing_docs(project_id)` — fill the gaps from the completeness checklist
- `draft_tasks_from_spec(project_id)` — turn the spec doc into board tasks

---

## 10. Frontend views

- **Portfolio**: lifecycle columns (ideas / active / archived); active sub-grouped by
  mode (developing vs maintaining). Quick-add idea box.
- **Project dashboard**: tabs for Overview · Docs · Board. Doc completeness indicator
  driven by `get_missing_docs`.
- **Doc editor**: markdown edit + preview, per doc type.
- **Board**: kanban with dnd-kit; cards show priority, labels, linked-doc badge.

---

## 11. Build phases

1. **Domain core** — models + persistence interfaces + in-memory impls + services + the
   state machine, with unit tests on the lifecycle invariants.
2. **REST + UI** — chi API; React portfolio + project dashboard (docs library + board).
3. **MCP** — stdio server: tools first, then resources, then prompts.
4. **SQLite** — swap in-memory repos for SQLite-backed ones.
5. **Polish** — tags, filtering, search; optional per-project doc templates.

---

## 12. Deferred / open questions

- Multiple boards per project (schema allows it; UI deferred).
- Recurring/ongoing tasks for maintenance projects (start with a `recurring` label, not machinery).
- Document version history.
- Full-text search across docs and tasks.
- Auth + multi-user (only if it ever leaves the local machine).
- Remote HTTP/SSE MCP transport (the transport adapter is designed so this can drop in later).
