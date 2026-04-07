# kt — KarpathyTalk 命令行工具

[KarpathyTalk](https://karpathytalk.com) 的只读命令行客户端，由 Andrej Karpathy 创建的社交平台。可以在终端中浏览时间线、阅读帖子、查看用户主页。

## 功能特性

- **交互式 TUI** — 在终端中通过键盘快捷键浏览帖子
- **三种输出模式** — 人性化显示、JSON、Markdown，适用于脚本和 LLM 管道
- **Unix 可组合性** — 将 JSON 或 Markdown 输出管道传递给任意工具
- **分页支持** — 基于游标的分页，可浏览大量结果
- **响应式布局** — 自动适应终端宽度

## 安装

从 [releases 页面](https://github.com/dy/KarpathyTalk-CLI/releases) 下载对应平台的预编译二进制文件，放入 `$PATH` 即可使用。

**或从源码构建**（需要 Go 1.21+）：

```bash
git clone https://github.com/dy/KarpathyTalk-CLI.git
cd KarpathyTalk-CLI

make build          # → dist/kt  （当前平台）
make install        # → $GOPATH/bin/kt
make release        # → dist/kt-<os>-<arch>（全平台交叉编译）
```

## 使用方法

```
kt [--host <url>] <命令> [参数]
```

### 命令

#### `kt timeline`

浏览公开时间线（仅根帖子，不含回复）。

```bash
kt timeline
kt timeline --limit 50
kt timeline --before 230       # 加载 ID 230 之前的帖子
kt timeline --json             # 输出原始 JSON
kt timeline --markdown         # 输出 Markdown
```

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--limit <n>` | 20 | 每页帖子数（最大 100） |
| `--before <id>` | — | 分页游标 |
| `--json` | — | 输出原始 JSON |
| `--markdown` | — | 将帖子内容输出为 Markdown |

---

#### `kt user <用户名>`

查看用户主页及其帖子。

```bash
kt user karpathy
kt user karpathy --replies     # 只显示回复，而非根帖子
kt user karpathy --json
kt user karpathy --markdown    # 带 YAML frontmatter 的主页 Markdown
```

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--replies` | — | 显示回复而非根帖子 |
| `--limit <n>` | 20 | 每页帖子数 |
| `--before <id>` | — | 分页游标 |
| `--json` | — | 输出原始 JSON |
| `--markdown` | — | 输出用户主页 Markdown |

---

#### `kt post <id>`

查看单篇帖子及其直接回复。

```bash
kt post 231
kt post 231 --json
kt post 231 --markdown         # 带 YAML frontmatter 的帖子
kt post 231 --raw              # 原始 Markdown，不含 frontmatter
kt post 231 --revision 2       # 查看指定版本
```

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--limit <n>` | 20 | 每页回复数 |
| `--json` | — | 将帖子和回复输出为 JSON |
| `--markdown` | — | 带 YAML frontmatter 的帖子 Markdown |
| `--raw` | — | 原始帖子 Markdown（无 frontmatter） |
| `--revision <n>` | — | 查看指定版本（与 `--markdown` / `--raw` 配合使用） |

---

#### `kt docs`

以 Markdown 格式获取 KarpathyTalk API 文档，适合作为 LLM 的上下文输入。

```bash
kt docs
kt docs | llm "总结这个 API"
```

---

### 全局参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--host <url>` | `https://karpathytalk.com` | API 根 URL |

可用于指向本地实例：

```bash
kt --host http://localhost:8080 timeline
```

## TUI 键盘快捷键

在终端（TTY）中运行时，`kt timeline`、`kt user`、`kt post` 会启动交互式 TUI。

**列表视图：**

| 按键 | 操作 |
|------|------|
| `j` / `↓` | 向下移动 |
| `k` / `↑` | 向上移动 |
| `Enter` | 打开帖子 |
| `q` / `Ctrl+C` | 退出 |

**详情视图：**

| 按键 | 操作 |
|------|------|
| `j` / `↓` | 向下滚动 |
| `k` / `↑` | 向上滚动 |
| `Esc` / `q` | 返回列表 |

滚动到列表末尾附近时，会自动加载更多帖子。

## 输出模式

| 模式 | 触发条件 | 适用场景 |
|------|----------|----------|
| TUI | stdout 是 TTY 且未指定格式参数 | 交互式浏览 |
| 人类可读 | stdout 为管道且未指定格式参数 | 可读纯文本输出 |
| `--json` | 显式指定 | 脚本、`jq`、结构化数据 |
| `--markdown` | 显式指定 | LLM 管道、导出、归档 |

所有错误信息输出到 stderr。JSON 和 Markdown 只写入 stdout，可安全用于管道。

## 使用示例

```bash
# 交互式阅读最新帖子
kt timeline

# 将近期帖子导出为 Markdown 文件
kt timeline --markdown > posts.md

# 将帖子内容传给 LLM
kt post 231 --raw | llm "总结这篇帖子"

# 获取结构化数据
kt user karpathy --json | jq '.posts[].like_count'

# 通过 pager 阅读
kt timeline | less
```

## 设计理念

- **只读设计** — 无需认证，不支持写操作。
- **一切皆帖子** — 回复、引用和时间线均为带不同过滤条件的帖子。
- **内容可移植** — 服务器在 Markdown 中将相对 URL 展开为绝对 URL。
- **Unix 可组合性** — 每种输出模式均可管道传输；JSON 和 Markdown 输出纯净无杂质。
- **依赖极简** — 基于 Charmbracelet 生态（Bubble Tea、Glamour、Lipgloss）和 Go 标准库构建。
