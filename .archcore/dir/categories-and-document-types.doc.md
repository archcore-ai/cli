---
title: "Categories and Document Types"
status: accepted
---

## Overview

Archcore organizes documents into three virtual categories — **vision**, **knowledge**, and **experience**. The category is derived from the document type in the filename (`slug.type.md`), not from the directory path.

## Vision

Documents that describe the future: what we want to build and why.

| Type   | Purpose                                                    |
| ------ | ---------------------------------------------------------- |
| `prd`  | Product requirements — goals, scope, acceptance criteria   |
| `idea` | A concept worth exploring — problem, value, rough approach |
| `plan` | A concrete implementation plan with phased tasks           |

## Knowledge

Documents that capture what we know: decisions, standards, contracts, and reference material.

| Type      | Purpose                                                                                                     |
| --------- | ----------------------------------------------------------------------------------------------------------- |
| `adr`     | A technical decision that has been made, with context and alternatives                                      |
| `rfc`     | A proposal open for review before a decision is made                                                        |
| `rule`    | A mandatory standard — imperative statements with good/bad examples                                         |
| `guide`   | Step-by-step instructions for completing a task                                                             |
| `spec`    | Canonical normative contract — behavior, constraints, invariants, conformance for a specific technical boundary |
| `doc`     | Non-behavioral reference — tables, registries, glossaries, component lists                                  |

## Experience

Documents that encode proven patterns and lessons from practice.

| Type        | Purpose                                                                             |
| ----------- | ----------------------------------------------------------------------------------- |
| `task-type` | A proven workflow for a recurring implementation task — steps, examples, pitfalls   |
| `cpat`      | A code pattern change — how and why a convention or approach changed (was → became) |

## Choosing the Right Type

- **rule vs doc** — rule prescribes behavior ("Always do X") with enforcement. doc describes what exists (tables, registries). Descriptive, non-behavioral content → doc.
- **adr vs rfc** — adr = decision already final. rfc = proposal open for feedback.
- **guide vs doc** — guide has sequential steps to follow. doc is non-sequential reference to look up.
- **spec vs doc** — spec defines a canonical normative contract for a concrete technical boundary (behavior, constraints, invariants, conformance). doc describes what exists without normative requirements. Normative contract → spec; structural reference → doc.
- **spec vs rule** — spec is a technical contract for one component. rule is a cross-cutting team standard. Scoped to a named artifact → spec; applied team-wide → rule.
- **spec vs adr** — spec is the living canonical truth (present-tense: "it works this way"). adr is the decision record (past-tense: "we chose this because"). Both may exist for the same component.
- **task-type vs guide** — task-type is a reusable pattern for a class of tasks (e.g., "how to create a UI-kit component"). guide is instructions for a specific one-time procedure.
- **cpat vs adr** — cpat focuses on a code pattern change with before/after examples. adr records a broader architectural decision with alternatives and consequences.
