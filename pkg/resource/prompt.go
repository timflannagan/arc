package resource

func init() { Register(&PromptType{}) }

// PromptType implements ResourceType for Prompt resources.
type PromptType struct{}

func (p *PromptType) Kind() string     { return "Prompt" }
func (p *PromptType) Singular() string { return "prompt" }
func (p *PromptType) Plural() string   { return "prompts" }
func (p *PromptType) APIPath() string  { return "/prompts" }

func (p *PromptType) TableColumns() []string {
	return []string{"Name", "Version", "Description"}
}

func (p *PromptType) TableRow(data map[string]any) []string {
	prompt := extractNested(data, "prompt")
	if prompt == nil {
		prompt = data
	}
	return []string{
		str(prompt, "name"),
		str(prompt, "version"),
		str(prompt, "description"),
	}
}

func (p *PromptType) ToCreatePayload(r *Resource) (any, error) {
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

func (p *PromptType) ExtractItem(response map[string]any) map[string]any {
	return extractItemField(response, "prompt", "prompts")
}

func (p *PromptType) ExtractList(response map[string]any) []map[string]any {
	return extractListField(response, "prompts")
}

func (p *PromptType) ToResource(response map[string]any) *Resource {
	item := p.ExtractItem(response)
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
		Kind:       "Prompt",
		Metadata:   Metadata{Name: name, Version: version},
		Spec:       spec,
	}
}
