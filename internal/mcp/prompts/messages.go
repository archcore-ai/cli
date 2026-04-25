package prompts

import (
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// ConfirmationGate is the literal sentence appended after every non-final
// phase. Centralizing it lets tests assert presence with a single substring
// and keeps the wording identical across every track.
const ConfirmationGate = "Summarize the result and ask the user 'ok to continue?' before the next phase."

// scopeClause renders the trailing fragment appended to a phase 1 instruction:
// "." when scope is empty, " based on the scope: <scope>." otherwise. Each
// track shared this exact snippet — keep it in one place so wording cannot
// drift between tracks.
func scopeClause(scope string) string {
	if scope == "" {
		return "."
	}
	return " based on the scope: " + scope + "."
}

// framingMessage is the opening message of every track. mcp-go exposes only
// RoleUser and RoleAssistant — there is no system role — so we use RoleUser
// with explicit "You are running…" prose. The chain string (e.g.
// "BRS -> StRS -> SyRS -> SRS") appears verbatim so the model can echo it
// back when confirming with the user.
//
// The optional contextPhrases are spliced in after the feature name (with a
// leading space) so callers can anchor the framing to extra context — e.g.
// architecture_track passes "for the auth-service component" when the
// optional component_name argument is set. Empty entries are skipped.
func framingMessage(track, feature, chain string, contextPhrases ...string) mcp.PromptMessage {
	var b strings.Builder
	for _, c := range contextPhrases {
		if c != "" {
			b.WriteByte(' ')
			b.WriteString(c)
		}
	}
	text := fmt.Sprintf(
		"You are running the %s cascade for %q%s. Chain: %s. Follow phases sequentially. After EACH phase: %s",
		track, feature, b.String(), chain, ConfirmationGate,
	)
	return mcp.PromptMessage{
		Role:    mcp.RoleUser,
		Content: mcp.NewTextContent(text),
	}
}

// phaseMessage formats a single phase instruction. The header carries the
// doc type twice — uppercased label plus the literal `type="<docType>"` —
// so tests and downstream agents can match either form, and so callers do
// not have to remember to embed the literal in their instruction text.
func phaseMessage(phase int, docType, instruction string) mcp.PromptMessage {
	text := fmt.Sprintf(`Phase %d — %s (type=%q). %s`, phase, strings.ToUpper(docType), docType, instruction)
	return mcp.PromptMessage{
		Role:    mcp.RoleUser,
		Content: mcp.NewTextContent(text),
	}
}

// verificationMessage closes a track by asking the agent to call
// list_relations on the root document and confirm the chain summary. Both
// substrings ("list_relations" and the chain summary) must appear so tests
// can pin the contract.
func verificationMessage(rootDocLabel, chainSummary string) mcp.PromptMessage {
	text := fmt.Sprintf(
		"Final — call list_relations on the %s path and confirm the chain %s is complete.",
		rootDocLabel, chainSummary,
	)
	return mcp.PromptMessage{
		Role:    mcp.RoleUser,
		Content: mcp.NewTextContent(text),
	}
}
