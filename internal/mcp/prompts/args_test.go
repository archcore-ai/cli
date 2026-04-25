package prompts

import (
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func newRequest(args map[string]string) mcp.GetPromptRequest {
	req := mcp.GetPromptRequest{}
	req.Params.Arguments = args
	return req
}

func TestRequireStringArg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     map[string]string
		key      string
		wantVal  string
		wantErr  bool
		errMatch string
	}{
		{
			name:    "present",
			args:    map[string]string{"feature_name": "auth"},
			key:     "feature_name",
			wantVal: "auth",
		},
		{
			name:     "missing",
			args:     map[string]string{},
			key:      "feature_name",
			wantErr:  true,
			errMatch: "feature_name",
		},
		{
			name:     "empty string treated as missing",
			args:     map[string]string{"feature_name": ""},
			key:      "feature_name",
			wantErr:  true,
			errMatch: "feature_name",
		},
		{
			name:     "nil arguments map",
			args:     nil,
			key:      "feature_name",
			wantErr:  true,
			errMatch: "feature_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := requireStringArg(newRequest(tt.args), tt.key)
			if (err != nil) != tt.wantErr {
				t.Fatalf("requireStringArg err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.errMatch) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errMatch)
				}
				return
			}
			if got != tt.wantVal {
				t.Errorf("requireStringArg = %q, want %q", got, tt.wantVal)
			}
		})
	}
}

func TestOptionalStringArg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       map[string]string
		key        string
		defaultVal string
		want       string
	}{
		{
			name:       "present returns value",
			args:       map[string]string{"scope": "auth login"},
			key:        "scope",
			defaultVal: "default",
			want:       "auth login",
		},
		{
			name:       "missing returns default",
			args:       map[string]string{},
			key:        "scope",
			defaultVal: "default",
			want:       "default",
		},
		{
			name:       "empty returns default",
			args:       map[string]string{"scope": ""},
			key:        "scope",
			defaultVal: "default",
			want:       "default",
		},
		{
			name:       "nil map returns default",
			args:       nil,
			key:        "scope",
			defaultVal: "fallback",
			want:       "fallback",
		},
		{
			name:       "empty default with missing key",
			args:       map[string]string{},
			key:        "scope",
			defaultVal: "",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := optionalStringArg(newRequest(tt.args), tt.key, tt.defaultVal)
			if got != tt.want {
				t.Errorf("optionalStringArg = %q, want %q", got, tt.want)
			}
		})
	}
}
