package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/agentregistry-dev/ar/pkg/resource"
	"github.com/agentregistry-dev/ar/pkg/scheme"
	"github.com/spf13/cobra"
)

func newBuildCmd() *cobra.Command {
	var (
		image    string
		push     bool
		platform string
	)

	cmd := &cobra.Command{
		Use:   "build DIRECTORY",
		Short: "Build a container image for a resource project",
		Long: `Build a Docker container image from a scaffolded project directory.

Looks for a resource YAML file (agent.yaml, mcpserver.yaml, etc.) and a
Dockerfile in the given directory. Builds the image and optionally pushes it.

This is a local-only operation — it does not contact the registry.`,
		Example: `  # Build from a project directory
  ar build ./my-agent

  # Build with a custom image tag
  ar build ./my-agent --image ghcr.io/org/my-agent:v1.0

  # Build and push
  ar build ./my-agent --image ghcr.io/org/my-agent:v1.0 --push

  # Build for a specific platform
  ar build ./my-agent --platform linux/amd64`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := args[0]

			// Verify directory exists.
			info, err := os.Stat(dir)
			if err != nil || !info.IsDir() {
				return fmt.Errorf("%s is not a directory", dir)
			}

			// Find the resource YAML.
			yamlFile, res, err := findResourceYAML(dir)
			if err != nil {
				return err
			}

			// Verify Dockerfile exists.
			dockerfile := filepath.Join(dir, "Dockerfile")
			if _, err := os.Stat(dockerfile); err != nil {
				return fmt.Errorf("no Dockerfile found in %s", dir)
			}

			// Determine image name.
			buildImage := image
			if buildImage == "" {
				buildImage = extractImage(res)
			}
			if buildImage == "" {
				buildImage = fmt.Sprintf("localhost:5001/%s:latest", res.Metadata.Name)
			}

			fmt.Printf("Building %s/%s from %s\n", res.Kind, res.Metadata.Name, yamlFile)
			fmt.Printf("Image: %s\n", buildImage)

			// Build with docker.
			buildArgs := []string{"build", "-t", buildImage}
			if platform != "" {
				buildArgs = append(buildArgs, "--platform", platform)
			}
			buildArgs = append(buildArgs, dir)

			buildCmd := exec.Command("docker", buildArgs...)
			buildCmd.Stdout = os.Stdout
			buildCmd.Stderr = os.Stderr
			if err := buildCmd.Run(); err != nil {
				return fmt.Errorf("docker build failed: %w", err)
			}

			fmt.Printf("\nBuilt %s\n", buildImage)

			// Optionally push.
			if push {
				fmt.Printf("Pushing %s...\n", buildImage)
				pushCmd := exec.Command("docker", "push", buildImage)
				pushCmd.Stdout = os.Stdout
				pushCmd.Stderr = os.Stderr
				if err := pushCmd.Run(); err != nil {
					return fmt.Errorf("docker push failed: %w", err)
				}
				fmt.Printf("Pushed %s\n", buildImage)
			}

			fmt.Println("\nNext steps:")
			fmt.Printf("  ar apply -f %s\n", yamlFile)

			return nil
		},
	}

	cmd.Flags().StringVar(&image, "image", "", "Override the image name/tag")
	cmd.Flags().BoolVar(&push, "push", false, "Push the image after building")
	cmd.Flags().StringVar(&platform, "platform", "", "Target platform (e.g. linux/amd64)")

	return cmd
}

// findResourceYAML looks for a known resource YAML file in the directory.
func findResourceYAML(dir string) (string, *resource.Resource, error) {
	candidates := []string{
		"agent.yaml",
		"mcpserver.yaml",
		"skill.yaml",
		"prompt.yaml",
	}

	for _, name := range candidates {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err != nil {
			continue
		}

		resources, err := scheme.DecodeFile(path)
		if err != nil {
			return "", nil, fmt.Errorf("parsing %s: %w", path, err)
		}
		if len(resources) > 0 {
			return path, resources[0], nil
		}
	}

	return "", nil, fmt.Errorf("no resource YAML found in %s (looked for: %s)",
		dir, strings.Join(candidates, ", "))
}

// extractImage tries to get the image field from the resource spec.
func extractImage(r *resource.Resource) string {
	if img, ok := r.Spec["image"]; ok {
		if s, ok := img.(string); ok {
			return s
		}
	}
	// For MCPServer, try packages[0].identifier.
	if pkgs, ok := r.Spec["packages"]; ok {
		if slice, ok := pkgs.([]any); ok && len(slice) > 0 {
			if pkg, ok := slice[0].(map[string]any); ok {
				if id, ok := pkg["identifier"]; ok {
					if s, ok := id.(string); ok {
						return s
					}
				}
			}
		}
	}
	return ""
}
