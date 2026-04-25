package prompts

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	sourcesTrackName        = "sources_track"
	sourcesTrackDescription = "Run Sources discovery: MRD -> BRD -> URD requirement-source documents, linked via 'related'."
	sourcesTrackChain       = "MRD -> BRD -> URD"
)

// NewSourcesTrackPrompt returns the prompt definition for the Sources
// discovery flow (MRD -> BRD -> URD). Sources are peers per
// internal/mcp/server.go:123-134, so the cascade uses 'related' edges,
// not 'implements'.
func NewSourcesTrackPrompt() mcp.Prompt {
	return mcp.NewPrompt(sourcesTrackName,
		mcp.WithPromptDescription(sourcesTrackDescription),
		mcp.WithArgument("feature_name",
			mcp.ArgumentDescription("Short name of the feature being discovered."),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("scope",
			mcp.ArgumentDescription("One-line scope statement (optional)."),
		),
	)
}

// HandleSourcesTrack returns the message script for the Sources discovery
// cascade (MRD -> BRD -> URD).
func HandleSourcesTrack(_ context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	feature, err := requireStringArg(req, "feature_name")
	if err != nil {
		return nil, err
	}
	scope := optionalStringArg(req, "scope", "")

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Sources discovery for %q", feature),
		Messages:    sourcesTrackMessages(feature, scope),
	}, nil
}

func sourcesTrackMessages(feature, scope string) []mcp.PromptMessage {
	return []mcp.PromptMessage{
		framingMessage("Sources discovery", feature, sourcesTrackChain),
		phaseMessage(1, "mrd",
			"Call create_document(type=\"mrd\", filename=..., directory=..., title=...). "+
				"Capture market analysis: TAM/SAM/SOM, competitors, timing"+
				scopeClause(scope)+" "+ConfirmationGate,
		),
		phaseMessage(2, "brd",
			"Call create_document(type=\"brd\", ...) to create the BRD document, then call "+
				"add_relation(source=<brd path>, target=<mrd path>, type=\"related\"). "+
				ConfirmationGate,
		),
		phaseMessage(3, "urd",
			"Call create_document(type=\"urd\", ...) to create the URD document, then call "+
				"add_relation(source=<urd path>, target=<brd path>, type=\"related\"). "+
				ConfirmationGate,
		),
		verificationMessage("MRD", "MRD <- BRD <- URD"),
	}
}
