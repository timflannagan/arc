package cmd

import (
	"fmt"

	"github.com/agentregistry-dev/ar/pkg/scaffold"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init TYPE NAME",
		Short: "Scaffold a new resource project",
		Long: `Initialize a new project directory with a resource YAML file and
supporting files (Dockerfile, source code, etc.).

This is a local-only operation — it does not contact the registry.`,
		Example: `  # Scaffold a new ADK Python agent
  arc initagent adk python my-summarizer

  # Scaffold with options
  arc initagent adk python my-agent --model-provider openai --model-name gpt-4o

  # Scaffold an MCP server
  arc initmcpserver my-server

  # Scaffold a skill
  arc initskill my-skill --category nlp

  # Scaffold a prompt
  arc initprompt my-system-prompt`,
	}

	cmd.AddCommand(
		newInitAgentCmd(),
		newInitMCPServerCmd(),
		newInitSkillCmd(),
		newInitPromptCmd(),
	)

	return cmd
}

func newInitAgentCmd() *cobra.Command {
	var opts scaffold.AgentOptions

	cmd := &cobra.Command{
		Use:   "agent FRAMEWORK LANGUAGE NAME",
		Short: "Scaffold a new agent project",
		Long: `Scaffold a new agent project with the given framework, language, and name.

Supported frameworks: adk
Supported languages: python`,
		Example: `  # ADK Python agent
  arc initagent adk python my-agent

  # With model options
  arc initagent adk python my-agent --model-provider openai --model-name gpt-4o`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Framework = args[0]
			opts.Language = args[1]
			opts.Name = args[2]
			if err := scaffold.Agent(opts); err != nil {
				return err
			}
			fmt.Printf("Agent project scaffolded in ./%s/\n", opts.Dir())
			fmt.Println("\nNext steps:")
			fmt.Printf("  1. cd %s && edit agent.yaml\n", opts.Dir())
			fmt.Printf("  2. arc build ./%s\n", opts.Dir())
			fmt.Printf("  3. arc apply -f %s/agent.yaml\n", opts.Dir())
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.Version, "version", "0.1.0", "Initial version")
	cmd.Flags().StringVar(&opts.Description, "description", "", "Agent description")
	cmd.Flags().StringVar(&opts.ModelProvider, "model-provider", "Gemini", "Model provider")
	cmd.Flags().StringVar(&opts.ModelName, "model-name", "gemini-2.0-flash", "Model name")
	cmd.Flags().StringVar(&opts.Image, "image", "", "Docker image (default: localhost:5001/<name>:latest)")
	cmd.Flags().StringVar(&opts.OutputDir, "output-dir", "", "Output directory (default: ./<name>)")

	return cmd
}

func newInitMCPServerCmd() *cobra.Command {
	var opts scaffold.MCPServerOptions

	cmd := &cobra.Command{
		Use:   "mcpserver NAME",
		Short: "Scaffold a new MCP server project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = args[0]
			if err := scaffold.MCPServer(opts); err != nil {
				return err
			}
			fmt.Printf("MCP server project scaffolded in ./%s/\n", opts.Dir())
			fmt.Println("\nNext steps:")
			fmt.Printf("  1. cd %s && edit server.py\n", opts.Dir())
			fmt.Printf("  2. arc build ./%s\n", opts.Dir())
			fmt.Printf("  3. arc apply -f %s/mcpserver.yaml\n", opts.Dir())
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.Version, "version", "0.1.0", "Initial version")
	cmd.Flags().StringVar(&opts.Description, "description", "", "Server description")
	cmd.Flags().StringVar(&opts.Framework, "framework", "fastmcp-python", "MCP framework")
	cmd.Flags().StringVar(&opts.Transport, "transport", "stdio", "Transport type: stdio, streamable-http")
	cmd.Flags().StringVar(&opts.Image, "image", "", "Docker image (default: localhost:5001/<name>:latest)")
	cmd.Flags().StringVar(&opts.OutputDir, "output-dir", "", "Output directory (default: ./<name>)")

	return cmd
}

func newInitSkillCmd() *cobra.Command {
	var opts scaffold.SkillOptions

	cmd := &cobra.Command{
		Use:   "skill NAME",
		Short: "Scaffold a new skill project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = args[0]
			if err := scaffold.Skill(opts); err != nil {
				return err
			}
			fmt.Printf("Skill project scaffolded in ./%s/\n", opts.Dir())
			fmt.Println("\nNext steps:")
			fmt.Printf("  1. cd %s && edit SKILL.md\n", opts.Dir())
			fmt.Printf("  2. arc apply -f %s/skill.yaml\n", opts.Dir())
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.Version, "version", "0.1.0", "Initial version")
	cmd.Flags().StringVar(&opts.Description, "description", "", "Skill description")
	cmd.Flags().StringVar(&opts.Category, "category", "general", "Skill category")
	cmd.Flags().StringVar(&opts.OutputDir, "output-dir", "", "Output directory (default: ./<name>)")

	return cmd
}

func newInitPromptCmd() *cobra.Command {
	var opts scaffold.PromptOptions

	cmd := &cobra.Command{
		Use:   "prompt NAME",
		Short: "Scaffold a new prompt resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = args[0]
			if err := scaffold.Prompt(opts); err != nil {
				return err
			}
			fmt.Printf("Prompt scaffolded in ./%s/\n", opts.Dir())
			fmt.Println("\nNext steps:")
			fmt.Printf("  1. Edit %s/prompt.yaml\n", opts.Dir())
			fmt.Printf("  2. arc apply -f %s/prompt.yaml\n", opts.Dir())
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.Version, "version", "0.1.0", "Initial version")
	cmd.Flags().StringVar(&opts.Description, "description", "", "Prompt description")
	cmd.Flags().StringVar(&opts.Content, "content", "", "Initial prompt content")
	cmd.Flags().StringVar(&opts.OutputDir, "output-dir", "", "Output directory (default: ./<name>)")

	return cmd
}
