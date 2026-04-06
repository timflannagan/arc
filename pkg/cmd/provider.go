package cmd

import (
	"fmt"
	"time"

	"github.com/agentregistry-dev/ar/pkg/printer"
	"github.com/spf13/cobra"
)

func newProviderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider",
		Short: "Manage cloud provider connections",
		Long: `Manage cloud provider connections for AgentCore (AWS) and Vertex (GCP).

Use the add subcommand to create new provider connections, list to view
existing ones, and setup to bootstrap cloud infrastructure.`,
		Example: `  # List all provider connections
  arc provider list

  # List only GCP providers
  arc provider list --platform gcp

  # Add a new AWS provider
  arc provider add aws my-aws --role-arn arn:aws:iam::123:role/Role --external-id abc123

  # Add a new GCP provider
  arc provider add gcp my-gcp --project-id my-project

  # Delete a provider
  arc provider delete my-aws

  # Generate AWS CloudFormation setup
  arc provider setup aws --aws-account-id 123456789012`,
	}

	cmd.AddCommand(
		newProviderListCmd(),
		newProviderAddCmd(),
		newProviderDeleteCmd(),
		newProviderUpdateCmd(),
		newProviderSetupCmd(),
	)

	return cmd
}

func newProviderListCmd() *cobra.Command {
	var platform string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List provider connections",
		Long:  `List cloud provider connections, optionally filtered by platform (aws or gcp).`,
		Example: `  # List all providers
  arc provider list

  # List only AWS providers
  arc provider list --platform aws

  # List as JSON
  arc provider list -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if platform != "" && platform != "aws" && platform != "gcp" {
				return fmt.Errorf("unsupported platform type: %s (must be 'gcp' or 'aws')", platform)
			}

			response, err := apiClient.ListProviders(platform)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			switch outputFormat {
			case printer.FormatYAML:
				return printer.PrintYAML(w, response)
			case printer.FormatJSON:
				return printer.PrintJSON(w, response)
			default:
				providers := extractProviderList(response)
				if len(providers) == 0 {
					fmt.Fprintln(w, "No provider connections found.")
					return nil
				}

				columns := []string{"Platform", "ID", "Name", "Project/Region", "Location", "Created"}
				rows := make([][]string, 0, len(providers))
				for _, p := range providers {
					rows = append(rows, providerTableRow(p))
				}
				printer.PrintTable(w, columns, rows)
				return nil
			}
		},
	}

	cmd.Flags().StringVar(&platform, "platform", "", "Filter by platform (aws or gcp)")

	return cmd
}

func extractProviderList(response map[string]any) []map[string]any {
	items, ok := response["providers"]
	if !ok {
		return nil
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

func providerTableRow(p map[string]any) []string {
	platform := strVal(p, "platform")
	displayPlatform := platform
	switch platform {
	case "gcp":
		displayPlatform = "Vertex"
	case "aws":
		displayPlatform = "AgentCore"
	}

	projectOrRegion := "-"
	location := "-"

	if cfg, ok := p["config"].(map[string]any); ok {
		switch platform {
		case "gcp":
			if v := strVal(cfg, "projectId"); v != "" {
				projectOrRegion = v
			}
			if v := strVal(cfg, "location"); v != "" {
				location = v
			}
		case "aws":
			if v := strVal(cfg, "region"); v != "" {
				projectOrRegion = v
			}
		}
	}

	created := strVal(p, "createdAt")
	if t, err := time.Parse(time.RFC3339, created); err == nil {
		created = t.Format("2006-01-02 15:04")
	}

	return []string{
		displayPlatform,
		strVal(p, "id"),
		strVal(p, "name"),
		projectOrRegion,
		location,
		created,
	}
}

func strVal(data map[string]any, key string) string {
	if v, ok := data[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func providerPlatform(data map[string]any) string {
	return strVal(data, "platform")
}
