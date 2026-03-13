---
name: review-go
description: Run a comprehensive Go code review on git-changed files using the golang-pro agent with Context7 library docs
disable-model-invocation: true
---

# Go Code Review Skill

You are performing a comprehensive Go code review. Follow these steps exactly.

## Step 1: Identify Changed Go Files

Run these commands to find changed `.go` files:

```bash
git diff --name-only HEAD -- '*.go'
git diff --cached --name-only -- '*.go'
```

Combine and deduplicate the results. Exclude `_test.go` files from the primary review list but keep them for the test quality review area.

If NO changed `.go` files are found, use `AskUserQuestion` to ask the user which Go files or directories to review. Do not proceed without files.

## Step 2: Pre-fetch Library Documentation via Context7 MCP

This step is critical — the `golang-pro` agent does NOT have Context7 MCP tools, so you must fetch docs now and pass them along.

1. Read `go.mod` to identify third-party dependencies (non-stdlib, non-internal modules).
2. Scan the changed Go files for `import` statements. Identify which third-party libraries are actually used.
3. For each non-trivial third-party library found (skip stdlib, skip trivial/obvious ones):
   a. Call `mcp__context7__resolve-library-id` with the library's module path (e.g., `github.com/spf13/cobra`)
   b. If resolved, call `mcp__context7__get-library-docs` with the resolved ID and a topic query relevant to how the library is used in the changed code
4. Collect the fetched documentation snippets — you will pass them to the golang-pro agent.

Limit to the 3-5 most important libraries to avoid overwhelming the review context.

## Step 3: Launch golang-pro Agent for Review

Use the `Task` tool with `subagent_type: "golang-pro"` to run the review. Pass a prompt that includes:

- The list of changed `.go` file paths (both source and test files)
- The pre-fetched library documentation from Step 2
- The review instructions below

### Review Prompt Template

Construct the prompt as follows:

```
You are reviewing Go code changes. Read each file listed below and produce a structured review.

## Files to Review
{list of file paths}

## Library Documentation Context
{pre-fetched Context7 docs for third-party libraries used in these files}

## Review Areas

Analyze each file across these 8 areas:

1. **Go Idioms & Code Quality** — Non-idiomatic patterns, naming violations, struct design issues
2. **Error Handling** — Swallowed errors, missing error checks, incorrect error wrapping, sentinel vs type errors
3. **Memory & Allocations** — Unnecessary allocations, missing pre-allocation for known-size slices, string concatenation in loops. Only flag practical issues, no micro-optimizations.
4. **Performance Patterns** — Unoptimal algorithms, unnecessary work in hot paths, missing caching where obvious
5. **Dead Code** — Unused functions, unreachable branches, commented-out code left behind
6. **Code Duplication** — Repeated logic that would clearly benefit from extraction (3+ occurrences or complex blocks). Do NOT suggest extraction for simple 2-3 line patterns.
7. **Test Quality** — Missing edge cases, brittle assertions, test isolation issues, missing error case tests
8. **Library Usage** — Incorrect or suboptimal use of stdlib and third-party libraries. Use the library documentation provided above to verify correct API usage.

## Review Rules

- Be minimal and optimal — don't overcomplicate suggestions
- Be context-specific — no generic advice, only what matters for THIS code
- Simple > complex — never suggest an overengineered solution
- Every finding MUST include a file:line reference
- If a review area has no findings, omit it entirely
- Do NOT suggest adding comments, docstrings, or type annotations unless there is a clear correctness issue
- Do NOT suggest error handling for impossible scenarios
- Do NOT suggest abstractions for one-time code

## Output Format

Group findings by review area. For each finding:

**[Area Name]**
- `file.go:42` — Description of the issue and suggested fix

If everything looks good, say so briefly. Do not pad the review with praise or filler.
```

## Step 4: Present Results

The golang-pro agent will return the structured review. Present it directly to the user. Do not add extra commentary or re-analyze — the review stands as-is.
