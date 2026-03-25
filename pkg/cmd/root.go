// Package cmd defines the cobra command tree for ar.
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

// Root returns the top-level ar command.
func Root() *cobra.Command {
	root := &cobra.Command{
		Use:   "ar",
		Short: "Declarative CLI for the agent registry",
		Long: `ar is a kubectl-style CLI for managing agents, MCP servers, skills,
and prompts in an agent registry.

Define resources in YAML and apply them declaratively, or use get/delete
to inspect and manage resources.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Parse output format.
			outputFormat = printer.ParseFormat(flagOutput)

			// Skip client setup for local-only commands.
			for p := cmd; p != nil; p = p.Parent() {
				switch p.Name() {
				case "version", "config", "init", "build":
					return nil
				}
			}

			// Load config and resolve connection info.
			cfg, err := config.Load(flagConfig)
			if err != nil {
				return err
			}

			// Also check env vars as overrides.
			if flagServer == "" {
				if envServer := os.Getenv("AR_SERVER"); envServer != "" {
					flagServer = envServer
				}
			}
			if flagToken == "" {
				if envToken := os.Getenv("AR_TOKEN"); envToken != "" {
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
	root.PersistentFlags().StringVar(&flagConfig, "config", "", "Config file path (default ~/.ar/config)")

	// Register subcommands.
	root.AddCommand(
		newInitCmd(),
		newBuildCmd(),
		newApplyCmd(),
		newGetCmd(),
		newDeleteCmd(),
		newVersionCmd(),
		newConfigCmd(),
	)

	return root
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the ar version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(Version)
		},
	}
}
