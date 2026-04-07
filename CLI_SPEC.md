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
  --markdown       输出 Markdown 原文
```

---

#### `kt post <id>`

查看单篇帖子详情及其所有回复。

```
kt post <id> [flags]

Flags:
  --json           输出原始 JSON
  --markdown       输出帖子 Markdown 原文（含 frontmatter）
  --raw            输出帖子纯 Markdown 正文（无 frontmatter）
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

| 命令 | API 端点 |
|---|---|
| `kt timeline` | `GET /api/posts?has_parent=false` |
| `kt user <u>` | `GET /api/users/<u>` + `GET /api/posts?author=<u>` |
| `kt user <u> --replies` | `GET /api/posts?author=<u>&has_parent=true` |
| `kt post <id>` | `GET /api/posts/<id>` + `GET /api/posts?parent_post_id=<id>` |
| `kt post <id> --markdown` | `GET /posts/<id>/md` |
| `kt post <id> --raw` | `GET /posts/<id>/raw` |
| `kt export <u>` | 遍历 `GET /api/posts?author=<u>` + 分页 `before` |

---

## 数据结构（client/types.go）

直接对应服务端 `internal/app/api.go` 的 JSON 响应，主要结构：

```go
type User struct { ... }          // 对应 apiUser
type UserRef struct { ... }       // 对应 apiUserRef
type Post struct { ... }          // 对应 apiPost
type PostsResponse struct { ... } // 对应 apiPostsQueryResponse
type PostResponse struct { ... }  // 对应 apiPostResponse
type ThreadResponse struct { ... } // 对应 apiPostThreadResponse
```

---

## 开发顺序建议

1. `internal/client` — API 封装层（先写，先测）
2. `internal/display` — 格式化输出层
3. `cmd/kt/main.go` — 命令入口，组装以上两层
4. 第一期三个命令：`timeline` → `user` → `post`
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
