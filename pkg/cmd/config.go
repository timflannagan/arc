package cmd

import (
	"fmt"
	"os"

	"github.com/agentregistry-dev/ar/pkg/config"
	"github.com/agentregistry-dev/ar/pkg/printer"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage arc configuration",
	}

	cmd.AddCommand(
		newConfigViewCmd(),
		newConfigUseContextCmd(),
		newConfigSetContextCmd(),
	)

	return cmd
}

func newConfigViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "Display the current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flagConfig)
			if err != nil {
				return err
			}
			return printer.PrintYAML(os.Stdout, cfg)
		},
	}
}

func newConfigUseContextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use-context NAME",
		Short: "Switch the active context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := flagConfig
			if path == "" {
				path = config.DefaultPath()
			}

			cfg, err := config.Load(path)
			if err != nil {
				return err
			}

			// Verify context exists.
			found := false
			for _, ctx := range cfg.Contexts {
				if ctx.Name == args[0] {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("context %q not found", args[0])
			}

			cfg.CurrentContext = args[0]
			if err := config.Save(path, cfg); err != nil {
				return err
			}
			fmt.Printf("Switched to context %q\n", args[0])
			return nil
		},
	}
}

func newConfigSetContextCmd() *cobra.Command {
	var server, token string

	cmd := &cobra.Command{
		Use:   "set-context NAME",
		Short: "Add or update a context with server and auth",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			path := flagConfig
			if path == "" {
				path = config.DefaultPath()
			}

			cfg, err := config.Load(path)
			if err != nil {
				return err
			}

			// Upsert the cluster entry (keyed by context name).
			if server != "" {
				clusterFound := false
				for i := range cfg.Clusters {
					if cfg.Clusters[i].Name == name {
						cfg.Clusters[i].Server = server
						clusterFound = true
						break
					}
				}
				if !clusterFound {
					cfg.Clusters = append(cfg.Clusters, config.Cluster{
						Name:   name,
						Server: server,
					})
				}
			}

			// Upsert the context entry.
			ctxFound := false
			for i := range cfg.Contexts {
				if cfg.Contexts[i].Name == name {
					cfg.Contexts[i].Cluster = name
					if token != "" {
						cfg.Contexts[i].Token = token
					}
					ctxFound = true
					break
				}
			}
			if !ctxFound {
				cfg.Contexts = append(cfg.Contexts, config.Context{
					Name:    name,
					Cluster: name,
					Token:   token,
				})
			}

			if err := config.Save(path, cfg); err != nil {
				return err
			}
			if server != "" {
				fmt.Printf("Context %q configured (server: %s)\n", name, server)
			} else {
				fmt.Printf("Context %q updated\n", name)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&server, "server", "", "Registry server URL")
	cmd.Flags().StringVar(&token, "token", "", "Auth token")

	return cmd
}
