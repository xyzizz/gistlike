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

The bundled `docker-compose.yml` is intended for a simple single-node deployment:

- the app listens on `127.0.0.1:8080` on the server host
- `GIN_MODE=release` is enabled in the container
- the SQLite database is mounted from `./data`
- Docker health checks probe `http://127.0.0.1:8080/`

## Single-node deployment

For a small public deployment on one server, the simplest setup is:

1. Clone the repository onto the server
2. Start the app with `docker compose up -d --build`
3. Put a reverse proxy in front of `127.0.0.1:8080`
4. Keep `./data/snippets.db` backed up

### Reverse proxy recommendation

For this project, `Caddy` is the recommended reverse proxy over `Nginx`.

Why:

- `Caddy` can automatically provision and renew HTTPS certificates when your domain points to the server
- the config for a single upstream app is very small
- it reduces deployment steps for a one-service server

`Nginx` is still a good choice if you already run an `Nginx`-based stack or need more custom gateway behavior, but `Caddy` is the easier default here.

The repository also includes the current production-style Caddy deployment files:

- `deploy/caddy/Caddyfile`
- `deploy/caddy/docker-compose.yml`
- `scripts/deploy_hostdare.sh`

### Example Caddyfile

Replace `snippets.example.com` with your real domain:

```caddyfile
snippets.example.com {
	reverse_proxy 127.0.0.1:8080
}
```

### Example server flow

```bash
git clone https://github.com/xyzizz/gistlike.git
cd gistlike
docker compose up -d --build
```

After that:

- point your DNS record to the server IP
- install and start `Caddy`
- place the `Caddyfile` above in `/etc/caddy/Caddyfile`
- ensure ports `80` and `443` are open

### One-command redeploy

The repository includes a deployment helper for the current `hostdare` server:

```bash
./scripts/deploy_hostdare.sh
```

What it does:

- syncs the local repository to `/home/gistlike` over `ssh`
- preserves the server-side `data/` directory
- removes old deployment directories created during setup
- rebuilds and restarts the app container
- starts the Caddy reverse proxy from `deploy/caddy/`

Requirements:

- an `ssh` host alias named `hostdare`
- `rsync` available locally and on the server

### Operational notes

- This app is designed for one node and one SQLite database file
- SQLite is a good fit for a small single-server deployment, but not for multi-node horizontal scaling
- Private snippets are unlisted, not authenticated; do not treat this app as a secure private paste service without adding auth first

## Cloudflare + Caddy 排错（中文）

如果站点前面使用 Cloudflare 代理，源站使用 Caddy 自动处理 HTTPS，最常见的问题有两类：

### 1. 浏览器报 `DNS_PROBE_FINISHED_NXDOMAIN`

这通常不是源站程序故障，而是 DNS 传播或本地 DNS 缓存未刷新。

判断思路：

- 如果权威 DNS 或 `1.1.1.1` 已经能解析域名，但本机或手机仍然报 `NXDOMAIN`
- 说明问题更可能在本地网络使用的递归 DNS 缓存，而不是服务器本身

这类情况下：

- 先等待几分钟到几十分钟
- 可以切换网络，或临时改用 `1.1.1.1`
- 不要频繁删除再重建 DNS 记录，否则会拉长传播时间

### 2. 浏览器报 “redirected you too many times”

这通常不是 DNS 死循环，而是 Cloudflare SSL 模式和源站 HTTPS 策略冲突。

如果 Cloudflare 使用 `Flexible`，它会用 `HTTP` 回源到服务器；而 Caddy 默认会把 `HTTP` 重定向到 `HTTPS`。这样就会形成下面的循环：

1. 浏览器访问 `https://your-domain`
2. Cloudflare 用 `HTTP` 请求源站
3. Caddy 返回 `308`，要求跳转到 `https://your-domain`
4. Cloudflare 再把这个跳转返回给浏览器
5. 浏览器继续访问 `https://your-domain`
6. 重复以上过程，最终出现重定向过多

正确做法：

- 在 Cloudflare `SSL/TLS` 中把模式设置为 `Full (strict)`
- 不要使用 `Flexible`

推荐的 Cloudflare 配置：

- DNS 记录可以保持橙云代理
- SSL/TLS 模式使用 `Full (strict)`
- 源站继续由 Caddy 提供证书和反向代理

一句话总结：

- `NXDOMAIN` 更像 DNS 传播问题
- `Too many redirects` 更像 Cloudflare `Flexible` 和 Caddy 强制 HTTPS 之间的冲突

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
