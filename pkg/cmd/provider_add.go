package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newProviderAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new provider connection",
		Long:  `Add a new cloud provider connection for AgentCore (AWS) or Vertex (GCP).`,
	}

	cmd.AddCommand(
		newProviderAddAWSCmd(),
		newProviderAddGCPCmd(),
	)

	return cmd
}

func newProviderAddAWSCmd() *cobra.Command {
	var (
		roleARN    string
		externalID string
		region     string
	)

	cmd := &cobra.Command{
		Use:   "aws NAME",
		Short: "Add an AgentCore (AWS) provider connection",
		Example: `  # Add an AWS provider
  arc provider add aws my-aws-provider \
    --role-arn arn:aws:iam::123456789012:role/AgentRegistryAccessRole \
    --external-id abc123def456`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			id := toProviderID(name)

			payload := map[string]any{
				"id":       id,
				"name":     name,
				"platform": "aws",
				"config": map[string]any{
					"region":     region,
					"roleArn":    roleARN,
					"externalId": externalID,
				},
			}

			resp, err := apiClient.CreateProvider(payload)
			if err != nil {
				return fmt.Errorf("creating AWS provider: %w", err)
			}

			respID := strVal(resp, "id")
			if respID == "" {
				respID = id
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Successfully created AgentCore provider connection\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  ID:       %s\n", respID)
			fmt.Fprintf(cmd.OutOrStdout(), "  Name:     %s\n", name)
			fmt.Fprintf(cmd.OutOrStdout(), "  Region:   %s\n", region)
			fmt.Fprintf(cmd.OutOrStdout(), "  Role ARN: (configured)\n")
			return nil
		},
	}

	cmd.Flags().StringVar(&roleARN, "role-arn", "", "AWS IAM Role ARN to assume (required)")
	cmd.Flags().StringVar(&externalID, "external-id", "", "External ID for role assumption (required)")
	cmd.Flags().StringVar(&region, "region", "us-east-1", "AWS region")
	cmd.MarkFlagRequired("role-arn")
	cmd.MarkFlagRequired("external-id")

	return cmd
}

func newProviderAddGCPCmd() *cobra.Command {
	var (
		projectID      string
		location       string
		serviceAccount string
	)

	cmd := &cobra.Command{
		Use:   "gcp NAME",
		Short: "Add a Vertex (GCP) provider connection",
		Example: `  # Add a GCP provider with ADC
  arc provider add gcp my-gcp-provider --project-id my-project

  # Add a GCP provider with a service account key
  arc provider add gcp my-gcp-provider \
    --project-id my-project \
    --service-account sa-key.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			id := toProviderID(name)

			config := map[string]any{
				"projectId": projectID,
				"location":  location,
			}

			if serviceAccount != "" {
				saKey, err := os.ReadFile(serviceAccount)
				if err != nil {
					return fmt.Errorf("reading service account file: %w", err)
				}
				config["serviceAccountKey"] = string(saKey)
			}

			payload := map[string]any{
				"id":       id,
				"name":     name,
				"platform": "gcp",
				"config":   config,
			}

			resp, err := apiClient.CreateProvider(payload)
			if err != nil {
				return fmt.Errorf("creating GCP provider: %w", err)
			}

			respID := strVal(resp, "id")
			if respID == "" {
				respID = id
			}

			authMethod := "Application Default Credentials (ADC)"
			if serviceAccount != "" {
				authMethod = "Service Account Key"
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Successfully created Vertex provider connection\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  ID:         %s\n", respID)
			fmt.Fprintf(cmd.OutOrStdout(), "  Name:       %s\n", name)
			fmt.Fprintf(cmd.OutOrStdout(), "  Project ID: %s\n", projectID)
			fmt.Fprintf(cmd.OutOrStdout(), "  Location:   %s\n", location)
			fmt.Fprintf(cmd.OutOrStdout(), "  Auth:       %s\n", authMethod)
			return nil
		},
	}

	cmd.Flags().StringVar(&projectID, "project-id", "", "Google Cloud project ID (required)")
	cmd.Flags().StringVar(&location, "location", "us-central1", "Vertex region/location")
	cmd.Flags().StringVar(&serviceAccount, "service-account", "", "Path to service account key JSON file")
	cmd.MarkFlagRequired("project-id")

	return cmd
}

func toProviderID(name string) string {
	return strings.ReplaceAll(strings.ToLower(name), " ", "-")
}
