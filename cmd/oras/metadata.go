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
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"oras.land/oras/cmd/oras/internal/option"

	"github.com/spf13/cobra"
)

type metadataOptions struct {
	option.Common
	option.Remote

	targetRef string
}

const devcontainerCollectionMediaType = "application/vnd.devcontainers.collection.layer.v1+json"

func metadataCmd() *cobra.Command {
	var opts metadataOptions
	cmd := &cobra.Command{
		Use:   "metadata [options] <name:tag|name@digest>",
		Short: "Fetch dev container collection metadata",
		Long:  `Fetch all available dev container collection metadata`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.targetRef = args[0]
			return runMetadata(opts)
		},
	}

	option.ApplyFlags(&opts, cmd.Flags())
	return cmd
}

func runMetadata(opts metadataOptions) error {

	var pullOpts = pullOptions{
		Common:            opts.Common,
		Remote:            opts.Remote,
		targetRef:         opts.targetRef,
		RequiredMediaType: devcontainerCollectionMediaType,
		HideOutput:        true,
		Output:            "/tmp",
	}

	if !strings.Contains(opts.targetRef, ":") {
		pullOpts.targetRef = fmt.Sprintf("%s:latest", opts.targetRef)
	}

	err := runPull(pullOpts)

	if err != nil {
		return err
	}

	collectionJsonBytes, err := os.ReadFile("/tmp/devcontainer-collection.json")
	if err != nil {
		return err
	}

	// TODO: parse the metadata to get each feature ID
	var collection Collection
	err = json.Unmarshal(collectionJsonBytes, &collection)
	if err != nil {
		fmt.Println("error:", err)
	}

	var referenceWithoutTag = strings.Split(opts.targetRef, ":")[0]

	fmt.Printf("Source Code: https://github.com/%s/%s\n", collection.SourceInformation.Owner, collection.SourceInformation.Repo)
	fmt.Printf("Commit:      %s\n\n", collection.SourceInformation.Sha)

	fmt.Println("Available Features:")
	for _, feature := range collection.Features {

		fmt.Println(feature.ID)
		ref := fmt.Sprintf("%s/%s", referenceWithoutTag, feature.ID)
		err = runFetchTags(fetchTagsOptions{
			Common:    opts.Common,
			Remote:    opts.Remote,
			targetRef: ref,
			prefix:    "   ",
		})

		if err != nil {
			return err
		}

	}

	// Run fetch tags on it

	return nil
}

type Collection struct {
	SourceInformation struct {
		Owner string `json:"owner"`
		Repo  string `json:"repo"`
		Ref   string `json:"ref"`
		Sha   string `json:"sha"`
		Tag   string `json:"tag"`
	} `json:"sourceInformation"`
	Features []struct {
		ID          string `json:"id"`
		Version     string `json:"version"`
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"features"`
	Templates []interface{} `json:"templates"`
}
