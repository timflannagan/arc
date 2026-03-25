// Package config provides kubeconfig-style configuration for arc.
// Config is stored at ~/.arc/config by default and supports multiple
// clusters (registry instances) and contexts.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigDir  = ".arc"
	DefaultConfigFile = "config"
	DefaultServer     = "http://localhost:12121"
)

// Config is the top-level configuration, analogous to kubeconfig.
type Config struct {
	CurrentContext string    `yaml:"current-context"`
	Clusters       []Cluster `yaml:"clusters"`
	Contexts       []Context `yaml:"contexts"`
}

// Cluster represents a registry server endpoint.
type Cluster struct {
	Name   string `yaml:"name"`
	Server string `yaml:"server"`
}

// Context binds a name to a cluster and optional auth credentials.
type Context struct {
	Name    string `yaml:"name"`
	Cluster string `yaml:"cluster"`
	Token   string `yaml:"token,omitempty"`
}

// ResolvedContext is the fully resolved connection info for the active context.
type ResolvedContext struct {
	Name   string
	Server string
	Token  string
}

// DefaultPath returns the default config file path (~/.ar/config).
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", DefaultConfigDir, DefaultConfigFile)
	}
	return filepath.Join(home, DefaultConfigDir, DefaultConfigFile)
}

// Load reads config from the given path, or returns a sensible default
// if the file doesn't exist.
func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultPath()
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return defaultConfig(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}

	return &cfg, nil
}

// Save writes the config to the given path, creating directories as needed.
func Save(path string, cfg *Config) error {
	if path == "" {
		path = DefaultPath()
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return os.WriteFile(path, data, 0o600)
}

// Resolve returns the connection info for the active context.
// Flags (server, token) override config values if set.
func (c *Config) Resolve(serverOverride, tokenOverride string) (*ResolvedContext, error) {
	// If overrides provide everything, skip config lookup.
	if serverOverride != "" {
		return &ResolvedContext{
			Name:   "flags",
			Server: serverOverride,
			Token:  tokenOverride,
		}, nil
	}

	// Find current context.
	if c.CurrentContext == "" && len(c.Contexts) > 0 {
		c.CurrentContext = c.Contexts[0].Name
	}

	var ctx *Context
	for i := range c.Contexts {
		if c.Contexts[i].Name == c.CurrentContext {
			ctx = &c.Contexts[i]
			break
		}
	}
	if ctx == nil {
		// No context configured; fall back to defaults.
		return &ResolvedContext{
			Name:   "default",
			Server: DefaultServer,
			Token:  tokenOverride,
		}, nil
	}

	// Find referenced cluster.
	var cluster *Cluster
	for i := range c.Clusters {
		if c.Clusters[i].Name == ctx.Cluster {
			cluster = &c.Clusters[i]
			break
		}
	}

	server := DefaultServer
	if cluster != nil {
		server = cluster.Server
	}

	token := ctx.Token
	if tokenOverride != "" {
		token = tokenOverride
	}

	return &ResolvedContext{
		Name:   ctx.Name,
		Server: server,
		Token:  token,
	}, nil
}

func defaultConfig() *Config {
	return &Config{
		CurrentContext: "local",
		Clusters: []Cluster{
			{Name: "local", Server: DefaultServer},
		},
		Contexts: []Context{
			{Name: "local", Cluster: "local"},
		},
	}
}
