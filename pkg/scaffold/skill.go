package scaffold

import "fmt"

// SkillOptions extends Options with skill-specific fields.
type SkillOptions struct {
	Options
	Category string
	Image    string
}

const skillYAML = `apiVersion: ar.dev/v1alpha1
kind: Skill
metadata:
  name: {{ .Name }}
  version: "{{ .Version }}"
spec:
  title: {{ .Name }}
  category: {{ .Category }}
  description: "{{ .Description }}"
  status: active
`

const skillMD = `---
name: {{ .Name }}
description: {{ .Description }}
---

# {{ .Name }}

## Overview

Describe when and how to use this skill.

## Instructions

Step-by-step instructions for the agent to follow.

1. First, do this.
2. Then, do that.
3. Finally, return the result.
`

// Skill scaffolds a new skill project.
func Skill(opts SkillOptions) error {
	if opts.Version == "" {
		opts.Version = "0.1.0"
	}
	if opts.Category == "" {
		opts.Category = "general"
	}
	if opts.Description == "" {
		opts.Description = fmt.Sprintf("Skill %s", opts.Name)
	}

	dir := opts.Dir()
	if err := ensureDir(dir); err != nil {
		return err
	}

	files := []struct {
		name string
		tmpl string
	}{
		{"skill.yaml", skillYAML},
		{"SKILL.md", skillMD},
	}

	for _, f := range files {
		if err := writeTemplate(dir, f.name, f.tmpl, opts); err != nil {
			return err
		}
	}

	return nil
}
