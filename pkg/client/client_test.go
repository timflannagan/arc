package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUsesLatestVersionAndEscapesName(t *testing.T) {
	t.Parallel()

	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		_ = json.NewEncoder(w).Encode(map[string]any{
			"agent": map[string]any{"name": "acme/test", "version": "latest"},
		})
	}))
	t.Cleanup(server.Close)

	c := New(server.URL, "")
	_, err := c.Get("/agents", "acme/test")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	want := "/v0/agents/acme%2Ftest/versions/latest"
	if gotPath != want {
		t.Fatalf("Get() path = %q, want %q", gotPath, want)
	}
}

func TestGetVersionEscapesVersion(t *testing.T) {
	t.Parallel()

	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		_ = json.NewEncoder(w).Encode(map[string]any{
			"agent": map[string]any{"name": "acme/test", "version": "1.0.0+build"},
		})
	}))
	t.Cleanup(server.Close)

	c := New(server.URL, "")
	_, err := c.GetVersion("/agents", "acme/test", "1.0.0+build")
	if err != nil {
		t.Fatalf("GetVersion() error = %v", err)
	}

	want := "/v0/agents/acme%2Ftest/versions/1.0.0%2Bbuild"
	if gotPath != want {
		t.Fatalf("GetVersion() path = %q, want %q", gotPath, want)
	}
}

func TestDeleteEscapesSegments(t *testing.T) {
	t.Parallel()

	var gotPath string
	var gotMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)

	c := New(server.URL, "")
	if err := c.Delete("/servers", "acme/test", "1.0.0+build"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if gotMethod != http.MethodDelete {
		t.Fatalf("Delete() method = %q, want %q", gotMethod, http.MethodDelete)
	}

	want := "/v0/servers/acme%2Ftest/versions/1.0.0%2Bbuild"
	if gotPath != want {
		t.Fatalf("Delete() path = %q, want %q", gotPath, want)
	}
}

func TestPingUsesV0Path(t *testing.T) {
	t.Parallel()

	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	c := New(server.URL, "")
	if err := c.Ping(); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}

	if gotPath != "/v0/ping" {
		t.Fatalf("Ping() path = %q, want %q", gotPath, "/v0/ping")
	}
}

func TestGetAnyNormalizesRelativeAndV0Paths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "relative", path: "version", want: "/v0/version"},
		{name: "slash relative", path: "/version", want: "/v0/version"},
		{name: "explicit v0", path: "/v0/version", want: "/v0/version"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var gotPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.EscapedPath()
				_ = json.NewEncoder(w).Encode(map[string]any{"version": "v1.2.3"})
			}))
			t.Cleanup(server.Close)

			c := New(server.URL, "")
			_, err := c.GetAny(tt.path)
			if err != nil {
				t.Fatalf("GetAny() error = %v", err)
			}

			if gotPath != tt.want {
				t.Fatalf("GetAny() path = %q, want %q", gotPath, tt.want)
			}
		})
	}
}
