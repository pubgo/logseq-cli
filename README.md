# logseq-cli

一个用于操作 Logseq HTTP API 的命令行工具（CLI），支持页面、块、查询、搜索，并内置文档站、Web 可视化入口与 MCP 集成能力。

> 当前仓库主命令名为 `logseq`。

## 功能特性

- 页面管理：列出、获取、创建、删除、重命名页面
- 块管理：获取、插入、更新、删除、移动、页面头尾追加
- 查询能力：Datalog / Logseq DSL
- 图谱信息：查看当前 Graph 元数据
- 全文搜索：调用 `logseq.App.search`
- 开发者增强：
	- `doc`：启动交互式命令文档站
	- `web`：打开可视化命令执行页面
	- `webui`：启动简化 Logseq 操作页面（页面/块/搜索/查询 + 最近操作回放 + 连接信息诊断）
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

其中包含：

- 完整接入步骤
- Claude Desktop 配置示例
- 常见报错与排查

## 端到端集成测试（E2E）

项目提供**独立 E2E 可执行模块**：`cmd/e2e`，可直接二进制运行，不依赖 `go test`。

执行方式（需本地 Logseq 已开启 API）：

- 直接运行：`LOGSEQ_API_TOKEN='your-token' go run ./cmd/e2e`
- 编译后二进制运行：`go build -o ./bin/logseq-e2e ./cmd/e2e && LOGSEQ_API_TOKEN='your-token' ./bin/logseq-e2e`
- 指定已有 CLI 二进制：`go run ./cmd/e2e --cli-bin ./logseq --token your-token`

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

## 全局参数

| 参数              | 环境变量           | 默认值      | 说明                       |
| ----------------- | ------------------ | ----------- | -------------------------- |
| `-t, --token`     | `LOGSEQ_API_TOKEN` | 无（必填）  | Logseq API token           |
| `--host`          | `LOGSEQ_HOST`      | `127.0.0.1` | Logseq API 主机            |
| `-p, --port`      | `LOGSEQ_PORT`      | `12315`     | Logseq API 端口            |
| `-o, --output`    | -                  | `json`      | 输出格式：`json` / `text`  |
| `--raw-envelope`  | -                  | `false`     | 输出结构化 NDJSON envelope |
| `--list-commands` | -                  | `false`     | 列出全部命令（含子命令）   |
| `--list-flags`    | -                  | `false`     | 列出全部参数               |

## 命令总览

### 页面（page）

- `logseq page list`
- `logseq page get <name> [-b|--blocks]`
- `logseq page create <name>`
- `logseq page delete <name>`
- `logseq page rename <old-name> <new-name>`

### 块（block）

- `logseq block get <uuid> [-c|--children]`
- `logseq block insert <target-uuid> <content> [-s|--sibling]`
- `logseq block update <uuid> <content>`
- `logseq block remove <uuid>`
- `logseq block move <src-uuid> <target-uuid>`
- `logseq block prepend <page> <content>`
- `logseq block append <page> <content>`

### 图谱（graph）

- `logseq graph info`

### 查询（query）

- `logseq query datalog <query>`
- `logseq query dsl <query>`

### 搜索（search）

- `logseq search <query>`

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
- 获取页面及其块树：`logseq page get "Daily Notes" --blocks`
- 新建页面：`logseq page create "项目规划"`
- 在页面末尾追加块：`logseq block append "项目规划" "- [ ] 第一阶段完成"`
- 更新块内容：`logseq block update <uuid> "- [x] 第一阶段完成"`
- 执行 Datalog 查询：`logseq query datalog '[:find ?p :where [?b :block/name ?p]]'`
- 全文搜索：`logseq search "Go SDK"`

## 输出说明

- 默认输出格式为 `json`
- 可通过 `--output text` 切换为文本输出
- 对接自动化流程时建议使用默认 `json` 或 `--raw-envelope`

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
- `cmds/`：命令定义（page / block / graph / query / search）
- `pkg/logseq/`：Logseq API 客户端与数据类型
- `docs/`：分析文档与补充资料

## 参考文档

- 命令速查与详细参数说明：`docs/COMMANDS.md`
- API 与实现分析：`docs/ANALYSIS.md`
- LLM/MCP 接入指南：`docs/LLM_MCP.md`

## License

本项目使用 `LICENSE` 中声明的开源协议。
