package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Project struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(serverURL string) *Client {
	return &Client{
		BaseURL: serverURL + "/api/v1",
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) get(ctx context.Context, path string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		return fmt.Errorf("invalid response: %w", err)
	}
	return nil
}

func (c *Client) CheckHealth(ctx context.Context) error {
	var result struct {
		Ready bool `json:"ready"`
	}
	if err := c.get(ctx, "/status", &result); err != nil {
		return err
	}
	if !result.Ready {
		return fmt.Errorf("server is not ready")
	}
	return nil
}

func (c *Client) ListProjects(ctx context.Context) ([]Project, error) {
	var projects []Project
	if err := c.get(ctx, "/projects", &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func (c *Client) GetProject(ctx context.Context, id int64) (*Project, error) {
	var project Project
	if err := c.get(ctx, fmt.Sprintf("/projects/%d", id), &project); err != nil {
		return nil, err
	}
	return &project, nil
}
