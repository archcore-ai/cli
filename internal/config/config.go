package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	dirName  = ".archcore"
	fileName = "settings.json"

	SyncTypeNone   = "none"
	SyncTypeCloud  = "cloud"
	SyncTypeOnPrem = "on-prem"
)

// CloudServerURL is the hardcoded URL for cloud sync. Var for test override.
var CloudServerURL = "https://app.archcore.ai"

var subdirs = []string{"vision", "knowledge", "experience"}

type Settings struct {
	Sync        string `json:"sync"`
	ProjectID   *int   `json:"project_id,omitempty"`
	ArchcoreURL string `json:"archcore_url,omitempty"`
}

// NewNoneSettings creates settings with sync disabled.
func NewNoneSettings() *Settings {
	return &Settings{Sync: SyncTypeNone}
}

// NewCloudSettings creates settings for cloud sync.
func NewCloudSettings() *Settings {
	return &Settings{Sync: SyncTypeCloud}
}

// NewOnPremSettings creates settings for on-prem sync with the given URL.
func NewOnPremSettings(url string) *Settings {
	return &Settings{Sync: SyncTypeOnPrem, ArchcoreURL: url}
}

// Validate checks that the settings are internally consistent.
func (s *Settings) Validate() error {
	switch s.Sync {
	case SyncTypeNone:
		if s.ProjectID != nil {
			return fmt.Errorf("sync %q does not allow project_id", SyncTypeNone)
		}
		if s.ArchcoreURL != "" {
			return fmt.Errorf("sync %q does not allow archcore_url", SyncTypeNone)
		}
	case SyncTypeCloud:
		if s.ArchcoreURL != "" {
			return fmt.Errorf("sync %q does not allow archcore_url", SyncTypeCloud)
		}
	case SyncTypeOnPrem:
		if s.ArchcoreURL == "" {
			return fmt.Errorf("sync %q requires archcore_url", SyncTypeOnPrem)
		}
	default:
		return fmt.Errorf("unknown sync type %q", s.Sync)
	}
	return nil
}

// ServerURL returns the server URL for the current sync type.
func (s *Settings) ServerURL() string {
	switch s.Sync {
	case SyncTypeCloud:
		return CloudServerURL
	case SyncTypeOnPrem:
		return s.ArchcoreURL
	default:
		return ""
	}
}

// allowedFields defines which JSON fields are valid per sync type (besides "sync" itself).
var allowedFields = map[string]map[string]bool{
	SyncTypeNone:   {},
	SyncTypeCloud:  {"project_id": true},
	SyncTypeOnPrem: {"project_id": true, "archcore_url": true},
}

// requiredFields defines which JSON fields must be present per sync type.
var requiredFields = map[string][]string{
	SyncTypeNone:   {},
	SyncTypeCloud:  {"project_id"},
	SyncTypeOnPrem: {"project_id", "archcore_url"},
}

func (s Settings) MarshalJSON() ([]byte, error) {
	switch s.Sync {
	case SyncTypeNone:
		return json.Marshal(struct {
			Sync string `json:"sync"`
		}{Sync: s.Sync})

	case SyncTypeCloud:
		return json.Marshal(struct {
			Sync      string `json:"sync"`
			ProjectID *int   `json:"project_id"`
		}{Sync: s.Sync, ProjectID: s.ProjectID})

	case SyncTypeOnPrem:
		return json.Marshal(struct {
			Sync        string `json:"sync"`
			ProjectID   *int   `json:"project_id"`
			ArchcoreURL string `json:"archcore_url"`
		}{Sync: s.Sync, ProjectID: s.ProjectID, ArchcoreURL: s.ArchcoreURL})

	default:
		return nil, fmt.Errorf("unknown sync type %q", s.Sync)
	}
}

func (s *Settings) UnmarshalJSON(data []byte) error {
	// Decode into a raw map to check fields.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Extract and validate sync field.
	syncRaw, ok := raw["sync"]
	if !ok {
		return fmt.Errorf("missing required field \"sync\"")
	}
	var syncType string
	if err := json.Unmarshal(syncRaw, &syncType); err != nil {
		return fmt.Errorf("field \"sync\" must be a string")
	}

	allowed, knownType := allowedFields[syncType]
	if !knownType {
		return fmt.Errorf("unknown sync type %q", syncType)
	}

	// Check for unknown fields.
	for key := range raw {
		if key == "sync" {
			continue
		}
		if !allowed[key] {
			return fmt.Errorf("field %q is not allowed for sync type %q", key, syncType)
		}
	}

	// Check for required fields.
	for _, req := range requiredFields[syncType] {
		if _, ok := raw[req]; !ok {
			return fmt.Errorf("missing required field %q for sync type %q", req, syncType)
		}
	}

	s.Sync = syncType

	// Decode project_id if present.
	if pidRaw, ok := raw["project_id"]; ok {
		// Accept null or number.
		if string(pidRaw) == "null" {
			s.ProjectID = nil
		} else {
			var pid int
			if err := json.Unmarshal(pidRaw, &pid); err != nil {
				return fmt.Errorf("field \"project_id\" must be null or a number")
			}
			s.ProjectID = &pid
		}
	}

	// Decode archcore_url if present.
	if urlRaw, ok := raw["archcore_url"]; ok {
		var url string
		if err := json.Unmarshal(urlRaw, &url); err != nil {
			return fmt.Errorf("field \"archcore_url\" must be a string")
		}
		if url == "" {
			return fmt.Errorf("field \"archcore_url\" must not be empty")
		}
		s.ArchcoreURL = url
	}

	return nil
}

func settingsPath(baseDir string) string {
	return filepath.Join(baseDir, dirName, fileName)
}

func Load(baseDir string) (*Settings, error) {
	data, err := os.ReadFile(settingsPath(baseDir))
	if err != nil {
		return nil, err
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("invalid settings: %w", err)
	}
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("invalid settings: %w", err)
	}
	return &s, nil
}

func Save(baseDir string, s *Settings) error {
	if err := s.Validate(); err != nil {
		return fmt.Errorf("refusing to save invalid settings: %w", err)
	}
	dir := filepath.Join(baseDir, dirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(settingsPath(baseDir), data, 0o644)
}

func InitDir(baseDir string) error {
	root := filepath.Join(baseDir, dirName)
	for _, sub := range subdirs {
		if err := os.MkdirAll(filepath.Join(root, sub), 0o755); err != nil {
			return err
		}
	}
	return nil
}

func DirExists(baseDir string) bool {
	info, err := os.Stat(filepath.Join(baseDir, dirName))
	return err == nil && info.IsDir()
}
