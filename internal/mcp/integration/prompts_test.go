package integration

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// TestInitializeAdvertisesPromptsCapability proves that
// server.WithPromptCapabilities(true) flowed through NewServer and that the
// in-process client observes the advertised capability after the Initialize
// handshake performed by newTestClient. The harness throws away the
// InitializeResult, but the client persists serverCapabilities internally and
// exposes them via GetServerCapabilities.
func TestInitializeAdvertisesPromptsCapability(t *testing.T) {
	t.Parallel()

	base := initArchcore(t)
	c := newTestClient(t, base)

	caps := c.GetServerCapabilities()
	if caps.Prompts == nil {
		t.Fatalf("server did not advertise prompts capability; caps=%+v", caps)
	}
}

// TestPromptRegistrationCanary asserts that exactly the five expected track
// prompts are registered with the integration server, and that every prompt
// description carries a chain-arrow ("->") plus a required feature_name
// argument. This is the prompts analogue of TestToolRegistrationCanary.
func TestPromptRegistrationCanary(t *testing.T) {
	t.Parallel()

	expected := []string{
		"architecture_track",
		"iso_track",
		"product_track",
		"sources_track",
		"standard_track",
	}
	slices.Sort(expected)

	base := initArchcore(t)
	c := newTestClient(t, base)

	res, err := c.ListPrompts(context.Background(), mcp.ListPromptsRequest{})
	if err != nil {
		t.Fatalf("ListPrompts: %v", err)
	}

	if len(res.Prompts) != len(expected) {
		t.Fatalf("registered %d prompts, want %d; got %+v",
			len(res.Prompts), len(expected), res.Prompts)
	}

	got := make([]string, 0, len(res.Prompts))
	for _, p := range res.Prompts {
		got = append(got, p.Name)
	}
	slices.Sort(got)

	for i, name := range expected {
		if got[i] != name {
			t.Errorf("prompt[%d] = %q, want %q (full got: %v)", i, got[i], name, got)
		}
	}

	for _, p := range res.Prompts {
		if p.Description == "" {
			t.Errorf("prompt %q: empty description", p.Name)
		}
		if !strings.Contains(p.Description, "->") {
			t.Errorf("prompt %q description missing chain-arrow: %q", p.Name, p.Description)
		}

		var hasRequiredFeatureName bool
		for _, a := range p.Arguments {
			if a.Name == "feature_name" {
				hasRequiredFeatureName = a.Required
				break
			}
		}
		if !hasRequiredFeatureName {
			t.Errorf("prompt %q must declare feature_name as a required argument; args=%+v",
				p.Name, p.Arguments)
		}
	}
}

// TestGetPromptCascadeShape exercises GetPrompt for every track and pins the
// observable contract: message count, first-message role, doc-type and
// relation substrings, list_relations call, and chain-summary in the
// verification message. This catches drift in any track's message script
// without re-implementing the per-track unit tests.
func TestGetPromptCascadeShape(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                  string
		expectedMessageCount  int
		expectedDocTypes      []string
		expectedRelationTypes []string
		chainSummary          string
	}{
		{
			name:                  "iso_track",
			expectedMessageCount:  6,
			expectedDocTypes:      []string{"brs", "strs", "syrs", "srs"},
			expectedRelationTypes: []string{"implements", "implements", "implements"},
			chainSummary:          "BRS <- StRS <- SyRS <- SRS",
		},
		{
			name:                  "sources_track",
			expectedMessageCount:  5,
			expectedDocTypes:      []string{"mrd", "brd", "urd"},
			expectedRelationTypes: []string{"related", "related"},
			chainSummary:          "MRD <- BRD <- URD",
		},
		{
			name:                  "product_track",
			expectedMessageCount:  4,
			expectedDocTypes:      []string{"prd", "plan"},
			expectedRelationTypes: []string{"implements"},
			chainSummary:          "PRD <- Plan",
		},
		{
			name:                  "standard_track",
			expectedMessageCount:  5,
			expectedDocTypes:      []string{"adr", "rule", "guide"},
			expectedRelationTypes: []string{"implements", "related"},
			chainSummary:          "ADR <- Rule <- Guide",
		},
		{
			name:                  "architecture_track",
			expectedMessageCount:  5,
			expectedDocTypes:      []string{"adr", "spec", "plan"},
			expectedRelationTypes: []string{"implements", "implements"},
			chainSummary:          "ADR <- Spec <- Plan",
		},
	}

	base := initArchcore(t)
	c := newTestClient(t, base)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := mcp.GetPromptRequest{}
			req.Params.Name = tc.name
			req.Params.Arguments = map[string]string{"feature_name": "demo-feature"}

			result, err := c.GetPrompt(context.Background(), req)
			if err != nil {
				t.Fatalf("GetPrompt(%s): %v", tc.name, err)
			}
			if result == nil {
				t.Fatalf("GetPrompt(%s): nil result", tc.name)
			}

			if !strings.Contains(result.Description, "demo-feature") {
				t.Errorf("description = %q, want substring %q", result.Description, "demo-feature")
			}

			if got, want := len(result.Messages), tc.expectedMessageCount; got != want {
				t.Fatalf("message count = %d, want %d; messages=%+v", got, want, result.Messages)
			}

			if result.Messages[0].Role != mcp.RoleUser {
				t.Errorf("messages[0].Role = %q, want %q", result.Messages[0].Role, mcp.RoleUser)
			}

			var sb strings.Builder
			for i, msg := range result.Messages {
				txt, ok := mcp.AsTextContent(msg.Content)
				if !ok {
					t.Fatalf("messages[%d]: content is not TextContent (%T)", i, msg.Content)
				}
				sb.WriteString(txt.Text)
				sb.WriteByte('\n')
			}
			full := sb.String()

			for _, dt := range tc.expectedDocTypes {
				want := `type="` + dt + `"`
				if !strings.Contains(full, want) {
					t.Errorf("messages missing doc-type substring %q", want)
				}
			}
			for _, rt := range tc.expectedRelationTypes {
				want := `type="` + rt + `"`
				if !strings.Contains(full, want) {
					t.Errorf("messages missing relation-type substring %q", want)
				}
			}

			// Guard against doc-type / relation-type substring collisions by
			// pinning the literal add_relation( call count to the number of
			// edges declared by the track.
			if got, want := strings.Count(full, "add_relation("), len(tc.expectedRelationTypes); got != want {
				t.Errorf("add_relation( occurrences = %d, want %d", got, want)
			}

			if !strings.Contains(full, "list_relations") {
				t.Errorf("messages missing 'list_relations' (verification step)")
			}
			if !strings.Contains(full, tc.chainSummary) {
				t.Errorf("messages missing chain summary %q", tc.chainSummary)
			}
		})
	}
}

// TestGetPromptMissingRequiredArg pins the behavior of GetPrompt when the
// required feature_name argument is omitted. mcp-go v0.49.0 may either
// reject at the protocol layer (transport-level error) or pass through to
// the handler (which returns an error that surfaces as an empty/nil result).
// The branch is asserted permissively here; once the observed behavior is
// stable we can tighten the assertion. This is the open-question resolution
// from .archcore/mcp/mcp-prompts-implementation.plan.md (Implementation
// Notes #4).
func TestGetPromptMissingRequiredArg(t *testing.T) {
	t.Parallel()

	base := initArchcore(t)
	c := newTestClient(t, base)

	req := mcp.GetPromptRequest{}
	req.Params.Name = "iso_track"
	req.Params.Arguments = map[string]string{}

	result, err := c.GetPrompt(context.Background(), req)
	if err != nil {
		if !strings.Contains(err.Error(), "feature_name") {
			t.Errorf("error = %q, want it to mention feature_name", err.Error())
		}
		return
	}
	if result != nil && len(result.Messages) > 0 {
		t.Errorf("GetPrompt without feature_name unexpectedly returned %d messages: %+v",
			len(result.Messages), result.Messages)
	}
}

// TestGetPromptOptionalArgComponentName covers the architecture_track-only
// optional component_name argument: when present, the framing message must
// embed the component name verbatim; when absent, the framing message must
// not leak the literal "<component>" placeholder.
func TestGetPromptOptionalArgComponentName(t *testing.T) {
	t.Parallel()

	base := initArchcore(t)
	c := newTestClient(t, base)

	t.Run("with_component_name", func(t *testing.T) {
		t.Parallel()

		req := mcp.GetPromptRequest{}
		req.Params.Name = "architecture_track"
		req.Params.Arguments = map[string]string{
			"feature_name":   "demo-feature",
			"component_name": "payment-gateway",
		}
		result, err := c.GetPrompt(context.Background(), req)
		if err != nil {
			t.Fatalf("GetPrompt: %v", err)
		}
		if len(result.Messages) == 0 {
			t.Fatalf("no messages returned")
		}
		framing, ok := mcp.AsTextContent(result.Messages[0].Content)
		if !ok {
			t.Fatalf("framing content is not TextContent (%T)", result.Messages[0].Content)
		}
		if !strings.Contains(framing.Text, "payment-gateway") {
			t.Errorf("framing message missing component name; got: %q", framing.Text)
		}
	})

	t.Run("without_component_name", func(t *testing.T) {
		t.Parallel()

		req := mcp.GetPromptRequest{}
		req.Params.Name = "architecture_track"
		req.Params.Arguments = map[string]string{"feature_name": "demo-feature"}
		result, err := c.GetPrompt(context.Background(), req)
		if err != nil {
			t.Fatalf("GetPrompt: %v", err)
		}
		if len(result.Messages) == 0 {
			t.Fatalf("no messages returned")
		}
		framing, ok := mcp.AsTextContent(result.Messages[0].Content)
		if !ok {
			t.Fatalf("framing content is not TextContent (%T)", result.Messages[0].Content)
		}
		if strings.Contains(framing.Text, "<component>") {
			t.Errorf("framing message leaked literal placeholder; got: %q", framing.Text)
		}
	})
}
