package cmd

import (
	"fmt"
	"os"

	"github.com/agentregistry-dev/ar/pkg/resource"
	"github.com/agentregistry-dev/ar/pkg/scheme"
	"github.com/spf13/cobra"
)

func newApplyCmd() *cobra.Command {
	var filePaths []string

	cmd := &cobra.Command{
		Use:   "apply -f FILE",
		Short: "Create or update resources from YAML files",
		Long: `Apply a declarative configuration to the registry. Reads one or more
YAML files containing resource definitions and creates or updates them.

Supports multi-document YAML files (--- separated) and multiple -f flags.`,
		Example: `  # Apply a single agent
  ar apply -f agent.yaml

  # Apply multiple resources
  ar apply -f agent.yaml -f mcpserver.yaml

  # Apply a multi-document file defining an agent with its dependencies
  ar apply -f full-stack.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(filePaths) == 0 {
				return fmt.Errorf("at least one -f/--filename is required")
			}

			var allResources []*resource.Resource
			for _, path := range filePaths {
				resources, err := scheme.DecodeFile(path)
				if err != nil {
					return fmt.Errorf("file %s: %w", path, err)
				}
				allResources = append(allResources, resources...)
			}

			if err := scheme.Validate(allResources); err != nil {
				return err
			}

			for _, r := range allResources {
				rt, err := resource.LookupByKind(r.Kind)
				if err != nil {
					return err
				}

				payload, err := rt.ToCreatePayload(r)
				if err != nil {
					return fmt.Errorf("preparing %s/%s: %w", r.Kind, r.Metadata.Name, err)
				}

				_, err = apiClient.Create(rt.APIPath(), payload)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s/%s apply failed: %v\n", rt.Singular(), r.Metadata.Name, err)
					continue
				}

				fmt.Printf("%s/%s applied\n", rt.Singular(), r.Metadata.Name)
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&filePaths, "filename", "f", nil, "YAML file(s) containing resource definitions")
	cmd.MarkFlagRequired("filename")

	return cmd
}
