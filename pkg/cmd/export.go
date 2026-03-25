package cmd

import (
	"fmt"
	"os"

	"github.com/agentregistry-dev/ar/pkg/resource"
	"github.com/agentregistry-dev/ar/pkg/scheme"
	"github.com/spf13/cobra"
)

func newExportCmd() *cobra.Command {
	var outputFile string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export all resources from the registry to a YAML file",
		Long: `Export all resources (agents, MCP servers, skills, prompts) from the
registry into a single multi-document YAML file. This file can be used
with 'arc import' to restore or seed another registry instance.`,
		Example: `  # Export to stdout
  arc export

  # Export to a file
  arc export -f catalog.yaml

  # Export specific types
  arc export agents
  arc export mcpservers skills`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine which resource types to export.
			types, err := resolveTypes(args)
			if err != nil {
				return err
			}

			var allResources []*resource.Resource

			for _, rt := range types {
				response, err := apiClient.List(rt.APIPath())
				if err != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to list %s: %v\n", rt.Plural(), err)
					continue
				}

				items := rt.ExtractList(response)
				for _, item := range items {
					// Wrap in the response envelope the ToResource expects.
					wrapped := map[string]any{
						rt.Singular(): item,
					}
					res := rt.ToResource(wrapped)
					allResources = append(allResources, res)
				}
			}

			if len(allResources) == 0 {
				fmt.Fprintln(os.Stderr, "No resources found.")
				return nil
			}

			data, err := scheme.Encode(allResources...)
			if err != nil {
				return fmt.Errorf("encoding resources: %w", err)
			}

			if outputFile != "" {
				if err := os.WriteFile(outputFile, data, 0o644); err != nil {
					return fmt.Errorf("writing %s: %w", outputFile, err)
				}
				fmt.Printf("Exported %d resources to %s\n", len(allResources), outputFile)
			} else {
				os.Stdout.Write(data)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFile, "filename", "f", "", "Output file (default: stdout)")

	return cmd
}

// resolveTypes maps CLI args to resource types. If no args, returns all types.
func resolveTypes(args []string) ([]resource.ResourceType, error) {
	if len(args) == 0 {
		return resource.All(), nil
	}

	var types []resource.ResourceType
	for _, name := range args {
		rt, err := resource.Lookup(name)
		if err != nil {
			return nil, err
		}
		types = append(types, rt)
	}
	return types, nil
}
