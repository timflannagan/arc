package cmd

import (
	"fmt"
	"os"

	"github.com/agentregistry-dev/ar/pkg/printer"
	"github.com/agentregistry-dev/ar/pkg/resource"
	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get TYPE [NAME]",
		Short: "List or retrieve resources",
		Long: `Get one or many resources from the registry.

If a NAME is provided, retrieves that specific resource.
Otherwise, lists all resources of the given type.`,
		Example: `  # List all agents
  arc get agents

  # Get a specific agent
  arc get agent my-summarizer

  # Get with YAML output
  arc get agent my-summarizer -o yaml

  # List MCP servers as JSON
  arc get mcpservers -o json`,
		Args: cobra.RangeArgs(1, 2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				var types []string
				for _, rt := range resource.All() {
					types = append(types, rt.Plural(), rt.Singular())
				}
				return types, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			typeName := args[0]
			rt, err := resource.Lookup(typeName)
			if err != nil {
				return err
			}

			if len(args) == 2 {
				return getOne(rt, args[1])
			}
			return getList(rt)
		},
	}

	return cmd
}

func getOne(rt resource.ResourceType, name string) error {
	response, err := apiClient.Get(rt.APIPath(), name)
	if err != nil {
		return err
	}

	switch outputFormat {
	case printer.FormatYAML:
		return printer.PrintYAML(os.Stdout, response)
	case printer.FormatJSON:
		return printer.PrintJSON(os.Stdout, response)
	default:
		item := rt.ExtractItem(response)
		columns := rt.TableColumns()
		row := rt.TableRow(response)
		printer.PrintTable(os.Stdout, columns, [][]string{row})
		// Also print extra detail below the table.
		if desc := item["description"]; desc != nil && desc != "" {
			fmt.Printf("\nDescription: %v\n", desc)
		}
		return nil
	}
}

func getList(rt resource.ResourceType) error {
	response, err := apiClient.List(rt.APIPath())
	if err != nil {
		return err
	}

	switch outputFormat {
	case printer.FormatYAML:
		return printer.PrintYAML(os.Stdout, response)
	case printer.FormatJSON:
		return printer.PrintJSON(os.Stdout, response)
	default:
		items := rt.ExtractList(response)
		if len(items) == 0 {
			fmt.Printf("No %s found.\n", rt.Plural())
			return nil
		}

		columns := rt.TableColumns()
		rows := make([][]string, 0, len(items))
		for _, item := range items {
			rows = append(rows, rt.TableRow(item))
		}
		printer.PrintTable(os.Stdout, columns, rows)
		return nil
	}
}
