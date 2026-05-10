---
description: "运行 Logseq MCP 只读回归（Phase A smoke）"
agent: "Logseq MCP Regression"
---

# 测试 Logseq MCP（只读）

使用专用 Agent `Logseq MCP Regression` 执行只读回归（Phase A）。

## 默认参数

- scope: `readonly`
- mixed-write: `off`
- baseline: 可选（用于对比）

## 期望输出

1. 当前轮次总表（✅/⚠️/❌）
2. Capabilities 门禁快照与跳过项
3. 可用性探测（available/unavailable）
4. 有效测试覆盖率
5. 与上次结果对比（新增失败 / 已修复 / 状态变化）

> 规则细节统一维护在：
> `.github/agents/logseq-mcp-regression.agent.md`
