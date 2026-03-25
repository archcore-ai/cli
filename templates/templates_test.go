package templates

import (
	"strings"
	"testing"
)

func TestGenerateTemplate(t *testing.T) {
	tests := []struct {
		name         string
		documentType DocumentType
		wantContains []string
	}{
		{
			name:         "ADR template",
			documentType: TypeADR,
			wantContains: []string{"## Context", "## Decision", "## Consequences", "### Positive", "### Negative", "## References", "## Alternatives Considered"},
		},
		{
			name:         "RFC template",
			documentType: TypeRFC,
			wantContains: []string{"## Summary", "## Motivation", "## Detailed Design", "## Drawbacks", "## Alternatives", "## Unresolved Questions", "## Implementation Plan", "## Security Considerations"},
		},
		{
			name:         "Rule template",
			documentType: TypeRule,
			wantContains: []string{"## Description", "## Rule", "## Examples", "### Good", "### Bad", "## Exceptions", "## References", "## Enforcement"},
		},
		{
			name:         "Guide template",
			documentType: TypeGuide,
			wantContains: []string{"## Overview", "## Prerequisites", "## Steps", "### Step 1:", "### Step 2:", "### Step 3:", "## Common Issues", "## Related Resources", "## Verification"},
		},
		{
			name:         "Doc template",
			documentType: TypeDoc,
			wantContains: []string{"## Overview", "## Content", "## Examples", "## Related Resources", "## Best Practices", "## FAQ"},
		},
		{
			name:         "Task-Type template",
			documentType: TypeTaskType,
			wantContains: []string{"## What", "## When to Use", "## Steps", "## Example", "## Things to Watch Out For"},
		},
		{
			name:         "CPAT template",
			documentType: TypeCPAT,
			wantContains: []string{"## What Changed", "## Why", "## Before", "## After", "## Scope"},
		},
		{
			name:         "PRD template",
			documentType: TypePRD,
			wantContains: []string{"## Vision", "## Problem Statement", "## Goals and Success Metrics", "## Requirements", "## Constraints", "## Timeline"},
		},
		{
			name:         "Idea template",
			documentType: TypeIdea,
			wantContains: []string{"## Idea", "## Value", "## Possible Implementation", "## Risks and Constraints", "## Next Steps"},
		},
		{
			name:         "Plan template",
			documentType: TypePlan,
			wantContains: []string{"## Goal", "## Tasks", "## Acceptance Criteria", "## Dependencies", "## Notes"},
		},
		{
			name:         "Spec template",
			documentType: TypeSpec,
			wantContains: []string{"## Purpose", "## Scope", "## Authority", "## Subject", "## Contract Surface", "## Normative Behavior", "## Constraints", "## Invariants", "## Error Handling", "## Conformance"},
		},
		{
			name:         "MRD template",
			documentType: TypeMRD,
			wantContains: []string{"## Market Landscape", "## TAM / SAM / SOM", "## Competitive Analysis", "## Market Needs", "## Opportunity and Timing"},
		},
		{
			name:         "BRD template",
			documentType: TypeBRD,
			wantContains: []string{"## Business Objectives", "## Stakeholders", "## Business Rules and Constraints", "## Success Metrics and ROI", "## Dependencies"},
		},
		{
			name:         "URD template",
			documentType: TypeURD,
			wantContains: []string{"## User Personas", "## User Journeys", "## User Requirements", "## Usability Requirements", "## Acceptance Criteria"},
		},
		{
			name:         "Unknown type falls back to doc template",
			documentType: DocumentType("unknown"),
			wantContains: []string{"## Overview", "## Content", "## Examples", "## Related Resources"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := GenerateTemplate(tt.documentType)

			if got == "" {
				t.Errorf("GenerateTemplate(%q) returned empty string", tt.documentType)
				return
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("GenerateTemplate(%q) missing expected section: %q", tt.documentType, want)
				}
			}
		})
	}
}

func TestGenerateADRTemplate(t *testing.T) {
	template := generateADRTemplate()

	requiredSections := []string{
		"## Context",
		"## Decision",
		"## Consequences",
		"### Positive",
		"### Negative",
		"## References",
		"## Alternatives Considered",
		"### Rationale",
	}

	for _, section := range requiredSections {
		if !strings.Contains(template, section) {
			t.Errorf("ADR template missing section: %q", section)
		}
	}

	if !strings.Contains(template, "Describe") {
		t.Error("ADR template should include guidance text")
	}
}

func TestGenerateRFCTemplate(t *testing.T) {
	template := generateRFCTemplate()

	requiredSections := []string{
		"## Summary",
		"## Motivation",
		"## Detailed Design",
		"## Drawbacks",
		"## Alternatives",
		"## Unresolved Questions",
		"## Implementation Plan",
		"## Security Considerations",
		"## Testing Strategy",
		"## Rollout Plan",
	}

	for _, section := range requiredSections {
		if !strings.Contains(template, section) {
			t.Errorf("RFC template missing section: %q", section)
		}
	}

	if !strings.Contains(template, "- [ ]") {
		t.Error("RFC template should include checkboxes in Implementation Plan")
	}
}

func TestGenerateRuleTemplate(t *testing.T) {
	template := generateRuleTemplate()

	requiredSections := []string{
		"## Description",
		"## Rule",
		"## Examples",
		"### Good",
		"### Bad",
		"## Exceptions",
		"## References",
		"## Enforcement",
		"## Rationale",
	}

	for _, section := range requiredSections {
		if !strings.Contains(template, section) {
			t.Errorf("Rule template missing section: %q", section)
		}
	}

	if !strings.Contains(template, "```") {
		t.Error("Rule template should include code block markers")
	}
}

func TestGenerateGuideTemplate(t *testing.T) {
	template := generateGuideTemplate()

	requiredSections := []string{
		"## Overview",
		"## Prerequisites",
		"## Steps",
		"### Step 1:",
		"### Step 2:",
		"### Step 3:",
		"### Step 4:",
		"## Common Issues",
		"## Related Resources",
		"## Verification",
		"## Next Steps",
	}

	for _, section := range requiredSections {
		if !strings.Contains(template, section) {
			t.Errorf("Guide template missing section: %q", section)
		}
	}

	stepCount := strings.Count(template, "### Step")
	if stepCount < 4 {
		t.Errorf("Guide template should have at least 4 steps, got %d", stepCount)
	}
}

func TestGenerateDocTemplate(t *testing.T) {
	template := generateDocTemplate()

	requiredSections := []string{
		"## Overview",
		"## Content",
		"## Examples",
		"## Related Resources",
		"## Best Practices",
		"## FAQ",
	}

	for _, section := range requiredSections {
		if !strings.Contains(template, section) {
			t.Errorf("Doc template missing section: %q", section)
		}
	}
}

func TestGenerateSpecTemplate(t *testing.T) {
	template := generateSpecTemplate()

	requiredSections := []string{
		"## Purpose",
		"## Scope",
		"## Authority",
		"## Subject",
		"## Definitions",
		"## Contract Surface",
		"### Interfaces",
		"### Inputs",
		"### Outputs",
		"## Normative Behavior",
		"### Preconditions",
		"### Postconditions",
		"## Constraints",
		"## Invariants",
		"## Error Handling",
		"### Failure Semantics",
		"## Conformance",
	}

	for _, section := range requiredSections {
		if !strings.Contains(template, section) {
			t.Errorf("Spec template missing section: %q", section)
		}
	}

	// Spec must contain code blocks for conformance-critical examples.
	codeBlockCount := strings.Count(template, "```")
	if codeBlockCount%2 != 0 {
		t.Errorf("code block markers = %d, should be even", codeBlockCount)
	}
	if codeBlockCount < 2 {
		t.Errorf("Spec template should have at least 1 code block pair, got %d markers", codeBlockCount)
	}

	// Spec must contain tables for definitions, inputs, outputs, constraints, error handling.
	pipeCount := strings.Count(template, "|")
	if pipeCount < 30 {
		t.Errorf("Spec template should have substantial table content, got %d pipe characters", pipeCount)
	}

	// Spec must contain RFC 2119 normative keywords.
	for _, keyword := range []string{"MUST", "SHOULD", "MAY"} {
		if !strings.Contains(template, keyword) {
			t.Errorf("Spec template missing RFC 2119 keyword: %q", keyword)
		}
	}

	// Optional sections must be marked with "Include only if" guidance.
	optionalMarkers := strings.Count(template, "Include only if")
	if optionalMarkers < 3 {
		t.Errorf("Spec template should mark optional sections, got %d 'Include only if' markers", optionalMarkers)
	}
}

func TestGenerateTaskTypeTemplate(t *testing.T) {
	template := generateTaskTypeTemplate()

	requiredSections := []string{
		"## What",
		"## When to Use",
		"## Steps",
		"## Example",
		"## Things to Watch Out For",
	}

	for _, section := range requiredSections {
		if !strings.Contains(template, section) {
			t.Errorf("TaskType template missing section: %q", section)
		}
	}
}

func TestGenerateCPATTemplate(t *testing.T) {
	template := generateCPATTemplate()

	requiredSections := []string{
		"## What Changed",
		"## Why",
		"## Before",
		"## After",
		"## Scope",
		"## Notes",
	}

	for _, section := range requiredSections {
		if !strings.Contains(template, section) {
			t.Errorf("CPAT template missing section: %q", section)
		}
	}

	if !strings.Contains(template, "```") {
		t.Error("CPAT template should include code blocks")
	}
}

func TestGeneratePRDTemplate(t *testing.T) {
	template := generatePRDTemplate()

	requiredSections := []string{
		"## Vision",
		"### Product Vision Statement",
		"## Problem Statement",
		"### Target Users",
		"### User Stories",
		"## Goals and Success Metrics",
		"## Requirements",
		"### Functional Requirements",
		"### Non-Functional Requirements",
		"## Constraints",
		"## Solution Overview",
		"## Risks and Mitigations",
		"## Timeline",
		"## Open Questions",
	}

	for _, section := range requiredSections {
		if !strings.Contains(template, section) {
			t.Errorf("PRD template missing section: %q", section)
		}
	}

	if !strings.Contains(template, "|") {
		t.Error("PRD template should include tables")
	}

	if !strings.Contains(template, "P0") {
		t.Error("PRD template should include priority levels")
	}
}

func TestTemplateStructure(t *testing.T) {
	tests := []struct {
		name         string
		documentType DocumentType
		minLength    int
	}{
		{
			name:         "ADR has substantial content",
			documentType: TypeADR,
			minLength:    800,
		},
		{
			name:         "RFC has substantial content",
			documentType: TypeRFC,
			minLength:    1500,
		},
		{
			name:         "Rule has substantial content",
			documentType: TypeRule,
			minLength:    800,
		},
		{
			name:         "Guide has substantial content",
			documentType: TypeGuide,
			minLength:    1200,
		},
		{
			name:         "Doc has substantial content",
			documentType: TypeDoc,
			minLength:    600,
		},
		{
			name:         "Spec has substantial content",
			documentType: TypeSpec,
			minLength:    2000,
		},
		{
			name:         "TaskType has substantial content",
			documentType: TypeTaskType,
			minLength:    500,
		},
		{
			name:         "CPAT has substantial content",
			documentType: TypeCPAT,
			minLength:    300,
		},
		{
			name:         "PRD has substantial content",
			documentType: TypePRD,
			minLength:    2000,
		},
		{
			name:         "Idea has substantial content",
			documentType: TypeIdea,
			minLength:    400,
		},
		{
			name:         "Plan has substantial content",
			documentType: TypePlan,
			minLength:    300,
		},
		{
			name:         "MRD has substantial content",
			documentType: TypeMRD,
			minLength:    1500,
		},
		{
			name:         "BRD has substantial content",
			documentType: TypeBRD,
			minLength:    1500,
		},
		{
			name:         "URD has substantial content",
			documentType: TypeURD,
			minLength:    1500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := GenerateTemplate(tt.documentType)

			if len(template) < tt.minLength {
				t.Errorf("template length = %d, want at least %d", len(template), tt.minLength)
			}
		})
	}
}

func TestTemplateMarkdownFormatting(t *testing.T) {
	types := []DocumentType{TypeADR, TypeRFC, TypeRule, TypeGuide, TypeDoc, TypeSpec, TypeTaskType, TypeCPAT, TypePRD, TypeIdea, TypePlan, TypeMRD, TypeBRD, TypeURD}

	for _, typ := range types {
		t.Run(string(typ), func(t *testing.T) {
			template := GenerateTemplate(typ)

			if !strings.Contains(template, "##") {
				t.Error("template should contain markdown headers (##)")
			}

			if !strings.HasSuffix(template, "\n") {
				t.Error("template should end with newline")
			}

			if strings.Contains(template, "\n\n\n\n") {
				t.Error("template has excessive empty lines")
			}
		})
	}
}

func TestRuleTemplate_CodeBlocks(t *testing.T) {
	template := generateRuleTemplate()

	codeBlockCount := strings.Count(template, "```")

	if codeBlockCount%2 != 0 {
		t.Errorf("code block markers = %d, should be even", codeBlockCount)
	}

	if codeBlockCount < 8 {
		t.Errorf("code block markers = %d, should have at least 8 (4 blocks)", codeBlockCount)
	}
}

func TestRFCTemplate_Checkboxes(t *testing.T) {
	template := generateRFCTemplate()

	checkboxCount := strings.Count(template, "- [ ]")

	if checkboxCount < 5 {
		t.Errorf("checkbox count = %d, want at least 5", checkboxCount)
	}

	if !strings.Contains(template, "## Implementation Plan") {
		t.Error("RFC should have Implementation Plan section")
	}
}

func TestGuideTemplate_CodeBlocks(t *testing.T) {
	template := generateGuideTemplate()

	codeBlockCount := strings.Count(template, "```")

	if codeBlockCount < 6 {
		t.Errorf("Guide template code blocks = %d, should have at least 6", codeBlockCount)
	}
}

func TestPRDTemplate_Tables(t *testing.T) {
	template := generatePRDTemplate()

	tableCount := strings.Count(template, "|")

	if tableCount < 50 {
		t.Errorf("PRD table elements = %d, should have at least 50", tableCount)
	}
}

func TestGenerateIdeaTemplate(t *testing.T) {
	template := generateIdeaTemplate()

	requiredSections := []string{
		"## Idea",
		"### Problem / Opportunity",
		"## Value",
		"### For Users",
		"### For Business",
		"### For Team",
		"## Possible Implementation",
		"### Technical Approach",
		"### Integrations",
		"## Risks and Constraints",
		"### Potential Risks",
		"### Known Constraints",
		"## Next Steps",
		"## Related Materials",
	}

	for _, section := range requiredSections {
		if !strings.Contains(template, section) {
			t.Errorf("Idea template missing section: %q", section)
		}
	}

	// Verify checklist format
	if !strings.Contains(template, "- [ ]") {
		t.Error("Idea template should include task checkboxes")
	}

	// Verify it has guidance text
	if !strings.Contains(template, "Describe") || !strings.Contains(template, "What") {
		t.Error("Idea template should include guidance questions")
	}
}

func TestGeneratePlanTemplate(t *testing.T) {
	template := generatePlanTemplate()

	requiredSections := []string{
		"## Goal",
		"### Context",
		"## Tasks",
		"### Phase 1:",
		"### Phase 2:",
		"## Acceptance Criteria",
		"## Dependencies",
		"## Notes",
	}

	for _, section := range requiredSections {
		if !strings.Contains(template, section) {
			t.Errorf("Plan template missing section: %q", section)
		}
	}

	// Verify checklist format
	checkboxCount := strings.Count(template, "- [ ]")
	if checkboxCount < 4 {
		t.Errorf("Plan template should have at least 4 task checkboxes, got %d", checkboxCount)
	}

	// Verify table format for dependencies
	if !strings.Contains(template, "|") {
		t.Error("Plan template should include table for dependencies")
	}

	// Verify phase structure
	phaseCount := strings.Count(template, "### Phase")
	if phaseCount < 2 {
		t.Errorf("Plan template should have at least 2 phases, got %d", phaseCount)
	}
}
