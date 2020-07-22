// Package main implements the entry point for the Flow CLI.
package main

import (
	"github.com/spf13/cobra"

	cli "github.com/dapperlabs/flow-cli/flow"
	"github.com/dapperlabs/flow-cli/flow/accounts"
	"github.com/dapperlabs/flow-cli/flow/cadence"
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
