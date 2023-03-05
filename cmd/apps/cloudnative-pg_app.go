// Copyright (c) arkade author(s) 2022. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package apps

import (
	"errors"
	"fmt"
	"github.com/alexellis/arkade/pkg/env"
	"github.com/alexellis/arkade/pkg/get"
	"github.com/alexellis/arkade/pkg/k8s"
	execute "github.com/alexellis/go-execute/pkg/v1"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

const (
	cnpgVersion = "v1.19.0"
)

func MakeInstallCloudNativePG() *cobra.Command {
	var cloudNativePG = &cobra.Command{
		Use:          "cloudnative-pg",
		Short:        "Install CloudNativePG",
		Long:         "Install CloudNativePG operator to manage PostgreSQL clusters",
		Example:      "arkade install cloudnative-pg",
		SilenceUsage: true,
	}

	cloudNativePG.Flags().StringP("version", "v", cnpgVersion, "The version of CloudNativePG to install (latest for the latest version")

	cloudNativePG.RunE = func(cmd *cobra.Command, args []string) error {
		version, _ := cloudNativePG.Flags().GetString("version")
		if version == "latest" {
			version, _ = get.FindGitHubRelease("cloudnative-pg", "cloudnative-pg")
		}

		if len(version) == 0 {
			return fmt.Errorf("you must provide a version to install using --version flag")
		}

		tools := get.MakeTools()
		var tool *get.Tool
		for _, t := range tools {
			if t.Name == "kubectl-cnpg" {
				tool = &t
				break
			}
		}
		if tool == nil {
			return fmt.Errorf("unable to find kubectl cnpg tool")
		}

		if _, err := os.Stat(env.LocalBinary(tool.Name, "")); errors.Is(err, os.ErrNotExist) {
			arch, clientOS := env.GetClientArch()
			_, _, err := get.Download(tool, arch, clientOS, version, get.DownloadArkadeDir, false, false)
			if err != nil {
				return err
			}
		}

		cnpgPlugin := env.LocalBinary("kubectl-cnpg", "")

		task := execute.ExecTask{
			Command:     cnpgPlugin,
			Args:        []string{"install", "generate"},
			Env:         os.Environ(),
			StreamStdio: false,
		}
		execResult, err := task.Execute()
		if err != nil {
			return err
		}

		reader := strings.NewReader(execResult.Stdout)
		_, err = k8s.KubectlTaskStdin(reader, "apply", "-f", "-")

		return err
	}

	return cloudNativePG
}

// Confirm manifest exists for the specified version, following this format
// https://github.com/cloudnative-pg/cloudnative-pg/releases/download/v1.19.0/cnpg-1.19.0.yaml
func confirmVersionExists(version string) bool {

	return false
}

const CloudNativePGInfoMsg = `# Get started with CloudNativePG here:
# https://github.com/cloudnative-pg/cloudnative-pg#getting-started
`
