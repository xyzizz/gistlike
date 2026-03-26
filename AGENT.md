# AGENT.md

This file is a quick orientation guide for future agents working in this repository.

## Project summary

`gistlike` is a small GitHub Gist-inspired snippet sharing service.

Current scope:
- create, view, edit, delete snippets
- search public snippets
- open a Gist-style raw text view
- render pages with `html/template`
- store data in SQLite with a repository layer that is easy to swap later

Important product behavior:
- `private` currently means **unlisted**, not access-controlled
- direct links to private snippets still work because there is no auth system yet

## Tech stack

- Go `1.25`
- Gin for HTTP routing
- `database/sql` + `modernc.org/sqlite`
- server-rendered HTML via `html/template`
- plain CSS + plain JavaScript
- Docker + docker-compose

## Repository map

- `cmd/server/main.go`
  App entrypoint, config load, DB bootstrap, router startup.

- `internal/config`
  Environment-based config.

- `internal/model`
  Core data structures like `Snippet`.

- `internal/repository`
  Persistence layer. SQLite implementation lives here.

- `internal/service`
  Business rules, validation, orchestration.

- `internal/handler`
  Gin handlers and router setup.

- `web/templates`
  Server-rendered pages.

- `web/static/css`
  Global styling.

- `web/static/js`
  Page-level behavior for form submit and delete confirmation.

- `migrations`
  Reference SQL for schema creation.

- `DEVLOG.md`
  Development log. Update this whenever you complete a meaningful implementation round.

## Request flow

Standard flow:

`router -> handler -> service -> repository -> SQLite`

Keep responsibilities separated:
- handlers handle HTTP concerns
- services own validation and product rules
- repositories own SQL and persistence

Do not move business logic into templates or `main.go`.

## Current routes

Page routes:
- `GET /`
- `HEAD /`
- `GET /snippets/new`
- `GET /snippets/:id`
- `GET /snippets/:id/raw`
- `GET /snippets/:id/edit`

API routes:
- `GET /api/snippets`
- `GET /api/snippets/:id`
- `POST /api/snippets`
- `PUT /api/snippets/:id`
- `DELETE /api/snippets/:id`

## Local run

Prerequisite:
- Go `1.25+`

Commands:

```bash
go mod tidy
go run ./cmd/server
```

Default app URL:
- `http://localhost:8080`

Useful override:

```bash
APP_ADDR=:8081 go run ./cmd/server
```

## Docker run

```bash
docker compose up --build
```

## Data and runtime files

- SQLite DB path defaults to `data/snippets.db`
- this is runtime state and should not be committed
- schema is bootstrapped automatically at startup

## Frontend notes

Current UI direction:
- restrained, developer-tool-like surface
- full-bleed dark hero on the home page
- code-first detail page
- editor page with a main form column plus utility sidebar

If you touch the UI:
- preserve strong hierarchy
- avoid generic card grids
- keep code panels dominant on detail pages
- keep interactions lightweight and native

## Raw snippet behavior

`GET /snippets/:id/raw` should return:
- `text/plain; charset=utf-8`
- `Content-Disposition: inline`
- a language-based file extension when possible

If you modify snippet language handling, keep raw filename behavior aligned.

## Validation expectations

For normal code changes, prefer this minimum verification:

```bash
go test ./...
```

For UI or routing changes, also do a quick local HTTP smoke test.

## Working agreements

- keep the project lightweight; do not introduce heavy frameworks
- preserve the repository/service/handler layering
- prefer small, readable changes over abstraction for its own sake
- if you complete a meaningful round of work, append a concise entry to `DEVLOG.md`
