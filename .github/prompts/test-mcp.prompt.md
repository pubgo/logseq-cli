---
description: "测试 Logseq MCP 工具全量功能"
agent: "agent"
---

# 测试 Logseq MCP

对所有 Logseq MCP 工具执行功能验证，按类别逐组测试，最终输出汇总表格。

## 测试步骤

### 1. 连接与图谱基础

- `mcp_logseq_graph_info` — 确认图谱名称和路径
- `mcp_logseq_capabilities_get` — 检查 API 能力（DSL、tags、db_graph 等）
- `mcp_logseq_graph_favorites` — 获取收藏页列表
- `mcp_logseq_graph_recent` — 获取最近访问页列表
- `mcp_logseq_graph_config` / `mcp_logseq_graph_graph-config` — 图谱配置
- `mcp_logseq_graph_app-info` — 应用信息
- `mcp_logseq_graph_state` (key: "sidebar/blocks") — 状态查询

### 2. 页面操作

- `mcp_logseq_page_list` — 列出所有页面（检查返回量级）
- `mcp_logseq_page_get` (name: 从 favorites 或 recent 中取一个) — 获取页面详情
- `mcp_logseq_page_current` — 当前打开页面（可能为空）
- `mcp_logseq_page_current-tree` — 当前页面树
- `mcp_logseq_page_journal` — 今日日记页
- `mcp_logseq_page_refs` (name: 同上) — 页面引用
- `mcp_logseq_page_namespace` (name: 有命名空间的页面) — 命名空间子页
- `mcp_logseq_page_properties` (name: 同上) — 页面属性

### 3. 搜索与查询

- `mcp_logseq_search` (query: "golang") — 全文搜索
- `mcp_logseq_search-notes` (query: "golang") — 笔记搜索
- `mcp_logseq_query_datalog` (query: `[:find (count ?b) :where [?b :block/uuid]]`) — Datalog 查询
- `mcp_logseq_query_dsl` (query: `(page-tags golang)`) — DSL 查询

### 4. 标签与属性

- `mcp_logseq_tag_list` — 标签列表
- `mcp_logseq_tag_search` (query: "golang") — 标签搜索
- `mcp_logseq_property_list` — 属性列表

### 5. Block 操作（只读）

- `mcp_logseq_block_current` — 当前选中块
- `mcp_logseq_block_selected` — 多选块
- `mcp_logseq_block_get` (uuid: 从搜索结果中取一个) — 获取块内容

### 6. 输出汇总

测试完成后输出 Markdown 表格：

| 功能 | 状态 | 备注 |
|------|------|------|
| 工具名 | ✅ / ❌ / ⚠️ | 返回摘要或错误原因 |

状态说明：
- ✅ 正常返回
- ⚠️ 预期内的限制（如未打开页面、需要参数）
- ❌ 异常错误（需要排查）

对于 ❌ 的项目，分析是 Logseq API 版本限制还是 CLI 实现问题，并给出修复建议。
