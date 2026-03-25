package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/agentregistry-dev/ar/pkg/resource"
	"github.com/agentregistry-dev/ar/pkg/scheme"
	"github.com/spf13/cobra"
)

func newPullCmd() *cobra.Command {
	var (
		version   string
		outputDir string
	)

	cmd := &cobra.Command{
		Use:   "pull TYPE NAME",
		Short: "Fetch a resource from the registry as a local YAML file",
		Long: `Pull a resource from the registry and write it as a local YAML file.
This is the inverse of apply — registry to local files.

Useful for inspecting, forking, or templating off existing resources.`,
		Example: `  # Pull an agent to ./my-agent/agent.yaml
  ar pull agent my-agent

  # Pull a specific version
  ar pull agent my-agent --version 1.0.0

  # Pull to a custom directory
  ar pull agent my-agent -o ./staging/

  # Pull an MCP server
  ar pull mcpserver acme/fetch`,
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

			// Fetch the resource from the registry.
			var response map[string]any
			if version != "" {
				response, err = apiClient.GetVersion(rt.APIPath(), name, version)
			} else {
				response, err = apiClient.Get(rt.APIPath(), name)
			}
			if err != nil {
				return fmt.Errorf("fetching %s/%s: %w", rt.Singular(), name, err)
			}

			// Convert API response to a Resource.
			res := rt.ToResource(response)

			// Encode to YAML.
			data, err := scheme.Encode(res)
			if err != nil {
				return fmt.Errorf("encoding %s/%s: %w", rt.Singular(), name, err)
			}

			// Determine output path.
			dir := outputDir
			if dir == "" {
				// Use the resource name, replacing slashes (e.g. acme/fetch → acme-fetch).
				safeName := filepath.Base(name)
				dir = safeName
			}

			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("creating directory %s: %w", dir, err)
			}

			filename := rt.Singular() + ".yaml"
			outPath := filepath.Join(dir, filename)

			if err := os.WriteFile(outPath, data, 0o644); err != nil {
				return fmt.Errorf("writing %s: %w", outPath, err)
			}

			fmt.Printf("%s/%s pulled to %s\n", rt.Singular(), name, outPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "Specific version to pull (default: latest)")
	cmd.Flags().StringVarP(&outputDir, "output-dir", "d", "", "Output directory (default: ./<name>)")

	return cmd
}
