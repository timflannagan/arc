package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRootGetAllPrintsSummaryTable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.EscapedPath() {
		case "/v0/agents":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"agents": []any{
					map[string]any{
						"agent": map[string]any{
							"name":    "agent-one",
							"version": "1.0.0",
							"status":  "published",
						},
					},
				},
			})
		case "/v0/servers":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"servers": []any{
					map[string]any{
						"server": map[string]any{
							"name":    "server-one",
							"version": "2.0.0",
							"status":  "active",
						},
					},
				},
			})
		case "/v0/prompts":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"prompts": []any{
					map[string]any{
						"prompt": map[string]any{
							"name":    "prompt-one",
							"version": "3.0.0",
						},
					},
				},
			})
		case "/v0/skills":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"skills": []any{
					map[string]any{
						"skill": map[string]any{
							"name":    "skill-one",
							"version": "4.0.0",
							"status":  "ready",
						},
					},
				},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.EscapedPath())
		}
	}))
	t.Cleanup(server.Close)

	restore := snapshotCommandGlobals()
	t.Cleanup(restore)

	var out bytes.Buffer
	cmd := Root()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--server", server.URL, "get", "all"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	text := out.String()
	for _, needle := range []string{
		"KIND", "NAME", "VERSION", "STATUS",
		"Agent", "agent-one", "1.0.0", "published",
		"MCPServer", "server-one", "2.0.0", "active",
		"Prompt", "prompt-one", "3.0.0",
		"Skill", "skill-one", "4.0.0", "ready",
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf("output missing %q:\n%s", needle, text)
		}
	}
}

func TestRootGetAllJSONGroupsByType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.EscapedPath() {
		case "/v0/agents":
			_ = json.NewEncoder(w).Encode(map[string]any{"agents": []any{}})
		case "/v0/servers":
			_ = json.NewEncoder(w).Encode(map[string]any{"servers": []any{}})
		case "/v0/prompts":
			_ = json.NewEncoder(w).Encode(map[string]any{"prompts": []any{}})
		case "/v0/skills":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"skills": []any{
					map[string]any{
						"skill": map[string]any{
							"name":    "skill-one",
							"version": "4.0.0",
						},
					},
				},
				"metadata": map[string]any{"count": 1},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.EscapedPath())
		}
	}))
	t.Cleanup(server.Close)

	restore := snapshotCommandGlobals()
	t.Cleanup(restore)

	var out bytes.Buffer
	cmd := Root()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--server", server.URL, "get", "all", "-o", "json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v\noutput:\n%s", err, out.String())
	}

	for _, key := range []string{"agents", "mcpservers", "prompts", "skills"} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("output missing key %q: %#v", key, payload)
		}
	}
}
