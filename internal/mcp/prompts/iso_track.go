package prompts

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	isoTrackName        = "iso_track"
	isoTrackDescription = "Run ISO 29148 cascade: BRS -> StRS -> SyRS -> SRS, linking each level via 'implements'."
	isoTrackChain       = "BRS -> StRS -> SyRS -> SRS"
)

// NewISOTrackPrompt returns the prompt definition for the ISO 29148 cascade.
func NewISOTrackPrompt() mcp.Prompt {
	return mcp.NewPrompt(isoTrackName,
		mcp.WithPromptDescription(isoTrackDescription),
		mcp.WithArgument("feature_name",
			mcp.ArgumentDescription("Short name of the feature being specified."),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("scope",
			mcp.ArgumentDescription("One-line scope statement (optional)."),
		),
	)
}

// HandleISOTrack returns the message script for the ISO 29148 cascade.
func HandleISOTrack(_ context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	feature, err := requireStringArg(req, "feature_name")
	if err != nil {
		return nil, err
	}
	scope := optionalStringArg(req, "scope", "")

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("ISO 29148 cascade for %q", feature),
		Messages:    isoTrackMessages(feature, scope),
	}, nil
}

func isoTrackMessages(feature, scope string) []mcp.PromptMessage {
	return []mcp.PromptMessage{
		framingMessage("ISO 29148", feature, isoTrackChain),
		phaseMessage(1, "brs",
			"Call create_document(type=\"brs\", filename=..., directory=..., title=...). "+
				"Use the standard BRS template; fill Mission, Goals, Operational Concept, Success Criteria"+
				scopeClause(scope)+" "+ConfirmationGate,
		),
		phaseMessage(2, "strs",
			"Call create_document(type=\"strs\", ...) to create the StRS document, then call "+
				"add_relation(source=<strs path>, target=<brs path>, type=\"implements\"). "+
				ConfirmationGate,
		),
		phaseMessage(3, "syrs",
			"Call create_document(type=\"syrs\", ...) to create the SyRS document, then call "+
				"add_relation(source=<syrs path>, target=<strs path>, type=\"implements\"). "+
				ConfirmationGate,
		),
		phaseMessage(4, "srs",
			"Call create_document(type=\"srs\", ...) to create the SRS document, then call "+
				"add_relation(source=<srs path>, target=<syrs path>, type=\"implements\"). "+
				ConfirmationGate,
		),
		verificationMessage("BRS", "BRS <- StRS <- SyRS <- SRS"),
	}
}
