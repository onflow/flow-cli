package flix

import (
	"context"
	"fmt"
	"os"

	"github.com/onflow/flow-cli/internal/scripts"

	"github.com/onflow/flow-cli/internal/transactions"

	"github.com/onflow/flixkit-go"
	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/spf13/cobra"
)

type Flags struct {
	ArgsJSON    string   `default:"" flag:"args-json" info:"arguments in JSON-Cadence format"`
	BlockID     string   `default:"" flag:"block-id" info:"block ID to execute the script at"`
	BlockHeight uint64   `default:"" flag:"block-height" info:"block height to execute the script at"`
	Signer      string   `default:"" flag:"signer" info:"Account name from configuration used to sign the transaction as proposer, payer and suthorizer"`
	Proposer    string   `default:"" flag:"proposer" info:"Account name from configuration used as proposer"`
	Payer       string   `default:"" flag:"payer" info:"Account name from configuration used as payer"`
	Authorizers []string `default:"" flag:"authorizer" info:"Name of a single or multiple comma-separated accounts used as authorizers from configuration"`
	Include     []string `default:"" flag:"include" info:"Fields to include in the output"`
	Exclude     []string `default:"" flag:"exclude" info:"Fields to exclude from the output (events)"`
	GasLimit    uint64   `default:"1000" flag:"gas-limit" info:"transaction gas limit"`
}

var flags = Flags{}

var Cmd = &cobra.Command{
	Use:   "flix",
	Short: "Commands for the Flix functionality",
}

var idCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "id <id>",
		Short:   "Execute flix operation with given id",
		Args:    cobra.MinimumNArgs(1),
		Example: `flow flix id 123`,
	},
	Flags: &flags,
	RunS: func(args []string, _ command.GlobalFlags, _ output.Logger, flow flowkit.Services, state *flowkit.State) (result command.Result, err error) {
		flixService := flixkit.NewFlixService(&flixkit.Config{})
		flixID := args[0]
		ctx := context.Background()
		template, err := flixService.GetFlixByID(ctx, flixID)
		if err != nil {
			return nil, fmt.Errorf("could not find flix template")
		}

		cadenceWithImportsReplaced, err := template.GetAndReplaceCadenceImports(flow.Network().Name)
		if err != nil {
			return nil, fmt.Errorf("could not replace imports")
		}

		if template.IsScript() {
			scriptsFlags := scripts.Flags{
				ArgsJSON:    flags.ArgsJSON,
				BlockID:     flags.BlockID,
				BlockHeight: flags.BlockHeight,
			}
			return scripts.SendScript([]byte(cadenceWithImportsReplaced), args[1:], "", flow, scriptsFlags)
		} else {
			transactionFlags := transactions.Flags{
				ArgsJSON:    flags.ArgsJSON,
				Signer:      flags.Signer,
				Proposer:    flags.Proposer,
				Payer:       flags.Payer,
				Authorizers: flags.Authorizers,
				Include:     flags.Include,
				Exclude:     flags.Exclude,
				GasLimit:    flags.GasLimit,
			}
			return transactions.SendTransaction([]byte(cadenceWithImportsReplaced), args[1:], "", flow, state, transactionFlags)
		}
	},
}

var nameCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "name <name>",
		Short:   "Execute flix operation with given name",
		Args:    cobra.MinimumNArgs(1),
		Example: `flow flix name transfer-flow`,
	},
	Flags: &flags,
	RunS: func(args []string, _ command.GlobalFlags, _ output.Logger, flow flowkit.Services, state *flowkit.State) (result command.Result, err error) {
		flixService := flixkit.NewFlixService(&flixkit.Config{})
		flixName := args[0]
		ctx := context.Background()
		template, err := flixService.GetFlix(ctx, flixName)
		if err != nil {
			return nil, fmt.Errorf("could not find flix template")
		}

		cadenceWithImportsReplaced, err := template.GetAndReplaceCadenceImports(flow.Network().Name)
		if err != nil {
			return nil, fmt.Errorf("could not replace imports")
		}

		if template.IsScript() {
			scriptsFlags := scripts.Flags{
				ArgsJSON:    flags.ArgsJSON,
				BlockID:     flags.BlockID,
				BlockHeight: flags.BlockHeight,
			}
			return scripts.SendScript([]byte(cadenceWithImportsReplaced), args[1:], "", flow, scriptsFlags)
		} else {
			transactionFlags := transactions.Flags{
				ArgsJSON:    flags.ArgsJSON,
				Signer:      flags.Signer,
				Proposer:    flags.Proposer,
				Payer:       flags.Payer,
				Authorizers: flags.Authorizers,
				Include:     flags.Include,
				Exclude:     flags.Exclude,
				GasLimit:    flags.GasLimit,
			}
			return transactions.SendTransaction([]byte(cadenceWithImportsReplaced), args[1:], "", flow, state, transactionFlags)
		}
	},
}

var pathCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "path <path>",
		Short:   "Execute flix operation with given name",
		Args:    cobra.MinimumNArgs(1),
		Example: `flow flix path transfer-flow`,
	},
	Flags: &flags,
	RunS: func(args []string, _ command.GlobalFlags, _ output.Logger, flow flowkit.Services, state *flowkit.State) (result command.Result, err error) {
		filePath := args[0]
		file, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("could not read file")
		}

		template, err := flixkit.ParseFlix(string(file))
		if err != nil {
			return nil, fmt.Errorf("could not parse flix")
		}

		cadenceWithImportsReplaced, err := template.GetAndReplaceCadenceImports(flow.Network().Name)
		if err != nil {
			return nil, fmt.Errorf("could not replace imports")
		}

		if template.IsScript() {
			scriptsFlags := scripts.Flags{
				ArgsJSON:    flags.ArgsJSON,
				BlockID:     flags.BlockID,
				BlockHeight: flags.BlockHeight,
			}
			return scripts.SendScript([]byte(cadenceWithImportsReplaced), args[1:], "", flow, scriptsFlags)
		} else {
			transactionFlags := transactions.Flags{
				ArgsJSON:    flags.ArgsJSON,
				Signer:      flags.Signer,
				Proposer:    flags.Proposer,
				Payer:       flags.Payer,
				Authorizers: flags.Authorizers,
				Include:     flags.Include,
				Exclude:     flags.Exclude,
				GasLimit:    flags.GasLimit,
			}
			return transactions.SendTransaction([]byte(cadenceWithImportsReplaced), args[1:], "", flow, state, transactionFlags)
		}
	},
}

func init() {
	idCommand.AddToParent(Cmd)
	nameCommand.AddToParent(Cmd)
	pathCommand.AddToParent(Cmd)
}
