---
title: "Implementation Plan: Add ISO 29148 Document Types (BRS, StRS, SyRS, SRS)"
status: draft
---

## Goal

Add all four ISO/IEC/IEEE 29148:2018 requirements specification types — BRS, StRS, SyRS, SRS — to the vision category, enabling a detailed requirements engineering cascade alongside the existing simple PRD approach.

### Context

- Archcore currently has 10 document types (3 vision, 5 knowledge, 2 experience)
- ISO 29148 defines a cascade: BRS (business mission) → StRS (stakeholder needs) → SyRS (system behavior) → SRS (software specs)
- PRD remains the simple/default approach; ISO types are for detailed elaboration
- See companion idea: @.archcore/document-types/prd-vs-iso-29148-requirements-strategy.idea.md

## Tasks

### Phase 1: Core Type System

- [ ] Add constants to @templates/templates.go: `TypeBRS = "brs"`, `TypeStRS = "strs"`, `TypeSyRS = "syrs"`, `TypeSRS = "srs"`
- [ ] Add all 4 to `categoryMap` as `CategoryVision`
- [ ] Extend `ValidTypes()` return slice with all 4 slugs
- [ ] Add 4 `case` arms in `GenerateTemplate()` switch
- [ ] Implement `generateBRSTemplate()` — sections per ISO §9.3: Business Purpose/Scope, Business Overview (Stakeholders, Environment), Mission/Goals/Objectives, Business Model (Processes, Policies/Rules), Business Constraints, High-Level Operational Concept (Scenarios, Modes), Success Criteria, Assumptions/Dependencies
- [ ] Implement `generateStRSTemplate()` — sections per ISO §9.4: Purpose/Scope, Stakeholder Classes (Priorities), Operational Concept/ConOps (Current Ops, Proposed Ops, Scenarios), Stakeholder Requirements (User, Usability, Quality), Operational Policies/Rules, Operational Constraints (Modes/States), Compliance/Regulatory, Project Constraints, Traceability
- [ ] Implement `generateSyRSTemplate()` — sections per ISO §9.5: System Purpose/Scope (Boundary), System Overview, System Requirements (Functional, Usability, Performance, Security, Reliability), System Interfaces (User, System-to-System, Hardware), System Operations (Modes/States, Physical/Environmental), Policy/Regulation, Life Cycle Sustainment, Verification Approach, Traceability
- [ ] Implement `generateSRSTemplate()` — sections per ISO §9.6: Purpose/Scope (Component, Boundaries), Product Perspective (Functions, User Characteristics, Limitations), Software Requirements (Functional, Behavioral, Error Handling), External Interfaces (API Endpoints, Internal), Data Requirements (Logical Database, Data Flows), Performance, Design Constraints (Standards Compliance), Software Quality Attributes, Verification Matrix, Traceability

### Phase 2: MCP Integration

- [ ] Update @internal/mcp/server.go `mcpServerInstructions`: add types to vision listing, add 4 WHEN TO CREATE lines, add 6 TYPE SELECTION RULES
- [ ] Update @internal/mcp/tools/create_document.go tool description: add 4 type entries with required sections, add 6 disambiguation rules
- [ ] Update @internal/mcp/tools/list_documents.go: add new types to `types` parameter description

### Phase 3: Tests

- [ ] Add 4 entries to template table-driven tests in @templates/templates_test.go
- [ ] Add 4 entries to AllTypes test in @internal/mcp/tools/create_document_test.go
- [ ] Update ValidTypes count assertions and TypesByCategory vision count (3 → 7)

### Phase 4: Documentation

- [ ] Update @.archcore/dir/categories-and-document-types.doc.md: add 4 types to vision table with ISO references, add 6 disambiguation entries

## Acceptance Criteria

- [ ] `go build -o archcore .` succeeds
- [ ] `go test ./...` passes (all existing + new tests)
- [ ] `./archcore mcp` shows all 4 new types in MCP instructions
- [ ] Creating a document of each new type via MCP renders the correct ISO-aligned template
- [ ] Disambiguation rules clearly distinguish brs/strs/syrs/srs from existing prd/adr types
- [ ] Traceability sections in strs/syrs/srs templates enable the ISO requirement cascade

## Dependencies

| Dependency | Type | Status |
|------------|------|--------|
| ISO/IEC/IEEE 29148:2018 standard (reference) | External | Available (PDF in repo) |
| Existing template system in templates.go | Internal | Ready |
| MCP server instructions in server.go | Internal | Ready |

## Notes

### Disambiguation Strategy (6 Rules)

These structural cues help agents choose the right type:

1. **brs vs prd**: brs has ONLY business objectives/outcomes, no user stories or solution. prd has user stories, functional requirements, solution overview.
2. **strs vs prd**: strs groups requirements PER STAKEHOLDER CLASS with ConOps. prd lists by priority (P0/P1/P2).
3. **syrs vs adr**: syrs defines WHOLE SYSTEM BOUNDARY and interface contracts. adr records a single decision.
4. **srs vs prd**: srs has PER-ENDPOINT/PER-FUNCTION requirements with verification matrix. prd has product-level requirements.
5. **brs vs strs**: brs = WHY (business outcomes, technology-agnostic). strs = WHAT stakeholders need (operational scenarios, solution-aware).
6. **syrs vs srs**: syrs = WHOLE SYSTEM boundary. srs = SINGLE COMPONENT's detailed behavior.

### ISO 29148 Section References

- BRS content: ISO §9.3 (§9.3.2–§9.3.19)
- StRS content: ISO §9.4 (§9.4.1–§9.4.19)
- SyRS content: ISO §9.5 (§9.5.1–§9.5.19)
- SRS content: ISO §9.6 (§9.6.1–§9.6.20)
