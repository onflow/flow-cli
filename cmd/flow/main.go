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

import (
	"github.com/spf13/cobra"

	cli "github.com/dapperlabs/flow-cli/flow"
	"github.com/dapperlabs/flow-cli/flow/accounts"
	"github.com/dapperlabs/flow-cli/flow/blocks"
	"github.com/dapperlabs/flow-cli/flow/cadence"
	"github.com/dapperlabs/flow-cli/flow/collections"
	"github.com/dapperlabs/flow-cli/flow/emulator"
	"github.com/dapperlabs/flow-cli/flow/initialize"
	"github.com/dapperlabs/flow-cli/flow/keys"
	"github.com/dapperlabs/flow-cli/flow/scripts"
	"github.com/dapperlabs/flow-cli/flow/transactions"
	"github.com/dapperlabs/flow-cli/flow/version"
)

var cmd = &cobra.Command{
	Use:              "flow",
	TraverseChildren: true,
}

func init() {
	cmd.AddCommand(initialize.Cmd)
	cmd.AddCommand(accounts.Cmd)
	cmd.AddCommand(blocks.Cmd)
	cmd.AddCommand(collections.Cmd)
	cmd.AddCommand(keys.Cmd)
	cmd.AddCommand(emulator.Cmd)
	cmd.AddCommand(cadence.Cmd)
	cmd.AddCommand(scripts.Cmd)
	cmd.AddCommand(transactions.Cmd)
	cmd.AddCommand(version.Cmd)
	cmd.PersistentFlags().StringVarP(&cli.ConfigPath, "config-path", "f", cli.ConfigPath, "Path to flow configuration file")
}

func main() {
	if err := cmd.Execute(); err != nil {
		cli.Exit(1, err.Error())
	}
}
