# LLM 通过 MCP 操作 Logseq（实战指南）

本文说明如何把 `logseq-cli` 暴露给支持 MCP 的 LLM 客户端，让模型可直接执行页面/块/查询等操作。

如果你的目标是“稳定生产托管给 LLM”，请继续阅读：`docs/LLM_P0_API.md`（能力探测 / 安全写入 / 分页与统一 schema）。

## 1. 前提条件

- Logseq 桌面端已启动
- Logseq API 可用（默认 `127.0.0.1:12315`）
- 已拿到 token（`LOGSEQ_API_TOKEN`）
- 本项目可正常运行 `logseq` 命令

可先做连通性检查：

- `logseq --token <TOKEN> --host 127.0.0.1 --port 12315 graph info`

## 2. 启动 MCP Server

`logseq-cli` 已内置 MCP 子命令，服务模式为 `stdio`：

- `logseq --token <TOKEN> --host 127.0.0.1 --port 12315 mcp serve --transport stdio`

> 说明：MCP 客户端会接管 stdio，所以该命令通常不是你手工长期运行，而是由客户端按需拉起。

## 3. 推荐先构建本地二进制

避免客户端每次 `go run` 带来的编译开销，建议先构建：

- `go build -o ./bin/logseq .`

后续 MCP 客户端统一调用 `./bin/logseq`。

## 4. Claude Desktop 配置示例（macOS）

参考文件：`docs/examples/claude_desktop_logseq_mcp.json`

将 `mcpServers.logseq` 合并进 Claude 的配置（`claude_desktop_config.json`）：

- `command`：`/绝对路径/logseq-cli/bin/logseq`
- `args`：`["mcp", "serve", "--transport", "stdio"]`
- `env`：设置 `LOGSEQ_API_TOKEN`、`LOGSEQ_HOST`、`LOGSEQ_PORT`

## 5. 通用 MCP 客户端配置要点

大多数 MCP 客户端配置结构相似，通常是：

- `command`: 可执行文件路径
- `args`: 启动参数
- `env`: 环境变量（token/host/port）

可直接套用以下参数：

- command: `.../bin/logseq`
- args: `mcp serve --transport stdio`
- env:
  - `LOGSEQ_API_TOKEN=<你的token>`
  - `LOGSEQ_HOST=127.0.0.1`
  - `LOGSEQ_PORT=12315`

## 6. 给 LLM 的提示词建议

为了降低误操作，建议你在系统提示里明确：

- 会话启动先执行 `logseq capabilities get`，依据能力矩阵选择执行路径
- 优先读操作：`graph info`、`page get`、`query datalog`
- 写操作先确认目标页面
- 批量改写前先备份或在测试页验证
- 删除操作必须二次确认

## 7. 故障排查

### 7.1 `Post "http:///api": http: no Host in request URL`

通常是 `LOGSEQ_HOST/LOGSEQ_PORT` 未生效。请优先用**显式参数**验证：

- `logseq --token <TOKEN> --host 127.0.0.1 --port 12315 graph info`

确认可用后再写进 MCP 客户端的 `env`。

### 7.2 客户端能连上 MCP，但工具调用失败

- 确认 Logseq 桌面端仍在运行
- 确认 token 未过期/未变更
- 优先执行一次 `graph info` 作为健康检查

### 7.3 `mcp list` 构建异常（本地 replace 依赖场景）

当前仓库使用了本地 `replace github.com/pubgo/redant => /Users/.../redant`。若出现本地依赖编译异常：

- 先使用 `mcp serve` 验证服务可启动
- 再检查本地 `redant` 代码版本是否与当前仓库兼容

## 8. 最小接入验收（建议照抄）

当你把 MCP 配置接入客户端后，建议用下面 4 步做首轮验收：

1. 能力探测（只读）

- `logseq capabilities get`

2. 检索（分页）

- `logseq search-notes logseq --limit 5`

3. 页面上下文（裁剪）

- `logseq page get-context logseq --max-blocks 20 --max-depth 3 --include-properties false`

4. 安全写入预演（不落盘）

- `logseq page append-safe logseq "llm mcp dry-run check" --dry-run`

如果以上命令都返回统一 envelope（`ok/data/error/meta/hints`），说明“LLM 可稳定接入”的核心链路已打通。

## 9. 给 LLM 的系统提示词（可直接粘贴）

可把下面内容放到 MCP 客户端的系统提示中：

- 先执行 `logseq capabilities get`，再决定调用路径。
- 读操作优先：`search-notes`、`page get-context`、`query datalog`。
- 写操作必须先 `--dry-run`，确认后再执行真实写入。
- 删除类操作必须显式确认；默认优先 `block delete-safe --dry-run`。
- 分页读取时优先使用 `next_cursor/has_more`，避免一次请求过大。
- 若 `dsl_query` 不可用，优先回退到 datalog 或已封装命令。

推荐直接使用示例文件：

- 系统提示词：`docs/examples/llm_system_prompt_logseq_mcp.txt`
- 首轮用户提示模板：`docs/examples/llm_first_turn_template.txt`
