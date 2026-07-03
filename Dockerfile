# syntax=docker/dockerfile:1

# ---- builder: compile a static, pure-Go binary ----
FROM golang:1.25 AS builder
WORKDIR /src

# Download modules first so the layer caches across source-only changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# modernc.org/sqlite is pure Go, so CGO_ENABLED=0 yields a fully static binary
# that runs on a distroless/scratch base with no libc. -trimpath + -ldflags
# "-s -w" drop paths and debug info for a smaller, reproducible binary.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath -ldflags="-s -w" \
    -o /out/api ./cmd/api

# ---- datadir: bake a nonroot-owned /data for the SQLite db + WAL sidecars ----
# distroless has no shell to chown at runtime, so create the directory in a
# shelled stage and COPY it in with the right ownership. A fresh named volume
# mounted at /data inherits this mountpoint ownership on first creation, so
# uid 65532 can write the db (and its -wal/-shm sidecars) without extra flags.
FROM alpine:3.21 AS datadir
RUN mkdir -p /data && chown 65532:65532 /data

# ---- final: distroless static, nonroot ----
FROM gcr.io/distroless/static:nonroot
WORKDIR /app

COPY --from=builder /out/api /app/api
COPY --from=datadir --chown=65532:65532 /data /data

# Defaults; override via `docker run -e` / compose `environment:`. Real env vars
# always win over any baked .env (which .dockerignore excludes anyway).
ENV PORT=4523 \
    DB_PATH=/data/projectmanager.db \
    MCP_HTTP_ENABLED=true

EXPOSE 4523

# The SQLite db + WAL sidecars live here; mount a named volume at this directory
# (not a single-file mount) so the sidecars persist together.
VOLUME ["/data"]

# Exec form so the binary is PID 1 and receives SIGTERM directly for the
# graceful shutdown cmd/api already implements.
ENTRYPOINT ["/app/api"]
