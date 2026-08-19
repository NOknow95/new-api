---
name: run-local
description: >-
  Code-change rebuild + restart the local docker stack in .local-run/ (image
  new-api:local + redis + postgres, port 3000). Use when the user asks to run,
  restart, rebuild, or re-run the local new-api docker stack, or after changing
  backend (Go) / frontend (web/) code and wanting the containerized app updated.
  Covers building the local image from the root Dockerfile, recreating the
  container, checking status/logs, and understanding that postgres data survives
  rebuilds in the name-volume. Load and follow this skill before rebuilding or
  restarting the local stack.
---

# Local Run Stack (.local-run)

## What this is

`.local-run/docker-compose.yml` runs the project locally in docker using the **local image `new-api:local`** (NOT the official `calciumion/new-api`) plus:

- `new-api` service — port **3000** (web dashboards API + OpenAI-compatible relay at `/v1`)
- `postgres` (port 5429 on the host, internal 5432) — source of truth DB `new-api`
- `redis` — internal only (port 6379), requirepass `123456`

Mounts: `./data:/data`, `./logs:/app/logs`. `SQL_DSN=postgresql://root:123456@postgres:5432/new-api`.

## Rebuilding the image after code changes (the core flow)

The root `Dockerfile` is a **full multi-stage build**: it runs `bun install` + `bun run build` (frontend) AND `go build` (backend), embedding `web/dist` via `//go:embed`. So **one build covers BOTH Go and web/ changes** — do NOT run `make build-web` first.

```bash
# from the project root (/Users/noknow/develop/projects/go/new-api)
docker build -t new-api:local .
```

Notes:

- The build reuses layers: if `web/package.json`+`web/bun.lock` and `go.mod`+`go.sum` are unchanged, deps cache and only the code diffs recompile — normally fast.
- If the build is stale or you changed the base images, force a clean pull/no-cache build:
  `docker build --pull --no-cache -t new-api:local .`
- Version string is baked from `VERSION` via ldflags; bump `VERSION` if you need the version to change.

## Restart the stack with the freshly built image

```bash
cd /Users/noknow/develop/projects/go/new-api
docker compose -f .local-run/docker-compose.yml up -d --force-recreate new-api
```

- `--force-recreate` guarantees a container is created from the new image (a plain `up -d` may skip recreation if it thinks the config is unchanged).
- `--wait` waits until the service is healthy: `docker compose -f .local-run/docker-compose.yml up -d --force-recreate --wait new-api`
- Health check hits `http://localhost:3000/api/status` and greps for `"success": true`.

## Data survives rebuilds

The database lives in the **named volume `new-api_pg_data`** (plus bound `./data`, `./logs`), NOT inside the image. Rebuilding the image / recreating the container does **NOT** wipe data — no need to re-run setup or re-add channels after a rebuild. Only `docker compose down -v` removes the postgres volume (use only when a full reset is intended).

## Everyday commands

```bash
# Start (does not rebuild)
docker compose -f .local-run/docker-compose.yml up -d
# Current status
docker compose -f .local-run/docker-compose.yml ps
# Logs (basic) / follow
docker compose -f .local-run/docker-compose.yml logs --tail=100 new-api
docker compose -f .local-run/docker-compose.yml logs -f --tail=100 new-api
# Stop containers (keeps volumes)
docker compose -f .local-run/docker-compose.yml down
# Stop and wipe postgres data (FULL reset — destroys all DB data)
docker compose -f .local-run/docker-compose.yml down -v
```

## Verify it's up

```bash
curl -s http://localhost:3000/api/status      # expect {"success":true,...}
```

## Decision guardrail

- User asked to "rebuild / restart / re-run / apply my code change locally" → run the rebuild + recreate flow above.
- User only asked to see status or logs, or start/stop the existing stack → just run the relevant everyday command, do NOT rebuild.
- If the image `new-api:local` does not exist yet (first run), build it first (the flow above), then `up -d`.