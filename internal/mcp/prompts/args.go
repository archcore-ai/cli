package prompts

import (
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// requireStringArg returns the value of a required prompt argument or an
// error naming the argument when it is missing or empty. mcp-go v0.49 models
// GetPromptParams.Arguments as map[string]string, so the comma-ok read here
// is intentional even though the zero value is "" — it lets us distinguish
// "missing key" from "empty value" only via emptiness, which is the right
// behavior for a required argument anyway.
func requireStringArg(req mcp.GetPromptRequest, name string) (string, error) {
	v := req.Params.Arguments[name]
	if v == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return v, nil
}

// optionalStringArg returns the argument value when present and non-empty,
// otherwise the supplied default. Empty string is treated as "unset" so
// clients can pass "" to mean "use the default" without special-casing.
func optionalStringArg(req mcp.GetPromptRequest, name, defaultValue string) string {
	if v := req.Params.Arguments[name]; v != "" {
		return v
	}
	return defaultValue
}
