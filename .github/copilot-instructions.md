# 项目指南

## 架构

Logseq HTTP API 的 Go CLI 工具。单模块，三层结构：

| 层级 | 路径 | 职责 |
|------|------|------|
| CLI 命令 | `cmds/` | 按领域一个文件（`page.go`、`block.go`、`tag.go` 等） |
| SDK 客户端 | `pkg/logseq/` | HTTP JSON-RPC 客户端、类型定义、editor/app/tags/db API |
| WebUI | `internal/webui/` | 内嵌 SPA（`go:embed`），handler 按领域拆分文件 |

CLI 框架使用 **`github.com/pubgo/redant`**（非 cobra）。命令返回 `*redant.Command` 树。大多数使用 `ResponseHandler: redant.Unary(...)` 返回 `*llmEnvelope` 实现结构化 JSON 输出。

## 构建与测试

```sh
go build ./...            # 编译
go test ./... -count=1    # 单元测试（不使用缓存）
golangci-lint run ./...   # lint（配置：.golangci.yml）
go install -v .           # 安装为 `logseq` 二进制
```

本地开发需要 `go.work` 引用 `../redant`。CI 会自动去除本地 replace 指令。

E2E 测试（`cmd/e2e/`）需要运行中的 Logseq 实例，不属于常规 `go test` 范围。

## 约定

### 命令结构

每个领域导出 `XxxCmd() *redant.Command`（公开），内部以私有函数定义子命令：

```go
func PageCmd() *redant.Command {
    return &redant.Command{
        Use: "page", Short: "...",
        Children: []*redant.Command{pageListCmd(), pageGetCmd(), ...},
    }
}
```

参数使用 `redant.ArgSet` / `redant.OptionSet`，不使用 cobra 风格的 flags。

### LLM 安全

写操作通过 `llm_safety.go` 中的 `ensureWriteAllowed()` 进行保护。相关环境变量：
- `LOGSEQ_LLM_WRITE_MODE`：`read-only`（默认）/ `confirm` / `direct`
- `LOGSEQ_LLM_REQUIRE_CONFIRM_FOR_DELETE`：`true`（默认）

### WebUI 处理器

Handler 文件按领域拆分：`handler_page.go`、`handler_block.go`、`handler_search.go` 等。核心基础设施（Server 结构体、路由、写入辅助函数）保留在 `server.go` 中。

### 测试

- 单元测试：与源码同包的 `_test.go` 文件
- 纯函数（解析、过滤、envelope 构造）必须有测试覆盖
- E2E 冒烟测试通过 `LOGSEQ_E2E=1` 环境变量控制

## 关键环境变量

| 变量 | 必需 | 默认值 | 用途 |
|------|------|--------|------|
| `LOGSEQ_API_TOKEN` | 是 | — | Logseq API Bearer 令牌 |
| `LOGSEQ_HOST` | 否 | `127.0.0.1` | API 主机地址 |
| `LOGSEQ_PORT` | 否 | `12315` | API 端口 |
