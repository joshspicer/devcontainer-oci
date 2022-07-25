package main

import (
	"os"

	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:          "devcontainers [command]",
		SilenceUsage: true,
	}
	cmd.AddCommand(
		// pullCmd(),
		resolveCmd(),
		// pushCmd(),
		loginCmd(),
		logoutCmd(),
		fetchTagsCmd(),
		metadataCmd(),
		versionCmd(),
		// discoverCmd(),
		// copyCmd(),
		// attachCmd(),
	)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
