/*
Copyright The ORAS Authors.
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
	"context"
	"fmt"
	"os"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras/cmd/oras/internal/display"
	"oras.land/oras/cmd/oras/internal/option"
	"oras.land/oras/internal/cache"
)

type pullOptions struct {
	option.Common
	option.Remote

	targetRef         string
	cacheRoot         string
	KeepOldFiles      bool
	PathTraversal     bool
	Output            string
	ManifestConfigRef string

	RequiredMediaType string // Ignore if empty string
	HideOutput        bool
	LockFilePath      string
}

func pullCmd() *cobra.Command {
	var opts pullOptions
	cmd := &cobra.Command{
		Use:   "pull <name:tag|name@digest>",
		Short: "Pull files from remote registry",
		Long: `Pull files from remote registry

Example - Pull all files:
  oras pull localhost:5000/hello:latest

Example - Pull files from the insecure registry:
  oras pull localhost:5000/hello:latest --insecure

Example - Pull files from the HTTP registry:
  oras pull localhost:5000/hello:latest --plain-http

Example - Pull files with local cache:
  export ORAS_CACHE=~/.oras/cache
  oras pull localhost:5000/hello:latest
`,
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			opts.cacheRoot = os.Getenv("ORAS_CACHE")
			return opts.ReadPassword()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.targetRef = args[0]
			return runPull(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.KeepOldFiles, "keep-old-files", "k", false, "do not replace existing files when pulling, treat them as errors")
	cmd.Flags().BoolVarP(&opts.PathTraversal, "allow-path-traversal", "T", false, "allow storing files out of the output directory")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", ".", "output directory")
	cmd.Flags().StringVarP(&opts.ManifestConfigRef, "manifest-config", "", "", "output manifest config file")

	// cmd.Flags().StringVarP(&opts.RequiredMediaType, "required-media-type", "", "", "Optional media type that must be present in the tree.  Ignored when omitted.")
	// cmd.Flags().BoolVarP(&opts.HideOutput, "hide-output", "", false, "hide output")

	option.ApplyFlags(&opts, cmd.Flags())
	return cmd
}

func runPull(opts pullOptions) error {
	repo, err := opts.NewRepository(opts.targetRef, opts.Common)
	if err != nil {
		return err
	}
	if repo.Reference.Reference == "" {
		return newErrInvalidReference(repo.Reference)
	}
	var src oras.Target = repo
	if opts.cacheRoot != "" {
		ociStore, err := oci.New(opts.cacheRoot)
		if err != nil {
			return err
		}
		src = cache.New(repo, ociStore)
	}

	var hasRequiredMediaType = false
	if opts.RequiredMediaType == "" {
		hasRequiredMediaType = true
	}

	// Copy Options
	copyOptions := oras.DefaultCopyOptions
	configPath, configMediaType := parseFileReference(opts.ManifestConfigRef, oras.MediaTypeUnknownConfig)
	copyOptions.FindSuccessors = func(ctx context.Context, fetcher content.Fetcher, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
		successors, err := content.Successors(ctx, fetcher, desc)
		if err != nil {
			return nil, err
		}
		var ret []ocispec.Descriptor
		for _, s := range successors {

			if s.MediaType == opts.RequiredMediaType {
				hasRequiredMediaType = true
			}

			if s.MediaType == configMediaType {
				// Add annotation for manifest config
				if s.Annotations == nil {
					s.Annotations = make(map[string]string)
				}
				s.Annotations[ocispec.AnnotationTitle] = configPath
			} else if s.Annotations[ocispec.AnnotationTitle] == "" {
				ss, err := content.Successors(ctx, fetcher, s)
				if err != nil {
					return nil, err
				}
				if len(ss) == 0 {
					continue
				}
			}
			ret = append(ret, s)
		}

		if !hasRequiredMediaType {
			return nil, fmt.Errorf("when fetching metadata required media type '%s' is required in the pull tree", opts.RequiredMediaType)
		}

		return ret, nil
	}

	pulledEmpty := true
	copyOptions.PreCopy = func(ctx context.Context, desc ocispec.Descriptor) error {
		if !opts.HideOutput {
			display.StatusPrinter("Downloading", opts.Verbose)
		}
		return nil
	}

	var lockFileDigests []LockFileItem
	copyOptions.PostCopy = func(ctx context.Context, desc ocispec.Descriptor) error {
		name := desc.Annotations[ocispec.AnnotationTitle]
		if name == "" {
			return nil
		}
		pulledEmpty = false

		if opts.LockFilePath != "" {
			// append digest to lockFileDigests
			lockFileDigests = append(lockFileDigests,
				LockFileItem{
					Target:      opts.targetRef,
					Ref:         repo.Reference.Reference,
					Registry:    repo.Reference.Registry,
					Digest:      string(desc.Digest),
					Annotations: desc.Annotations,
					MediaType:   desc.MediaType,
				})
		}

		if !opts.HideOutput {
			return display.Print("Downloaded ", display.ShortDigest(desc), name)
		}

		return nil
	}

	ctx, _ := opts.SetLoggerLevel()
	var dst = file.New(opts.Output)
	dst.AllowPathTraversalOnWrite = opts.PathTraversal
	dst.DisableOverwrite = opts.KeepOldFiles

	// Copy
	desc, err := oras.Copy(ctx, src, repo.Reference.Reference, dst, repo.Reference.Reference, copyOptions)
	if err != nil {
		return err
	}
	if pulledEmpty && !opts.HideOutput {
		fmt.Println("Downloaded empty artifact")
	}

	if !opts.HideOutput {
		fmt.Println("Pulled", opts.targetRef)
		fmt.Println("Digest:", desc.Digest)
	}

	if opts.LockFilePath != "" {
		// write lock file
		lockFile, err := os.OpenFile(opts.LockFilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		defer lockFile.Close()
		for _, item := range lockFileDigests {
			entry := fmt.Sprintf("%s\n      %s\n      %s\n      %s\n      %s\n      %s\n\n", item.Target, item.Ref, item.Registry, item.Digest, item.Annotations[ocispec.AnnotationTitle], item.MediaType)
			lockFile.WriteString(entry)
		}
	}

	return nil
}
