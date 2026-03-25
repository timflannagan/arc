// Package resource defines the resource type registry pattern.
// Adding a new resource type means implementing ResourceType and calling Register().
package resource

import (
	"fmt"
	"strings"
	"sync"
)

// Resource is the universal declarative YAML/JSON structure, modeled after
// Kubernetes resource manifests.
type Resource struct {
	APIVersion string         `yaml:"apiVersion" json:"apiVersion"`
	Kind       string         `yaml:"kind" json:"kind"`
	Metadata   Metadata       `yaml:"metadata" json:"metadata"`
	Spec       map[string]any `yaml:"spec" json:"spec"`
}

// Metadata holds identity and labeling information for a resource.
type Metadata struct {
	Name    string            `yaml:"name" json:"name"`
	Version string            `yaml:"version,omitempty" json:"version,omitempty"`
	Labels  map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// ResourceType defines how a particular kind of resource maps to the registry
// API. Each first-class resource (Agent, MCPServer, Skill, Prompt) implements
// this interface and registers itself at init time.
type ResourceType interface {
	// Kind returns the resource kind string used in YAML (e.g. "Agent").
	Kind() string

	// Singular returns the lowercase singular name (e.g. "agent").
	Singular() string

	// Plural returns the lowercase plural name (e.g. "agents").
	Plural() string

	// APIPath returns the v0 API path segment (e.g. "/v0/agents").
	APIPath() string

	// TableColumns returns column headers for tabular output.
	TableColumns() []string

	// TableRow extracts a single row of table values from an API response item.
	TableRow(data map[string]any) []string

	// ToCreatePayload converts a resource spec + metadata into the JSON body
	// the registry API expects for creation/publish.
	ToCreatePayload(r *Resource) (any, error)

	// ExtractItem pulls the inner resource object from an API response
	// (e.g. response["agent"] for agent responses).
	ExtractItem(response map[string]any) map[string]any

	// ExtractList pulls the list of resource objects from a list API response.
	ExtractList(response map[string]any) []map[string]any
}

var (
	mu       sync.RWMutex
	registry = map[string]ResourceType{}
	aliases  = map[string]string{} // singular/plural → kind
)

// Register adds a resource type to the global registry.
func Register(rt ResourceType) {
	mu.Lock()
	defer mu.Unlock()

	kind := rt.Kind()
	registry[kind] = rt
	aliases[strings.ToLower(kind)] = kind
	aliases[rt.Singular()] = kind
	aliases[rt.Plural()] = kind
}

// Lookup finds a resource type by kind, singular, or plural name.
// The lookup is case-insensitive.
func Lookup(name string) (ResourceType, error) {
	mu.RLock()
	defer mu.RUnlock()

	lower := strings.ToLower(name)
	if kind, ok := aliases[lower]; ok {
		return registry[kind], nil
	}
	return nil, fmt.Errorf("unknown resource type %q; known types: %s", name, knownTypes())
}

// LookupByKind finds a resource type by its exact Kind string.
func LookupByKind(kind string) (ResourceType, error) {
	mu.RLock()
	defer mu.RUnlock()

	if rt, ok := registry[kind]; ok {
		return rt, nil
	}
	return nil, fmt.Errorf("unknown resource kind %q; known types: %s", kind, knownTypes())
}

// All returns all registered resource types.
func All() []ResourceType {
	mu.RLock()
	defer mu.RUnlock()

	result := make([]ResourceType, 0, len(registry))
	for _, rt := range registry {
		result = append(result, rt)
	}
	return result
}

func knownTypes() string {
	names := make([]string, 0, len(registry))
	for _, rt := range registry {
		names = append(names, rt.Plural())
	}
	return strings.Join(names, ", ")
}
