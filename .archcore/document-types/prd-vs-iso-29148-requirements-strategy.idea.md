---
title: "PRD as Simple Alternative to ISO 29148 Requirements Cascade"
status: draft
---

## Idea

Archcore should support two requirements engineering approaches side by side: a **simple path** (PRD) and a **detailed path** (ISO 29148 cascade: BRS → StRS → SyRS → SRS). Users choose based on project complexity, team size, and regulatory needs. Archcore should actively guide this choice.

### Problem / Opportunity

- The current PRD type is a pragmatic hybrid that covers business context, stakeholder needs, and functional requirements in a single document. This works well for most projects.
- ISO/IEC/IEEE 29148:2018 defines four specialized requirements specification types (BRS, StRS, SyRS, SRS) that decompose requirements into progressively detailed levels. This is essential for complex, regulated, or multi-team systems.
- Without guidance, users won't know when to use PRD vs. the ISO cascade — leading to either over-engineering simple projects or under-specifying complex ones.

## Value

### For Users

- **Simple projects** stay simple: one PRD covers everything needed. No pressure to create 4 separate documents.
- **Complex projects** get proper structure: the ISO cascade ensures traceability from business objectives through to testable software specs.
- Clear guidance on when to upgrade from PRD to the full cascade.

### For AI Agents

- Agents can read a PRD and immediately start implementing for simple features.
- For complex features, agents follow the cascade: BRS provides alignment context, StRS provides operational scenarios, SyRS provides system boundaries, SRS provides per-endpoint specifications they can directly translate to code and tests.
- The `implements` relation type naturally connects: `strs implements brs`, `syrs implements strs`, `srs implements syrs`, `plan implements srs`.

### For Business

- Positions archcore as a tool that scales from indie developer to enterprise — the same platform supports both simple and ISO-compliant requirements workflows.

## Possible Implementation

### Two Paths, One Platform

**Simple path (PRD)**:

- Single document covering vision, problem, requirements, and solution overview
- Best for: individual features, small teams, rapid prototyping, internal tools
- The PRD template remains unchanged — it already works well

**Detailed path (ISO 29148 cascade)**:

- Four documents with formal traceability between levels
- Best for: regulated systems, multi-team projects, hardware+software integration, external contracts, complex distributed systems
- Each level adds precision: BRS (why) → StRS (who needs what) → SyRS (system behavior) → SRS (software behavior)

### How PRD Relates to ISO Types

The PRD is essentially a condensed version that merges all four ISO levels:

| PRD Section                                   | Equivalent ISO Type | ISO Section                                            |
| --------------------------------------------- | ------------------- | ------------------------------------------------------ |
| Vision, Strategic Alignment                   | BRS                 | §9.3.7 Mission, Goals and Objectives                   |
| Problem Statement, Target Users, User Stories | StRS                | §9.4.15 User Requirements, §9.4.16 Operational Concept |
| Non-Functional Requirements, Constraints      | SyRS                | §9.5.5-§9.5.13 System Requirements                     |
| Functional Requirements (P0/P1/P2)            | SRS                 | §9.6.10-§9.6.12 Specified Requirements                 |
| Solution Overview                             | SyRS/SRS            | §9.5.4 System Overview, §9.6.4 Product Perspective     |

Users can start with a PRD and later decompose it into ISO documents when complexity demands it.

### Archcore Guidance System (Future)

Archcore could suggest the appropriate approach based on signals:

**Suggest ISO cascade when:**

- Project has multiple stakeholder classes with conflicting priorities
- System integrates with external systems (needs formal interface specs)
- Regulatory or compliance requirements exist
- Multiple teams will implement different parts
- Requirements need formal traceability (defense, medical, automotive)

**Suggest PRD when:**

- Single team, single product
- Internal tool or feature
- Rapid iteration expected
- No regulatory requirements
- Requirements are well-understood

This could be implemented as:

- A prompt during `archcore create` when user selects requirements-related types
- A `doctor` check that suggests decomposing large PRDs into ISO cascade
- MCP server instructions that guide agents to recommend the right approach

### User Control

Users should be able to:

- Mix approaches freely (some features use PRD, others use full cascade)
- Start with PRD and gradually decompose into ISO types as complexity grows
- Use partial cascade (e.g., only BRS + SRS, skipping StRS and SyRS)
- Configure a project-level preference in `settings.json` (e.g., `"requirements_approach": "simple" | "iso-29148"`)

## Risks and Constraints

### Potential Risks

- **Type confusion**: 7 vision types (prd + idea + plan + brs + strs + syrs + srs) increase cognitive load. Mitigated by strong disambiguation rules in MCP instructions.
- **Over-engineering**: Teams might feel pressured to use the full cascade when PRD would suffice. Mitigated by clearly positioning PRD as the default/recommended approach.
- **Agent selection errors**: Agents might create ISO documents when PRD would be more appropriate. Mitigated by structural disambiguation cues (section-based, not abstract definitions).

### Known Constraints

- ISO 29148:2018 is a copyrighted standard — templates should be inspired by the structure but not reproduce copyrighted text verbatim.
- The guidance system (suggesting simple vs. detailed approach) is a future enhancement, not part of the initial type implementation.
- Traceability between ISO documents relies on the existing `implements` relation type — no new infrastructure needed.

## Next Steps

- [ ] Implement BRS, StRS, SyRS, SRS document types (templates, MCP descriptions, disambiguation rules)
- [ ] Add disambiguation rules to MCP server instructions and create_document tool
- [ ] Update categories-and-document-types.doc.md with both approaches documented
- [ ] Future: Add guidance prompts in CLI suggesting simple vs. detailed approach
- [ ] Future: Add `requirements_approach` setting to settings.json
- [ ] Future: Add `doctor` check suggesting ISO decomposition for large PRDs

## Related Materials

- ISO/IEC/IEEE 29148:2018 — Systems and software engineering — Life cycle processes — Requirements engineering
- Existing PRD template in @templates/templates.go (lines 765-951)
- MCP server instructions in @internal/mcp/server.go (lines 11-90)
