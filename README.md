# GistLike

GistLike is a small MVP web service for creating, browsing, editing, deleting, and searching code snippets. It uses Go, Gin, HTML templates, SQLite, and a lightweight repository-service-handler structure so the app stays easy to extend.

## Features

- Create snippets with title, description, language, content, and public/private visibility
- Browse public snippets ordered by newest creation time
- View snippet detail pages with readable code blocks
- Open a raw text view for any snippet, similar to a Gist raw link
- Edit and delete snippets
- Search public snippets by title or description
- Friendly 404 and 500 pages
- JSON API for CRUD operations
- Docker and docker-compose support

## Tech choices

- **Gin** keeps routing and middleware simple without over-abstracting the app
- **database/sql + modernc SQLite driver** keeps the data layer light while preserving a clean path toward PostgreSQL later
- **Repository -> Service -> Handler** keeps HTTP, business logic, and persistence responsibilities separated
- **html/template + native JS** keeps the frontend simple, fast, and easy to evolve

## Project structure

```text
.
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── handler/
│   │   ├── api_handler.go
│   │   ├── page_handler.go
│   │   ├── router.go
│   │   └── view_data.go
│   ├── model/
│   │   └── snippet.go
│   ├── repository/
│   │   ├── errors.go
│   │   └── snippet_repository.go
│   └── service/
│       └── snippet_service.go
├── migrations/
│   └── 001_create_snippets.sql
├── web/
│   ├── static/
│   │   ├── css/
│   │   │   └── app.css
│   │   └── js/
│   │       ├── snippet-detail.js
│   │       └── snippet-form.js
│   └── templates/
│       ├── error.html
│       ├── index.html
│       ├── layout.html
│       ├── snippet_detail.html
│       └── snippet_form.html
├── .dockerignore
├── DEVLOG.md
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
```

## Data model

The app stores a single `snippets` table:

| Column | Type | Notes |
| --- | --- | --- |
| `id` | `TEXT` | UUID string primary key |
| `title` | `TEXT` | Required |
| `description` | `TEXT` | Optional, stored as empty string when omitted |
| `language` | `TEXT` | Required, constrained in service layer |
| `content` | `TEXT` | Required code body |
| `is_public` | `INTEGER` | `1` for public, `0` for private/unlisted |
| `created_at` | `TEXT` | RFC3339 timestamp |
| `updated_at` | `TEXT` | RFC3339 timestamp |

Notes:

- Private snippets are **unlisted**, not access-controlled. Direct links still work.
- The repository layer is interface-based, so replacing SQLite with PostgreSQL later mainly affects repository implementation and SQL placeholders.

## Routes

### Page routes

- `GET /` — public snippet list and search
- `GET /snippets/new` — create page
- `GET /snippets/:id` — detail page
- `GET /snippets/:id/raw` — raw text view for code content
- `GET /snippets/:id/edit` — edit page

### API routes

- `GET /api/snippets` — list public snippets
- `GET /api/snippets/:id` — get a snippet by id
- `POST /api/snippets` — create a snippet
- `PUT /api/snippets/:id` — update a snippet
- `DELETE /api/snippets/:id` — delete a snippet

## Local development

### Prerequisites

- Go 1.25+

### Run the app

```bash
go mod tidy
go run ./cmd/server
```

The service starts on [http://localhost:8080](http://localhost:8080).

On first run it creates the SQLite database automatically at `data/snippets.db`.

## Docker

### Build and run with Docker Compose

```bash
docker compose up --build
```

The app will be available at [http://localhost:8080](http://localhost:8080).

The SQLite database is persisted to `./data`.

## How to use

1. Open the home page and click **New snippet**
2. Fill in title, description, language, content, and visibility
3. Submit the form to create the snippet
4. Use the detail page to open the `Raw` view, edit, or delete it
5. Use the search box on the home page to search public snippets

## Future extensions

- Add authentication and ownership for real private access control
- Add version history per snippet
- Integrate syntax highlighting libraries
- Add better search and pagination
- Replace SQLite repository with a PostgreSQL repository
- Add permissions, rate limiting, and moderation workflows
