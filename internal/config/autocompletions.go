/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/pkg/flowkit/output"
)

var CompletionCmd = &cobra.Command{
	Use:                   "setup-completions [powershell]",
	Short:                 "Setup command autocompletion",
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"powershell"},
	Args:                  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		shell := ""
		if len(args) == 1 {
			shell = args[0]
		}

		if shell == "powershell" {
			_ = cmd.Root().GenPowerShellCompletion(os.Stdout)
		} else {
			shell, shellOS := output.AutocompletionPrompt()

			if shell == "bash" && shellOS == "MacOS" {
				_ = cmd.Root().GenBashCompletionFile("/usr/local/etc/bash_completion.d/flow")

				fmt.Printf("Flow command completions installed in: /usr/local/etc/bash_completion.d/flow\n")
				fmt.Printf("You will need to start a new shell for this setup to take effect.\n\n")
			} else if shell == "bash" && shellOS == "Linux" {
				_ = cmd.Root().GenBashCompletionFile("/etc/bash_completion.d/flow")

				fmt.Printf("Flow command completions installed in: /etc/bash_completion.d/flow\n")
				fmt.Printf("You will need to start a new shell for this setup to take effect.\n\n")
			} else if shell == "zsh" {
				c := exec.Command("zsh", "-c ", `echo -n ${fpath[1]}`)
				path, _ := c.Output()
				_ = cmd.Root().GenZshCompletionFile(fmt.Sprintf("%s/_flow", path))

				fmt.Printf("Flow command completions installed in: '%s/_flow'\n", path)
				fmt.Printf("You will need to start a new shell for this setup to take effect.\n\n")
			}
		}
	},
}
