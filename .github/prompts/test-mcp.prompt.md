---
description: "运行 Logseq MCP 回归测试（委托专用 Agent 执行）"
agent: "Logseq MCP Regression"
---

# 测试 Logseq MCP

使用专用 Agent `Logseq MCP Regression` 执行完整回归。

## 调用参数建议

- scope: `full`（默认，执行 Phase A/B/C）或 `readonly`（仅 Phase A）
- baseline: 可选，上一次结果摘要或日期（用于对比）
- mixed-write: `auto` / `on` / `off`

## 期望输出

1. 当前轮次总表（✅/⚠️/❌）
2. Capabilities 门禁快照与跳过项
3. 可用性探测（available/unavailable）
4. 写入链路通过率与残留对象
5. 与上次结果对比（新增失败 / 已修复 / 状态变化）

> 规则细节已统一维护在 Agent 文件：
> `.github/agents/logseq-mcp-regression.agent.md`
