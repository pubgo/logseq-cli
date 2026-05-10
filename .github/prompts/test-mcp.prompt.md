---
description: "测试 Logseq MCP 工具全量功能"
agent: "agent"
---

# 测试 Logseq MCP

对 Logseq MCP 工具执行功能验证，按类别逐组测试，并输出：

- 当前轮次汇总结果
- 与上一次测试结果的对比（新增失败 / 已修复 / 状态变化）

## 执行规则（必须遵守）

1. 按下面的分组顺序执行，不要跳步。
2. 优先并行执行互不依赖的读操作。
3. 参数必须使用工具要求的字段名（例如 `tag_search` 用 `name`，不是 `query`）。
4. 同一工具若失败，允许重试 1 次；重试仍失败则记录为最终状态。
5. 对于“预期限制”要标记为 ⚠️，不要误判为 ❌。
6. **必须先完成只读测试，再进行混合测试**；若只读阶段出现 ❌，混合阶段默认跳过并先排障。

## 前置步骤：工具可用性探测（新增）

在正式执行前，先快速探测本轮可用的 MCP 工具集合，并生成两份列表：

- `available_tools`：本轮可直接调用的工具
- `unavailable_tools`：本轮未暴露/不可调用的工具

执行策略：
- 对 `unavailable_tools` 不做失败判定，不计入 ❌；
- 在总表中标记为 ⚠️，备注统一写“本轮工具集未暴露”；
- 统计时单独给出“有效测试覆盖率”（仅按 `available_tools` 计算）。

## 前置步骤：Capabilities 驱动门禁（新增）

在拿到 `mcp_logseq_capabilities_get` 结果后，提取并缓存以下字段作为运行时门禁：

- `api.app_info.supported`
- `api.tag_search.supported`
- `api.property_list.supported`
- `api.property_get.supported`

动态执行规则：
- 若上述字段为 `false`，对应测试项直接标记 ⚠️（"当前 Logseq 能力不可用"），不再执行该项、也不计入 ❌。
- 若字段缺失（旧版 capabilities 输出），按“未知能力”处理：允许执行 1 次，失败后再判定。
- 若 `connection_info.tokenConfigured=false`，直接判定 Phase A 阻塞。

建议映射关系：
- `graph_app-info` ← `api.app_info.supported`
- `tag_search` ← `api.tag_search.supported`
- `property_list` ← `api.property_list.supported`
- `property_get/upsert/remove`（若执行）← `api.property_get.supported`

## 测试模式（完整流程）

### Phase A：只读基线测试（必跑）

- 目标：验证连接、查询、搜索、只读页面/块读取链路是否稳定。
- 范围：本文“测试步骤”里的 1~5 章节现有读操作。
- 门禁：
	- 若 `❌ = 0`，进入 Phase B。
	- 若 `❌ > 0`，先输出失败归因与修复建议，再决定是否继续。

补充：
- 仅对 `available_tools` 做门禁判定；
- 若某关键项不可用（例如 `capabilities_get`），直接判定 Phase A 阻塞。
- 若 Capabilities 明确声明某能力不支持，则相关测试项按 ⚠️ 跳过，不作为失败。

### Phase B：混合测试（读 + 安全写）

- 目标：验证关键写路径与读回一致性（优先安全写接口）。
- 原则：
	- 优先 `dry-run`，再最小写入验证。
	- 若写策略为 `read-only` 且无法提升权限，标记 ⚠️ 并跳过写入。
	- 测试数据使用固定前缀：`mcp-test-`，便于识别与清理。
	- 若写接口可用但被策略降级为 dry-run，记为“受策略限制通过”（⚠️，非 ❌）。

建议最小写入用例（按顺序）：
1. `mcp_logseq_page_append-safe`：
	 - `dry-run=true`（应成功预检）
	 - 再 `confirm=true` 写入一条测试 block
2. `mcp_logseq_block_append`：在 `mcp-test-*` 页面追加 block
3. `mcp_logseq_block_update`：更新刚写入的 block 内容
4. `mcp_logseq_block_get`：读回核对内容是否一致
5. `mcp_logseq_block_remove`：删除测试 block（若支持）

注意：
- 若 `page_append-safe` 目标页不存在，可先用 `block_append` 创建测试页后重试；
- 若 `block_remove` 不可用，允许保留残留并记录 UUID。

清理策略（best-effort）：
- 优先删除测试 block；
- 无法删除时，至少在备注里记录残留对象（page / block uuid）。

### Phase C：完整回归结论

- 合并 Phase A + Phase B 结果，输出统一总表与趋势对比。
- 额外输出：
	- 写入链路通过率（写入相关 ✅ / 总写入项）
	- 是否存在残留测试数据
	- 有效测试覆盖率（可用工具中已执行项 / 可用工具总数）

## 状态判定标准

- ✅ 正常返回，且结果结构符合预期
- ⚠️ 预期内限制（例如：未打开页面、未选择块、DSL 有能力限制、需要参数）
- ❌ 异常错误（需排查，可能是 API 版本限制或 CLI 实现问题）

补充判定：
- 工具未暴露/不可调用：⚠️（不计入失败率）
- 策略限制导致仅 dry-run：⚠️（不计入实现失败）
- capabilities 明确不支持：⚠️（不计入实现失败）

## 测试步骤

### 1. 连接与图谱基础

- `mcp_logseq_graph_info` — 确认图谱名称和路径
- `mcp_logseq_capabilities_get` — 检查 API 能力（DSL、tags、db_graph 等）
- `mcp_logseq_graph_favorites` — 获取收藏页列表
- `mcp_logseq_graph_recent` — 获取最近访问页列表
- `mcp_logseq_graph_config` / `mcp_logseq_graph_graph-config` — 图谱配置
- `mcp_logseq_graph_app-info` — 应用信息
- `mcp_logseq_graph_state` (key: "sidebar/blocks") — 状态查询

补充校验：
- 检查 `capabilities` 中 `connection_info.tokenConfigured` 是否为 `true`
- 记录 `dsl_query.supported` 与 `db_graph.supported` 的布尔值和原因

### 2. 页面操作

- `mcp_logseq_page_list` — 列出所有页面（检查返回量级）
- `mcp_logseq_page_get` (name: 从 favorites 或 recent 中取一个) — 获取页面详情
- `mcp_logseq_page_current` — 当前打开页面（可能为空）
- `mcp_logseq_page_current-tree` — 当前页面树
- `mcp_logseq_page_journal` — 今日日记页
- `mcp_logseq_page_refs` (name: 同上) — 页面引用
- `mcp_logseq_page_namespace` (name: 有命名空间的页面) — 命名空间子页
- `mcp_logseq_page_properties` (name: 同上) — 页面属性

补充校验：
- `page_get` 若 favorites 第一个页面不存在，则自动降级到 recent 第一个页面
- `page_current` / `page_current-tree` 返回空时标记为 ⚠️（无当前页面）
- `page_journal` 成功时记录 `name` 与 `journalDay`

### 3. 搜索与查询

- `mcp_logseq_search` (query: "golang") — 全文搜索
- `mcp_logseq_search-notes` (query: "golang") — 笔记搜索
- `mcp_logseq_query_datalog` (query: `[:find (count ?b) :where [?b :block/uuid]]`) — Datalog 查询
- `mcp_logseq_query_dsl` (query: `(page-tags golang)`) — DSL 查询

补充校验：
- 对 `search-notes` 再跑一次分页请求（使用 `next_cursor`）验证分页
- `query_dsl` 若返回空数组但无报错，判定为 ✅（结果为空，不是失败）

### 4. 标签与属性

- `mcp_logseq_tag_list` — 标签列表
- `mcp_logseq_tag_search` (name: "golang") — 标签搜索
- `mcp_logseq_property_list` — 属性列表

补充校验：
- 统计 `tag_search` 与 `property_list` 的结果数量
- 若数量为 0 但无错误，判定 ✅ 并在备注中注明“空结果”

### 5. Block 操作（只读）

- `mcp_logseq_block_current` — 当前选中块
- `mcp_logseq_block_selected` — 多选块
- `mcp_logseq_block_get` (uuid: 从搜索结果中取一个) — 获取块内容

补充校验：
- `block_selected = null` 时标记 ⚠️（当前无多选）
- `block_get` 的 uuid 优先取 `search` 结果的第一个 block uuid

### 6. 输出汇总

#### A. 当前轮次总表

输出 Markdown 表格：

| 功能 | 状态 | 备注 |
|------|------|------|
| 工具名 | ✅ / ❌ / ⚠️ | 返回摘要或错误原因 |

并给出总计：`✅ x / ⚠️ y / ❌ z`。

额外给出：`写入链路通过率 = write_pass / write_total`（若未执行写入则写 N/A）。
额外给出：`有效测试覆盖率 = executed_available / total_available`。

#### B. 与上次结果对比

若本会话中存在上一轮结果，追加输出：

| 对比项 | 数量 | 详情 |
|---|---:|---|
| 新增失败 | n | 工具名列表 |
| 已修复 | n | 工具名列表 |
| 状态变化（含 ⚠️↔✅） | n | 工具名 + 变化方向 |

若无上次结果，写明“无可对比基线”。

对于 ❌ 的项目，分析是 Logseq API 版本限制还是 CLI 实现问题，并给出修复建议。

#### C. 失败归因模板（用于每个 ❌）

- 工具：
- 错误原文：
- 归因：`API 版本限制` / `参数问题` / `CLI 实现问题` / `环境问题`
- 建议修复：

#### D. 本轮执行轨迹（简版）

- Phase A：通过 / 阻塞（原因）
- Phase B：已执行 / 跳过（原因）
- 清理结果：成功 / 部分成功 / 未执行

#### E. 可用性探测结果（新增）

- available_tools: [ ... ]
- unavailable_tools: [ ... ]
- 未暴露工具数量: n（这些项按 ⚠️ 记录，但不计入 ❌）

#### F. Capabilities 门禁快照（新增）

- api.app_info.supported: true/false
- api.tag_search.supported: true/false
- api.property_list.supported: true/false
- api.property_get.supported: true/false
- 因门禁跳过的测试项: [ ... ]
