/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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

// Package main implements the entry point for the Flow CLI.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/onflow/flow-cli/internal/completion"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/accounts"
	"github.com/onflow/flow-cli/internal/blocks"
	"github.com/onflow/flow-cli/internal/cadence"
	"github.com/onflow/flow-cli/internal/collections"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/config"
	"github.com/onflow/flow-cli/internal/emulator"
	"github.com/onflow/flow-cli/internal/events"
	"github.com/onflow/flow-cli/internal/keys"
	"github.com/onflow/flow-cli/internal/project"
	"github.com/onflow/flow-cli/internal/scripts"
	"github.com/onflow/flow-cli/internal/transactions"
	"github.com/onflow/flow-cli/internal/version"
	"github.com/onflow/flow-cli/pkg/flowcli/util"
)

func main() {
	var cmd = &cobra.Command{
		Use:              "flow",
		TraverseChildren: true,
	}

	autocompletion(cmd)
	// hot commands
	config.InitCommand.AddToParent(cmd)

	// structured commands
	cmd.AddCommand(cadence.Cmd)
	cmd.AddCommand(version.Cmd)
	cmd.AddCommand(emulator.Cmd)
	cmd.AddCommand(accounts.Cmd)
	cmd.AddCommand(scripts.Cmd)
	cmd.AddCommand(transactions.Cmd)
	cmd.AddCommand(keys.Cmd)
	cmd.AddCommand(events.Cmd)
	cmd.AddCommand(blocks.Cmd)
	cmd.AddCommand(collections.Cmd)
	cmd.AddCommand(project.Cmd)

	cmd.AddCommand(completion.Cmd)

	command.InitFlags(cmd)

	if err := cmd.Execute(); err != nil {
		util.Exit(1, err.Error())
	}
}

func autocompletion(cmd *cobra.Command) {
	shell := os.Getenv("SHELL")

	if strings.Contains(shell, "zsh") {
		c := exec.Command("zsh", "-c ", `echo -n ${fpath[1]}`)
		path, err := c.Output()
		if err != nil {
			return
		}

		cmd.GenZshCompletionFile(fmt.Sprintf("%s/_dflow", path))
	}
}
