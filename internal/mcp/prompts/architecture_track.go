package prompts

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	architectureTrackName        = "architecture_track"
	architectureTrackDescription = "Run Architecture track: ADR -> Spec -> Plan; spec implements adr, plan implements spec."
	architectureTrackChain       = "ADR -> Spec -> Plan"
)

// NewArchitectureTrackPrompt returns the prompt definition for the
// Architecture track (ADR -> Spec -> Plan). The optional component_name
// argument lets callers anchor the framing message to a specific component.
func NewArchitectureTrackPrompt() mcp.Prompt {
	return mcp.NewPrompt(architectureTrackName,
		mcp.WithPromptDescription(architectureTrackDescription),
		mcp.WithArgument("feature_name",
			mcp.ArgumentDescription("Short name of the feature being specified."),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("scope",
			mcp.ArgumentDescription("One-line scope statement (optional)."),
		),
		mcp.WithArgument("component_name",
			mcp.ArgumentDescription("Specific component the architecture targets (optional)."),
		),
	)
}

// HandleArchitectureTrack returns the message script for the Architecture
// track (ADR -> Spec -> Plan).
func HandleArchitectureTrack(_ context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	feature, err := requireStringArg(req, "feature_name")
	if err != nil {
		return nil, err
	}
	scope := optionalStringArg(req, "scope", "")
	component := optionalStringArg(req, "component_name", "")

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Architecture track for %q", feature),
		Messages:    architectureTrackMessages(feature, scope, component),
	}, nil
}

func architectureTrackMessages(feature, scope, component string) []mcp.PromptMessage {
	var componentPhrase string
	if component != "" {
		componentPhrase = "for the " + component + " component"
	}

	return []mcp.PromptMessage{
		framingMessage("Architecture", feature, architectureTrackChain, componentPhrase),
		phaseMessage(1, "adr",
			"Call create_document(type=\"adr\", filename=..., directory=..., title=...). "+
				"Capture the decision context, options, and chosen architectural approach"+
				scopeClause(scope)+" "+ConfirmationGate,
		),
		phaseMessage(2, "spec",
			"Call create_document(type=\"spec\", ...) to create the Spec document, then call "+
				"add_relation(source=<spec path>, target=<adr path>, type=\"implements\"). "+
				ConfirmationGate,
		),
		phaseMessage(3, "plan",
			"Call create_document(type=\"plan\", ...) to create the Plan document, then call "+
				"add_relation(source=<plan path>, target=<spec path>, type=\"implements\"). "+
				ConfirmationGate,
		),
		verificationMessage("ADR", "ADR <- Spec <- Plan"),
	}
}
