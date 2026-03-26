# Development Log

## Round 1 - Initial scaffold and MVP implementation

### Goal
- Bootstrap the project from an empty workspace into a runnable Gist-like MVP.

### Changes
- Created a Go module and entrypoint under `cmd/server`.
- Added `config`, `model`, `repository`, `service`, and `handler` layers under `internal`.
- Implemented SQLite-backed snippet CRUD with schema bootstrap on startup.
- Added page routes and RESTful API routes with Gin.
- Built server-rendered templates for list, create, edit, detail, and error pages.
- Added native CSS and JavaScript for form submission and delete confirmation.
- Added `Dockerfile`, `docker-compose.yml`, `.dockerignore`, `README.md`, and SQL migration reference.

### Key decisions
- Used `database/sql` plus `modernc.org/sqlite` to keep the app lightweight and reduce Docker complexity.
- Modeled `private` as unlisted visibility because the MVP intentionally does not include authentication.
- Used UUID snippet IDs so direct links are harder to guess than incremental numeric IDs.

### Validation
- Code scaffold completed.
- Formatting, dependency resolution, and runtime verification are queued for the next round.

## Round 2 - Formatting, dependency locking, and build validation

### Goal
- Normalize code style and verify that the generated project really builds.

### Changes
- Ran `gofmt` across all Go source files.
- Generated `go.sum` via `go mod tidy`.
- Enabled `HandleMethodNotAllowed` so custom 405 handling is active.
- Fixed the frontend form script to use `form.elements.namedItem(...)` instead of relying on property access that can clash with native form fields.

### Validation
- `go test ./...` passed successfully.
- During startup verification, a template composition issue was discovered and moved into the next round for correction.

## Round 3 - Runtime fixes and smoke testing

### Goal
- Resolve startup issues and verify the app responds over HTTP.

### Changes
- Reworked the layout template to select page partials via explicit conditionals instead of invalid dynamic template invocation.
- Added `HEAD /` so simple health-check style requests no longer return `405`.
- Updated Docker and README to match the resolved Go toolchain requirement.

### Validation
- Confirmed the server starts successfully.
- Confirmed `GET /` returns rendered HTML for the homepage.
- Ran an API smoke test:
  - created a snippet through `POST /api/snippets`
  - confirmed it appeared in `GET /api/snippets`
  - deleted the temporary validation snippet and confirmed the list was empty again

## Round 4 - Final regression pass

### Goal
- Re-run verification after the last routing and environment alignment changes.

### Changes
- Added an explicit `HEAD /` route for lightweight homepage checks.
- Aligned the Docker builder image and README prerequisites with the resolved Go 1.25 toolchain requirement.

### Validation
- Re-ran `go test ./...` successfully.
- Confirmed `HEAD /` now returns `200 OK`.

## Round 5 - Frontend redesign and layout repair

### Goal
- Replace the generic card-heavy UI with a cleaner, more deliberate product surface and fix layout issues that made pages feel crowded or misaligned.

### Analysis
- The home page hierarchy was weak: the hero looked like another card instead of the product's first impression.
- Listing, detail, and form pages were all using nearly the same boxed treatment, which flattened the information hierarchy.
- The code-reading surface did not stand out strongly enough from the rest of the UI.
- The editor layout had a structural issue where the title field could become too narrow on desktop because of the initial grid allocation.

### Changes
- Reworked `layout.html` so the header and page body have clearer global structure and page-specific styling hooks.
- Rebuilt the home page into a full-bleed dark hero plus a cleaner feed layout with timeline-style snippet rows instead of stacked cards.
- Reworked the editor page into a main working column plus a utility sidebar for visibility rules and supported languages.
- Reworked the detail page into a code-first layout with a dominant code panel and a quieter metadata rail.
- Rewrote the global CSS system with:
  - new typography hierarchy
  - cleaner spacing and dividers
  - reduced card treatment
  - stronger code-surface contrast
  - restrained entrance and hover motion
  - responsive layout rules for mobile
- Fixed the form grid so the title field spans the full row and no longer gets squeezed by the side controls.

### Validation
- Re-ran `go test ./...` successfully after the template changes.
- Ran the app on port `8081` for isolated validation because port `8080` was already in use by another process.
- Confirmed redesigned HTML responses for:
  - `/`
  - `/snippets/new`
  - `/snippets/1502edb9-e108-44b5-82c2-e7e6c66341d7`

## Round 6 - Raw snippet view

### Goal
- Add a Gist-like raw code view so each snippet can be opened directly as plain text.

### Changes
- Added a new page route: `GET /snippets/:id/raw`.
- Implemented a plain text response handler with:
  - `text/plain; charset=utf-8`
  - `Content-Disposition: inline`
  - language-based filename extensions such as `.go`, `.py`, `.js`, `.sql`, `.json`, and `.txt`
- Added `Raw` entry points to the snippet detail page so users can open or share the raw URL directly.
- Updated the README route list and usage notes.

### Validation
- Re-ran `go test ./...` successfully.
- Ran the app on port `8082` for isolated validation.
- Confirmed `GET /snippets/1502edb9-e108-44b5-82c2-e7e6c66341d7/raw` returns:
  - `200 OK`
  - `Content-Type: text/plain; charset=utf-8`
  - `Content-Disposition: inline; filename="aaa.py"`
  - the plain text code body instead of HTML

## Round 7 - Repository agent guide and Git hygiene

### Goal
- Add a durable repository guide for future agents and avoid committing runtime database files.

### Changes
- Added `AGENT.md` summarizing:
  - project purpose
  - stack and structure
  - request flow
  - key routes
  - runtime behavior
  - validation expectations
  - working agreements for future contributors and agents
- Added `.gitignore` to exclude runtime data such as `data/` and `*.db`.

### Validation
- Git working tree was reviewed before preparing the commit.
- No application code changed in this round.
