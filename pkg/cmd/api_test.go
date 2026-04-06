package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/agentregistry-dev/ar/pkg/client"
	"github.com/agentregistry-dev/ar/pkg/printer"
)

func TestAPIVersionCommandPrintsTable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() != "/v0/version" {
			t.Fatalf("path = %q, want %q", r.URL.EscapedPath(), "/v0/version")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"version":    "v1.2.3",
			"git_commit": "abc1234",
			"build_time": "2026-04-06T12:00:00Z",
		})
	}))
	t.Cleanup(server.Close)

	oldClient := apiClient
	oldFormat := outputFormat
	t.Cleanup(func() {
		apiClient = oldClient
		outputFormat = oldFormat
	})

	apiClient = client.New(server.URL, "")
	outputFormat = printer.FormatTable

	var out bytes.Buffer
	cmd := newAPICmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	text := out.String()
	for _, needle := range []string{"VERSION", "GIT COMMIT", "BUILD TIME", "v1.2.3", "abc1234"} {
		if !strings.Contains(text, needle) {
			t.Fatalf("output missing %q:\n%s", needle, text)
		}
	}
}

func TestAPIJWKSCommandPrintsJSONWhenFormatIsTable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() != "/v0/token-provider/jwks.json" {
			t.Fatalf("path = %q, want %q", r.URL.EscapedPath(), "/v0/token-provider/jwks.json")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"keys": []any{
				map[string]any{
					"kid": "key-1",
					"kty": "RSA",
				},
			},
		})
	}))
	t.Cleanup(server.Close)

	oldClient := apiClient
	oldFormat := outputFormat
	t.Cleanup(func() {
		apiClient = oldClient
		outputFormat = oldFormat
	})

	apiClient = client.New(server.URL, "")
	outputFormat = printer.FormatTable

	var out bytes.Buffer
	cmd := newAPICmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"jwks"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	text := out.String()
	for _, needle := range []string{`"keys"`, `"kid": "key-1"`, `"kty": "RSA"`} {
		if !strings.Contains(text, needle) {
			t.Fatalf("output missing %q:\n%s", needle, text)
		}
	}
}

func TestAPIPingCommandPrintsOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() != "/v0/ping" {
			t.Fatalf("path = %q, want %q", r.URL.EscapedPath(), "/v0/ping")
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	oldClient := apiClient
	oldFormat := outputFormat
	t.Cleanup(func() {
		apiClient = oldClient
		outputFormat = oldFormat
	})

	apiClient = client.New(server.URL, "")
	outputFormat = printer.FormatTable

	var out bytes.Buffer
	cmd := newAPICmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"ping"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	text := out.String()
	for _, needle := range []string{"STATUS", "ok"} {
		if !strings.Contains(text, needle) {
			t.Fatalf("output missing %q:\n%s", needle, text)
		}
	}
}
