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
		newConfigSetClusterCmd(),
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

func newConfigSetClusterCmd() *cobra.Command {
	var server string

	cmd := &cobra.Command{
		Use:   "set-cluster NAME",
		Short: "Add or update a cluster entry",
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

			// Update existing or append new.
			found := false
			for i := range cfg.Clusters {
				if cfg.Clusters[i].Name == args[0] {
					cfg.Clusters[i].Server = server
					found = true
					break
				}
			}
			if !found {
				cfg.Clusters = append(cfg.Clusters, config.Cluster{
					Name:   args[0],
					Server: server,
				})
			}

			if err := config.Save(path, cfg); err != nil {
				return err
			}
			fmt.Printf("Cluster %q set to %s\n", args[0], server)
			return nil
		},
	}

	cmd.Flags().StringVar(&server, "server", "", "Registry server URL")
	cmd.MarkFlagRequired("server")

	return cmd
}

func newConfigSetContextCmd() *cobra.Command {
	var cluster, token string

	cmd := &cobra.Command{
		Use:   "set-context NAME",
		Short: "Add or update a context entry",
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

			found := false
			for i := range cfg.Contexts {
				if cfg.Contexts[i].Name == args[0] {
					if cluster != "" {
						cfg.Contexts[i].Cluster = cluster
					}
					if token != "" {
						cfg.Contexts[i].Token = token
					}
					found = true
					break
				}
			}
			if !found {
				cfg.Contexts = append(cfg.Contexts, config.Context{
					Name:    args[0],
					Cluster: cluster,
					Token:   token,
				})
			}

			if err := config.Save(path, cfg); err != nil {
				return err
			}
			fmt.Printf("Context %q configured\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&cluster, "cluster", "", "Cluster name to reference")
	cmd.Flags().StringVar(&token, "token", "", "Auth token for this context")

	return cmd
}
