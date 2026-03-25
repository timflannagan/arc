package resource

import "fmt"

func init() { Register(&AgentType{}) }

// AgentType implements ResourceType for Agent resources.
type AgentType struct{}

func (a *AgentType) Kind() string    { return "Agent" }
func (a *AgentType) Singular() string { return "agent" }
func (a *AgentType) Plural() string  { return "agents" }
func (a *AgentType) APIPath() string { return "/agents" }

func (a *AgentType) TableColumns() []string {
	return []string{"Name", "Version", "Framework", "Model", "Status"}
}

func (a *AgentType) TableRow(data map[string]any) []string {
	agent := extractNested(data, "agent")
	if agent == nil {
		agent = data
	}
	return []string{
		str(agent, "agentName"),
		str(agent, "version"),
		str(agent, "framework"),
		str(agent, "modelName"),
		str(agent, "status"),
	}
}

func (a *AgentType) ToCreatePayload(r *Resource) (any, error) {
	payload := map[string]any{
		"agentName": r.Metadata.Name,
		"version":   r.Metadata.Version,
	}
	// Copy spec fields into payload.
	for k, v := range r.Spec {
		payload[k] = v
	}
	// Ensure name and version from metadata take precedence.
	payload["agentName"] = r.Metadata.Name
	if r.Metadata.Version != "" {
		payload["version"] = r.Metadata.Version
	}
	return payload, nil
}

func (a *AgentType) ExtractItem(response map[string]any) map[string]any {
	if item, ok := response["agent"]; ok {
		if m, ok := item.(map[string]any); ok {
			return m
		}
	}
	return response
}

func (a *AgentType) ExtractList(response map[string]any) []map[string]any {
	return extractListField(response, "agents")
}

func (a *AgentType) ToResource(response map[string]any) *Resource {
	item := a.ExtractItem(response)
	name := str(item, "agentName")
	version := str(item, "version")

	spec := make(map[string]any)
	for k, v := range item {
		switch k {
		case "agentName", "version":
			// These go in metadata, not spec.
		default:
			spec[k] = v
		}
	}

	return &Resource{
		APIVersion: "ar.dev/v1alpha1",
		Kind:       "Agent",
		Metadata:   Metadata{Name: name, Version: version},
		Spec:       spec,
	}
}

// --- helpers shared across resource types ---

func str(data map[string]any, key string) string {
	if v, ok := data[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func extractNested(data map[string]any, key string) map[string]any {
	if v, ok := data[key]; ok {
		if m, ok := v.(map[string]any); ok {
			return m
		}
	}
	return nil
}

func extractListField(data map[string]any, key string) []map[string]any {
	items, ok := data[key]
	if !ok {
		// Try "items" as fallback.
		items, ok = data["items"]
		if !ok {
			return nil
		}
	}
	slice, ok := items.([]any)
	if !ok {
		return nil
	}
	result := make([]map[string]any, 0, len(slice))
	for _, item := range slice {
		if m, ok := item.(map[string]any); ok {
			result = append(result, m)
		}
	}
	return result
}
