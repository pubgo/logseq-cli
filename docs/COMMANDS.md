# logseq-cli 命令参考（中文）

本文档提供 `logseq` 命令的中文速查与参数说明，内容基于当前代码与命令帮助输出整理。

## 根命令

- 命令：`logseq`
- 描述：Logseq CLI - command line tool for Logseq

---

## capabilities：LLM 能力探测

### `logseq capabilities get`

探测当前运行时对 LLM/MCP 关键链路的可用性，返回能力矩阵与建议提示。

返回内容包含（节选）：

- API：`datascript_query`、`dsl_query`、`search`、`tags_list`、`state_store`
- 图谱：`db_graph`、`:block/tags` 可用性、`:block/refs` 可用性
- 写策略：`LOGSEQ_LLM_WRITE_MODE`、删除确认策略、最大结果数

建议在 LLM 工作流开始时先调用一次此命令，用于选择执行路径与回退策略。

### 全局参数

| 参数              | 类型                | 默认值      | 环境变量           | 说明                                        |
| ----------------- | ------------------- | ----------- | ------------------ | ------------------------------------------- |
| `-t, --token`     | string              | 无（必填）  | `LOGSEQ_API_TOKEN` | Logseq API token                            |
| `--host`          | string              | `127.0.0.1` | `LOGSEQ_HOST`      | Logseq API host                             |
| `-p, --port`      | string              | `12315`     | `LOGSEQ_PORT`      | Logseq API port                             |
| `--raw-envelope`  | bool                | `false`     | -                  | 输出结构化 NDJSON envelope                  |
| `--list-commands` | bool                | `false`     | -                  | 列出全部命令（含子命令）                    |
| `--list-flags`    | bool                | `false`     | -                  | 列出全部参数                                |
| `--list-format`   | enum(`text`,`json`) | `text`      | -                  | `--list-commands` / `--list-flags` 输出格式 |
| `-h, --help`      | bool                | `false`     | -                  | 显示帮助                                    |

---

## page：页面管理

### `logseq page list`

列出所有页面。

### `logseq page current`

获取当前焦点页面。

### `logseq page current-tree`

获取当前焦点页面的块树。

### `logseq page get <name>`

获取页面信息。

参数：

- `name`（string，必填）：页面名称

选项：

- `-b, --blocks`（bool）：包含页面块树

### `logseq page get-context <name>`

获取面向 LLM 的页面上下文视图（带块树裁剪与统计信息）。

参数：

- `name`（string，必填）：页面名称

选项：

- `--max-blocks`（int，默认 `200`）：最多返回的块数（同时受 `LOGSEQ_LLM_MAX_RESULTS` 上限约束）
- `--max-depth`（int，默认 `6`）：块树最大深度
- `--include-properties`（bool，默认 `true`）：是否返回页面属性

返回字段（`data` 节选）：

- `page`：页面基础信息
- `properties`：页面属性（当 `include-properties=true` 时）
- `outline_blocks[]`：裁剪后的块树（`uuid/content/marker/priority/level/properties/children`）
- `limits`：本次生效的裁剪参数
- `stats`：`returned_blocks / clipped_by_depth / clipped_by_limit`

### `logseq page create <name>`

创建页面。

参数：

- `name`（string，必填）：页面名称

选项：

- `-c, --content`（string）：初始块内容；传 `-` 时从 stdin 读取

### `logseq page append-safe <name> <content>`

安全追加块（面向 LLM 的受控写入入口）。

参数：

- `name`（string，必填）：页面名称
- `content`（string，必填）：块内容（传 `-` 时从 stdin 读取）

选项：

- `--dry-run`（bool）：仅预览，不写入
- `--confirm`（bool）：在 `LOGSEQ_LLM_WRITE_MODE=confirm` 时必需
- `--idempotency-key`（string）：幂等键，避免重复写入

### `logseq page journal [date]`

按日期创建 Journal 页面。

参数：

- `date`（string，可选）：日期字符串，格式 `YYYY-MM-DD`；不传则使用今天

### `logseq page delete <name>`

删除页面。

参数：

- `name`（string，必填）：页面名称

### `logseq page rename <old-name> <new-name>`

重命名页面。

参数：

- `old-name`（string，必填）：旧名称
- `new-name`（string，必填）：新名称

### `logseq page refs <name>`

获取页面的反向引用（backlinks / linked references）。

参数：

- `name`（string，必填）：页面名称

### `logseq page namespace <name>`

列出某命名空间下的页面。

参数：

- `name`（string，必填）：命名空间前缀

选项：

- `--tree`（bool）：以树结构返回

### `logseq page properties <name> [key=value ...]`

读取或写入页面属性。

- 不传 `key=value`：返回当前页面属性
- 传入一个或多个 `key=value`：批量写入属性

参数：

- `name`（string，必填）：页面名称

---

## block：块管理

### `logseq block get <uuid>`

通过 UUID 获取块。

参数：

- `uuid`（string，必填）：块 UUID

选项：

- `-c, --children`（bool）：包含子块

### `logseq block current`

获取当前焦点块。

### `logseq block selected`

获取当前选中的块列表。

### `logseq block clear-selected`

清空当前选中块。

### `logseq block new-uuid`

生成新的块 UUID（`logseq.Editor.newBlockUUID`）。

### `logseq block prev-sibling <uuid>`

获取前一个同级块。

参数：

- `uuid`（string，必填）：块 UUID

### `logseq block next-sibling <uuid>`

获取后一个同级块。

参数：

- `uuid`（string，必填）：块 UUID

### `logseq block insert <target-uuid> <content>`

向目标块插入新块。

参数：

- `target-uuid`（string，必填）：目标块 UUID
- `content`（string，必填）：块内容（传 `-` 时从 stdin 读取）

选项：

- `-s, --sibling`（bool）：作为同级块插入（默认作为子块）

### `logseq block insert-batch <target-uuid> <blocks-json>`

批量插入块树。

参数：

- `target-uuid`（string，必填）：目标块 UUID
- `blocks-json`（string，必填）：JSON 数组（传 `-` 时从 stdin 读取）

选项：

- `-s, --sibling`（bool）：作为同级块插入（默认作为子块）

示例：

- `echo '[{"content":"- item1"},{"content":"- item2"}]' | logseq block insert-batch <target-uuid> -`

### `logseq block update <uuid> <content>`

更新块内容。

参数：

- `uuid`（string，必填）：块 UUID
- `content`（string，必填）：新内容（传 `-` 时从 stdin 读取）

### `logseq block remove <uuid>`

删除块。

参数：

- `uuid`（string，必填）：块 UUID

### `logseq block delete-safe <uuid>`

安全删除块（先预览影响，再确认执行）。

参数：

- `uuid`（string，必填）：块 UUID

选项：

- `--dry-run`（bool）：仅预览，不删除
- `--confirm`（bool）：危险删除确认（默认策略下必需）

### `logseq block move <src-uuid> <target-uuid>`

移动块。

参数：

- `src-uuid`（string，必填）：源块 UUID
- `target-uuid`（string，必填）：目标块 UUID

选项：

- `--before`（bool）：移动到目标块之前（默认是之后）

### `logseq block prepend <page> <content>`

在页面顶部插入块。

参数：

- `page`（string，必填）：页面名称
- `content`（string，必填）：块内容（传 `-` 时从 stdin 读取）

### `logseq block append <page> <content>`

在页面底部追加块。

参数：

- `page`（string，必填）：页面名称
- `content`（string，必填）：块内容（传 `-` 时从 stdin 读取）

### `logseq block property`

块属性操作分组。

- `logseq block property get <uuid>`：获取块全部属性
- `logseq block property set <uuid> <key> <value>`：设置块属性
- `logseq block property remove <uuid> <key>`：删除块属性

> `set` 的 `value` 会尝试按 JSON 解析；解析成功则按结构化值写入。

### `logseq block collapse <uuid>`

折叠/展开块。

参数：

- `uuid`（string，必填）：块 UUID

选项：

- `-e, --expand`（bool）：展开而不是折叠

---

## graph：图谱信息

### `logseq graph info`

获取当前图谱信息。

### `logseq graph config`

获取用户配置（`logseq.App.getUserConfigs`）。

### `logseq graph app-info`

获取应用信息（`logseq.App.getInfo`）。

### `logseq graph user-info`

获取用户信息（`logseq.App.getUserInfo`）。

### `logseq graph db-graph`

检查当前图谱是否为 DB Graph（`logseq.App.checkCurrentIsDbGraph`）。

### `logseq graph graph-config`

获取当前图谱配置（`logseq.App.getCurrentGraphConfigs`）。

### `logseq graph favorites`

获取当前图谱收藏（`logseq.App.getCurrentGraphFavorites`）。

### `logseq graph recent`

获取当前图谱最近访问项（`logseq.App.getCurrentGraphRecent`）。

### `logseq graph templates`

获取当前图谱模板集合（`logseq.App.getCurrentGraphTemplates`）。

### `logseq graph state <key>`

获取应用状态存储中的键值（`logseq.App.getStateFromStore`）。

参数：

- `key`（string，必填）：状态键名

### `logseq graph state-set <key> <value>`

设置应用状态存储中的键值（`logseq.App.setStateFromStore`）。

参数：

- `key`（string，必填）：状态键名
- `value`（string，必填）：状态值（优先按 JSON 字面量解析，失败则按字符串写入）

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

## search-notes：LLM 友好检索（分页）

### `logseq search-notes <query>`

面向 LLM 的结构化检索输出，支持分页与游标。

参数：

- `query`（string，必填）：搜索关键词

选项：

- `--tag`（string，可选）：标签过滤（文本匹配）
- `--limit`（int，默认 `20`，最大 `200`）：分页大小
- `--cursor`（string，可选）：游标（当前实现为 offset 字符串）
- `--include`（string，默认 `pages,blocks`）：返回类型，支持 `pages,blocks,files`

返回字段（节选）：

- `items[]`：统一条目（`type/title/snippet/page/uuid`）
- `next_cursor` / `has_more`
- `total`

---

## tag：标签管理

### `logseq tag list`

列出当前图谱中的全部标签。

### `logseq tag get <name-or-id>`

按标签名或实体 id 获取标签信息。

参数：

- `name-or-id`（string，必填）：标签名或数字 id

### `logseq tag search <name>`

按名称搜索标签（`logseq.Editor.getTagsByName`）。

参数：

- `name`（string，必填）：标签名关键词

### `logseq tag create <name>`

创建标签（`logseq.Editor.createTag`）。

参数：

- `name`（string，必填）：标签名

选项：

- `--uuid`（string）：自定义标签页面 UUID

### `logseq tag objects <name>`

获取标签对象块（`logseq.Editor.getTagObjects`）。

参数：

- `name`（string，必填）：标签名

### `logseq tag property add <tag-id> <property-id-or-name>`

给标签添加属性关系。

### `logseq tag property remove <tag-id> <property-id-or-name>`

移除标签属性关系。

### `logseq tag extends add <tag-id> <parent-tag-id-or-name>`

给标签添加父标签关系。

### `logseq tag extends remove <tag-id> <parent-tag-id-or-name>`

移除标签父标签关系。

### `logseq tag block add <block-id> <tag-id>`

给块添加标签。

### `logseq tag block remove <block-id> <tag-id>`

从块移除标签。

---

## property：属性 schema 管理

### `logseq property list`

列出全部属性实体（`logseq.Editor.getAllProperties`）。

### `logseq property get <key>`

按 key 获取属性实体（`logseq.Editor.getProperty`）。

参数：

- `key`（string，必填）：属性 key

### `logseq property upsert <key> [schema-json]`

创建或更新属性 schema（`logseq.Editor.upsertProperty`）。

参数：

- `key`（string，必填）：属性 key
- `schema-json`（string，可选）：JSON 对象，如 `'{"type":"number","cardinality":"one"}'`

### `logseq property remove <key>`

删除属性 schema（`logseq.Editor.removeProperty`）。

参数：

- `key`（string，必填）：属性 key

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
- `GET /api/search`：支持 `q`，并可用 `tag` 做附加过滤

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
- 自动化脚本建议使用默认 JSON 输出或 `--raw-envelope`。
- 若遇到 API 版本差异导致的方法不可用，先用小范围命令验证（如 `graph info`、`page list`）。
