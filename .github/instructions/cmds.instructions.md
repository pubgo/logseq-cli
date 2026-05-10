---
applyTo: "cmds/**"
---

# cmds 包编码指南

## 命令结构

每个领域一个文件，导出 `XxxCmd() *redant.Command`，子命令用私有函数：

```go
func PageCmd() *redant.Command {
    return &redant.Command{
        Use: "page", Short: "页面管理",
        Children: []*redant.Command{pageListCmd(), pageGetCmd()},
    }
}

func pageListCmd() *redant.Command {
    return &redant.Command{
        Use:   "list",
        Short: "列出所有页面",
        ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) ([]logseq.Page, error) {
            client := NewClient()
            return client.GetAllPages(ctx)
        }),
    }
}
```

小型子命令可直接在 `Children` 内联定义，不必抽独立函数。

## 参数定义

位置参数用 `ArgSet`，标志参数用 `OptionSet`（闭包外声明变量，指针绑定）：

```go
var sibling bool
return &redant.Command{
    Args: redant.ArgSet{
        {Name: "name", Required: true, Value: redant.StringOf(new(string)), Description: "页面名称"},
    },
    Options: redant.OptionSet{
        {Flag: "sibling", Shorthand: "s", Description: "作为兄弟节点插入", Value: redant.BoolOf(&sibling)},
    },
}
```

Value 构造器：`redant.StringOf(&s)`、`redant.BoolOf(&b)`。

## ResponseHandler 选择

- **主流**：`ResponseHandler: redant.Unary(...)` — 自动 JSON 序列化，返回结构体/slice/map/`*llmEnvelope` 均可
- **手动输出**：`Handler: func(ctx, inv) error` — 用于需要 `json.RawMessage` 或自定义格式的场景

## Client 获取

每个 handler 中调用 `client := NewClient()`，不复用全局实例。

## LLM Envelope（结构化输出）

需要结构化输出的命令返回 `*llmEnvelope`：

```go
start := time.Now()
// 业务逻辑...
return envelopeSuccess(start, data,
    withCapabilityUsed("logseq.Editor.getPage"),
    withHints("提示信息"),
), nil
```

错误时：

```go
return envelopeFailure(start, err, "BAD_REQUEST", "请提供页面名",
    withCapabilityUsed("logseq.Editor.getPage"),
), nil
```

错误码：`SAFETY_BLOCKED`、`BAD_REQUEST`、`RESOURCE_NOT_FOUND`、`CAPABILITY_UNAVAILABLE`、`TIMEOUT`、`UPSTREAM_ERROR`。空字符串由 `classifyEnvelopeErrorCode(err)` 自动分类。

Option 函数：`withCapabilityUsed()`、`withHints()`、`withPageMeta(cursor, hasMore)`、`withFallbackUsed()`。

## LLM 写操作安全

写操作必须调用 `ensureWriteAllowed`，失败时返回 `SAFETY_BLOCKED` envelope：

```go
if err := ensureWriteAllowed("page.append-safe", dryRun, confirm, false); err != nil {
    return envelopeFailure(start, err, "SAFETY_BLOCKED", "可先使用 --dry-run"), nil
}
```

- `action` 格式：`"domain.operation"`
- `dangerous=true` 用于删除操作
- 配合 `--dry-run` 和 `--confirm` OptionSet 标志

## 测试规范

- 文件：同包 `_test.go`（如 `llm_safety_test.go`）
- 风格：表驱动（table-driven）+ `t.Run` 子测试
- 范围：测试纯函数（解析、分类、构造、过滤），不测试需要 Logseq 实例的 handler
- 环境变量：用 `t.Setenv()` 隔离

## 常用 Import

```go
import (
    "context"
    "fmt"
    "time"
    "strings"
    "encoding/json"

    "github.com/pubgo/redant"
    "github.com/pubgo/logseq-cli/pkg/logseq"
)
```
