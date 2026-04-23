---
title: "Implementation Plan: search_documents MCP Tool"
status: accepted
---

## Goal

Add a deterministic MCP tool `search_documents` to the archcore CLI so agents can find `.archcore/` documents matching a code path reference, a content substring, or metadata filters — in one call, without grep loops or guessing.

The tool is a pure matching primitive: it answers "which documents match X" and returns ranked evidence. It does not classify user input, group results by category, or render presentation. Those concerns belong to any downstream caller (LLM agent, skill, hook, CI script).

**Problem.** `list_documents` only filters on frontmatter metadata; `get_document` fetches one known document. Neither answers "which of these 200 docs reference `src/payments/`" or "which rules mention this topic". Agents fall back to ad-hoc grep loops that are slow, non-portable, and consume tokens.

## Key Design Decisions

- **Matching is deterministic Go.** Regexes, specificity math, UTF-8 safety, filter composition — all in code with unit tests. No LLM reasoning in the matching path.
- **`sort: "relevance" | "mtime"` parameter on day one.** Default `relevance` = `max(match.specificity)` DESC → hardcoded type priority (`rule > adr > spec > cpat > guide > plan > idea`) → `mtime` DESC. Keeps ranking fully in Go.
- **Tight path-ref matching.** Explicit `@path` accepted unconditionally. Bare substrings require ≥ 1 `/` AND one of: trailing `/`, ≥ 2 `/` separators, or final segment has a source-file extension. Rejects URLs, comment prose, manifest examples.
- **Content mode is strict substring** (case-insensitive) against `title + body`. No stemming, no fuzzy, no OR-split. Documented explicitly in the tool description so callers do not fabricate fallbacks.
- **`mtime` plumbed through `LocalDocument`** via `fs.DirEntry.Info().ModTime()`; JSON tag `omitzero` (Go 1.24) — `omitempty` does not work on `time.Time`.
- **Lazy body loading.** Pure metadata queries (no `path_ref` / `content`) use `ScanDocuments` (no bodies retained). Queries needing body access use `ScanDocumentsFull`. I/O identical; heap differs.
- **Relations enrichment in-tool.** Response includes `incoming_relations` / `outgoing_relations` per doc from the manifest. Avoids N extra `list_relations` calls by callers.
- **UTF-8-safe excerpts.** Byte offsets snapped to rune boundaries before slicing. Excerpts always valid UTF-8.
- **Stderr, not stdout.** Manifest load failures logged to stderr — stdout is the MCP protocol channel.
- **No new manifest fields, no index file.** Cold scan per call. Pre-built index and hook integration are P2.

## Sprint Status

### ✅ Sprint 1 — search_documents primitive — DONE

Delivered:

- `mtime` field on `LocalDocument`, plumbed via `fs.DirEntry.Info().ModTime()` (scan) and `os.Stat` (single-doc read). JSON tag `omitzero` (Go 1.24).
- `ScanDocumentsFull(baseDir)` — one-pass scan with `Content`. Shared private `scanDocuments(baseDir, includeContent)`; zero extra I/O vs. previous `ScanDocuments`.
- `templates/source_extensions.go` + unit test — curated list, case-insensitive `IsSourceExtension`.
- `internal/mcp/tools/search_documents.go` — tool factory, handler, all helpers (`parseMtimeAfter`, `extractPathRefs` with overlap suppression, `filterBareMentions`, `computeSpecificity`, `extractContentMatch`, `buildExcerpt` with rune-boundary snapping, `sortResults`).
- `internal/mcp/tools/search_documents_test.go` — 22 table/case-driven tests incl. URL-like negative cases, AND-combined filters, both sort modes, manifest-relations enrichment, UTF-8 excerpt safety, lazy body loading regression.
- Registered in @internal/mcp/server.go; `WHEN TO SEARCH CONTENT` block added to `mcpServerInstructions`.

Deviations from the original draft (all intentional):

- `incoming_relations` / `outgoing_relations` always present (non-nil), not `omitempty`. Callers can rely on the fields being present.
- Manifest load failures go to stderr (`fmt.Fprintf(os.Stderr, ...)`). Stdout would corrupt the MCP stdio channel.
- `limit=0` maps to default 50 (JSON-number path cannot distinguish "omitted" from "explicit 0"). Documented in the `limit` parameter description.
- Overlap suppression inside `extractPathRefs` — `@src/foo` does not emit both an explicit ref and a bare `src/foo` ref. Prevents double-counting in ranking.
- Post-review fixes: lazy body loading for metadata-only queries (avoids N × body_size heap retention); rune-boundary snapping in `buildExcerpt` (avoids slicing mid-UTF-8-rune on non-ASCII content). Both have dedicated regression tests.
- Benchmark (fixture + P95 assertion) deferred — will add if real-world usage reveals performance concerns.

## Acceptance Criteria — Sprint 1

All of the following hold:

- `search_documents(path_ref)` returns ranked matches for paths using both `@`-notation and qualified bare mentions.
- `search_documents(content)` returns case-insensitive substring matches across title + body.
- `search_documents(sort="relevance")` sorts deterministically by specificity → type priority → mtime; `sort="mtime"` sorts by mtime only.
- Response includes `matches`, `mtime`, `incoming_relations`, `outgoing_relations` per doc.
- URL-like strings and single-separator prose do not produce false positives.
- All excerpts satisfy `utf8.ValidString`.
- Pure metadata queries do not load document bodies.
- Go unit tests pass with ≥ 20 cases including regression guards for the documented risks.

## Out of Scope (Sprint 1 boundary)

- PreToolUse push-injection hook (P2).
- Persisted path-index in `.sync-state.json` (P2 — only when hook exists).
- Symbol / feature resolvers (`AuthMiddleware` → class → docs) — P2.
- Violation detection (`archcore check`) — separate workstream.
- Semantic / embedding search — P3+.
- Stemming or fuzzy matching in content mode — P2 if usage demands.
- Downstream presentation tooling (grouping into sections, markdown rendering, user-facing skills) — out of this repository's scope.

## Dependencies

- Reuses `WalkArchcoreFiles`, `SplitDocument`, `sync.LoadManifest` — no new internal APIs required.
- No new Go module dependencies.

## Risks and Mitigations

### §1 — Path-reference regex false positives

`[\w\-_]+/[\w\-_./]+` alone is too permissive. It would catch URLs (`docs.example/com`), git refs, prose like `see/CONTRIBUTING.md`, manifest examples.

**Mitigation (shipped):** heuristic accept-list for bare mentions — require one of: trailing `/`, ≥ 2 `/` separators, OR final segment has a source-file extension. Explicit `@path` notation bypasses these checks (intentional). Unit tests include URL-like strings and single-separator prose as negative cases.

### §2 — Topic (content) mode is strict substring

`content="money rounding"` will not match a title "Money Arithmetic in Decimals". This is a deliberate limitation of the primitive: no stemming, no fuzzy matching. Callers must not hide this from users.

**Mitigation (shipped):** the tool description explicitly states "strict substring — singular/plural forms do not match". Empty results are honest — callers are expected to surface them plainly, not fabricate silent retries.

### §3 — Payload overhead from enriched relations

`incoming_relations` / `outgoing_relations` are emitted for every matched doc up to `limit`. On a dense relation graph, payload can grow to ~10+ KB.

**Mitigation:** monitor in practice. If it becomes a problem, add an optional `include_relations: boolean` (default true) or `include_relations_for_types: string[]` parameter. Not blocking.

### §4 — Ambient rules without `@path` are not findable

A rule like "HTTP handlers return JSON errors" without any `@src/api/` reference in its body will not appear in path-mode results for `src/api/`. This is intentional (determinism > recall), but users may misread as "tool is broken".

**Mitigation (partial):** the tool description and spec document this explicitly. A future `archcore doctor` hygiene warning for rules with zero `@path` references is tracked separately.

### §5 — ToLower byte-offset drift for non-ASCII content queries

`extractContentMatch` searches `strings.ToLower(body)` and uses the returned byte index to slice the original `body`. For certain non-ASCII case mappings (e.g., Turkish `İ → i̇`) the lowered string differs in byte length from the original, which can shift the match position. In practice the excerpt still renders valid UTF-8 (post-fix via rune-boundary snapping) but may lose a few adjacent characters.

**Mitigation:** known limitation. Accept for P1 (content is mostly ASCII/English in target repos). Revisit if dogfood reveals real impact.

## Follow-Up Backlog

Captured from Sprint 1 deferrals and review notes — not blocking, revisit when signal justifies:

- **Benchmark fixture** in `internal/mcp/tools/` — add if a real repo crosses ~500 docs.
- **`include_relations` parameter** — add if Risks §3 payload bloat reported.
- **`archcore doctor` ambient-rule hygiene** — surface rules with zero `@path` references as a warning.
- **Path-ref index in `.sync-state.json`** — only when a `PreToolUse` hook needs sub-ms lookup.
- **Proper case-fold content search** — via `golang.org/x/text`, if Risks §5 surfaces.
