/*
 * Flow CLI
 *
 * Copyright 2019-2020 Dapper Labs, Inc.
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

import "C"
import (
	"fmt"
	"github.com/onflow/flow-cli/cmd"
	"github.com/onflow/flow-cli/cmd/accounts"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/flow/blocks"
	"github.com/onflow/flow-cli/flow/cadence"
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/flow/collections"
	"github.com/onflow/flow-cli/flow/emulator"
	"github.com/onflow/flow-cli/flow/events"
	"github.com/onflow/flow-cli/flow/initialize"
	"github.com/onflow/flow-cli/flow/keys"
	"github.com/onflow/flow-cli/flow/project"
	"github.com/onflow/flow-cli/flow/scripts"
	"github.com/onflow/flow-cli/flow/transactions"
	"github.com/onflow/flow-cli/flow/version"
)

var c = &cobra.Command{
	Use:              "flow",
	TraverseChildren: true,
}

func init() {
	c.AddCommand(project.Cmd)
	c.AddCommand(initialize.Cmd)
	c.AddCommand(blocks.Cmd)
	c.AddCommand(collections.Cmd)
	c.AddCommand(keys.Cmd)
	c.AddCommand(emulator.Cmd)
	c.AddCommand(events.Cmd)
	c.AddCommand(cadence.Cmd)
	c.AddCommand(scripts.Cmd)
	c.AddCommand(transactions.Cmd)
	c.AddCommand(version.Cmd)

	addCommand(c, accounts.Init())

	c.PersistentFlags().StringSliceVarP(&cli.ConfigPath, "config-path", "f", cli.ConfigPath, "Path to flow configuration file")
}

func addCommand(c *cobra.Command, command cmd.Command) {
	command.GetCmd().RunE = func(cmd *cobra.Command, args []string) error {

		// validation of flags
		err := command.ValidateFlags()
		if err != nil {
			fmt.Println("flag validation error", err)
		}

		// initialize project but ignore error since config can be missing
		project, _ := cli.LoadProject(cli.ConfigPath)

		// run command
		result, err := command.Run(cmd, args, project)
		if err != nil {
			fmt.Println("Error: ", err)
			return nil
		}

		// TODO check flag for json
		fmt.Println(result.String())

		return nil
	}

	bindFlags(command.SetFlags())
	c.AddCommand(command.GetCmd())
}

func bindFlags(config *sconfig.Config) {
	err := config.
		FromEnvironment(cli.EnvPrefix).
		BindFlags(c.PersistentFlags()).
		Parse()
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	if err := c.Execute(); err != nil {
		cli.Exit(1, err.Error())
	}
}
