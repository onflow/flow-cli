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
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flow"

	"github.com/onflow/flow-cli/internal/blocks"
	"github.com/onflow/flow-cli/internal/emulator"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/accounts"
	"github.com/onflow/flow-cli/internal/cadence"
	"github.com/onflow/flow-cli/internal/collections"
	"github.com/onflow/flow-cli/internal/events"
	"github.com/onflow/flow-cli/internal/keys"
	"github.com/onflow/flow-cli/internal/project"
	"github.com/onflow/flow-cli/internal/scripts"
	"github.com/onflow/flow-cli/internal/transactions"
	"github.com/onflow/flow-cli/internal/version"
)

var c = &cobra.Command{
	Use:              "flow",
	TraverseChildren: true,
}

func init() {

	c.AddCommand(cadence.Cmd)
	c.AddCommand(version.Cmd)
	c.AddCommand(emulator.Cmd)

	c.AddCommand(accounts.Cmd)
	command.Add(accounts.Cmd, accounts.NewGetCmd())
	command.Add(accounts.Cmd, accounts.NewAddCmd())
	command.Add(accounts.Cmd, accounts.NewCreateCmd())
	command.Add(accounts.Cmd, accounts.NewStakingInfoCmd())
	command.Add(accounts.Cmd, accounts.NewAddContractCmd())
	command.Add(accounts.Cmd, accounts.NewRemoveContractCmd())
	command.Add(accounts.Cmd, accounts.NewUpdateContractCmd())

	c.AddCommand(scripts.Cmd)
	command.Add(scripts.Cmd, scripts.NewExecuteScriptCmd())

	c.AddCommand(transactions.Cmd)
	command.Add(transactions.Cmd, transactions.NewSendCmd())
	command.Add(transactions.Cmd, transactions.NewGetCmd())

	c.AddCommand(keys.Cmd)
	command.Add(keys.Cmd, keys.NewGenerateCmd())
	command.Add(keys.Cmd, keys.NewCmdDecode())

	c.AddCommand(events.Cmd)
	command.Add(events.Cmd, events.NewGetCmd())

	c.AddCommand(blocks.Cmd)
	command.Add(blocks.Cmd, blocks.NewGetCmd())

	c.AddCommand(collections.Cmd)
	command.Add(collections.Cmd, collections.NewGetCmd())

	c.AddCommand(project.Cmd)
	command.Add(project.Cmd, project.NewInitCmd())
	command.Add(project.Cmd, project.NewDeployCmd())

	c.PersistentFlags().StringVarP(
		&command.HostFlag,
		"host",
		"",
		command.HostFlag,
		"Flow Access API host address",
	)

	c.PersistentFlags().StringVarP(
		&command.FilterFlag,
		"filter",
		"x",
		command.FilterFlag,
		"Filter result values by property name",
	)

	c.PersistentFlags().StringVarP(
		&command.FormatFlag,
		"output",
		"o",
		command.FormatFlag,
		"Output format",
	)

	c.PersistentFlags().StringVarP(
		&command.SaveFlag,
		"save",
		"s",
		command.SaveFlag,
		"Save result to a filename",
	)

	c.PersistentFlags().StringVarP(
		&command.LogFlag,
		"log",
		"l",
		command.LogFlag,
		"Log level verbosity",
	)

	c.PersistentFlags().BoolVarP(
		&command.RunEmulatorFlag,
		"emulator",
		"e",
		command.RunEmulatorFlag,
		"Run in-memory emulator",
	)

	c.PersistentFlags().StringSliceVarP(
		&flow.ConfigPath,
		"conf",
		"f",
		flow.ConfigPath,
		"Path to flow configuration file",
	)

	c.PersistentFlags().StringVarP(
		&command.NetworkFlag,
		"network",
		"n",
		command.NetworkFlag,
		"Network from configuration file",
	)
}

func main() {
	if err := c.Execute(); err != nil {
		flow.Exit(1, err.Error())
	}
}
