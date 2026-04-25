package prompts

import (
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// messageText extracts the text payload of a PromptMessage. PromptMessage.Content
// is the Content interface; for our prompts it is always TextContent.
func messageText(t *testing.T, m mcp.PromptMessage) string {
	t.Helper()
	tc, ok := mcp.AsTextContent(m.Content)
	if !ok {
		t.Fatalf("message content is not TextContent: %T", m.Content)
	}
	return tc.Text
}

func TestFramingMessage(t *testing.T) {
	t.Parallel()

	msg := framingMessage("ISO 29148", "auth", "BRS -> StRS -> SyRS -> SRS")

	if msg.Role != mcp.RoleUser {
		t.Errorf("role = %q, want %q", msg.Role, mcp.RoleUser)
	}
	text := messageText(t, msg)

	wants := []string{
		"ISO 29148",
		"auth",
		"BRS -> StRS -> SyRS -> SRS",
		ConfirmationGate,
	}
	for _, w := range wants {
		if !strings.Contains(text, w) {
			t.Errorf("framing message missing %q: %s", w, text)
		}
	}
}

func TestPhaseMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		phase       int
		docType     string
		instruction string
		wantSubs    []string
	}{
		{
			name:        "brs phase 1",
			phase:       1,
			docType:     "brs",
			instruction: "Call create_document.",
			wantSubs:    []string{"Phase 1", "BRS", `type="brs"`, "Call create_document."},
		},
		{
			name:        "strs phase 2",
			phase:       2,
			docType:     "strs",
			instruction: `add_relation(source=<strs path>, target=<brs path>, type="implements").`,
			wantSubs:    []string{"Phase 2", "STRS", `type="strs"`, `type="implements"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			msg := phaseMessage(tt.phase, tt.docType, tt.instruction)
			if msg.Role != mcp.RoleUser {
				t.Errorf("role = %q, want %q", msg.Role, mcp.RoleUser)
			}
			text := messageText(t, msg)
			for _, sub := range tt.wantSubs {
				if !strings.Contains(text, sub) {
					t.Errorf("phase message missing %q: %s", sub, text)
				}
			}
		})
	}
}

func TestVerificationMessage(t *testing.T) {
	t.Parallel()

	msg := verificationMessage("BRS", "BRS <- StRS <- SyRS <- SRS")

	if msg.Role != mcp.RoleUser {
		t.Errorf("role = %q, want %q", msg.Role, mcp.RoleUser)
	}
	text := messageText(t, msg)

	wants := []string{"list_relations", "BRS <- StRS <- SyRS <- SRS"}
	for _, w := range wants {
		if !strings.Contains(text, w) {
			t.Errorf("verification message missing %q: %s", w, text)
		}
	}
}
