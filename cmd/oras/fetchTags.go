/*
Copyright GitHub

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"strings"

	"oras.land/oras/cmd/oras/internal/option"

	"github.com/spf13/cobra"
)

type fetchTagsOptions struct {
	option.Common
	option.Remote

	targetRef string
	prefix    string
}

func fetchTagsCmd() *cobra.Command {
	var opts fetchTagsOptions
	cmd := &cobra.Command{
		Use:   "fetchTags [options] <name:tag|name@digest>",
		Short: "Fetch all published tags for target ref",
		Long:  `Fetch all published tags for target ref`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.targetRef = args[0]
			return runFetchTags(opts)
		},
	}

	option.ApplyFlags(&opts, cmd.Flags())
	return cmd
}

func runFetchTags(opts fetchTagsOptions) error {
	if len(strings.Split(opts.targetRef, ":")) > 1 {
		return fmt.Errorf("target repository must not container a version tag")
	}

	ctx, _ := opts.SetLoggerLevel()
	repo, err := opts.NewRepository(opts.targetRef, opts.Common)

	if err != nil {
		return err
	}

	repo.Tags(ctx, "", func(tags []string) error {
		for _, t := range tags {
			fmt.Printf("%s%s\n", opts.prefix, t)
		}
		return nil
	})

	return nil
}
