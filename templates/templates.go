package templates

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type DocumentType string

const (
	TypeADR      DocumentType = "adr"
	TypeRFC      DocumentType = "rfc"
	TypeRule     DocumentType = "rule"
	TypeGuide    DocumentType = "guide"
	TypeDoc      DocumentType = "doc"
	TypeTaskType DocumentType = "task-type"
	TypeCPAT     DocumentType = "cpat"
	TypePRD      DocumentType = "prd"
	TypeIdea     DocumentType = "idea"
	TypePlan     DocumentType = "plan"
)

const (
	CategoryVision     = "vision"
	CategoryKnowledge  = "knowledge"
	CategoryExperience = "experience"
)

const (
	StatusDraft    = "draft"
	StatusAccepted = "accepted"
	StatusRejected = "rejected"
)

// SkipFiles are non-document meta files that live in .archcore/ and should be
// skipped during scanning, validation, and sync operations.
var SkipFiles = map[string]bool{
	"settings.json":    true,
	".sync-state.json": true,
}

// ValidStatuses returns all valid document status strings.
func ValidStatuses() []string {
	return []string{StatusDraft, StatusAccepted, StatusRejected}
}

// IsValidStatus checks whether the given string is a valid document status.
func IsValidStatus(s string) bool {
	switch s {
	case StatusDraft, StatusAccepted, StatusRejected:
		return true
	}
	return false
}

var categoryMap = map[DocumentType]string{
	TypePRD:  CategoryVision,
	TypeIdea: CategoryVision,
	TypePlan: CategoryVision,

	TypeADR:     CategoryKnowledge,
	TypeRFC:     CategoryKnowledge,
	TypeRule:    CategoryKnowledge,
	TypeGuide:   CategoryKnowledge,
	TypeDoc: CategoryKnowledge,

	TypeTaskType: CategoryExperience,
	TypeCPAT:     CategoryExperience,
}

// CategoryForType returns the category directory for a document type.
func CategoryForType(docType DocumentType) string {
	if cat, ok := categoryMap[docType]; ok {
		return cat
	}
	return CategoryKnowledge
}

// ValidTypes returns all valid document type strings.
func ValidTypes() []string {
	return []string{
		string(TypeADR),
		string(TypeRFC),
		string(TypeRule),
		string(TypeGuide),
		string(TypeDoc),
		string(TypeTaskType),
		string(TypeCPAT),
		string(TypePRD),
		string(TypeIdea),
		string(TypePlan),
	}
}

// TypesByCategory returns types grouped by category.
func TypesByCategory() map[string][]string {
	result := map[string][]string{}
	for dt, cat := range categoryMap {
		result[cat] = append(result[cat], string(dt))
	}
	return result
}

// IsValidType checks whether the given string is a valid document type.
func IsValidType(t string) bool {
	_, ok := categoryMap[DocumentType(t)]
	return ok
}

// ExtractDocType extracts the type from a filename like "use-postgres.adr.md".
func ExtractDocType(filename string) string {
	name := strings.TrimSuffix(filename, ".md")
	parts := strings.Split(name, ".")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return ""
}

// ExtractSlug extracts the slug from a filename like "use-postgres.adr.md".
func ExtractSlug(filename string) string {
	name := strings.TrimSuffix(filename, ".md")
	parts := strings.Split(name, ".")
	if len(parts) >= 2 {
		return strings.Join(parts[:len(parts)-1], ".")
	}
	return name
}

// SplitDocument splits raw document bytes into frontmatter fields and body.
// It returns the title, status, and the markdown body after the closing "---".
func SplitDocument(data []byte) (title, status, body string) {
	s := strings.ReplaceAll(string(data), "\r\n", "\n")
	if !strings.HasPrefix(s, "---\n") {
		return "", "", s
	}

	end := strings.Index(s[4:], "\n---\n")
	if end == -1 {
		return "", "", s
	}
	end += 4 // adjust for the offset

	frontmatter := s[4:end]
	for _, line := range strings.Split(frontmatter, "\n") {
		if strings.HasPrefix(line, "title: ") {
			title = strings.TrimPrefix(line, "title: ")
			// Remove surrounding quotes added by buildDocumentFile (%q).
			if len(title) >= 2 && title[0] == '"' && title[len(title)-1] == '"' {
				if unq, err := strconv.Unquote(title); err == nil {
					title = unq
				}
			}
		}
		if strings.HasPrefix(line, "status: ") {
			status = strings.TrimPrefix(line, "status: ")
		}
	}

	body = s[end+5:] // skip past "\n---\n"
	body = strings.TrimPrefix(body, "\n")

	return title, status, body
}

// WalkArchcoreFiles walks archcoreDir recursively, calling fn for each .md
// document file found. It skips hidden directories, non-.md files, and known
// meta files (settings.json, .sync-state.json).
func WalkArchcoreFiles(archcoreDir string, fn func(path string, d fs.DirEntry) error) error {
	return filepath.WalkDir(archcoreDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		name := d.Name()

		// Skip hidden directories (but not .archcore itself).
		if d.IsDir() && strings.HasPrefix(name, ".") && path != archcoreDir {
			return filepath.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		// Skip non-.md files and known meta files.
		if !strings.HasSuffix(name, ".md") || SkipFiles[name] {
			return nil
		}

		return fn(path, d)
	})
}

func GenerateTemplate(documentType DocumentType) string {
	switch documentType {
	case TypeADR:
		return generateADRTemplate()
	case TypeRFC:
		return generateRFCTemplate()
	case TypeRule:
		return generateRuleTemplate()
	case TypeGuide:
		return generateGuideTemplate()
	case TypeDoc:
		return generateDocTemplate()
	case TypeTaskType:
		return generateTaskTypeTemplate()
	case TypeCPAT:
		return generateCPATTemplate()
	case TypePRD:
		return generatePRDTemplate()
	case TypeIdea:
		return generateIdeaTemplate()
	case TypePlan:
		return generatePlanTemplate()
	default:
		return generateDocTemplate()
	}
}

func generateADRTemplate() string {
	return `## Context

Describe the context and problem statement that motivates this decision.

### Current State

- What is the current situation?
- What constraints exist?
- What pain points are we experiencing?

### Problem Statement

Clear, concise description of the problem that needs to be solved.

## Decision

State the decision that was made.

### Rationale

Explain why this decision was chosen over alternatives:

- Key factors that influenced the decision
- Trade-offs that were considered
- Assumptions that were made

## Alternatives Considered

### Alternative 1: [Name]

- Description of the alternative
- Why it was not chosen

### Alternative 2: [Name]

- Description of the alternative
- Why it was not chosen

## Consequences

### Positive

- Benefit 1: Description
- Benefit 2: Description
- Benefit 3: Description

### Negative

- Trade-off 1: Description and mitigation
- Trade-off 2: Description and mitigation

### Risks

- Risk 1: Description and mitigation strategy
- Risk 2: Description and mitigation strategy

## Implementation Notes

Key implementation considerations:

- Migration path (if applicable)
- Dependencies affected
- Timeline considerations

## References

- Link to relevant documentation
- Link to discussions or RFCs
- Related ADRs
`
}

func generateRFCTemplate() string {
	return `## Summary

One-paragraph summary of the proposal.

## Motivation

### Problem Statement

What problem does this proposal solve? Be specific about the pain points.

### Goals

- Goal 1: Description
- Goal 2: Description

### Non-Goals

What is explicitly out of scope:

- Non-goal 1: Why it's excluded
- Non-goal 2: Why it's excluded

## Detailed Design

### Overview

High-level description of the proposed solution.

### Architecture

Describe the architectural changes:

- Components affected
- New components introduced
- Data flow changes

### API Changes

Describe any API changes:

` + "```" + `
// Example API or interface changes
` + "```" + `

### Data Model

Describe data model changes if applicable:

- New entities
- Schema changes
- Migration requirements

### Implementation Details

Step-by-step implementation approach:

1. Phase 1: Description
2. Phase 2: Description
3. Phase 3: Description

## Drawbacks

### Technical Drawbacks

- Drawback 1: Description and severity
- Drawback 2: Description and severity

### Operational Drawbacks

- Drawback 1: Description
- Drawback 2: Description

## Alternatives

### Alternative 1: [Name]

- Description of approach
- Pros and cons
- Why not chosen

### Alternative 2: [Name]

- Description of approach
- Pros and cons
- Why not chosen

## Security Considerations

- Security implication 1
- Security implication 2
- Mitigation strategies

## Testing Strategy

- Unit testing approach
- Integration testing approach
- Performance testing requirements

## Rollout Plan

### Phases

1. Phase 1: Description and success criteria
2. Phase 2: Description and success criteria
3. Phase 3: Description and success criteria

### Rollback Plan

How to rollback if issues are discovered.

## Unresolved Questions

- Question 1: Context and potential answers
- Question 2: Context and potential answers

## Implementation Plan

- [ ] Step 1: Description
- [ ] Step 2: Description
- [ ] Step 3: Description
- [ ] Step 4: Description
- [ ] Step 5: Description

## References

- Related RFCs or ADRs
- External documentation
- Research or benchmarks
`
}

func generateRuleTemplate() string {
	return `## Description

Brief description of what this rule covers and why it exists.

## Rule

State the rule clearly as imperative statements:

1. [Rule as imperative statement]
2. [Rule as imperative statement]
3. [Rule as imperative statement]

## Rationale

Why this rule exists and what problems it prevents.

## Examples

### Good

` + "```" + `
// Example of correct usage
` + "```" + `

` + "```" + `
// Another example of correct usage
` + "```" + `

### Bad

` + "```" + `
// Example of incorrect usage
` + "```" + `

` + "```" + `
// Another example of incorrect usage
` + "```" + `

## Exceptions

- Exception 1: When this rule does not apply
- Exception 2: When this rule does not apply

## Enforcement

How this rule is enforced:

- Enforcement method 1: Description
- Enforcement method 2: Description

## References

- Link to related ADR/RFC
- Link to related documentation
`
}

func generateGuideTemplate() string {
	return `## Overview

Brief overview of what this guide covers and what the reader will accomplish.

### Target Audience

Who should read this guide:

- Role 1: What they'll learn
- Role 2: What they'll learn

### Time Estimate

Approximate time to complete: X minutes

## Prerequisites

### Required Knowledge

- Prerequisite 1: Brief description
- Prerequisite 2: Brief description

### Required Tools

- Tool 1: Version and installation link
- Tool 2: Version and installation link

### Required Access

- Access 1: How to obtain
- Access 2: How to obtain

## Steps

### Step 1: [Title]

Description of what this step accomplishes.

` + "```" + `
# Commands or code for this step
` + "```" + `

**Expected result:** What you should see after this step.

### Step 2: [Title]

Description of what this step accomplishes.

` + "```" + `
# Commands or code for this step
` + "```" + `

**Expected result:** What you should see after this step.

### Step 3: [Title]

Description of what this step accomplishes.

` + "```" + `
# Commands or code for this step
` + "```" + `

**Expected result:** What you should see after this step.

### Step 4: [Title]

Description of what this step accomplishes.

**Expected result:** What you should see after this step.

## Verification

How to verify everything is working correctly:

1. Verification step 1
2. Verification step 2
3. Verification step 3

## Common Issues

### Issue 1: [Error message or symptom]

**Cause:** Why this happens

**Solution:**

` + "```" + `
# Commands to fix
` + "```" + `

### Issue 2: [Error message or symptom]

**Cause:** Why this happens

**Solution:** Steps to resolve

### Issue 3: [Error message or symptom]

**Cause:** Why this happens

**Solution:** Steps to resolve

## Next Steps

What to do after completing this guide:

- Next step 1: Link or description
- Next step 2: Link or description

## Related Resources

- Link to related guide
- Link to reference documentation
- Link to troubleshooting guide
`
}

func generateDocTemplate() string {
	return `## Overview

Brief description of what this document covers and its purpose.

### Scope

What this document includes:

- Topic 1
- Topic 2
- Topic 3

What this document does not cover:

- Out of scope 1
- Out of scope 2

## Content

### Section 1: [Title]

Main content for this section.

Key points:

- Point 1: Description
- Point 2: Description
- Point 3: Description

### Section 2: [Title]

Main content for this section.

Key points:

- Point 1: Description
- Point 2: Description

### Section 3: [Title]

Main content for this section.

## Examples

### Example 1: [Title]

Context for when to use this example.

` + "```" + `
// Code or configuration example
` + "```" + `

### Example 2: [Title]

Context for when to use this example.

` + "```" + `
// Code or configuration example
` + "```" + `

## Best Practices

- Best practice 1: Description
- Best practice 2: Description
- Best practice 3: Description

## FAQ

### Q: Common question 1?

Answer to the question.

### Q: Common question 2?

Answer to the question.

## Related Resources

- Link to related documentation
- Link to API reference
- Link to tutorials
`
}

func generateTaskTypeTemplate() string {
	return `## What

What this typical task covers and what the end result looks like.

## When to Use

Use when:

- Condition 1
- Condition 2

Do NOT use when:

- Condition: use [alternative] instead

## Steps

1. Step one — what to do and where (@path/to/file)
2. Step two — what to do next
3. Step three — final checks

## Example

` + "```" + `
// Small code snippet or @-reference to a real implementation
` + "```" + `

## Things to Watch Out For

- Pitfall or gotcha 1
- Edge case to keep in mind
- Common mistake to avoid
`
}

func generateCPATTemplate() string {
	return `## What Changed

The pattern, convention, or approach that changed.

## Why

What problem the old way caused and why the change was needed.

## Before

` + "```" + `
// Old pattern
` + "```" + `

## After

` + "```" + `
// New pattern
` + "```" + `

## Scope

Affected files and modules:

- @path/to/affected/module
- @path/to/another/file

## Notes

- Exceptions where the old pattern is still acceptable
- Migration notes or timeline
`
}

func generatePRDTemplate() string {
	return `## Vision

### Product Vision Statement

One-sentence vision of what this product/feature will achieve.

### Strategic Alignment

How this aligns with company/team goals:

- Strategic goal 1: How this product supports it
- Strategic goal 2: How this product supports it

## Problem Statement

### Current State

Describe the current situation and its limitations:

- Pain point 1: Description and impact
- Pain point 2: Description and impact
- Pain point 3: Description and impact

### Target Users

| User Segment | Description | Key Needs |
|--------------|-------------|-----------|
| Segment 1 | Description | Needs |
| Segment 2 | Description | Needs |

### User Stories

As a [user type], I want [goal] so that [benefit].

- User story 1
- User story 2
- User story 3

## Goals and Success Metrics

### Goals

| Goal | Description | Priority |
|------|-------------|----------|
| Goal 1 | Description | P0/P1/P2 |
| Goal 2 | Description | P0/P1/P2 |
| Goal 3 | Description | P0/P1/P2 |

### Success Metrics

| Metric | Current | Target | Timeline |
|--------|---------|--------|----------|
| Metric 1 | Value | Value | Date |
| Metric 2 | Value | Value | Date |

### Non-Goals

Explicitly out of scope for this version:

- Non-goal 1: Reason
- Non-goal 2: Reason

## Requirements

### Functional Requirements

#### P0 (Must Have)

| ID | Requirement | Acceptance Criteria |
|----|-------------|---------------------|
| FR-001 | Description | Criteria |
| FR-002 | Description | Criteria |
| FR-003 | Description | Criteria |

#### P1 (Should Have)

| ID | Requirement | Acceptance Criteria |
|----|-------------|---------------------|
| FR-101 | Description | Criteria |
| FR-102 | Description | Criteria |

#### P2 (Nice to Have)

| ID | Requirement | Acceptance Criteria |
|----|-------------|---------------------|
| FR-201 | Description | Criteria |

### Non-Functional Requirements

| Category | Requirement | Target |
|----------|-------------|--------|
| Performance | Description | Metric |
| Scalability | Description | Metric |
| Security | Description | Metric |
| Reliability | Description | Metric |

## Constraints

### Technical Constraints

- Constraint 1: Description and impact
- Constraint 2: Description and impact

### Business Constraints

- Constraint 1: Description and impact
- Constraint 2: Description and impact

### Dependencies

| Dependency | Type | Owner | Status |
|------------|------|-------|--------|
| Dependency 1 | Internal/External | Team | Status |
| Dependency 2 | Internal/External | Team | Status |

## Solution Overview

### Proposed Approach

High-level description of the proposed solution.

### Key Components

- Component 1: Purpose
- Component 2: Purpose
- Component 3: Purpose

### User Experience

Key UX considerations:

- UX consideration 1
- UX consideration 2

### Technical Considerations

Key technical considerations (defer details to RFC/design doc):

- Consideration 1
- Consideration 2

## Risks and Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Risk 1 | High/Med/Low | High/Med/Low | Strategy |
| Risk 2 | High/Med/Low | High/Med/Low | Strategy |

## Timeline

### Milestones

| Milestone | Target Date | Description |
|-----------|-------------|-------------|
| Milestone 1 | Date | Description |
| Milestone 2 | Date | Description |
| Milestone 3 | Date | Description |

### Phases

- Phase 1: Scope and timeline
- Phase 2: Scope and timeline
- Phase 3: Scope and timeline

## Open Questions

| Question | Context | Decision Owner | Due Date |
|----------|---------|----------------|----------|
| Question 1 | Context | Owner | Date |
| Question 2 | Context | Owner | Date |

## Appendix

### Glossary

| Term | Definition |
|------|------------|
| Term 1 | Definition |
| Term 2 | Definition |

### References

- Reference 1: Link
- Reference 2: Link
- Related PRDs: Link
`
}

func generateIdeaTemplate() string {
	return `## Idea

Describe the core idea in 2-3 sentences.

### Problem / Opportunity

- What problem does it solve?
- What opportunity does it open?

## Value

### For Users

### For Business

### For Team

## Possible Implementation

### Technical Approach

### Integrations

## Risks and Constraints

### Potential Risks

### Known Constraints

## Next Steps

- [ ] Step 1
- [ ] Step 2

## Related Materials
`
}

func generatePlanTemplate() string {
	return `## Goal

Describe the desired outcome in one sentence.

### Context

- What motivated creating this plan?

## Tasks

### Phase 1: [Name]

- [ ] Task 1
- [ ] Task 2

### Phase 2: [Name]

- [ ] Task 1

## Acceptance Criteria

- [ ] Criterion 1
- [ ] Criterion 2

## Dependencies

| Dependency | Type | Status |
|------------|------|--------|

## Notes
`
}
