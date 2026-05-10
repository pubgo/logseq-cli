# Logseq CLI 项目分析文档

## 1. Logseq HTTP API 概述

### 1.1 通信协议

Logseq 桌面版内建了 HTTP API Server，通过 JSON-RPC 风格调用暴露插件 API 给外部应用。

| 项目         | 说明                            |
| ------------ | ------------------------------- |
| 基础 URL     | `http://127.0.0.1:12315/api`    |
| 请求方法     | `POST`                          |
| 认证方式     | `Authorization: Bearer <token>` |
| Content-Type | `application/json`              |

### 1.2 请求格式

```json
{
  "method": "logseq.Namespace.methodName",
  "args": [arg1, arg2, ...]
}
```

### 1.3 响应格式

成功时返回 JSON 结果（具体数据结构取决于方法），失败时返回：

```json
{
  "error": "MethodNotExist: method_name"
}
```

---

## 2. 已验证的 API 方法清单

### 2.1 Editor 命名空间 (`logseq.Editor.*`)

#### Page 相关

| 方法                        | 参数                            | 返回      | 说明                                          |
| --------------------------- | ------------------------------- | --------- | --------------------------------------------- |
| `getAllPages`               | `[]`                            | `[]Page`  | 获取所有页面列表                              |
| `getPage`                   | `[nameOrUUID]`                  | `Page`    | 按名称或 UUID 获取单页元数据                  |
| `createPage`                | `[name, properties?, options?]` | `Page`    | 创建页面。options: `{createFirstBlock: true}` |
| `deletePage`                | `[name]`                        | `null`    | 删除页面                                      |
| `renamePage`                | `[oldName, newName]`            | `null`    | 重命名页面并更新引用                          |
| `getPageBlocksTree`         | `[nameOrUUID]`                  | `[]Block` | 获取页面完整 Block 树                         |
| `getPageLinkedReferences`   | `[name]`                        | `[]`      | 获取页面反向链接                              |
| `getPagesFromNamespace`     | `[namespace]`                   | `[]Page`  | 获取命名空间下所有页面（平铺）                |
| `getPagesTreeFromNamespace` | `[namespace]`                   | `[]`      | 获取命名空间下页面树结构                      |
| `setPageProperties`         | `[name, properties]`            | —         | 设置页面级属性                                |

#### Block 相关

| 方法                  | 参数                              | 返回          | 说明                                                   |
| --------------------- | --------------------------------- | ------------- | ------------------------------------------------------ |
| `getBlock`            | `[uuid, {includeChildren: bool}]` | `Block`       | 获取单个 Block                                         |
| `insertBlock`         | `[targetUUID, content, options?]` | `Block`       | 插入 Block。options: `{sibling: bool, properties: {}}` |
| `insertBatchBlock`    | `[srcUUID, blocks[], {sibling}]`  | `null`        | 批量插入 Block 树（当前 SDK 暴露 `sibling bool` 参数） |
| `updateBlock`         | `[uuid, content]`                 | `Block\|null` | 更新 Block 内容                                        |
| `removeBlock`         | `[uuid]`                          | `null`        | 删除 Block                                             |
| `moveBlock`           | `[srcUUID, targetUUID, options?]` | —             | 移动 Block                                             |
| `prependBlockInPage`  | `[page, content]`                 | `Block`       | 在页面头部插入 Block                                   |
| `appendBlockInPage`   | `[page, content, options?]`       | `Block`       | 在页面尾部追加 Block                                   |
| `setBlockCollapsed`   | `[uuid, {flag: bool}]`            | `null`        | 设置 Block 折叠状态                                    |
| `upsertBlockProperty` | `[uuid, key, value]`              | `null`        | 设置 Block 属性                                        |
| `removeBlockProperty` | `[uuid, key]`                     | `null`        | 删除 Block 属性                                        |
| `getBlockProperties`  | `[uuid]`                          | `dict`        | 获取 Block 属性                                        |
| `getCurrentBlock`     | `[]`                              | `Block`       | 获取当前焦点 Block                                     |
| `getCurrentPage`      | `[]`                              | `Page`        | 获取当前焦点页面                                       |

### 2.2 App 命名空间 (`logseq.App.*`)

| 方法                | 参数      | 返回           | 说明                             |
| ------------------- | --------- | -------------- | -------------------------------- |
| `getCurrentGraph`   | `[]`      | `GraphInfo`    | 获取当前图谱信息                 |
| `getStateFromStore` | `[key]`   | `any`          | 获取应用状态                     |
| `getUserConfigs`    | `[]`      | `Config`       | 获取用户配置                     |
| `search`            | `[query]` | `SearchResult` | 全文搜索（⚠️ 部分版本可能不可用） |

### 2.3 DB 命名空间 (`logseq.DB.*`)

| 方法              | 参数                  | 返回              | 说明                              |
| ----------------- | --------------------- | ----------------- | --------------------------------- |
| `datascriptQuery` | `[query, ...inputs?]` | `json.RawMessage` | 执行 Datalog 查询（原始 JSON）    |
| `q`               | `[dslQuery]`          | `json.RawMessage` | 执行 Logseq DSL 查询（原始 JSON） |

### 2.4 已确认不可用的方法

| 方法                              | 说明                                                            |
| --------------------------------- | --------------------------------------------------------------- |
| `logseq.Editor.appendBlock`       | 返回 `MethodNotExist`                                           |
| `logseq.Editor.getPageProperties` | 返回 `MethodNotExist`（需用 getPageBlocksTree 的首 block 代替） |
| `logseq.Editor.createJournalPage` | 返回 `MethodNotExist`（需用 createPage + 日期格式名代替）       |

---

## 3. 数据模型

### 3.1 Page

```go
type Page struct {
    Name         string            `json:"name"`
    OriginalName string            `json:"originalName"`
    UUID         string            `json:"uuid"`
    Properties   map[string]any    `json:"properties,omitempty"`
    IsJournal    bool              `json:"journal?"`
    JournalDay   int               `json:"journalDay,omitempty"`
    CreatedAt    int64             `json:"createdAt,omitempty"`
    UpdatedAt    int64             `json:"updatedAt,omitempty"`
}
```

### 3.2 Block

```go
type Block struct {
    UUID       string            `json:"uuid"`
    Content    string            `json:"content"`
    Page       *PageRef          `json:"page,omitempty"`       // 可能是 int 或 {id: int}
    Properties map[string]any    `json:"properties,omitempty"`
    Children   []Block           `json:"children,omitempty"`
    Level      int               `json:"level,omitempty"`
    Format     string            `json:"format,omitempty"`     // "markdown" 或 "org"
    Marker     string            `json:"marker,omitempty"`     // "TODO","DOING","DONE" 等
    Priority   string            `json:"priority,omitempty"`
}

type PageRef struct {
    ID int64 `json:"id"`
}
```

### 3.3 GraphInfo

```go
type GraphInfo struct {
    Name string `json:"name"`
    Path string `json:"path"`
    URL  string `json:"url"`
}
```

---

## 4. Go SDK 设计方案

### 4.1 目录结构

```
logseq-cli/
├── go.mod
├── main.go                     # CLI 入口（根命令与全局参数）
├── cmd/
│   └── e2e/                    # 独立 E2E 可执行程序
├── cmds/                       # CLI 子命令定义
│   ├── cmds.go                 # 共享配置与客户端创建
│   ├── page.go                 # page list/get/create/delete/rename/refs/namespace/properties
│   ├── tag.go                  # tag list
│   ├── block.go                # block get/insert/update/remove/move/prepend/append/property/collapse
│   ├── graph.go                # graph/query/search 命令
│   └── webui.go                # webui 命令入口
├── internal/
│   └── webui/                  # WebUI 服务与静态页面
├── pkg/
│   └── logseq/                 # Logseq Go SDK
│       ├── client.go           # HTTP 客户端（底层 RPC 调用）
│       ├── types.go            # 数据模型定义
│       ├── editor.go           # Editor 命名空间 API
│       ├── app.go              # App 命名空间 API
│       ├── db.go               # DB 命名空间 API（Datalog/DSL 查询）
│       └── tags.go             # 标签聚合逻辑（含多路径回退）
└── docs/
    └── ANALYSIS.md             # 本文档
```

### 4.2 SDK Client 核心接口

```go
package logseq

import (
    "context"
    "net/http"
)

// Client 是 Logseq HTTP API 的 Go 封装
type Client struct {
    baseURL    string
    token      string
    httpClient *http.Client
}

// NewClient 创建 Logseq API 客户端
func NewClient(opts ...Option) *Client

// CallAPI 是底层 JSON-RPC 调用
func (c *Client) CallAPI(ctx context.Context, method string, args ...any) (json.RawMessage, error)
```

### 4.3 Editor API 封装

```go
// === Page 操作 ===
func (c *Client) GetAllPages(ctx context.Context) ([]Page, error)
func (c *Client) GetPage(ctx context.Context, nameOrUUID string) (*Page, error)
func (c *Client) CreatePage(ctx context.Context, name string, properties map[string]any, opts *CreatePageOptions) (*Page, error)
func (c *Client) DeletePage(ctx context.Context, name string) error
func (c *Client) RenamePage(ctx context.Context, oldName, newName string) error
func (c *Client) GetPageBlocksTree(ctx context.Context, nameOrUUID string) ([]Block, error)
func (c *Client) GetPageLinkedReferences(ctx context.Context, name string) (json.RawMessage, error)
func (c *Client) GetPagesFromNamespace(ctx context.Context, ns string) ([]Page, error)
func (c *Client) GetPagesTreeFromNamespace(ctx context.Context, namespace string) (json.RawMessage, error)
func (c *Client) SetPageProperties(ctx context.Context, name string, properties map[string]any) error

// === Block 操作 ===
func (c *Client) GetBlock(ctx context.Context, uuid string, includeChildren bool) (*Block, error)
func (c *Client) InsertBlock(ctx context.Context, targetUUID, content string, opts *InsertBlockOptions) (*Block, error)
func (c *Client) InsertBatchBlock(ctx context.Context, srcUUID string, blocks []BatchBlock, sibling bool) error
func (c *Client) UpdateBlock(ctx context.Context, uuid, content string) error
func (c *Client) RemoveBlock(ctx context.Context, uuid string) error
func (c *Client) MoveBlock(ctx context.Context, srcUUID, targetUUID string, opts map[string]any) error
func (c *Client) PrependBlockInPage(ctx context.Context, page, content string) (*Block, error)
func (c *Client) AppendBlockInPage(ctx context.Context, page, content string) (*Block, error)
func (c *Client) UpsertBlockProperty(ctx context.Context, uuid, key string, value any) error
func (c *Client) RemoveBlockProperty(ctx context.Context, uuid, key string) error
func (c *Client) GetBlockProperties(ctx context.Context, uuid string) (map[string]any, error)
func (c *Client) SetBlockCollapsed(ctx context.Context, uuid string, collapsed bool) error

// === 当前焦点 ===
func (c *Client) GetCurrentPage(ctx context.Context) (*Page, error)
func (c *Client) GetCurrentBlock(ctx context.Context) (*Block, error)
```

### 4.4 App API 封装

```go
func (c *Client) GetCurrentGraph(ctx context.Context) (*GraphInfo, error)
func (c *Client) GetUserConfigs(ctx context.Context) (map[string]any, error)
func (c *Client) Search(ctx context.Context, query string) (*SearchResult, error)
```

### 4.5 DB API 封装

```go
func (c *Client) DatascriptQuery(ctx context.Context, query string, inputs ...any) (json.RawMessage, error)
func (c *Client) DSLQuery(ctx context.Context, query string) (json.RawMessage, error)
```

---

## 5. CLI 命令设计（基于 redant 框架）

### 5.1 全局 Flags

| Flag              | 环境变量           | 默认值      | 说明                                                     |
| ----------------- | ------------------ | ----------- | -------------------------------------------------------- |
| `-t, --token`     | `LOGSEQ_API_TOKEN` | —           | API 认证 token（必须）                                   |
| `--host`          | `LOGSEQ_HOST`      | `127.0.0.1` | Logseq 主机地址                                          |
| `-p, --port`      | `LOGSEQ_PORT`      | `12315`     | Logseq 端口                                              |
| `--raw-envelope`  | —                  | `false`     | 输出结构化 NDJSON envelope                               |
| `--list-commands` | —                  | `false`     | 列出全部命令（含子命令）                                 |
| `--list-flags`    | —                  | `false`     | 列出全部参数                                             |
| `--list-format`   | —                  | `text`      | `--list-commands` / `--list-flags` 输出格式（text/json） |

### 5.2 命令树

```
logseq
├── page                        # 页面管理
│   ├── list                    # 列出所有页面
│   ├── get <name>              # 获取页面内容
│   ├── create <name>           # 创建页面（支持 --content / stdin）
│   ├── delete <name>           # 删除页面
│   ├── rename <old> <new>      # 重命名页面
│   ├── refs <name>             # 获取页面反向引用
│   ├── namespace <name>        # 命名空间页面查询（支持 --tree）
│   └── properties <name>       # 页面属性读写
├── block                       # Block 管理
│   ├── get <uuid>              # 获取 Block
│   ├── insert <uuid> <content> # 插入 Block
│   ├── update <uuid> <content> # 更新 Block
│   ├── remove <uuid>           # 删除 Block
│   ├── move <src> <target>     # 移动 Block（支持 --before）
│   ├── prepend <page> <content># 页面头部插入
│   ├── append <page> <content> # 页面尾部追加
│   ├── property                # Block 属性操作（get/set/remove）
│   └── collapse <uuid>         # Block 折叠/展开（--expand）
├── graph                       # 图谱信息
│   ├── info                    # 当前图谱信息
│   └── config                  # 用户配置
├── query                       # 查询
│   ├── datalog <query>         # Datascript/Datalog 查询
│   └── dsl <query>             # Logseq DSL 查询
├── search <query>              # 全文搜索
├── tag                         # 标签管理
│   └── list                    # 列出所有标签
├── completion <shell>          # shell 自动补全
├── doc                         # 交互式命令文档站
├── web                         # 可视化命令执行页面
├── webui                       # 简易 Logseq 操作台（页面/块/搜索/查询/过滤/连接诊断）
├── mcp                         # MCP 集成命令
│   ├── list                    # 列出 MCP 工具元数据
│   └── serve                   # 启动 MCP 服务
└── llms-txt                    # LLM 友好文档导出
```

### 5.3 redant 框架核心用法

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/pubgo/redant"
)

func main() {
    var token string
    var host string
    var port string

    root := redant.Command{
        Use:   "logseq",
        Short: "Logseq CLI - 命令行操作 Logseq 图谱",
        Options: redant.OptionSet{
            {
                Flag:        "token",
                Description: "Logseq API token",
                Shorthand:   "t",
                Envs:        []string{"LOGSEQ_API_TOKEN"},
                Required:    true,
                Value:       redant.StringOf(&token),
                Inherit:     true,
            },
            {
                Flag:        "host",
                Description: "Logseq host",
                Envs:        []string{"LOGSEQ_HOST"},
                Default:     "127.0.0.1",
                Value:       redant.StringOf(&host),
                Inherit:     true,
            },
            {
                Flag:        "port",
                Shorthand:   "p",
                Description: "Logseq port",
                Envs:        []string{"LOGSEQ_PORT"},
                Default:     "12315",
                Value:       redant.StringOf(&port),
                Inherit:     true,
            },
        },
        Children: []*redant.Command{
            pageCmd(),    // page 子命令
            blockCmd(),   // block 子命令
            graphCmd(),   // graph 子命令
            queryCmd(),   // query 子命令
            searchCmd(),  // search 命令
            tagCmd(),     // tag 子命令
            webuiCmd(),   // webui 命令
            llmstxtcmd.New(),
            doccmd.New(),
        },
    }

    webcmd.AddWebCommand(&root)
    mcpcmd.AddMCPCommand(&root)
    completioncmd.AddCompletionCommand(&root)

    if err := root.Invoke().WithOS().Run(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

---

## 6. redant 框架关键特性（用于本项目）

| 特性         | 说明                                                    |
| ------------ | ------------------------------------------------------- |
| 命令树       | `Children []*Command` 嵌套子命令                        |
| 选项声明     | `OptionSet` 支持 flag/env/default 三来源                |
| Handler      | `func(ctx, *Invocation) error` 标准签名                 |
| 中间件       | `Middleware: Chain(...)` 链式编排                       |
| MCP 集成     | `app mcp serve` 暴露命令为 MCP Tools                    |
| 结构化输出   | `ResponseHandler` / `ResponseStreamHandler` 支持 NDJSON |
| 位置参数     | `inv.Args` 获取位置参数                                 |
| Stdin/Stdout | `inv.Stdin` / `inv.Stdout` / `inv.Stderr`               |
| Web 控制台   | 可视化命令调用                                          |
| 文档生成     | `llms-txt` 生成 LLM 友好文档                            |

---

## 7. 关键实现注意事项

1. **Block 追加的 Workaround**: `logseq.Editor.appendBlock` 不存在，需通过 `getPageBlocksTree` + `insertBlock(lastBlockUUID, content, {sibling: true})` 实现。

2. **页面属性获取**: `logseq.Editor.getPageProperties` 在常见版本中不可用；CLI 当前通过 `getPage` 读取页面对象中的 `properties`，并通过 `setPageProperties` 写入。

3. **Journal 创建**: `createJournalPage` 不可用，需调用 `createPage("YYYY_MM_DD")` + 通过 `journalDay` 查找确认。

4. **Page 字段中的 `page` 引用**: Block 中的 `page` 字段可能是 `int` 或 `{id: int}` 两种格式，Go SDK 需做兼容解析。

5. **Search 方法**: `logseq.App.search` 在部分版本可用、部分不可用；必要时可用 Datascript 查询做降级搜索方案。

6. **标签聚合与过滤策略**: 某些图谱中 `:block/tags` 结果偏少，需回退 `:block/refs` 与页面属性聚合；WebUI 的标签过滤与搜索标签过滤已采用多路径匹配。

7. **Token 安全**: Token 应优先从环境变量读取，避免暴露在命令行参数中。

---

## 8. 参考项目对比

| 项目                     | 语言         | 定位       | 关键实现                               |
| ------------------------ | ------------ | ---------- | -------------------------------------- |
| logseq-cli (Python)      | Python/Typer | 命令行工具 | asyncio + httpx，NDJSON 输出，管道组合 |
| logseq-mcp-server (Rust) | Rust         | MCP 服务器 | reqwest，强类型数据模型，MCP 工具暴露  |
| mcp-logseq (Python)      | Python       | MCP 服务器 | requests 同步，Markdown 解析，批量写入 |
| **本项目 (Go)**          | Go/redant    | CLI + MCP  | net/http，redant 命令树，MCP 内建      |

---

## 9. 下一步计划

1. 增补与维护中文文档（README / 命令参考 / API 分析）
2. 增加集成测试与示例脚本（需可连接 Logseq 实例）
3. 补充错误码与常见故障排查说明
4. 评估 API 版本差异并提供能力矩阵
5. 持续完善 MCP 使用场景与示例配置
