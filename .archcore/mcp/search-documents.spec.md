---
title: "search_documents MCP Tool Contract"
status: accepted
---

## Purpose

This specification defines the canonical contract for the `search_documents` MCP tool exposed by the archcore CLI server.

It is normative for:

- The tool implementation in @internal/mcp/tools/search_documents.go.
- Any MCP client calling the tool (LLM agents, future hooks, CI scripts).
- Any downstream consumer that depends on the tool's response shape.

## Scope

### Covers

- Tool signature, parameter validation, enum constraints.
- Path-reference extraction algorithm (explicit `@path` + qualified bare mentions).
- Content (topic) matching algorithm.
- Specificity computation.
- Ranking modes (`relevance`, `mtime`).
- Excerpt construction (including UTF-8 safety).
- Relations enrichment from the manifest.
- Lazy body loading behavior.
- Response JSON shape.
- Error conditions and messages.

### Does Not Cover

- Persistent path index in `.sync-state.json` — deferred (see plan §Out of Scope).
- Hook-time invocation semantics — deferred to a future hook spec.
- Semantic / embedding search — out of scope.

## Authority

This document is the normative specification for the behavior of `search_documents`.

If the implementation, tests, or downstream consumers diverge from this specification, this specification takes precedence until it is amended.

### Related Artifacts

- Implementation: @internal/mcp/tools/search_documents.go
- Tool registration: @internal/mcp/server.go
- Tests: @internal/mcp/tools/search_documents_test.go
- Shared scan helpers: @internal/mcp/tools/common.go (`ScanDocuments`, `ScanDocumentsFull`)
- Manifest loading: @internal/sync/manifest.go
- Source-extension list: @templates/source_extensions.go
- Implementation plan: `.archcore/mcp/search-documents-implementation.plan.md`

## Subject

- **Name:** `search_documents`
- **Kind:** MCP tool (JSON-RPC over stdio transport).
- **Primary responsibility:** Return `.archcore/` documents matching the caller's filters, with deterministic matching, ranking, and per-match evidence.
- **Consumers / dependents:** LLM agents invoking MCP tools directly, future PreToolUse hooks, any third-party MCP client.

## Definitions

| Term               | Definition                                                                                                                                                      |
| ------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Document body      | Markdown content following the YAML frontmatter block of an `.archcore/*.md` file.                                                                              |
| Frontmatter        | YAML block bounded by `---` at the top of an `.archcore/*.md` file, carrying `title`, `status`, `tags`.                                                         |
| Explicit reference | A path reference in a document body matching the regex `@[\w./\-_]+`.                                                                                           |
| Bare mention       | A path-like token in a document body matching `[\w\-_]+/[\w\-_./]+`, accepted only when one of the heuristic conditions defined in Normative Behavior §5 holds. |
| Specificity        | A non-negative integer measure of how precisely a candidate match relates to the filter.                                                                        |
| Manifest           | The JSON file at `.archcore/.sync-state.json` that stores document relations, loaded via `sync.LoadManifest`.                                                   |
| Source extension   | An extension listed in `templates.IsSourceExtension` (e.g., `.go`, `.ts`, `.py`, `.md`).                                                                        |

## Contract Surface

### Interface

Exposed over MCP using `github.com/mark3labs/mcp-go`. Registered from `NewServer(baseDir)` in @internal/mcp/server.go. Annotated `ReadOnlyHint: true`.

### Inputs

| Name          | Type     | Required    | Description                                                                                                                                   |
| ------------- | -------- | ----------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| `path_ref`    | string   | conditional | Path reference to match in document bodies. Leading `@` is stripped during comparison.                                                        |
| `content`     | string   | conditional | Case-insensitive substring matched against `title + body`. No stemming, no OR-split, no fuzzy matching.                                       |
| `types`       | string[] | conditional | Filter by document type. OR semantics across the list.                                                                                        |
| `status`      | string   | conditional | Filter by frontmatter status. Enum: `draft`, `accepted`, `rejected`.                                                                          |
| `mtime_after` | string   | no          | Inclusive lower bound on document mtime. Accepts RFC3339 (ISO-8601) or a positive relative duration: `<N>h`, `<N>d`, `<N>w`, `<N>mo`, `<N>y`. |
| `sort`        | string   | no          | Ordering mode. Enum: `relevance` (default), `mtime`.                                                                                          |
| `limit`       | number   | no          | Maximum number of results. Default 50, maximum 200. Values above 200 are clamped; `0` or omitted both map to the default.                     |

At least one of `path_ref`, `content`, `types`, or `status` MUST be provided.

### Outputs

A JSON array of `searchResult` objects. Each object has the following fields:

| Field                | Type             | Nullability             | Description                                                      |
| -------------------- | ---------------- | ----------------------- | ---------------------------------------------------------------- |
| `path`               | string           | always                  | Relative path beginning with `.archcore/`.                       |
| `title`              | string           | always                  | Frontmatter title (may be empty for malformed docs).             |
| `type`               | string           | always                  | Document type (e.g., `rule`, `adr`, `spec`).                     |
| `status`             | string           | omit if empty           | Frontmatter status.                                              |
| `mtime`              | string (RFC3339) | always                  | File modification time.                                          |
| `tags`               | string[]         | omit if empty           | Frontmatter tags.                                                |
| `matches`            | Match[]          | always                  | Evidence array. Always present; empty for pure metadata queries. |
| `incoming_relations` | Relation[]       | always (possibly empty) | Manifest edges where this doc is the target.                     |
| `outgoing_relations` | Relation[]       | always (possibly empty) | Manifest edges where this doc is the source.                     |

`Match` fields:

| Field         | Type    | Description                                                                                   |
| ------------- | ------- | --------------------------------------------------------------------------------------------- |
| `kind`        | string  | Enum: `path_ref_explicit`, `path_ref_mention`, `content`.                                     |
| `ref`         | string  | The raw matched token (e.g., `@src/payments/` or `money rounding`).                           |
| `specificity` | integer | Per-match specificity value (see Normative Behavior §6).                                      |
| `excerpt`     | string  | ≤ ~120-character window around the match, padded with `...` if truncated. Always valid UTF-8. |

`Relation` fields (reuses `DocumentRelation` from @internal/mcp/tools/common.go): `path`, `type` (one of `related`, `implements`, `extends`, `depends_on`).

## Normative Behavior

### §1 Filter validation

1. The handler MUST reject calls where all of `path_ref`, `content`, `types`, and `status` are empty or absent, returning the MCP error `specify at least one filter (path_ref, content, types, or status)`.
2. The handler MUST reject `limit < 0` with the error `limit must be non-negative`.
3. The handler MUST treat `limit == 0` as equivalent to omitted — both produce the default of 50.
4. The handler MUST clamp `limit > 200` to 200 without emitting an error.
5. `mtime_after` MUST accept both RFC3339 timestamps and positive relative durations of the form `<N>h`, `<N>d`, `<N>w`, `<N>mo`, `<N>y`. Invalid input MUST return the error `invalid mtime_after: <reason>`.

### §2 Document loading

1. When `path_ref` and `content` are both empty, the handler MUST call `ScanDocuments(baseDir)`, which does not retain document bodies in memory beyond the frontmatter parse.
2. When either `path_ref` or `content` is non-empty, the handler MUST call `ScanDocumentsFull(baseDir)`, which retains bodies on each `LocalDocument`.
3. Scanner I/O cost MUST be identical in both cases; only heap retention differs.

### §3 Manifest loading

1. The handler SHOULD call `sync.LoadManifest(baseDir)` to obtain relation data.
2. If manifest loading fails, the handler MUST continue processing and emit all results with empty `incoming_relations` and `outgoing_relations` arrays.
3. Manifest load failures MUST be logged to `os.Stderr` (never `os.Stdout`, which is the MCP protocol channel).

### §4 Metadata filters (AND semantics)

Filters are evaluated in the following order; a document MUST be excluded if any active filter fails:

1. `types`: if non-empty, doc's `type` MUST be present in the list.
2. `status`: if non-empty, doc's `status` MUST equal the filter exactly.
3. `mtime_after`: if non-zero, doc's mtime MUST be strictly after the threshold.

### §5 Path-reference extraction

For each document's body (populated only when `path_ref` is set):

1. Apply regex `@[\w./\-_]+` to extract explicit references. Each match yields a candidate with `Kind: "explicit"`.
2. Apply regex `[\w\-_]+/[\w\-_./]+` to extract bare mention candidates. Each match yields a candidate with `Kind: "mention_candidate"`.
3. Remove overlap: if a bare candidate's byte range is contained within an explicit candidate's range, the bare candidate MUST be dropped. This prevents double-counting when an `@`-reference syntactically contains a bare path.
4. Filter bare candidates via `filterBareMentions`: a candidate is retained only if at least one of the following holds:
   - The candidate ends with `/`.
   - The candidate contains ≥ 2 `/` separators.
   - The final path segment's extension passes `templates.IsSourceExtension`.
5. For each surviving candidate, compute `computeSpecificity(candidate, pathRefFilter)` as the number of left-aligned `/`-segments shared between the normalized candidate (leading `@` and trailing `/` stripped) and the filter (same normalization). Specificity of `0` MUST NOT produce a match.
6. Each surviving candidate with specificity ≥ 1 MUST produce a `Match` with `kind` = `path_ref_explicit` or `path_ref_mention` depending on its origin regex.

### §6 Content matching

When `content` is non-empty, for each document:

1. Lowercase the query once.
2. Search first in the document title: if found, emit a `Match` with `kind: "content"`, `specificity: 3`, `excerpt` built from the title.
3. Otherwise, search the document body: if found, emit a `Match` with `kind: "content"`, `specificity: 1`, `excerpt` built from the body.
4. If neither title nor body contains the query, emit no content match for that document.

### §7 Ranking (`sort` parameter)

1. `sort="relevance"` (default) MUST order results by three keys, in order:
   - `max(matches.specificity)` DESC (documents with no matches use specificity 0).
   - Type priority ASC, per the map: `rule=1, adr=2, spec=3, cpat=4, guide=5, plan=6, idea=7, other=100`.
   - `mtime` DESC.
2. `sort="mtime"` MUST order purely by `mtime` DESC, ignoring specificity and type.
3. The sort MUST be stable: ties preserve input order (filesystem walk order), yielding deterministic output across runs.

### §8 Truncation

After sorting, the handler MUST truncate the result slice to `limit` entries.

### §9 Excerpt construction

1. For path-ref matches, `pos` is the byte offset of the matched token in the source body.
2. For content matches, `pos` is the byte offset of the match in the source (title or body).
3. The window width is 120 characters (`excerptWindow`), split evenly around `pos`.
4. Byte offsets MUST be snapped to rune boundaries before slicing: `start` moves backward until `utf8.RuneStart(source[start])` is true; `end` moves forward until `utf8.RuneStart(source[end])` is true or EOF is reached. Resulting excerpts MUST be valid UTF-8.
5. Internal whitespace runs in the excerpt MUST be collapsed to a single space.
6. `...` MUST be prepended when the excerpt's start is > 0 and appended when the excerpt's end is < `len(source)`.

### §10 Relations enrichment

For every emitted result, the handler MUST populate `incoming_relations` and `outgoing_relations` from the manifest. Both arrays MUST be serialized as `[]` when empty, never `null`.

### §11 Response shape invariants

1. The response is a JSON array. An empty array MUST be returned when no documents match (not `null`).
2. `matches` MUST be an empty array, never `null`, when a document passed metadata filters but has no per-match evidence (pure metadata query).
3. `mtime` MUST always be present in RFC3339 format. The `omitzero` JSON tag applies at the `LocalDocument` layer but the result-level struct serializes `mtime` unconditionally.

## Constraints

| Constraint                          | Value                     | Rationale                                                                                |
| ----------------------------------- | ------------------------- | ---------------------------------------------------------------------------------------- |
| Default limit                       | 50                        | Balances coverage and payload size for typical LLM invocations.                          |
| Maximum limit                       | 200                       | Caps payload to bounded size even when caller requests more.                             |
| Excerpt window                      | 120 chars                 | Keeps per-match payload small while carrying enough context for user/LLM disambiguation. |
| Cold-scan target                    | ≤ 500 ms P95 for 200 docs | Informal target from plan; not enforced by test harness yet.                             |
| Heap growth for pure-metadata query | O(N × frontmatter_size)   | Bodies are not retained when `path_ref` and `content` are both empty.                    |
| Heap growth for content query       | O(N × body_size)          | Bodies are retained; bounded by the `.archcore/` directory size.                         |

## Invariants

- The handler is read-only: no filesystem writes, no manifest mutations.
- Two identical calls against an unchanged `.archcore/` tree MUST produce byte-identical JSON output (excluding `mtime` serialization if clock-driven sources intrude — they do not in current implementation).
- All emitted `excerpt` strings MUST satisfy `utf8.ValidString(excerpt)`.
- Every returned result's `path` MUST begin with `.archcore/` and be a valid clean relative path.
- Results MUST NOT contain duplicate documents. One document yields at most one result row; multiple match candidates for the same doc collapse into the `matches` array.

## Error Handling

| Condition             | Response                                                                   | Recovery                                                                                            |
| --------------------- | -------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------- |
| All filters empty     | MCP error: `specify at least one filter (...)`                             | Caller adds at least one filter.                                                                    |
| `limit < 0`           | MCP error: `limit must be non-negative`                                    | Caller supplies non-negative limit.                                                                 |
| Invalid `mtime_after` | MCP error: `invalid mtime_after: <reason>`                                 | Caller supplies valid RFC3339 or relative duration.                                                 |
| Manifest read fails   | Continue with empty relations, log to stderr                               | Caller treats empty relations as absence of data.                                                   |
| Malformed frontmatter | Silently skip frontmatter parse; doc still indexed with empty title/status | Caller cannot distinguish malformed frontmatter from absent frontmatter — acceptable for this spec. |
| I/O error during scan | Wrapped error returned to MCP client                                       | Caller retries or falls back.                                                                       |

### Failure semantics

- Non-retriable: filter validation errors, `limit` validation.
- Retriable: transient I/O errors during scan (wrapped and surfaced).
- The handler is idempotent: no observable state changes between calls.

## Conformance

An implementation conforms to this specification if it satisfies:

- All MUST and MUST NOT statements in §§1–11.
- All stated invariants.
- All applicable error-handling rows.
- The recorded unit tests in @internal/mcp/tools/search_documents_test.go (these tests are the executable acceptance harness for this spec).

## Examples

### Path-mode query with mixed matches

```txt
// Input
search_documents({
  path_ref: "src/payments/utils.ts",
  limit: 50
})

// Output (abbreviated)
[
  {
    "path": ".archcore/rules/money-arithmetic.rule.md",
    "type": "rule",
    "status": "accepted",
    "mtime": "2026-03-12T09:14:00Z",
    "matches": [
      {
        "kind": "path_ref_explicit",
        "ref": "@src/payments/",
        "specificity": 2,
        "excerpt": "...amounts in @src/payments/ MUST use Decimal..."
      }
    ],
    "incoming_relations": [],
    "outgoing_relations": []
  }
]

// Notes
// Ordered by (specificity=2 DESC) → (type=rule rank=1) → (mtime DESC).
```

### Pure metadata query (no body scan)

```txt
// Input
search_documents({
  types: ["plan", "idea"],
  status: "draft",
  limit: 10
})

// Output (per-doc matches array is empty)
[
  {
    "path": ".archcore/payments/stripe-3ds-rollout.plan.md",
    "type": "plan",
    "status": "draft",
    "mtime": "2026-04-20T16:00:00Z",
    "matches": [],
    "incoming_relations": [],
    "outgoing_relations": []
  }
]

// Notes
// ScanDocuments is used (not ScanDocumentsFull) — bodies are not loaded.
```

### Invalid filter

```txt
// Input
search_documents({})

// Output
{ "error": "specify at least one filter (path_ref, content, types, or status)" }
```

## Security Considerations

- The tool operates only under `baseDir/.archcore/` and never outside. No path traversal is possible via the filters (no filter accepts a filesystem path as input — `path_ref` is matched against document body content, not dereferenced).
- The tool is read-only; no content can be written, modified, or deleted through it.
- Manifest load failures MUST be logged to stderr, not stdout, to protect the MCP protocol channel from corruption.

## Compatibility

- The response shape is additive: new fields MAY be added to `searchResult` or `Match` without breaking existing consumers that ignore unknown fields.
- Existing field names, types, and enum values MUST NOT be removed or repurposed in a minor CLI release. A breaking change requires a new major version.
- The `sort` enum MAY grow additional values (e.g., `created_at`) in a minor release; existing `relevance` and `mtime` semantics MUST remain stable.
