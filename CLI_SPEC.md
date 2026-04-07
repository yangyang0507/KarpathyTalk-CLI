# kt — KarpathyTalk CLI 开发说明

## 概述

`kt` 是 KarpathyTalk 的命令行读取客户端，使用 Go 编写，与服务端同仓库。
目标是让开发者和 LLM Agent 能够在终端里流畅浏览 KarpathyTalk 的内容。

---

## 目录结构

```
cmd/
  kt/
    main.go              ← 入口，解析顶层命令
internal/
  client/
    client.go            ← HTTP 客户端，封装所有 API 请求
    types.go             ← API 响应结构体（对应服务端 api.go）
  display/
    human.go             ← 人类友好的格式化输出
    json.go              ← 原始 JSON 透传输出
    markdown.go          ← Markdown 格式输出
```

> `internal/client` 与服务端完全解耦，只依赖公开 API，未来桌面/移动端可直接复用。

---

## 第一期功能

### 命令

#### `kt timeline`

浏览公共时间线（最新根帖子）。

```
kt timeline [flags]

Flags:
  --limit <n>      每页条数，默认 20，最大 100
  --before <id>    分页游标，加载此 ID 之前的内容
  --json           输出原始 JSON
  --markdown       输出 Markdown 原文
```

示例输出（默认）：
```
─────────────────────────────────────────────
 #42  karpathy  (Andrej)          2h ago  ♥ 12  ↺ 3  💬 5
─────────────────────────────────────────────
 ## Hello World
 This is my first post on KarpathyTalk...
─────────────────────────────────────────────
```

---

#### `kt user <username>`

查看用户资料和帖子列表。

```
kt user <username> [flags]

Flags:
  --replies        只看该用户的回复，默认只看根帖子
  --limit <n>      每页条数，默认 20
  --before <id>    分页游标
  --json           输出原始 JSON
  --markdown       输出用户资料 Markdown（含 YAML frontmatter，调用 /user/<u>/md）
```

---

#### `kt post <id>`

查看单篇帖子详情及其直接回复（两次 API 调用：帖子本身 + 直接回复列表）。

> 服务端目前未注册 `/api/posts/{id}/thread` 端点，嵌套回复需额外递归请求，
> 当前版本仅展示第一层直接回复。

```
kt post <id> [flags]

Flags:
  --limit <n>      回复每页条数，默认 20
  --json           输出原始 JSON
  --markdown       输出帖子 Markdown 原文（含 frontmatter，调用 /posts/<id>/md）
  --raw            输出帖子纯 Markdown 正文（无 frontmatter，调用 /posts/<id>/raw）
  --revision <n>   查看指定修订版本（配合 --markdown / --raw 使用）
```

#### `kt docs`

获取平台 API 文档的 Markdown 原文，便于 LLM Agent 了解平台能力。

```
kt docs
```

---

### 全局 Flag

```
  --host <url>     API 根地址，默认 https://karpathytalk.com
                   支持本地实例：--host http://localhost:8080
```

---

## 第二期功能

### 命令

#### `kt watch <username>`

轮询某用户，有新帖子时在终端打印通知。

```
kt watch <username> [flags]

Flags:
  --interval <s>   轮询间隔（秒），默认 60
```

#### `kt open <id>`

在系统默认浏览器中打开帖子页面。

```
kt open <id>
```

#### `kt export <username>`

将某用户所有帖子导出为本地 Markdown 文件。

```
kt export <username> [flags]

Flags:
  --out <dir>      输出目录，默认 ./<username>
```

---

## API 映射

服务端实际已注册的公开只读端点（`app.go` 路由表）：

| 命令 | API 端点 | 备注 |
|---|---|---|
| `kt timeline` | `GET /api/posts?has_parent=false` | `limit` / `before` 参数均支持 |
| `kt user <u>` | `GET /api/users/<u>` + `GET /api/posts?author=<u>&has_parent=false` | 两次调用；`has_parent=false` 过滤掉回复 |
| `kt user <u> --replies` | `GET /api/posts?author=<u>&has_parent=true` | |
| `kt user <u> --markdown` | `GET /user/<u>/md` | 返回含 YAML frontmatter 的用户资料 Markdown |
| `kt post <id>` | `GET /api/posts/<id>` + `GET /api/posts?parent_post_id=<id>` | 仅直接回复；`/api/posts/{id}/thread` 未注册 |
| `kt post <id> --markdown` | `GET /posts/<id>/md` | 支持 `?revision=N` |
| `kt post <id> --raw` | `GET /posts/<id>/raw` | 支持 `?revision=N` |
| `kt docs` | `GET /docs.md` | |
| `kt export <u>` | 遍历 `GET /api/posts?author=<u>&has_parent=false` + 逐条 `GET /posts/<id>/md` | |

---

## 数据结构（client/types.go）

直接对应服务端 `internal/app/api.go` 的 JSON 响应，主要结构：

```go
type User struct { ... }          // 对应 apiUser（GET /api/users/{username}）
type UserRef struct { ... }       // 对应 apiUserRef（嵌入在 Post 中）
type Post struct { ... }          // 对应 apiPost
type PostsResponse struct { ... } // 对应 apiPostsQueryResponse（GET /api/posts，无 User 字段）
type PostResponse struct { ... }  // 对应 apiPostResponse（GET /api/posts/{id}）
```

> **注意**：服务端的 `apiPostThreadResponse`（含 `Post` + `Replies` 的一体结构）
> 对应的路由 `/api/posts/{id}/thread` **尚未注册**，客户端不应依赖该结构；
> `kt post` 通过两次独立请求自行组合帖子与回复数据。
>
> 同理，`apiPostListResponse`（含嵌套 User 字段的帖子列表）来自未注册的
> `/api/users/{username}/posts` 端点，也不需要在 types.go 中声明。

---

## 开发顺序建议

1. `internal/client` — API 封装层（先写，先测）
2. `internal/display` — 格式化输出层
3. `cmd/kt/main.go` — 命令入口，组装以上两层
4. 第一期四个命令：`timeline` → `user` → `post` → `docs`
5. 第二期功能按需迭代

---

## 依赖

- Go 标准库（`net/http`、`encoding/json`、`flag` 或 `cobra`）
- 可选：`github.com/fatih/color` 用于终端着色
- 可选：`github.com/spf13/cobra` 用于更丰富的命令行解析

---

## 备注

- 所有 API 端点均为公开只读，无需认证
- `--host` flag 支持指向本地实例，便于开发调试
- `internal/client` 包不依赖任何 UI 层，保证可复用性
