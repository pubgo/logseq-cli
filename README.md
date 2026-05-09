# logseq-cli

一个用于操作 Logseq HTTP API 的命令行工具（CLI），支持页面、块、查询、搜索，并内置文档站、Web 可视化入口与 MCP 集成能力。

> 当前仓库主命令名为 `logseq`。

## 功能特性

- 页面管理：列出、获取、创建（日记页）、删除、重命名、命名空间、页面属性、反向引用、当前焦点页
- 标签管理：列出、查询、创建标签，维护 tag-property / tag-extends / block-tag 关系
- 块管理：获取、当前选中块、清空选中、生成 UUID、前后兄弟块、插入、批量插入、更新、删除、移动、页面头尾追加、折叠/展开、属性读写、当前焦点块
- 查询能力：Datalog / Logseq DSL
- 属性管理：属性 schema 的查看、创建/更新、删除
- 图谱信息：查看 Graph 基础信息、用户配置、应用信息、用户信息、收藏/最近/模板、DB 模式与状态存储键值（读/写）
- 全文搜索：调用 `logseq.App.search`
- 开发者增强：
  - `doc`：启动交互式命令文档站
  - `web`：打开可视化命令执行页面
  - `webui`：启动简化 Logseq 操作页面（页面/块/搜索/查询 + 标签/属性 schema 管理 + Graph 状态读写 + 最近操作回放 + 连接信息诊断）
  - `mcp`：以 MCP 方式暴露命令树
  - `completion`：生成 shell 自动补全

## 环境要求

- Go `1.26.1` 或更高版本
- 已运行 Logseq 桌面应用
- 已开启 Logseq HTTP API，并获取 API Token

## 安装方式

### 方式一：源码运行（推荐开发调试）

在仓库根目录执行：

- `go run . --help`

### 方式二：安装为本地命令

- `go install ./...`

安装后可直接使用：

- `logseq --help`

## 快速开始

### 1) 配置环境变量

- `LOGSEQ_API_TOKEN`：必填，Logseq API token
- `LOGSEQ_HOST`：可选，默认 `127.0.0.1`
- `LOGSEQ_PORT`：可选，默认 `12315`

示例：

- `export LOGSEQ_API_TOKEN='your-token'`
- `export LOGSEQ_HOST='127.0.0.1'`
- `export LOGSEQ_PORT='12315'`

### 2) 验证连通性

- `logseq graph info`

如果返回当前图谱信息（名称、路径、URL），说明配置成功。

## LLM / MCP 集成

如果你希望让 LLM 直接操作 Logseq，可使用内置 MCP 服务：

- `logseq mcp serve --transport stdio`

推荐先阅读：`docs/LLM_MCP.md`

若你计划把本项目长期托管给 LLM（而不只是手工调用），建议继续阅读：`docs/LLM_P0_API.md`

其中包含：

- 面向 LLM 的 P0 工具契约（能力探测、安全写入、分页）
- 统一 JSON 响应信封与错误码建议
- 与现有命令树的映射和最小落地顺序

其中包含：

- 完整接入步骤
- Claude Desktop 配置示例
- 常见报错与排查

建议在 LLM 会话开始时先执行：

- `logseq capabilities get`

用于获取当前实例能力矩阵（DSL/Search/Tag 字段可用性等），让模型优先走正确路径并减少试错。

随后可优先使用：

- `logseq search-notes <query> --limit 20 --cursor <cursor>`

该命令提供 LLM 友好结构化结果与分页游标，避免一次性返回过大结果集。

写入场景建议使用：

- `logseq page append-safe <name> <content> --dry-run`
- `logseq block delete-safe <uuid> --dry-run`

并在确认后加 `--confirm` 执行实际写入/删除。

## 端到端集成测试（E2E）

项目提供**独立 E2E 可执行模块**：`cmd/e2e`，可直接二进制运行，不依赖 `go test`。

执行方式（需本地 Logseq 已开启 API）：

- 直接运行：`LOGSEQ_API_TOKEN='your-token' go run ./cmd/e2e`
- 编译后二进制运行：`go build -o ./bin/logseq-e2e ./cmd/e2e && LOGSEQ_API_TOKEN='your-token' ./bin/logseq-e2e`
- 指定已有 CLI 二进制：`go run ./cmd/e2e --cli-bin ./logseq --token your-token`
- 通过 `go test` 触发 smoke（默认跳过，需显式开启）：`LOGSEQ_E2E=1 LOGSEQ_API_TOKEN='your-token' go test ./cmd/e2e -run TestE2ESmoke -v`

常用参数：

- `--token`：API Token（默认读 `LOGSEQ_API_TOKEN`）
- `--host` / `--port`：默认 `127.0.0.1:12315`
- `--page-prefix`：临时页面名前缀
- `--keep-page`：保留测试页面用于排查
- `--timeout`：单条命令超时时间（默认 `60s`）

覆盖链路：

- `graph info` 连通性检查
- `page create`
- `block append/get/update/get/remove`
- `page delete`
- `query datalog` 清理校验

说明：

- 若未显式提供 `--cli-bin`，运行器会自动构建当前仓库 CLI 再执行
- 运行器会优先读取环境变量，也会自动加载项目根目录 `.env`
- 默认会自动清理临时页面，避免污染现有笔记
- `go test ./...` 默认不会执行该 e2e（`LOGSEQ_E2E` 未设置时会 skip），适合本地/CI 分层运行

### CI 快速检查（GitHub Actions）

仓库已提供 workflow：`.github/workflows/ci-quick.yml`，用于 PR / 主分支快速回归。

触发方式：

- `pull_request`
- `push` 到 `main` / `master`

执行内容：

- 标准 `go test ./...`
- 自动移除 `go.mod` 中本地绝对路径 `replace`（仅 CI 进程内生效）

### CI Smoke（GitHub Actions / self-hosted）

仓库已提供 workflow：`.github/workflows/e2e-smoke.yml`，用于定时或手动执行 e2e smoke。

触发方式：

- 手动：`workflow_dispatch`（可传入 `host` / `port` / `page_prefix` / `timeout` / `keep_page`）
- 定时：每天 UTC 03:00

前置条件：

- 使用 `self-hosted` runner（需要能访问本地 Logseq API）
- runner 上已安装 Go（workflow 会自动 setup）
- Logseq 桌面端已运行并开启 HTTP API
- runner 环境已配置 `LOGSEQ_API_TOKEN`

可选配置：

- `workflow_dispatch` 输入参数：
  - `host` / `port`：覆盖默认连接地址（默认 `127.0.0.1:12315`）
  - `page_prefix`：临时页面名前缀（默认 `__copilot_e2e_ci`）
  - `timeout`：单命令超时（默认 `60s`）
  - `keep_page`：是否保留测试页面用于排障（默认 `false`）
- runner 环境变量（可选）：`LOGSEQ_HOST` / `LOGSEQ_PORT` / `LOGSEQ_E2E_CLI_BIN`

## 全局参数

| 参数              | 环境变量           | 默认值      | 说明                                                           |
| ----------------- | ------------------ | ----------- | -------------------------------------------------------------- |
| `-t, --token`     | `LOGSEQ_API_TOKEN` | 无（必填）  | Logseq API token                                               |
| `--host`          | `LOGSEQ_HOST`      | `127.0.0.1` | Logseq API 主机                                                |
| `-p, --port`      | `LOGSEQ_PORT`      | `12315`     | Logseq API 端口                                                |
| `--raw-envelope`  | -                  | `false`     | 输出结构化 NDJSON envelope                                     |
| `--list-commands` | -                  | `false`     | 列出全部命令（含子命令）                                       |
| `--list-flags`    | -                  | `false`     | 列出全部参数                                                   |
| `--list-format`   | -                  | `text`      | `--list-commands` / `--list-flags` 输出格式（`text` / `json`） |

## 命令总览

### 页面（page）

- `logseq page list`
- `logseq page current`
- `logseq page current-tree`
- `logseq page get <name> [-b|--blocks]`
- `logseq page create <name> [--content <text|->]`
- `logseq page journal [date]`
- `logseq page delete <name>`
- `logseq page rename <old-name> <new-name>`
- `logseq page refs <name>`
- `logseq page namespace <name> [--tree]`
- `logseq page properties <name> [key=value ...]`

### 块（block）

- `logseq block get <uuid> [-c|--children]`
- `logseq block current`
- `logseq block selected`
- `logseq block clear-selected`
- `logseq block new-uuid`
- `logseq block prev-sibling <uuid>`
- `logseq block next-sibling <uuid>`
- `logseq block insert <target-uuid> <content|-> [-s|--sibling]`
- `logseq block insert-batch <target-uuid> <blocks-json|-> [-s|--sibling]`
- `logseq block update <uuid> <content|->`
- `logseq block remove <uuid>`
- `logseq block move <src-uuid> <target-uuid> [--before]`
- `logseq block prepend <page> <content|->`
- `logseq block append <page> <content|->`
- `logseq block property get <uuid>`
- `logseq block property set <uuid> <key> <value>`
- `logseq block property remove <uuid> <key>`
- `logseq block collapse <uuid> [--expand]`

### 图谱（graph）

- `logseq graph info`
- `logseq graph config`
- `logseq graph app-info`
- `logseq graph user-info`
- `logseq graph db-graph`
- `logseq graph graph-config`
- `logseq graph favorites`
- `logseq graph recent`
- `logseq graph templates`
- `logseq graph state <key>`
- `logseq graph state-set <key> <value>`

### 查询（query）

- `logseq query datalog <query>`
- `logseq query dsl <query>`

### 搜索（search）

- `logseq search <query>`

### 标签（tag）

- `logseq tag list`
- `logseq tag get <name-or-id>`
- `logseq tag search <name>`
- `logseq tag create <name> [--uuid <uuid>]`
- `logseq tag objects <name>`
- `logseq tag property add <tag-id> <property-id-or-name>`
- `logseq tag property remove <tag-id> <property-id-or-name>`
- `logseq tag extends add <tag-id> <parent-tag-id-or-name>`
- `logseq tag extends remove <tag-id> <parent-tag-id-or-name>`
- `logseq tag block add <block-id> <tag-id>`
- `logseq tag block remove <block-id> <tag-id>`

### 属性（property）

- `logseq property list`
- `logseq property get <key>`
- `logseq property upsert <key> [schema-json]`
- `logseq property remove <key>`

### 其他能力

- `logseq completion <bash|zsh|fish>`
- `logseq doc [--addr 127.0.0.1:18081] [--open true|false]`
- `logseq web [--addr 127.0.0.1:18080] [--open true|false]`
- `logseq webui [--addr 127.0.0.1:18090] [--open true|false]`
- `logseq mcp list [--format json|text]`
- `logseq mcp serve [--transport stdio]`
- `logseq llms-txt [-f markdown|json|skill] [-d <depth>] [-o <dir>]`

## 常用示例

- 查看所有页面：`logseq page list`
- 查看当前焦点页面：`logseq page current`
- 查看当前焦点页块树：`logseq page current-tree`
- 创建今天的 Journal 页面：`logseq page journal`
- 获取页面及其块树：`logseq page get "Daily Notes" --blocks`
- 新建页面并从 stdin 读取首块内容：`echo "- [ ] todo" | logseq page create "项目规划" --content -`
- 获取当前选中块：`logseq block selected`
- 在页面末尾追加块：`logseq block append "项目规划" "- [ ] 第一阶段完成"`
- 查看块的前后兄弟：`logseq block prev-sibling <uuid>` / `logseq block next-sibling <uuid>`
- 批量插入块（JSON）：`echo '[{"content":"- item1"},{"content":"- item2"}]' | logseq block insert-batch <target-uuid> -`
- 更新块内容（stdin）：`echo "- [x] 第一阶段完成" | logseq block update <uuid> -`
- 设置块属性：`logseq block property set <uuid> priority A`
- 读取状态存储：`logseq graph state 'ui/sidebar-open?'`
- 写入状态存储：`logseq graph state-set 'ui/sidebar-open?' true`
- 查看当前图谱收藏：`logseq graph favorites`
- 执行 Datalog 查询：`logseq query datalog '[:find ?p :where [?b :block/name ?p]]'`
- 全文搜索：`logseq search "Go SDK"`
- 查看全部标签：`logseq tag list`
- 新建标签：`logseq tag create project --uuid <uuid>`
- 关联标签与属性：`logseq tag property add <tag-id> priority`
- 写入属性 schema：`logseq property upsert priority '{"type":"default","cardinality":"one","public":true}'`
- 启动 webui：`logseq webui --addr 127.0.0.1:18090 --open true`

## WebUI 过滤能力（标签 + 元数据）

`webui` 页面左侧支持组合过滤，便于快速定位页面：

- 标签过滤（来自 `/api/tags`）
- 页面名模糊匹配（`name`）
- 元数据键值过滤（`property` + `value`）
- 匹配模式：`contains` / `equals`
- 是否包含 Journal 页面（`includeJournal=true|false`）

对应后端接口：

- `GET /api/tags`
- `GET /api/pages/filter?tag=&name=&property=&value=&mode=contains&includeJournal=true`
- `GET /api/search?q=<kw>&tag=<optional-tag>`（搜索结果按标签可选过滤）

## WebUI 增强能力（新增）

当前 `webui` 还支持以下增强操作：

- 页面：创建 Journal 页面（可指定日期）
- 块：当前焦点块、当前选中块、清空选中、生成 UUID、查询前后兄弟块
- 图谱：`app-info/user-info/current-config/favorites/recent/templates`，以及 `state` 读写
- 标签：按名称/ID 查询、搜索、创建标签，维护 `tag-property` / `tag-extends` / `block-tag` 关系
- 属性 schema：`list/get/upsert/remove`
- 快捷预设：Graph `state key` 常用键一键填充、Property schema 模板一键填充
- 一键验证：串行执行轻量 smoke（health/graph/config/tags/property-list）快速检查联通性

对应后端接口（节选）：

- `POST /api/page/journal`
- `GET /api/block/current`
- `GET /api/block/selected`
- `POST /api/block/selected/clear`
- `GET /api/block/new-uuid`
- `GET /api/block/prev-sibling?uuid=`
- `GET /api/block/next-sibling?uuid=`
- `GET /api/graph/app-info`
- `GET /api/graph/user-info`
- `GET /api/graph/current-config`
- `GET /api/graph/favorites`
- `GET /api/graph/recent`
- `GET /api/graph/templates`
- `GET /api/graph/state?key=`
- `POST /api/graph/state`
- `GET /api/tag?nameOrID=`
- `GET /api/tag/search?name=`
- `POST /api/tag/create`
- `GET /api/tag/objects?name=`
- `POST /api/tag/property`
- `POST /api/tag/extends`
- `POST /api/tag/block`
- `GET /api/property/list`
- `GET /api/property?key=`
- `POST /api/property/upsert`
- `POST /api/property/remove`

## 输出说明

- 默认输出为 JSON（便于自动化处理）
- 对接流水线建议使用默认 JSON 或 `--raw-envelope`

## 常见问题排查

### 提示缺少 token

请确认已设置 `LOGSEQ_API_TOKEN`，或在命令中显式传入 `--token`。

### 无法连接到 `127.0.0.1:12315`

- 确认 Logseq 正在运行
- 确认 HTTP API 服务已开启
- 确认 `LOGSEQ_HOST` / `LOGSEQ_PORT` 与 Logseq 配置一致

### 某些 API 返回 `MethodNotExist`

不同 Logseq 版本的 API 可用性存在差异。建议：

- 升级 Logseq 到较新稳定版
- 优先使用本项目已封装并在代码中使用的方法
- 对搜索等能力在你的版本上先做小样本验证

## 项目结构

- `main.go`：CLI 入口与全局参数
- `cmds/`：命令定义（page / block / graph / query / search / tag / webui）
- `pkg/logseq/`：Logseq API 客户端与数据类型
- `internal/webui/`：WebUI 服务端与静态前端
- `docs/`：分析文档与补充资料

## 参考文档

- 命令速查与详细参数说明：`docs/COMMANDS.md`
- API 与实现分析：`docs/ANALYSIS.md`
- LLM/MCP 接入指南：`docs/LLM_MCP.md`

## License

本项目使用 `LICENSE` 中声明的开源协议。
