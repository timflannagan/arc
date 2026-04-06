package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func newProviderSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Generate platform bootstrap configuration",
		Long: `Setup subcommands are bootstrap helpers for preparing cloud infrastructure
before creating a provider connection.`,
	}

	cmd.AddCommand(
		newProviderSetupAWSCmd(),
		newProviderSetupGCPCmd(),
	)

	return cmd
}

func newProviderSetupAWSCmd() *cobra.Command {
	var (
		roleName     string
		awsAccountID string
	)

	cmd := &cobra.Command{
		Use:   "aws",
		Short: "Generate AWS CloudFormation template for AgentCore setup",
		Long: `Generate a CloudFormation template and External ID for setting up
AWS IAM roles required by AgentCore.

The CloudFormation YAML is printed to stdout so it can be redirected to a file.`,
		Example: `  # Generate CloudFormation template
  arc provider setup aws --aws-account-id 123456789012 > cf-template.yaml

  # With a custom role name
  arc provider setup aws --aws-account-id 123456789012 --role-name MyRole`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if roleName == "" {
				suffix, err := generateRandomSuffix(6)
				if err != nil {
					return fmt.Errorf("generating role name suffix: %w", err)
				}
				roleName = "AgentRegistryAccessRole-" + suffix
			}

			body := map[string]any{
				"awsAccountId": awsAccountID,
				"roleName":     roleName,
			}

			resp, err := apiClient.PostSetup("/v0/platforms/aws/setup", body)
			if err != nil {
				return fmt.Errorf("getting AWS setup: %w", err)
			}

			cfYAML := strVal(resp, "cloudFormationYaml")
			externalID := strVal(resp, "externalId")
			expectedRole := strVal(resp, "expectedRoleName")

			// Print CloudFormation to stdout for redirection.
			fmt.Fprint(cmd.OutOrStdout(), cfYAML)

			// Print metadata to stderr so it doesn't mix with YAML.
			fmt.Fprintf(cmd.ErrOrStderr(), "# External ID: %s\n", externalID)
			fmt.Fprintf(cmd.ErrOrStderr(), "# Role Name: %s\n", expectedRole)

			return nil
		},
	}

	cmd.Flags().StringVar(&awsAccountID, "aws-account-id", "", "AWS account ID for AgentCore onboarding (required)")
	cmd.Flags().StringVar(&roleName, "role-name", "", "Name for the IAM role (auto-generated if not provided)")
	cmd.MarkFlagRequired("aws-account-id")

	return cmd
}

func newProviderSetupGCPCmd() *cobra.Command {
	var (
		projectID      string
		serviceAccount string
		keyFile        string
		dryRun         bool
	)

	cmd := &cobra.Command{
		Use:   "gcp",
		Short: "Bootstrap GCP service account for Vertex deployment",
		Long: `Create a GCP service account with the IAM roles required for Vertex
deployment. Requires the gcloud CLI to be installed and authenticated.`,
		Example: `  # Set up a GCP service account
  arc provider setup gcp --project-id my-project

  # Custom service account name and key file
  arc provider setup gcp --project-id my-project \
    --service-account my-deployer --key-file my-key.json

  # Dry run to see commands without executing
  arc provider setup gcp --project-id my-project --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()

			// Check prerequisites.
			if err := checkGcloudAvailable(); err != nil {
				return err
			}
			if err := checkGcloudAuth(); err != nil {
				return err
			}
			if err := checkGcloudProject(projectID); err != nil {
				return err
			}

			saEmail := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", serviceAccount, projectID)

			apis := []string{
				"aiplatform.googleapis.com",
				"serviceusage.googleapis.com",
				"iamcredentials.googleapis.com",
				"cloudbuild.googleapis.com",
				"artifactregistry.googleapis.com",
				"run.googleapis.com",
				"secretmanager.googleapis.com",
			}

			// Enable APIs.
			for _, api := range apis {
				if err := runGcloud(w, dryRun, "services", "enable", api, "--project", projectID); err != nil {
					return err
				}
			}

			// Create service account.
			if err := runGcloud(w, dryRun, "iam", "service-accounts", "create", serviceAccount,
				"--display-name", "Agent Registry Deployer",
				"--project", projectID); err != nil {
				// Ignore "already exists" errors.
				if !strings.Contains(err.Error(), "already exists") {
					return err
				}
			}

			// Standard IAM role bindings.
			roles := []string{
				"roles/aiplatform.admin",
				"roles/storage.objectAdmin",
				"roles/serviceusage.serviceUsageConsumer",
				"roles/artifactregistry.admin",
				"roles/run.admin",
				"roles/iam.serviceAccountUser",
				"roles/cloudbuild.builds.editor",
				"roles/logging.logWriter",
			}

			for _, role := range roles {
				if err := runGcloud(w, dryRun, "projects", "add-iam-policy-binding", projectID,
					"--member", "serviceAccount:"+saEmail,
					"--role", role,
					"--condition", "None"); err != nil {
					return err
				}
			}

			// Custom IAM roles.
			customRoles := []struct {
				id          string
				title       string
				permissions []string
			}{
				{
					id:    "AgentRegistrySecretManager",
					title: "Agent Registry Secret Manager",
					permissions: []string{
						"secretmanager.secrets.create",
						"secretmanager.secrets.delete",
						"secretmanager.versions.add",
						"secretmanager.secrets.getIamPolicy",
						"secretmanager.secrets.setIamPolicy",
					},
				},
				{
					id:    "AgentRegistryIAMManager",
					title: "Agent Registry IAM Manager",
					permissions: []string{
						"resourcemanager.projects.getIamPolicy",
						"resourcemanager.projects.setIamPolicy",
					},
				},
			}

			for _, cr := range customRoles {
				perms := strings.Join(cr.permissions, ",")
				if err := runGcloud(w, dryRun, "iam", "roles", "create", cr.id,
					"--project", projectID,
					"--title", cr.title,
					"--permissions", perms); err != nil {
					// Ignore "already exists" errors.
					if !strings.Contains(err.Error(), "already exists") {
						return err
					}
				}

				customRoleRef := fmt.Sprintf("projects/%s/roles/%s", projectID, cr.id)
				if err := runGcloud(w, dryRun, "projects", "add-iam-policy-binding", projectID,
					"--member", "serviceAccount:"+saEmail,
					"--role", customRoleRef,
					"--condition", "None"); err != nil {
					return err
				}
			}

			// Create service account key.
			if err := runGcloud(w, dryRun, "iam", "service-accounts", "keys", "create", keyFile,
				"--iam-account", saEmail,
				"--project", projectID); err != nil {
				return err
			}

			fmt.Fprintf(w, "\nVertex setup complete!\n\n")
			fmt.Fprintf(w, "Service account key saved to: %s\n\n", keyFile)
			fmt.Fprintf(w, "Next steps:\n\n")
			fmt.Fprintf(w, "  arc provider add gcp my-vertex-connection \\\n")
			fmt.Fprintf(w, "    --project-id %s \\\n", projectID)
			fmt.Fprintf(w, "    --location us-central1 \\\n")
			fmt.Fprintf(w, "    --service-account %s\n", keyFile)

			return nil
		},
	}

	cmd.Flags().StringVar(&projectID, "project-id", "", "Google Cloud project ID (required)")
	cmd.Flags().StringVar(&serviceAccount, "service-account", "agentregistry-deployer", "Name for the service account")
	cmd.Flags().StringVar(&keyFile, "key-file", "sa-gcp-deployer.json", "Output path for service account key")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print commands without executing")
	cmd.MarkFlagRequired("project-id")

	return cmd
}

func checkGcloudAvailable() error {
	if _, err := exec.LookPath("gcloud"); err != nil {
		return fmt.Errorf("gcloud CLI not found; install it from https://cloud.google.com/sdk/docs/install")
	}
	return nil
}

func checkGcloudAuth() error {
	out, err := exec.Command("gcloud", "auth", "print-identity-token").CombinedOutput()
	if err != nil {
		return fmt.Errorf("gcloud not authenticated; run 'gcloud auth login' first: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

func checkGcloudProject(projectID string) error {
	out, err := exec.Command("gcloud", "projects", "describe", projectID).CombinedOutput()
	if err != nil {
		return fmt.Errorf("cannot access project %s; check permissions: %s", projectID, strings.TrimSpace(string(out)))
	}
	return nil
}

func runGcloud(w io.Writer, dryRun bool, args ...string) error {
	cmdArgs := append([]string{}, args...)
	fullCmd := "gcloud " + strings.Join(cmdArgs, " ")

	if dryRun {
		fmt.Fprintf(w, "  %s\n", fullCmd)
		return nil
	}

	fmt.Fprintf(w, "-> %s\n", fullCmd)
	out, err := exec.Command("gcloud", cmdArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", fullCmd, strings.TrimSpace(string(out)))
	}
	return nil
}

func generateRandomSuffix(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}
