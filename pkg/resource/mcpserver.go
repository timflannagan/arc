package resource

func init() { Register(&MCPServerType{}) }

// MCPServerType implements ResourceType for MCP Server resources.
type MCPServerType struct{}

func (m *MCPServerType) Kind() string     { return "MCPServer" }
func (m *MCPServerType) Singular() string { return "mcpserver" }
func (m *MCPServerType) Plural() string   { return "mcpservers" }
func (m *MCPServerType) APIPath() string  { return "/servers" }

func (m *MCPServerType) TableColumns() []string {
	return []string{"Name", "Version", "Status"}
}

func (m *MCPServerType) TableRow(data map[string]any) []string {
	server := extractNested(data, "server")
	if server == nil {
		server = data
	}
	return []string{
		str(server, "name"),
		str(server, "version"),
		str(server, "status"),
	}
}

func (m *MCPServerType) ToCreatePayload(r *Resource) (any, error) {
	payload := map[string]any{
		"name":    r.Metadata.Name,
		"version": r.Metadata.Version,
	}
	for k, v := range r.Spec {
		payload[k] = v
	}
	payload["name"] = r.Metadata.Name
	if r.Metadata.Version != "" {
		payload["version"] = r.Metadata.Version
	}
	return payload, nil
}

func (m *MCPServerType) ExtractItem(response map[string]any) map[string]any {
	if item, ok := response["server"]; ok {
		if data, ok := item.(map[string]any); ok {
			return data
		}
	}
	return response
}

func (m *MCPServerType) ExtractList(response map[string]any) []map[string]any {
	return extractListField(response, "servers")
}

func (m *MCPServerType) ToResource(response map[string]any) *Resource {
	item := m.ExtractItem(response)
	name := str(item, "name")
	version := str(item, "version")

	spec := make(map[string]any)
	for k, v := range item {
		switch k {
		case "name", "version":
		default:
			spec[k] = v
		}
	}

	return &Resource{
		APIVersion: "ar.dev/v1alpha1",
		Kind:       "MCPServer",
		Metadata:   Metadata{Name: name, Version: version},
		Spec:       spec,
	}
}
