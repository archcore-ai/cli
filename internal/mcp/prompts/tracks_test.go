package prompts

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// trackEdge describes one expected add_relation edge declared by a phase
// instruction. source must appear textually before target so the assertion
// can pin source/target ordering without parsing.
type trackEdge struct {
	phaseIdx int
	source   string
	target   string
	relType  string
}

type trackTestCase struct {
	name                 string
	newPrompt            func() mcp.Prompt
	handler              func(context.Context, mcp.GetPromptRequest) (*mcp.GetPromptResult, error)
	descriptionEndpoints []string // both substrings must appear in the prompt description
	optionalArgs         []string // additional optional argument names beyond "scope"
	msgCount             int
	phaseDocTypes        []string // index i corresponds to message index i+1
	edges                []trackEdge
	chainSummary         string
}

// trackCases enumerates every track exposed by RegisterAll. Add a new track
// by appending one row — every assertion below derives its expectations from
// these fields.
var trackCases = []trackTestCase{
	{
		name:                 "iso_track",
		newPrompt:            NewISOTrackPrompt,
		handler:              HandleISOTrack,
		descriptionEndpoints: []string{"BRS", "SRS"},
		msgCount:             6,
		phaseDocTypes:        []string{"brs", "strs", "syrs", "srs"},
		edges: []trackEdge{
			{phaseIdx: 2, source: "strs", target: "brs", relType: "implements"},
			{phaseIdx: 3, source: "syrs", target: "strs", relType: "implements"},
			{phaseIdx: 4, source: "srs", target: "syrs", relType: "implements"},
		},
		chainSummary: "BRS <- StRS <- SyRS <- SRS",
	},
	{
		name:                 "sources_track",
		newPrompt:            NewSourcesTrackPrompt,
		handler:              HandleSourcesTrack,
		descriptionEndpoints: []string{"MRD", "URD"},
		msgCount:             5,
		phaseDocTypes:        []string{"mrd", "brd", "urd"},
		edges: []trackEdge{
			{phaseIdx: 2, source: "brd", target: "mrd", relType: "related"},
			{phaseIdx: 3, source: "urd", target: "brd", relType: "related"},
		},
		chainSummary: "MRD <- BRD <- URD",
	},
	{
		name:                 "product_track",
		newPrompt:            NewProductTrackPrompt,
		handler:              HandleProductTrack,
		descriptionEndpoints: []string{"PRD", "Plan"},
		msgCount:             4,
		phaseDocTypes:        []string{"prd", "plan"},
		edges: []trackEdge{
			{phaseIdx: 2, source: "plan", target: "prd", relType: "implements"},
		},
		chainSummary: "PRD <- Plan",
	},
	{
		name:                 "standard_track",
		newPrompt:            NewStandardTrackPrompt,
		handler:              HandleStandardTrack,
		descriptionEndpoints: []string{"ADR", "Guide"},
		msgCount:             5,
		phaseDocTypes:        []string{"adr", "rule", "guide"},
		edges: []trackEdge{
			{phaseIdx: 2, source: "rule", target: "adr", relType: "implements"},
			{phaseIdx: 3, source: "guide", target: "rule", relType: "related"},
		},
		chainSummary: "ADR <- Rule <- Guide",
	},
	{
		name:                 "architecture_track",
		newPrompt:            NewArchitectureTrackPrompt,
		handler:              HandleArchitectureTrack,
		descriptionEndpoints: []string{"ADR", "Plan"},
		optionalArgs:         []string{"component_name"},
		msgCount:             5,
		phaseDocTypes:        []string{"adr", "spec", "plan"},
		edges: []trackEdge{
			{phaseIdx: 2, source: "spec", target: "adr", relType: "implements"},
			{phaseIdx: 3, source: "plan", target: "spec", relType: "implements"},
		},
		chainSummary: "ADR <- Spec <- Plan",
	},
}

func TestTrackPromptDefinitions(t *testing.T) {
	t.Parallel()
	for _, tc := range trackCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p := tc.newPrompt()
			if p.Name != tc.name {
				t.Errorf("name = %q, want %q", p.Name, tc.name)
			}
			if p.Description == "" {
				t.Error("description must not be empty")
			}
			for _, sub := range tc.descriptionEndpoints {
				if !strings.Contains(p.Description, sub) {
					t.Errorf("description should mention %q, got %q", sub, p.Description)
				}
			}

			args := map[string]mcp.PromptArgument{}
			for _, a := range p.Arguments {
				args[a.Name] = a
			}
			feature, ok := args["feature_name"]
			if !ok {
				t.Fatal("missing feature_name argument")
			}
			if !feature.Required {
				t.Error("feature_name must be required")
			}
			scope, ok := args["scope"]
			if !ok {
				t.Fatal("missing scope argument")
			}
			if scope.Required {
				t.Error("scope must be optional")
			}
			for _, name := range tc.optionalArgs {
				a, ok := args[name]
				if !ok {
					t.Errorf("missing optional argument %q", name)
					continue
				}
				if a.Required {
					t.Errorf("optional argument %q must not be required", name)
				}
			}
		})
	}
}

func TestTrackHandlerMissingFeatureName(t *testing.T) {
	t.Parallel()
	for _, tc := range trackCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := tc.handler(context.Background(), newRequest(map[string]string{}))
			if err == nil {
				t.Fatal("expected error when feature_name is missing")
			}
			if !strings.Contains(err.Error(), "feature_name") {
				t.Errorf("error %q should mention feature_name", err.Error())
			}
		})
	}
}

func TestTrackHandlerMessages(t *testing.T) {
	t.Parallel()
	for _, tc := range trackCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			res, err := tc.handler(context.Background(), newRequest(map[string]string{"feature_name": "demo"}))
			if err != nil {
				t.Fatalf("handler error: %v", err)
			}
			if res == nil {
				t.Fatal("nil result")
			}
			if !strings.Contains(res.Description, "demo") {
				t.Errorf("description = %q, want substring %q", res.Description, "demo")
			}
			if len(res.Messages) != tc.msgCount {
				t.Fatalf("message count = %d, want %d", len(res.Messages), tc.msgCount)
			}

			first := res.Messages[0]
			if first.Role != mcp.RoleUser {
				t.Errorf("first message role = %q, want %q", first.Role, mcp.RoleUser)
			}
			firstText := messageText(t, first)
			if !strings.Contains(firstText, "phases sequentially") {
				t.Errorf("first message missing 'phases sequentially': %s", firstText)
			}

			for i, dt := range tc.phaseDocTypes {
				msgIdx := i + 1
				text := messageText(t, res.Messages[msgIdx])
				want := `type="` + dt + `"`
				if !strings.Contains(text, want) {
					t.Errorf("message %d missing %q: %s", msgIdx, want, text)
				}
			}

			for _, e := range tc.edges {
				text := messageText(t, res.Messages[e.phaseIdx])
				if !strings.Contains(text, "add_relation(") {
					t.Errorf("phase %d missing add_relation(: %s", e.phaseIdx, text)
				}
				relLiteral := `type="` + e.relType + `"`
				if !strings.Contains(text, relLiteral) {
					t.Errorf("phase %d missing %s: %s", e.phaseIdx, relLiteral, text)
				}
				srcIdx := strings.Index(text, e.source)
				tgtIdx := strings.Index(text, e.target)
				if srcIdx < 0 || tgtIdx < 0 {
					t.Fatalf("phase %d missing source=%q (idx %d) or target=%q (idx %d): %s",
						e.phaseIdx, e.source, srcIdx, e.target, tgtIdx, text)
				}
				if srcIdx > tgtIdx {
					t.Errorf("phase %d: source %q must appear before target %q in text: %s",
						e.phaseIdx, e.source, e.target, text)
				}
			}

			for i := range len(res.Messages) - 1 {
				text := messageText(t, res.Messages[i])
				if !strings.Contains(text, ConfirmationGate) {
					t.Errorf("message %d missing confirmation gate: %s", i, text)
				}
			}
			final := messageText(t, res.Messages[len(res.Messages)-1])
			if strings.Contains(final, ConfirmationGate) {
				t.Errorf("final message must not contain confirmation gate: %s", final)
			}
			if !strings.Contains(final, "list_relations") {
				t.Errorf("final message must reference list_relations: %s", final)
			}
			if !strings.Contains(final, tc.chainSummary) {
				t.Errorf("final message missing chain summary %q: %s", tc.chainSummary, final)
			}
		})
	}
}

func TestTrackHandlerWithScope(t *testing.T) {
	t.Parallel()
	for _, tc := range trackCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			res, err := tc.handler(context.Background(), newRequest(map[string]string{
				"feature_name": "demo",
				"scope":        "checkout flow",
			}))
			if err != nil {
				t.Fatalf("handler error: %v", err)
			}
			found := false
			for _, m := range res.Messages {
				if strings.Contains(messageText(t, m), "checkout flow") {
					found = true
					break
				}
			}
			if !found {
				t.Error("scope value must appear in some message")
			}
		})
	}
}

// TestArchitectureTrackComponentName covers the architecture-only optional
// component_name argument: when set it must surface in the framing message;
// when absent, the framing must not leak the literal "<component>" or
// "{component_name}" placeholder syntax.
func TestArchitectureTrackComponentName(t *testing.T) {
	t.Parallel()

	t.Run("with_component", func(t *testing.T) {
		t.Parallel()
		res, err := HandleArchitectureTrack(context.Background(), newRequest(map[string]string{
			"feature_name":   "demo",
			"component_name": "auth-service",
		}))
		if err != nil {
			t.Fatalf("handler error: %v", err)
		}
		first := messageText(t, res.Messages[0])
		if !strings.Contains(first, "auth-service") {
			t.Errorf("framing message must mention component_name: %s", first)
		}
	})

	t.Run("without_component_no_placeholder_leak", func(t *testing.T) {
		t.Parallel()
		res, err := HandleArchitectureTrack(context.Background(), newRequest(map[string]string{
			"feature_name": "demo",
		}))
		if err != nil {
			t.Fatalf("handler error: %v", err)
		}
		first := messageText(t, res.Messages[0])
		if strings.Contains(first, "<component") || strings.Contains(first, "{component_name}") {
			t.Errorf("framing message leaked placeholder: %s", first)
		}
	})
}
