package resource

func init() { Register(&SkillType{}) }

// SkillType implements ResourceType for Skill resources.
type SkillType struct{}

func (s *SkillType) Kind() string     { return "Skill" }
func (s *SkillType) Singular() string { return "skill" }
func (s *SkillType) Plural() string   { return "skills" }
func (s *SkillType) APIPath() string  { return "/skills" }

func (s *SkillType) TableColumns() []string {
	return []string{"Name", "Version", "Category", "Status"}
}

func (s *SkillType) TableRow(data map[string]any) []string {
	skill := extractNested(data, "skill")
	if skill == nil {
		skill = data
	}
	return []string{
		str(skill, "name"),
		str(skill, "version"),
		str(skill, "category"),
		str(skill, "status"),
	}
}

func (s *SkillType) ToCreatePayload(r *Resource) (any, error) {
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

func (s *SkillType) ExtractItem(response map[string]any) map[string]any {
	if item, ok := response["skill"]; ok {
		if data, ok := item.(map[string]any); ok {
			return data
		}
	}
	return response
}

func (s *SkillType) ExtractList(response map[string]any) []map[string]any {
	return extractListField(response, "skills")
}
