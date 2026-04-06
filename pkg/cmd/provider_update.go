package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newProviderUpdateCmd() *cobra.Command {
	var (
		newName        string
		projectID      string
		location       string
		serviceAccount string
		roleARN        string
		externalID     string
		region         string
	)

	cmd := &cobra.Command{
		Use:   "update NAME",
		Short: "Update a provider connection",
		Long:  `Update an existing cloud provider connection. At least one update flag must be provided.`,
		Example: `  # Update a GCP provider's location
  arc provider update my-gcp --location us-east1

  # Update an AWS provider's role ARN
  arc provider update my-aws --role-arn arn:aws:iam::123:role/NewRole

  # Rename a provider
  arc provider update my-provider --name "New Display Name"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			provider, err := apiClient.GetProvider(id)
			if err != nil {
				return fmt.Errorf("provider connection '%s' not found: %w", id, err)
			}

			platform := providerPlatform(provider)
			switch platform {
			case "gcp":
				return updateGCPProvider(cmd, id, provider, newName, projectID, location, serviceAccount)
			case "aws":
				return updateAWSProvider(cmd, id, provider, newName, region, roleARN, externalID)
			default:
				return fmt.Errorf("unknown provider platform: %s", platform)
			}
		},
	}

	cmd.Flags().StringVar(&newName, "name", "", "Update display name")
	cmd.Flags().StringVar(&projectID, "project-id", "", "Update Google Cloud project ID (GCP)")
	cmd.Flags().StringVar(&location, "location", "", "Update location/region (GCP)")
	cmd.Flags().StringVar(&serviceAccount, "service-account", "", "Path to new service account key JSON file (GCP)")
	cmd.Flags().StringVar(&roleARN, "role-arn", "", "Update AWS IAM Role ARN (AWS)")
	cmd.Flags().StringVar(&externalID, "external-id", "", "Update external ID (AWS)")
	cmd.Flags().StringVar(&region, "region", "", "Update AWS region (AWS)")

	return cmd
}

func updateGCPProvider(cmd *cobra.Command, id string, existing map[string]any, newName, projectID, location, serviceAccount string) error {
	if newName == "" && projectID == "" && location == "" && serviceAccount == "" {
		return fmt.Errorf("no updates specified; use flags like --name, --project-id, --location, or --service-account")
	}

	existingConfig, _ := existing["config"].(map[string]any)
	if existingConfig == nil {
		existingConfig = map[string]any{}
	}

	config := map[string]any{}
	if projectID != "" {
		config["projectId"] = projectID
	} else if v := strVal(existingConfig, "projectId"); v != "" {
		config["projectId"] = v
	}
	if location != "" {
		config["location"] = location
	} else if v := strVal(existingConfig, "location"); v != "" {
		config["location"] = v
	}
	if serviceAccount != "" {
		saKey, err := readFileContent(serviceAccount)
		if err != nil {
			return err
		}
		config["serviceAccountKey"] = saKey
	}

	payload := map[string]any{
		"config": config,
	}
	if newName != "" {
		payload["name"] = newName
	}

	resp, err := apiClient.UpdateProvider(id, payload)
	if err != nil {
		return fmt.Errorf("updating GCP provider: %w", err)
	}

	displayName := strVal(resp, "name")
	if displayName == "" {
		displayName = strVal(existing, "name")
	}

	respConfig, _ := resp["config"].(map[string]any)
	if respConfig == nil {
		respConfig = config
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Successfully updated Vertex provider connection\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  ID:         %s\n", id)
	fmt.Fprintf(cmd.OutOrStdout(), "  Name:       %s\n", displayName)
	fmt.Fprintf(cmd.OutOrStdout(), "  Project ID: %s\n", strVal(respConfig, "projectId"))
	fmt.Fprintf(cmd.OutOrStdout(), "  Location:   %s\n", strVal(respConfig, "location"))
	return nil
}

func updateAWSProvider(cmd *cobra.Command, id string, existing map[string]any, newName, region, roleARN, externalID string) error {
	if newName == "" && region == "" && roleARN == "" && externalID == "" {
		return fmt.Errorf("no updates specified; use flags like --name, --region, --role-arn, or --external-id")
	}

	existingConfig, _ := existing["config"].(map[string]any)
	if existingConfig == nil {
		existingConfig = map[string]any{}
	}

	config := map[string]any{}
	if region != "" {
		config["region"] = region
	} else if v := strVal(existingConfig, "region"); v != "" {
		config["region"] = v
	}
	if roleARN != "" {
		config["roleArn"] = roleARN
	} else if v := strVal(existingConfig, "roleArn"); v != "" {
		config["roleArn"] = v
	}
	if externalID != "" {
		config["externalId"] = externalID
	} else if v := strVal(existingConfig, "externalId"); v != "" {
		config["externalId"] = v
	}

	payload := map[string]any{
		"config": config,
	}
	if newName != "" {
		payload["name"] = newName
	}

	resp, err := apiClient.UpdateProvider(id, payload)
	if err != nil {
		return fmt.Errorf("updating AWS provider: %w", err)
	}

	displayName := strVal(resp, "name")
	if displayName == "" {
		displayName = strVal(existing, "name")
	}

	respConfig, _ := resp["config"].(map[string]any)
	if respConfig == nil {
		respConfig = config
	}

	var updated []string
	if newName != "" {
		updated = append(updated, "name")
	}
	if region != "" {
		updated = append(updated, "region")
	}
	if roleARN != "" {
		updated = append(updated, "role-arn")
	}
	if externalID != "" {
		updated = append(updated, "external-id")
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Successfully updated AgentCore provider connection\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  ID:      %s\n", id)
	fmt.Fprintf(cmd.OutOrStdout(), "  Name:    %s\n", displayName)
	fmt.Fprintf(cmd.OutOrStdout(), "  Region:  %s\n", strVal(respConfig, "region"))
	if len(updated) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "  Updated: %s\n", joinWords(updated))
	}
	return nil
}

func readFileContent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading file %s: %w", path, err)
	}
	return string(data), nil
}

func joinWords(words []string) string {
	if len(words) == 0 {
		return ""
	}
	result := words[0]
	for i := 1; i < len(words); i++ {
		result += ", " + words[i]
	}
	return result
}
