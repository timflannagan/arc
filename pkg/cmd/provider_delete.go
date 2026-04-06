package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newProviderDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a provider connection",
		Example: `  # Delete with confirmation prompt
  arc provider delete my-provider

  # Delete without confirmation
  arc provider delete my-provider --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Detect which platform this provider belongs to.
			provider, err := apiClient.GetProvider(name)
			if err != nil {
				return fmt.Errorf("provider connection '%s' not found: %w", name, err)
			}

			platform := providerPlatform(provider)
			displayName := strVal(provider, "name")
			if displayName == "" {
				displayName = name
			}

			platformLabel := platform
			switch platform {
			case "gcp":
				platformLabel = "Vertex"
			case "aws":
				platformLabel = "AgentCore"
			}

			if !force {
				fmt.Fprintf(cmd.OutOrStderr(), "Are you sure you want to delete %s provider connection '%s'? [y/N]: ", platformLabel, displayName)
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(answer)
				if answer != "y" && answer != "Y" && answer != "yes" && answer != "Yes" {
					fmt.Fprintln(cmd.OutOrStdout(), "Cancelled.")
					return nil
				}
			}

			if err := apiClient.DeleteProvider(name); err != nil {
				return fmt.Errorf("deleting provider: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Successfully deleted %s provider connection '%s'\n", platformLabel, displayName)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")

	return cmd
}
