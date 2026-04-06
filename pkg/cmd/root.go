// Package cmd defines the cobra command tree for arc.
package cmd

import (
	"fmt"
	"os"

	"github.com/agentregistry-dev/ar/pkg/client"
	"github.com/agentregistry-dev/ar/pkg/config"
	"github.com/agentregistry-dev/ar/pkg/printer"
	"github.com/spf13/cobra"

	// Register all resource types.
	_ "github.com/agentregistry-dev/ar/pkg/resource"
)

// Version is set at build time via ldflags.
var Version = "dev"

// Shared state across commands, populated in PersistentPreRun.
var (
	apiClient    *client.Client
	outputFormat printer.Format
)

// Flag values.
var (
	flagServer string
	flagToken  string
	flagOutput string
	flagConfig string
)

// Root returns the top-level arc command.
func Root() *cobra.Command {
	root := &cobra.Command{
		Use:   "arc",
		Short: "Declarative CLI for the agent registry",
		Long: `arc is a kubectl-style CLI for managing agents, MCP servers, skills,
and prompts in an agent registry.

Define resources in YAML and apply them declaratively, or use get/delete
to inspect and manage resources.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Parse output format.
			outputFormat = printer.ParseFormat(flagOutput)

			// Skip client setup for local-only commands.
			if isLocalOnlyCommand(cmd) {
				return nil
			}

			// Load config and resolve connection info.
			cfg, err := config.Load(flagConfig)
			if err != nil {
				return err
			}

			// Also check env vars as overrides.
			if flagServer == "" {
				if envServer := os.Getenv("ARC_SERVER"); envServer != "" {
					flagServer = envServer
				}
			}
			if flagToken == "" {
				if envToken := os.Getenv("ARC_TOKEN"); envToken != "" {
					flagToken = envToken
				}
			}

			resolved, err := cfg.Resolve(flagServer, flagToken)
			if err != nil {
				return err
			}

			apiClient = client.New(resolved.Server, resolved.Token)
			return nil
		},
	}

	// Global flags.
	root.PersistentFlags().StringVar(&flagServer, "server", "", "Registry server URL (overrides config)")
	root.PersistentFlags().StringVar(&flagToken, "token", "", "Auth token (overrides config)")
	root.PersistentFlags().StringVarP(&flagOutput, "output", "o", "table", "Output format: table, yaml, json")
	root.PersistentFlags().StringVar(&flagConfig, "config", "", "Config file path (default ~/.arc/config)")

	// Register subcommands.
	root.AddCommand(
		newAPICmd(),
		newInitCmd(),
		newBuildCmd(),
		newApplyCmd(),
		newGetCmd(),
		newPullCmd(),
		newDeleteCmd(),
		newExportCmd(),
		newImportCmd(),
		newVersionCmd(),
		newConfigCmd(),
	)

	return root
}

func isLocalOnlyCommand(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}

	parents := cmd.CommandPath()
	switch parents {
	case "arc version", "arc config", "arc config view", "arc config use-context",
		"arc config set-context":
		return true
	}

	if cmd.HasParent() && cmd.Parent().CommandPath() == "arc init" {
		return true
	}

	if cmd.CommandPath() == "arc build" {
		return true
	}

	return false
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the arc version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(Version)
		},
	}
}
