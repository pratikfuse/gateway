package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/provider"
)

var (
	// cfgPath is the path to the EnvoyGateway configuration file.
	cfgPath string
)

// getServerCommand returns the server cobra command to be executed.
func getServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "server",
		Aliases: []string{"serve"},
		Short:   "Serve Envoy Gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			return server()
		},
	}
	cmd.PersistentFlags().StringVarP(&cfgPath, "config-path", "c", "",
		"The path to the configuration file.")

	return cmd
}

// server serves Envoy Gateway.
func server() error {
	// Initialize with default config parameters.
	cfg, err := config.NewDefaultServer()
	if err != nil {
		return err
	}
	log := cfg.Logger

	// Read the config file.
	if cfgPath == "" {
		// Use default config parameters
		log.Info("No config file provided, using default parameters")
	} else {
		// Load the config file.
		eg, err := config.Decode(cfgPath)
		if err != nil {
			log.Error(err, "failed to decode config file", "name", cfgPath)
			os.Exit(1)
		}
		// Set defaults for unset fields
		eg.SetDefaults()
		cfg.EnvoyGateway = eg
	}

	k8sTable := new(provider.ResourceTable)
	if err := provider.Start(cfg, k8sTable); err != nil {
		return err
	}
	// TODO: while the provider.Start goroutine writes to the k8sTable, a (not-yet-existent)
	// translator goroutine will read from it.

	return nil
}
