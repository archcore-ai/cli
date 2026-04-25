package prompts

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	productTrackName        = "product_track"
	productTrackDescription = "Run Product track: PRD -> Plan, plan implements prd."
	productTrackChain       = "PRD -> Plan"
)

// NewProductTrackPrompt returns the prompt definition for the Product flow
// (PRD -> Plan).
func NewProductTrackPrompt() mcp.Prompt {
	return mcp.NewPrompt(productTrackName,
		mcp.WithPromptDescription(productTrackDescription),
		mcp.WithArgument("feature_name",
			mcp.ArgumentDescription("Short name of the feature being specified."),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("scope",
			mcp.ArgumentDescription("One-line scope statement (optional)."),
		),
	)
}

// HandleProductTrack returns the message script for the Product flow
// (PRD -> Plan).
func HandleProductTrack(_ context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	feature, err := requireStringArg(req, "feature_name")
	if err != nil {
		return nil, err
	}
	scope := optionalStringArg(req, "scope", "")

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Product track for %q", feature),
		Messages:    productTrackMessages(feature, scope),
	}, nil
}

func productTrackMessages(feature, scope string) []mcp.PromptMessage {
	return []mcp.PromptMessage{
		framingMessage("Product", feature, productTrackChain),
		phaseMessage(1, "prd",
			"Call create_document(type=\"prd\", filename=..., directory=..., title=...). "+
				"Use the standard PRD template; cover problem, users, requirements, and solution overview"+
				scopeClause(scope)+" "+ConfirmationGate,
		),
		phaseMessage(2, "plan",
			"Call create_document(type=\"plan\", ...) to create the Plan document, then call "+
				"add_relation(source=<plan path>, target=<prd path>, type=\"implements\"). "+
				ConfirmationGate,
		),
		verificationMessage("PRD", "PRD <- Plan"),
	}
}
