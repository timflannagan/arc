package cmd

import (
	"fmt"
	"io"
	"sort"

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

  # List all first-class resource types
  arc get all

  # Get a specific agent
  arc get agent my-summarizer

  # Get with YAML output
  arc get agent my-summarizer -o yaml

  # List MCP servers as JSON
  arc get mcpservers -o json`,
		Args: cobra.RangeArgs(1, 2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				types := []string{"all"}
				for _, rt := range resource.All() {
					types = append(types, rt.Plural(), rt.Singular())
				}
				return types, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			typeName := args[0]
			if typeName == "all" {
				if len(args) != 1 {
					return fmt.Errorf("get all does not accept a NAME")
				}
				return getAll(cmd.OutOrStdout())
			}

			rt, err := resource.Lookup(typeName)
			if err != nil {
				return err
			}

			if len(args) == 2 {
				return getOne(cmd.OutOrStdout(), rt, args[1])
			}
			return getList(cmd.OutOrStdout(), rt)
		},
	}

	return cmd
}

func getOne(w io.Writer, rt resource.ResourceType, name string) error {
	response, err := apiClient.Get(rt.APIPath(), name)
	if err != nil {
		return err
	}

	switch outputFormat {
	case printer.FormatYAML:
		return printer.PrintYAML(w, response)
	case printer.FormatJSON:
		return printer.PrintJSON(w, response)
	default:
		item := rt.ExtractItem(response)
		columns := rt.TableColumns()
		row := rt.TableRow(response)
		printer.PrintTable(w, columns, [][]string{row})
		if desc := item["description"]; desc != nil && desc != "" {
			fmt.Fprintf(w, "\nDescription: %v\n", desc)
		}
		return nil
	}
}

func getList(w io.Writer, rt resource.ResourceType) error {
	response, err := apiClient.List(rt.APIPath())
	if err != nil {
		return err
	}

	switch outputFormat {
	case printer.FormatYAML:
		return printer.PrintYAML(w, response)
	case printer.FormatJSON:
		return printer.PrintJSON(w, response)
	default:
		items := rt.ExtractList(response)
		if len(items) == 0 {
			fmt.Fprintf(w, "No %s found.\n", rt.Plural())
			return nil
		}

		columns := rt.TableColumns()
		rows := make([][]string, 0, len(items))
		for _, item := range items {
			rows = append(rows, rt.TableRow(item))
		}
		printer.PrintTable(w, columns, rows)
		return nil
	}
}

func getAll(w io.Writer) error {
	resourceTypes := resource.All()
	sort.Slice(resourceTypes, func(i, j int) bool {
		return resourceTypes[i].Plural() < resourceTypes[j].Plural()
	})

	grouped := make(map[string]any, len(resourceTypes))
	rows := make([][]string, 0)

	for _, rt := range resourceTypes {
		response, err := apiClient.List(rt.APIPath())
		if err != nil {
			return fmt.Errorf("listing %s: %w", rt.Plural(), err)
		}

		grouped[rt.Plural()] = response

		for _, item := range rt.ExtractList(response) {
			summary := summarizeResource(rt, item)
			rows = append(rows, []string{
				rt.Kind(),
				summary["name"],
				summary["version"],
				summary["status"],
			})
		}
	}

	switch outputFormat {
	case printer.FormatYAML:
		return printer.PrintYAML(w, grouped)
	case printer.FormatJSON:
		return printer.PrintJSON(w, grouped)
	default:
		if len(rows) == 0 {
			fmt.Fprintln(w, "No resources found.")
			return nil
		}
		printer.PrintTable(w, []string{"Kind", "Name", "Version", "Status"}, rows)
		return nil
	}
}

func summarizeResource(rt resource.ResourceType, response map[string]any) map[string]string {
	item := rt.ExtractItem(response)
	return map[string]string{
		"name":    valueOrEmpty(item["name"]),
		"version": valueOrEmpty(item["version"]),
		"status":  valueOrEmpty(item["status"]),
	}
}

func valueOrEmpty(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
