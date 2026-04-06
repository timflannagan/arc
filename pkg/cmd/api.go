package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/agentregistry-dev/ar/pkg/printer"
	"github.com/spf13/cobra"
)

func newAPICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Access non-resource registry API endpoints",
		Long: `Access registry API endpoints that do not map cleanly to declarative
resources.

Use this for endpoints like ping, version, and token-provider JWKS, or for
inspecting newer enterprise-only endpoints without adding first-class resource
types to arctl yet.`,
	}

	cmd.AddCommand(
		newAPIGetCmd(),
		newAPIPingCmd(),
		newAPIVersionCmd(),
		newAPIJWKSCommand(),
	)

	return cmd
}

func newAPIGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get PATH",
		Short: "GET an arbitrary registry API path",
		Long: `Fetch a registry API endpoint and print the decoded JSON response.

PATH may be relative to /v0, like "version" or "/version", or it may be a
full /v0 path like "/v0/providers".`,
		Example: `  # Get registry version info
  arc api get version

  # Get the token-provider JWKS
  arc api get /token-provider/jwks.json

  # Get an enterprise endpoint directly
  arc api get /v0/providers`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := apiClient.GetAny(args[0])
			if err != nil {
				return err
			}
			return printAPIResponse(cmd.OutOrStdout(), data)
		},
	}
}

func newAPIPingCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ping",
		Short: "Check registry connectivity",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := apiClient.Ping(); err != nil {
				return err
			}

			status := map[string]any{"status": "ok"}
			if outputFormat == printer.FormatTable {
				printer.PrintTable(cmd.OutOrStdout(), []string{"status"}, [][]string{{"ok"}})
				return nil
			}
			return printAPIResponse(cmd.OutOrStdout(), status)
		},
	}
}

func newAPIVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Fetch registry version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := apiClient.GetAny("/version")
			if err != nil {
				return err
			}

			if outputFormat == printer.FormatTable {
				versionInfo, err := asObject(data)
				if err != nil {
					return err
				}
				printer.PrintTable(cmd.OutOrStdout(),
					[]string{"Version", "Git Commit", "Build Time"},
					[][]string{{
						stringValue(versionInfo["version"]),
						stringValue(versionInfo["git_commit"]),
						stringValue(versionInfo["build_time"]),
					}},
				)
				return nil
			}

			return printAPIResponse(cmd.OutOrStdout(), data)
		},
	}
}

func newAPIJWKSCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "jwks",
		Short: "Fetch the token-provider JWKS document",
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := apiClient.GetAny("/token-provider/jwks.json")
			if err != nil {
				return err
			}
			return printAPIResponse(cmd.OutOrStdout(), data)
		},
	}
}

func printAPIResponse(w io.Writer, data any) error {
	switch outputFormat {
	case printer.FormatYAML:
		return printer.PrintYAML(w, data)
	default:
		return printer.PrintJSON(w, data)
	}
}

func asObject(data any) (map[string]any, error) {
	obj, ok := data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected JSON object response, got %T", data)
	}
	return obj, nil
}

func stringValue(v any) string {
	if v == nil {
		return ""
	}
	text := fmt.Sprintf("%v", v)
	return strings.TrimSpace(text)
}
