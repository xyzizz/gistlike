# GistLike

GistLike 是一个轻量的代码片段分享服务，支持创建、浏览、编辑、删除和搜索 snippet。项目使用 Go、Gin、HTML 模板和 SQLite 实现，适合个人项目、内部工具或单机部署场景。

## 功能概览

- 创建 snippet，支持标题、描述、语言、代码内容和公开性设置
- 浏览公开 snippet 列表，并按创建时间倒序展示
- 查看详情页与原始文本页
- 编辑和删除 snippet
- 按标题或描述搜索公开 snippet
- 提供页面路由和 JSON API
- 支持 Docker 单机部署

## 重要说明

- 所谓“私有 snippet”当前只是 `unlisted`，不是鉴权保护。只要知道链接，仍然可以访问。
- 当前存储使用 SQLite，适合单节点部署，不适合多节点水平扩展。

## 技术栈

- `Go + Gin`
- `html/template + 原生 JavaScript`
- `SQLite`，驱动为 `modernc.org/sqlite`
- `Repository -> Service -> Handler` 分层结构

## 目录结构

```text
.
├── cmd/server/               # 程序入口
├── internal/
│   ├── config/               # 配置
│   ├── handler/              # 页面与 API 路由处理
│   ├── model/                # 数据模型
│   ├── repository/           # 数据访问层
│   └── service/              # 业务层
├── migrations/               # 数据库初始化 SQL
├── web/
│   ├── static/               # CSS / JS 静态资源
│   └── templates/            # HTML 模板
├── deploy/caddy/             # Caddy 反向代理配置
├── scripts/                  # 部署脚本
├── Dockerfile
├── docker-compose.yml
└── README.md
```

## 数据结构

项目当前只有一张 `snippets` 表，核心字段如下：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `TEXT` | UUID 主键 |
| `title` | `TEXT` | 标题，必填 |
| `description` | `TEXT` | 描述，可空 |
| `language` | `TEXT` | 代码语言，必填 |
| `content` | `TEXT` | 代码内容，必填 |
| `is_public` | `INTEGER` | `1` 表示公开，`0` 表示 unlisted |
| `created_at` | `TEXT` | 创建时间，RFC3339 |
| `updated_at` | `TEXT` | 更新时间，RFC3339 |

## 路由说明

### 页面路由

- `GET /`：首页与公开 snippet 列表
- `GET /snippets/new`：新建页面
- `GET /snippets/:id`：详情页
- `GET /snippets/:id/raw`：原始文本页
- `GET /snippets/:id/edit`：编辑页

### API 路由

- `GET /api/snippets`：获取公开 snippet 列表
- `GET /api/snippets/:id`：获取单条 snippet
- `POST /api/snippets`：创建 snippet
- `PUT /api/snippets/:id`：更新 snippet
- `DELETE /api/snippets/:id`：删除 snippet

## 本地开发

### 依赖

- `Go 1.25+`

### 启动方式

```bash
go mod tidy
go run ./cmd/server
```

启动后可访问 [http://localhost:8080](http://localhost:8080)。

首次运行时，程序会自动在 `data/snippets.db` 创建 SQLite 数据库。

## Docker 运行

```bash
docker compose up --build
```

启动后可访问 [http://localhost:8080](http://localhost:8080)。

当前仓库内的 [`docker-compose.yml`](./docker-compose.yml) 适合单机部署，默认行为如下：

- 应用监听宿主机 `127.0.0.1:8080`
- 容器内启用 `GIN_MODE=release`
- SQLite 数据挂载到 `./data`
- 健康检查访问 `http://127.0.0.1:8080/healthz`

当前还提供了一个轻量健康检查接口：

- `GET /healthz`：仅用于存活探测，返回 `200 OK`

## 单机部署建议

如果你要把它部署到一台公网服务器，推荐按下面的方式做：

1. 应用本身只监听 `127.0.0.1:8080`
2. 用 `Caddy` 反代对外提供 `80/443`
3. 定期备份 `data/snippets.db`

推荐 `Caddy` 的原因很简单：

- 配置更少，适合单服务项目
- 域名解析正确时可以自动申请和续签 HTTPS 证书
- 比较适合当前这种单节点、单应用的部署方式

## 仓库内的部署文件

仓库已经包含当前部署所需的关键文件：

- [`deploy/caddy/Caddyfile`](./deploy/caddy/Caddyfile)
- [`deploy/caddy/docker-compose.yml`](./deploy/caddy/docker-compose.yml)
- [`scripts/deploy_hostdare.sh`](./scripts/deploy_hostdare.sh)

其中 [`scripts/deploy_hostdare.sh`](./scripts/deploy_hostdare.sh) 是当前 `hostdare` 服务器的一键重部署脚本。

## 一键重部署

```bash
./scripts/deploy_hostdare.sh
```

这个脚本会执行以下操作：

- 通过 `ssh` 和 `rsync` 把本地仓库同步到服务器 `/home/gistlike`
- 保留服务器上的 `data/` 数据目录
- 清理之前部署过程中遗留的旧目录
- 重新构建并启动应用容器
- 启动仓库内的 Caddy 容器配置

使用前提：

- 本机已经配置 `ssh hostdare`
- 本机安装了 `ssh` 和 `rsync`
- 服务器可正常使用 Docker 与 Docker Compose

## Cloudflare + Caddy 常见问题

如果域名走 Cloudflare 代理，源站由 Caddy 提供 HTTPS，最常见的问题有两类。

### 1. 浏览器报 `DNS_PROBE_FINISHED_NXDOMAIN`

这通常不是程序本身出错，而是 DNS 传播或本地 DNS 缓存还没刷新。

判断思路：

- 如果 Cloudflare 权威 DNS 或 `1.1.1.1` 已经能解析
- 但你的电脑或手机仍然提示 `NXDOMAIN`
- 那更可能是本地网络正在使用旧缓存

建议处理方式：

- 等待几分钟到几十分钟
- 切换网络后重试
- 临时改用 `1.1.1.1`
- 不要频繁删除并重建 DNS 记录

### 2. 浏览器报 “redirected you too many times”

这通常不是 DNS 死循环，而是 Cloudflare 的 SSL 模式与 Caddy 的强制 HTTPS 策略冲突。

典型原因：

- Cloudflare 使用了 `Flexible`
- Cloudflare 回源时走 `HTTP`
- Caddy 又把 `HTTP` 重定向到 `HTTPS`
- 最终形成无限跳转

正确做法：

- 在 Cloudflare 后台把 `SSL/TLS` 模式改为 `Full (strict)`
- 不要使用 `Flexible`

推荐配置：

- DNS 记录保持橙云代理
- `SSL/TLS` 使用 `Full (strict)`
- 源站继续由 Caddy 提供证书和反向代理

一句话总结：

- `NXDOMAIN` 更像 DNS 传播或本地缓存问题
- `Too many redirects` 更像 Cloudflare `Flexible` 与 Caddy 强制 HTTPS 冲突
