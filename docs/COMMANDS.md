# logseq-cli 命令参考（中文）

本文档提供 `logseq` 命令的中文速查与参数说明，内容基于当前代码与命令帮助输出整理。

## 根命令

- 命令：`logseq`
- 描述：Logseq CLI - command line tool for Logseq

### 全局参数

| 参数              | 类型                | 默认值      | 环境变量           | 说明                       |
| ----------------- | ------------------- | ----------- | ------------------ | -------------------------- |
| `-t, --token`     | string              | 无（必填）  | `LOGSEQ_API_TOKEN` | Logseq API token           |
| `--host`          | string              | `127.0.0.1` | `LOGSEQ_HOST`      | Logseq API host            |
| `-p, --port`      | string              | `12315`     | `LOGSEQ_PORT`      | Logseq API port            |
| `-o, --output`    | enum(`json`,`text`) | `json`      | -                  | 输出格式                   |
| `--raw-envelope`  | bool                | `false`     | -                  | 输出结构化 NDJSON envelope |
| `--list-commands` | bool                | `false`     | -                  | 列出全部命令（含子命令）   |
| `--list-flags`    | bool                | `false`     | -                  | 列出全部参数               |
| `-h, --help`      | bool                | `false`     | -                  | 显示帮助                   |

---

## page：页面管理

### `logseq page list`

列出所有页面。

### `logseq page get <name>`

获取页面信息。

参数：

- `name`（string，必填）：页面名称

选项：

- `-b, --blocks`（bool）：包含页面块树

### `logseq page create <name>`

创建页面。

参数：

- `name`（string，必填）：页面名称

### `logseq page delete <name>`

删除页面。

参数：

- `name`（string，必填）：页面名称

### `logseq page rename <old-name> <new-name>`

重命名页面。

参数：

- `old-name`（string，必填）：旧名称
- `new-name`（string，必填）：新名称

---

## block：块管理

### `logseq block get <uuid>`

通过 UUID 获取块。

参数：

- `uuid`（string，必填）：块 UUID

选项：

- `-c, --children`（bool）：包含子块

### `logseq block insert <target-uuid> <content>`

向目标块插入新块。

参数：

- `target-uuid`（string，必填）：目标块 UUID
- `content`（string，必填）：块内容

选项：

- `-s, --sibling`（bool）：作为同级块插入（默认作为子块）

### `logseq block update <uuid> <content>`

更新块内容。

参数：

- `uuid`（string，必填）：块 UUID
- `content`（string，必填）：新内容

### `logseq block remove <uuid>`

删除块。

参数：

- `uuid`（string，必填）：块 UUID

### `logseq block move <src-uuid> <target-uuid>`

移动块。

参数：

- `src-uuid`（string，必填）：源块 UUID
- `target-uuid`（string，必填）：目标块 UUID

### `logseq block prepend <page> <content>`

在页面顶部插入块。

参数：

- `page`（string，必填）：页面名称
- `content`（string，必填）：块内容

### `logseq block append <page> <content>`

在页面底部追加块。

参数：

- `page`（string，必填）：页面名称
- `content`（string，必填）：块内容

---

## graph：图谱信息

### `logseq graph info`

获取当前图谱信息。

---

## query：查询

### `logseq query datalog <query>`

执行 Datalog 查询。

参数：

- `query`（string，必填）：Datalog 查询字符串

### `logseq query dsl <query>`

执行 Logseq DSL 查询。

参数：

- `query`（string，必填）：DSL 查询字符串

---

## search：全文搜索

### `logseq search <query>`

参数：

- `query`（string，必填）：搜索关键词

---

## tag：标签管理

### `logseq tag list`

列出当前图谱中的全部标签。

---

## completion：自动补全

### `logseq completion <shell>`

生成 shell 自动补全脚本。

参数：

- `shell`（必填）：`bash` / `zsh` / `fish`

---

## doc：交互式文档站

### `logseq doc`

启动交互式命令文档站。

选项：

- `--addr`（string，默认 `127.0.0.1:18081`）：监听地址
- `--open`（bool，默认 `true`）：启动后自动打开浏览器

---

## web：可视化命令执行页面

### `logseq web`

启动可视化命令执行页面。

选项：

- `--addr`（string，默认 `127.0.0.1:18080`）：监听地址
- `--open`（bool，默认 `true`）：启动后自动打开浏览器

---

## webui：简易 Logseq 操作台

### `logseq webui`

启动专注于 Logseq 数据验证的轻量页面，支持页面/块操作、搜索、Datalog/DSL 查询、历史回放与连接诊断。

选项：

- `--addr`（string，默认 `127.0.0.1:18090`）：监听地址
- `--open`（bool，默认 `true`）：启动后自动打开浏览器

页面内置过滤能力（由 webui 后端提供）：

- `GET /api/tags`：获取标签列表
- `GET /api/pages/filter`：按标签、页面名、元数据键值过滤页面
	- 查询参数：`tag`、`name`、`property`、`value`、`mode(contains|equals)`、`includeJournal(true|false)`

---

## mcp：MCP 集成

### `logseq mcp list`

列出 MCP 工具元数据。

选项：

- `--format`（enum: `json` / `text`，默认 `json`）

### `logseq mcp serve`

启动 MCP 服务。

选项：

- `--transport`（当前默认 `stdio`）

---

## llms-txt：LLM 文档导出

### `logseq llms-txt`

导出命令树文档，便于 LLM 消费。

选项：

- `-f, --format`：`markdown` / `json` / `skill`（默认 `markdown`）
- `-d, --depth`：命令树最大深度（`0` 表示不限制）
- `-o, --output-dir`：skill 模式输出目录

---

## 推荐用法

- 优先使用环境变量传递 token，避免在 shell 历史中泄露敏感信息。
- 自动化脚本建议使用默认 JSON 输出。
- 若遇到 API 版本差异导致的方法不可用，先用小范围命令验证（如 `graph info`、`page list`）。
