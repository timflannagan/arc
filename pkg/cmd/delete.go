package cmd

import (
	"fmt"

	"github.com/agentregistry-dev/ar/pkg/resource"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	var version string

	cmd := &cobra.Command{
		Use:   "delete TYPE NAME",
		Short: "Delete a resource from the registry",
		Long: `Delete a resource by type and name. A version must be specified
since the registry tracks versioned resources.`,
		Example: `  # Delete a specific agent version
  arc delete agent my-summarizer --version 1.0.0

  # Delete an MCP server version
  arc delete mcpserver my-server --version 2.1.0`,
		Args: cobra.ExactArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				var types []string
				for _, rt := range resource.All() {
					types = append(types, rt.Singular())
				}
				return types, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			typeName := args[0]
			name := args[1]

			rt, err := resource.Lookup(typeName)
			if err != nil {
				return err
			}

			if version == "" {
				return fmt.Errorf("--version is required for delete")
			}

			if err := apiClient.Delete(rt.APIPath(), name, version); err != nil {
				return err
			}

			fmt.Printf("%s/%s (version %s) deleted\n", rt.Singular(), name, version)
			return nil
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "Version to delete (required)")
	cmd.MarkFlagRequired("version")

	return cmd
}
