package cmd

import (
	"fmt"
	"os"

	"github.com/agentregistry-dev/ar/pkg/resource"
	"github.com/agentregistry-dev/ar/pkg/scheme"
	"github.com/spf13/cobra"
)

func newImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import -f FILE",
		Short: "Import resources into the registry from a YAML file",
		Long: `Import resources from a multi-document YAML file into the registry.
This is the bulk counterpart of 'arc apply' and the inverse of 'arc export'.

Resources are created in document order, so list dependencies before
the resources that reference them.`,
		Example: `  # Import from an export file
  arc import -f catalog.yaml

  # Import from a full-stack definition
  arc import -f full-stack.yaml`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			filePaths, err := cmd.Flags().GetStringArray("filename")
			if err != nil || len(filePaths) == 0 {
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

			var succeeded, failed int
			for _, r := range allResources {
				rt, err := resource.LookupByKind(r.Kind)
				if err != nil {
					fmt.Fprintf(os.Stderr, "skip %s/%s: %v\n", r.Kind, r.Metadata.Name, err)
					failed++
					continue
				}

				payload, err := rt.ToCreatePayload(r)
				if err != nil {
					fmt.Fprintf(os.Stderr, "skip %s/%s: %v\n", rt.Singular(), r.Metadata.Name, err)
					failed++
					continue
				}

				_, err = apiClient.Create(rt.APIPath(), payload)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s/%s failed: %v\n", rt.Singular(), r.Metadata.Name, err)
					failed++
					continue
				}

				fmt.Printf("%s/%s imported\n", rt.Singular(), r.Metadata.Name)
				succeeded++
			}

			fmt.Printf("\nImported %d resources", succeeded)
			if failed > 0 {
				fmt.Printf(" (%d failed)", failed)
			}
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().StringArrayP("filename", "f", nil, "YAML file(s) to import")
	cmd.MarkFlagRequired("filename")

	return cmd
}
