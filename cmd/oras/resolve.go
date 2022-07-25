package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"oras.land/oras/cmd/oras/internal/option"
)

type resolveOptions struct {
	option.Common
	option.Remote

	referencesList []string
}

var lockFilePath = "devcontainers.lock"

func resolveCmd() *cobra.Command {
	var opts resolveOptions
	cmd := &cobra.Command{
		Use:   "resolve <references-list>",
		Short: "Internal commmand to resolve a list of references",
		Long:  `Resolves a list of references`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.referencesList = strings.Split(args[0], ",")
			return runResolve(opts)
		},
	}
	option.ApplyFlags(&opts, cmd.Flags())
	return cmd
}

func runResolve(opts resolveOptions) error {

	for _, reference := range opts.referencesList {
		fmt.Println("Fetching reference:", reference)
		var pullOpts = pullOptions{
			Common:       opts.Common,
			Remote:       opts.Remote,
			targetRef:    reference,
			LockFilePath: lockFilePath,
		}

		runPull(pullOpts)
	}

	return nil
}
