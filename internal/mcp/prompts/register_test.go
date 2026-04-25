package prompts

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// TestRegisterAll wires RegisterAll into a fresh MCPServer and uses the
// in-process client to introspect the registered prompts via the public
// prompts/list path. mcp-go does not export the internal prompts map, so the
// in-process client is the canonical way to verify registration.
func TestRegisterAll(t *testing.T) {
	t.Parallel()

	s := server.NewMCPServer(
		"archcore-test",
		"0.0.0",
		server.WithPromptCapabilities(true),
	)
	RegisterAll(s)

	c, err := client.NewInProcessClient(s)
	if err != nil {
		t.Fatalf("NewInProcessClient: %v", err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Errorf("client.Close: %v", err)
		}
	})

	ctx := context.Background()
	if err := c.Start(ctx); err != nil {
		t.Fatalf("client.Start: %v", err)
	}
	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{
		Name:    "archcore-prompts-register-test",
		Version: "0.0.0",
	}
	if _, err := c.Initialize(ctx, initReq); err != nil {
		t.Fatalf("client.Initialize: %v", err)
	}

	res, err := c.ListPrompts(ctx, mcp.ListPromptsRequest{})
	if err != nil {
		t.Fatalf("ListPrompts: %v", err)
	}

	// Phase 2 ships exactly five prompts.
	wantNames := map[string]bool{
		"iso_track":          false,
		"sources_track":      false,
		"product_track":      false,
		"standard_track":     false,
		"architecture_track": false,
	}
	if len(res.Prompts) != len(wantNames) {
		t.Fatalf("prompt count = %d, want %d; got %+v", len(res.Prompts), len(wantNames), res.Prompts)
	}
	for _, p := range res.Prompts {
		if _, expected := wantNames[p.Name]; !expected {
			t.Errorf("unexpected prompt registered: %q", p.Name)
			continue
		}
		wantNames[p.Name] = true

		if p.Description == "" {
			t.Errorf("prompt %q has empty description", p.Name)
		}

		var hasRequiredFeatureName bool
		for _, a := range p.Arguments {
			if a.Name == "feature_name" {
				hasRequiredFeatureName = a.Required
				break
			}
		}
		if !hasRequiredFeatureName {
			t.Errorf("prompt %q must declare feature_name as a required argument", p.Name)
		}
	}
	for name, seen := range wantNames {
		if !seen {
			t.Errorf("expected prompt %q not found in ListPrompts result", name)
		}
	}
}
