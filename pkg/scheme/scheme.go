// Package scheme handles parsing YAML documents with apiVersion/kind dispatch.
// It supports multi-document YAML files (--- separated).
package scheme

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/agentregistry-dev/ar/pkg/resource"
	"gopkg.in/yaml.v3"
)

const (
	// APIVersion is the current API version for resource manifests.
	APIVersion = "ar.dev/v1alpha1"
)

// DecodeFile reads a YAML file and returns all resources found in it.
// Supports multi-document files separated by "---".
func DecodeFile(path string) ([]*resource.Resource, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return Decode(data)
}

// Decode parses YAML bytes into resources. Supports multi-document YAML.
func Decode(data []byte) ([]*resource.Resource, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	var resources []*resource.Resource
	for {
		var r resource.Resource
		err := decoder.Decode(&r)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parsing YAML: %w", err)
		}
		if r.Kind == "" {
			return nil, fmt.Errorf("resource missing required field 'kind'")
		}
		if r.APIVersion == "" {
			return nil, fmt.Errorf("resource %q missing required field 'apiVersion'", r.Metadata.Name)
		}
		if r.Metadata.Name == "" {
			return nil, fmt.Errorf("resource of kind %q missing metadata.name", r.Kind)
		}
		resources = append(resources, &r)
	}

	if len(resources) == 0 {
		return nil, fmt.Errorf("no resources found in input")
	}

	return resources, nil
}

// Validate checks that all resources reference known kinds and a supported API version.
func Validate(resources []*resource.Resource) error {
	for _, r := range resources {
		if r.APIVersion != APIVersion {
			return fmt.Errorf("unsupported apiVersion %q for %s/%s; expected %s",
				r.APIVersion, r.Kind, r.Metadata.Name, APIVersion)
		}
		if _, err := resource.LookupByKind(r.Kind); err != nil {
			return err
		}
	}
	return nil
}

// Encode serializes resources to YAML bytes, separated by "---".
func Encode(resources ...*resource.Resource) ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	for _, r := range resources {
		if err := encoder.Encode(r); err != nil {
			return nil, fmt.Errorf("encoding resource %s/%s: %w", r.Kind, r.Metadata.Name, err)
		}
	}

	if err := encoder.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
