# logseq-cli 面向 LLM 的 P0 API 规范（草案）

> 目标：把当前“可用”的 CLI/MCP 能力升级为“可稳定托管给 LLM”的接口层。

## 1. 设计目标

P0 仅覆盖三件事：

1. **可判定能力边界**：不同 Logseq 版本/图谱差异可被探测与显式返回。
2. **可控写入风险**：默认只读、写入可预演（dry-run）、危险操作强确认。
3. **可控上下文体积**：统一分页游标与结果裁剪，避免 LLM 上下文爆炸。

---

## 2. 统一响应信封（所有工具）

所有 MCP 工具响应建议统一为：

```json
{
  "ok": true,
  "data": {},
  "error": null,
  "meta": {
    "request_id": "uuid",
    "duration_ms": 12,
    "capability_used": ["logseq.DB.datascriptQuery"],
    "fallback_used": ["refs_over_tags"],
    "next_cursor": "opaque-cursor",
    "has_more": false
  },
  "hints": [
    "当前图谱 :block/tags 为空，已使用 :block/refs 回退"
  ]
}
```

错误时：

- `ok=false`
- `error` 结构化：`code` / `message` / `retryable` / `suggestion`

建议错误码：

- `BAD_REQUEST`
- `AUTH_REQUIRED`
- `CAPABILITY_UNAVAILABLE`
- `RESOURCE_NOT_FOUND`
- `CONFLICT`
- `TIMEOUT`
- `UPSTREAM_ERROR`
- `SAFETY_BLOCKED`

---

## 3. P0 工具集合

## 3.1 `logseq.capabilities.get`

### 目的

返回当前实例能力矩阵，指导 LLM 选择路径（避免盲试）。

### 输入

```json
{}
```

### 输出（示例）

```json
{
  "ok": true,
  "data": {
    "api": {
      "search": true,
      "dsl_query": false,
      "datascript_query": true,
      "tag_relation_ops": true
    },
    "graph": {
      "db_graph": true,
      "tags_field_available": false,
      "refs_field_available": true
    },
    "write_policy": {
      "default_mode": "read-only",
      "dangerous_requires_confirm": true
    }
  },
  "meta": {
    "request_id": "req-...",
    "duration_ms": 9,
    "capability_used": [
      "logseq.DB.datascriptQuery",
      "logseq.DB.q",
      "logseq.App.search"
    ]
  },
  "hints": [
    "dsl query unsupported in current runtime"
  ]
}
```

### 实现映射（现有能力）

- `graph db-graph`
- `query datalog`
- `query dsl`
- `tag list`

---

## 3.2 `logseq.search.notes`

### 目的

统一“全文检索 + 标签过滤 + 分页裁剪”的读取入口。

### 输入

```json
{
  "query": "golang logseq",
  "tag": "project",
  "limit": 20,
  "cursor": "",
  "include": ["pages", "blocks"]
}
```

### 输出

- `items[]`：统一条目结构（`type`, `title`, `snippet`, `page`, `uuid`）
- `next_cursor` / `has_more`
- 外层统一信封：`ok/data/error/meta/hints`

### 实现映射

- `search`
- `query datalog`（必要时兜底）

当前已落地的 CLI 入口：`logseq search-notes <query> --limit --cursor --include --tag`

典型返回（节选）：

```json
{
  "ok": true,
  "data": {
    "query": "golang",
    "limit": 20,
    "next_cursor": "20",
    "has_more": true,
    "total": 83,
    "items": []
  },
  "meta": {
    "request_id": "req-...",
    "duration_ms": 5,
    "capability_used": ["logseq.App.search"],
    "next_cursor": "20",
    "has_more": true
  }
}
```

---

## 3.3 `logseq.page.get_context`

### 目的

返回页面上下文的“LLM可消费视图”（元数据 + 有限块树）。

### 输入

```json
{
  "page": "Project Alpha",
  "max_blocks": 200,
  "max_depth": 6,
  "include_properties": true
}
```

### 输出

- `page`：基础信息
- `properties`
- `outline_blocks[]`：裁剪后的块树（含 `uuid/content/marker/priority/level`）

### 实现映射

- `page get --blocks`
- `page properties`

当前已落地的 CLI 入口：`logseq page get-context <name> --max-blocks --max-depth --include-properties`

返回同样使用统一信封，并在 `data.stats` 中提供：

- `returned_blocks`
- `clipped_by_depth`
- `clipped_by_limit`

---

## 3.4 `logseq.page.append_safe`

### 目的

安全追加内容；支持幂等与 dry-run。

### 输入

```json
{
  "page": "Project Alpha",
  "content": "- [ ] Draft proposal",
  "idempotency_key": "sha256:...",
  "dry_run": false
}
```

### 行为约束

- `dry_run=true`：仅返回将执行动作，不写入。
- `idempotency_key` 相同且近期已执行：返回已存在结果，不重复写。

### 实现映射

- `block append`
- （建议新增轻量幂等缓存，内存/文件均可）

当前已落地的 CLI 入口：`logseq page append-safe <name> <content> --dry-run --confirm --idempotency-key`

返回同样使用统一信封，`data.action` 为：

- `dry-run`
- `deduped`
- `appended`

---

## 3.5 `logseq.task.upsert`

### 目的

给 LLM 一个稳定“任务写入”语义层，避免直接拼装底层块调用。

### 输入

```json
{
  "page": "Project Alpha",
  "title": "Draft proposal",
  "marker": "TODO",
  "priority": "A",
  "properties": {
    "owner": "barry",
    "due": "2026-05-12"
  },
  "dedupe_by": "title"
}
```

### 输出

- `action`: `created | updated | noop`
- `block_uuid`

### 实现映射

- `page get --blocks`（查重）
- `block append` 或 `block update`
- `block property set`

---

## 3.6 `logseq.block.delete_safe`

### 目的

对危险删除进行双保险。

### 输入

```json
{
  "uuid": "...",
  "confirm": false,
  "dry_run": true
}
```

### 行为约束

- `dry_run=true`：返回块摘要、子树规模与影响范围。
- 实际删除要求 `confirm=true`，否则 `SAFETY_BLOCKED`。

### 实现映射

- `block get --children`
- `block remove`

当前已落地的 CLI 入口：`logseq block delete-safe <uuid> --dry-run --confirm`

返回同样使用统一信封，`data.action` 为：

- `dry-run`
- `deleted`

---

## 4. 分页与裁剪约定

- 默认 `limit=20`，最大 `limit=200`。
- `cursor` 使用不透明字符串（内部可编码 offset/anchor）。
- 每个 `snippet` 建议裁剪到 `<= 500` 字符。
- 查询输出建议优先结构化字段，避免超长原始 JSON。

---

## 5. 安全策略（P0）

建议新增运行时策略（可来自环境变量）：

- `LOGSEQ_LLM_WRITE_MODE=read-only|confirm|direct`（默认 `read-only`）
- `LOGSEQ_LLM_REQUIRE_CONFIRM_FOR_DELETE=true`（默认 true）
- `LOGSEQ_LLM_MAX_RESULTS=200`

策略应出现在 `capabilities.get` 返回中，便于 LLM 决策。

---

## 6. 与现有命令树映射总览

- 读取层：`search` / `page get` / `query datalog` / `tag list`
- 写入层：`block append` / `block update` / `block remove` / `block property set`
- 状态层：`graph state` / `graph state-set`

即：P0 不要求推翻现有命令，只需要在 MCP 暴露层增加“更稳的工具契约”。

---

## 7. 最小落地顺序（建议）

1. 先实现 `capabilities.get`。
2. 再实现 `search.notes` + `page.get_context`（含分页）。
3. 最后实现 `append_safe` + `delete_safe`（含 dry-run / confirm）。

完成以上三步后，LLM 侧稳定性会有明显提升。
