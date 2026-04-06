package resource

import "testing"

func TestExtractItemFieldFallsBackToSingleListEntry(t *testing.T) {
	t.Parallel()

	response := map[string]any{
		"servers": []any{
			map[string]any{
				"server": map[string]any{
					"name":    "acme/test",
					"version": "1.0.0",
				},
			},
		},
	}

	item := (&MCPServerType{}).ExtractItem(response)
	if got := str(item, "name"); got != "acme/test" {
		t.Fatalf("ExtractItem() name = %q, want %q", got, "acme/test")
	}
	if got := str(item, "version"); got != "1.0.0" {
		t.Fatalf("ExtractItem() version = %q, want %q", got, "1.0.0")
	}
}
