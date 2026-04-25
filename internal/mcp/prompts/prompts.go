// Package prompts registers MCP prompts that orchestrate Archcore's
// multi-document tracks (per .archcore/mcp/mcp-prompts-for-tracks-only.adr.md
// and .archcore/mcp/mcp-prompts-orchestration-only.rule.md).
//
// Each prompt walks an MCP-aware client through a fixed cascade of
// create_document calls plus the relations that link them — exactly the
// surface that cannot be expressed cleanly as a single tool plus instruction
// text. Single-document workflows belong in templates, diagnostic flows
// belong in plugin skills; both are explicitly out of scope here.
package prompts

import (
	"github.com/mark3labs/mcp-go/server"
)

// RegisterAll wires every track prompt into s. Phase 2 lights up the full
// surface: ISO 29148 cascade, Sources discovery, Product, Standard, and
// Architecture tracks.
func RegisterAll(s *server.MCPServer) {
	s.AddPrompt(NewISOTrackPrompt(), HandleISOTrack)
	s.AddPrompt(NewSourcesTrackPrompt(), HandleSourcesTrack)
	s.AddPrompt(NewProductTrackPrompt(), HandleProductTrack)
	s.AddPrompt(NewStandardTrackPrompt(), HandleStandardTrack)
	s.AddPrompt(NewArchitectureTrackPrompt(), HandleArchitectureTrack)
}
