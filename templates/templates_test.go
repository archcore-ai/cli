package templates

import (
	"strings"
	"testing"
)

func TestGenerateTemplate(t *testing.T) {
	tests := []struct {
		name         string
		documentType DocumentType
		wantEmpty    bool
		wantContains []string
	}{
		{
			name:         "ADR template",
			documentType: TypeADR,
			wantEmpty:    false,
			wantContains: []string{"## Context", "## Decision", "## Consequences", "### Positive", "### Negative", "## References", "## Alternatives Considered"},
		},
		{
			name:         "RFC template",
			documentType: TypeRFC,
			wantEmpty:    false,
			wantContains: []string{"## Summary", "## Motivation", "## Detailed Design", "## Drawbacks", "## Alternatives", "## Unresolved Questions", "## Implementation Plan", "## Security Considerations"},
		},
		{
			name:         "Rule template",
			documentType: TypeRule,
			wantEmpty:    false,
			wantContains: []string{"## Description", "## Rule", "## Examples", "### Good", "### Bad", "## Exceptions", "## References", "## Enforcement"},
		},
		{
			name:         "Guide template",
			documentType: TypeGuide,
			wantEmpty:    false,
			wantContains: []string{"## Overview", "## Prerequisites", "## Steps", "### Step 1:", "### Step 2:", "### Step 3:", "## Common Issues", "## Related Resources", "## Verification"},
		},
		{
			name:         "Doc template",
			documentType: TypeDoc,
			wantEmpty:    false,
			wantContains: []string{"## Overview", "## Content", "## Examples", "## Related Resources", "## Best Practices", "## FAQ"},
		},
		{
			name:         "Project template",
			documentType: TypeProject,
			wantEmpty:    false,
			wantContains: []string{"## Overview", "## Purpose", "## Architecture", "## Getting Started", "## Key Components", "## Related Resources", "## Development"},
		},
		{
			name:         "Task-Type template",
			documentType: TypeTaskType,
			wantEmpty:    false,
			wantContains: []string{"## Description", "## When to Use", "## Fields", "## Workflow", "## Examples"},
		},
		{
			name:         "CPAT template",
			documentType: TypeCPAT,
			wantEmpty:    false,
			wantContains: []string{"## Overview", "## Context", "## Problem", "## Action", "## Timeline"},
		},
		{
			name:         "PRD template",
			documentType: TypePRD,
			wantEmpty:    false,
			wantContains: []string{"## Vision", "## Problem Statement", "## Goals and Success Metrics", "## Requirements", "## Constraints", "## Timeline"},
		},
		{
			name:         "Idea template",
			documentType: TypeIdea,
			wantEmpty:    false,
			wantContains: []string{"## Idea", "## Value", "## Possible Implementation", "## Risks and Constraints", "## Next Steps"},
		},
		{
			name:         "Plan template",
			documentType: TypePlan,
			wantEmpty:    false,
			wantContains: []string{"## Goal", "## Tasks", "## Acceptance Criteria", "## Dependencies", "## Notes"},
		},
		{
			name:         "Unknown type falls back to doc template",
			documentType: DocumentType("unknown"),
			wantEmpty:    false,
			wantContains: []string{"## Overview", "## Content", "## Examples", "## Related Resources"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateTemplate(tt.documentType)

			if tt.wantEmpty {
				if got != "" {
					t.Errorf("GenerateTemplate(%q) = %q, want empty string", tt.documentType, got)
				}
				return
			}

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

func TestGenerateProjectTemplate(t *testing.T) {
	template := generateProjectTemplate()

	requiredSections := []string{
		"## Overview",
		"## Purpose",
		"## Architecture",
		"## Getting Started",
		"## Key Components",
		"## Related Resources",
		"## Development",
	}

	for _, section := range requiredSections {
		if !strings.Contains(template, section) {
			t.Errorf("Project template missing section: %q", section)
		}
	}

	if !strings.Contains(template, "|") {
		t.Error("Project template should include tables")
	}
}

func TestGenerateTaskTypeTemplate(t *testing.T) {
	template := generateTaskTypeTemplate()

	requiredSections := []string{
		"## Description",
		"## When to Use",
		"## Fields",
		"## Workflow",
		"## Examples",
		"### Required Fields",
		"### States",
	}

	for _, section := range requiredSections {
		if !strings.Contains(template, section) {
			t.Errorf("TaskType template missing section: %q", section)
		}
	}

	if !strings.Contains(template, "|") {
		t.Error("TaskType template should include tables")
	}
}

func TestGenerateCPATTemplate(t *testing.T) {
	template := generateCPATTemplate()

	requiredSections := []string{
		"## Overview",
		"## Context",
		"## Problem",
		"## Action",
		"## Timeline",
		"### Root Cause",
		"### Corrective Actions",
		"### Preventive Actions",
	}

	for _, section := range requiredSections {
		if !strings.Contains(template, section) {
			t.Errorf("CPAT template missing section: %q", section)
		}
	}

	if !strings.Contains(template, "|") {
		t.Error("CPAT template should include tables")
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
			name:         "Project has substantial content",
			documentType: TypeProject,
			minLength:    1200,
		},
		{
			name:         "TaskType has substantial content",
			documentType: TypeTaskType,
			minLength:    1000,
		},
		{
			name:         "CPAT has substantial content",
			documentType: TypeCPAT,
			minLength:    1000,
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
	types := []DocumentType{TypeADR, TypeRFC, TypeRule, TypeGuide, TypeDoc, TypeProject, TypeTaskType, TypeCPAT, TypePRD, TypeIdea, TypePlan}

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
