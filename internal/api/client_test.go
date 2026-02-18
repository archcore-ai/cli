package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	c := NewClient("http://example.com")
	want := "http://example.com/api/v1"
	if c.BaseURL != want {
		t.Errorf("BaseURL = %q, want %q", c.BaseURL, want)
	}
}

func TestCheckHealth(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr bool
	}{
		{
			name: "healthy",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"ready":true}`))
			},
			wantErr: false,
		},
		{
			name: "not ready",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"ready":false}`))
			},
			wantErr: true,
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr: true,
		},
		{
			name: "bad JSON",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`not json`))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.handler)
			defer srv.Close()

			c := NewClient(srv.URL)
			err := c.CheckHealth(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckHealth error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckHealth_ConnectionRefused(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close() // close immediately so connection is refused

	c := NewClient(srv.URL)
	if err := c.CheckHealth(context.Background()); err == nil {
		t.Fatal("expected error for closed server")
	}
}

func TestCheckHealth_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"ready":true}`))
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before request

	c := NewClient(srv.URL)
	if err := c.CheckHealth(ctx); err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestListProjects(t *testing.T) {
	tests := []struct {
		name      string
		handler   http.HandlerFunc
		wantErr   bool
		wantCount int
	}{
		{
			name: "success with projects",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/projects" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("unexpected method: %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`[{"id":1,"name":"project-a"},{"id":2,"name":"project-b"}]`))
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name: "empty list",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`[]`))
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr: true,
		},
		{
			name: "bad JSON",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`not json`))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.handler)
			defer srv.Close()

			c := NewClient(srv.URL)
			projects, err := c.ListProjects(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("ListProjects error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(projects) != tt.wantCount {
				t.Errorf("got %d projects, want %d", len(projects), tt.wantCount)
			}
		})
	}
}

func TestListProjects_ValidatesFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id":42,"name":"my-project"}]`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	projects, err := c.ListProjects(context.Background())
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("got %d projects, want 1", len(projects))
	}
	if projects[0].ID != 42 {
		t.Errorf("ID = %d, want 42", projects[0].ID)
	}
	if projects[0].Name != "my-project" {
		t.Errorf("Name = %q, want %q", projects[0].Name, "my-project")
	}
}

func TestGetProject(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr bool
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/projects/42" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"id":42,"name":"my-project"}`))
			},
			wantErr: false,
		},
		{
			name: "not found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			wantErr: true,
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr: true,
		},
		{
			name: "bad JSON",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`not json`))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.handler)
			defer srv.Close()

			c := NewClient(srv.URL)
			project, err := c.GetProject(context.Background(), 42)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetProject error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if project.ID != 42 {
					t.Errorf("ID = %d, want 42", project.ID)
				}
				if project.Name != "my-project" {
					t.Errorf("Name = %q, want %q", project.Name, "my-project")
				}
			}
		})
	}
}

func TestGetProject_ConnectionRefused(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close()

	c := NewClient(srv.URL)
	_, err := c.GetProject(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error for closed server")
	}
}

func TestListProjects_ConnectionRefused(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close()

	c := NewClient(srv.URL)
	_, err := c.ListProjects(context.Background())
	if err == nil {
		t.Fatal("expected error for closed server")
	}
}
