package prompts

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	standardTrackName        = "standard_track"
	standardTrackDescription = "Run Standard track: ADR -> Rule -> Guide; rule implements adr, guide related rule."
	standardTrackChain       = "ADR -> Rule -> Guide"
)

// NewStandardTrackPrompt returns the prompt definition for the Standard track
// (ADR -> Rule -> Guide). The required argument is named feature_name for
// API consistency with the other tracks but semantically represents the
// standard's topic (e.g., "logging-conventions").
func NewStandardTrackPrompt() mcp.Prompt {
	return mcp.NewPrompt(standardTrackName,
		mcp.WithPromptDescription(standardTrackDescription),
		mcp.WithArgument("feature_name",
			mcp.ArgumentDescription("Short name of the standard topic (e.g., \"logging-conventions\")."),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("scope",
			mcp.ArgumentDescription("One-line scope statement (optional)."),
		),
	)
}

// HandleStandardTrack returns the message script for the Standard track
// (ADR -> Rule -> Guide).
func HandleStandardTrack(_ context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	topic, err := requireStringArg(req, "feature_name")
	if err != nil {
		return nil, err
	}
	scope := optionalStringArg(req, "scope", "")

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Standard track for %q", topic),
		Messages:    standardTrackMessages(topic, scope),
	}, nil
}

func standardTrackMessages(topic, scope string) []mcp.PromptMessage {
	return []mcp.PromptMessage{
		framingMessage("Standard", topic, standardTrackChain),
		phaseMessage(1, "adr",
			"Call create_document(type=\"adr\", filename=..., directory=..., title=...). "+
				"Capture the decision context, options, and chosen approach for the standard topic"+
				scopeClause(scope)+" "+ConfirmationGate,
		),
		phaseMessage(2, "rule",
			"Call create_document(type=\"rule\", ...) to create the Rule document, then call "+
				"add_relation(source=<rule path>, target=<adr path>, type=\"implements\"). "+
				ConfirmationGate,
		),
		phaseMessage(3, "guide",
			"Call create_document(type=\"guide\", ...) to create the Guide document, then call "+
				"add_relation(source=<guide path>, target=<rule path>, type=\"related\"). "+
				ConfirmationGate,
		),
		verificationMessage("ADR", "ADR <- Rule <- Guide"),
	}
}
