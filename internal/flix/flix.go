package flix

import (
	"context"
	"fmt"
	"os"

	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"

	"github.com/onflow/flow-cli/flowkit"

	"github.com/onflow/flixkit-go"

	"github.com/onflow/flow-cli/internal/scripts"
	"github.com/onflow/flow-cli/internal/transactions"
	"github.com/spf13/cobra"
)

type flixFlags struct {
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
	ID          string   `default:"" flag:"id" info:"id of the flix"`
	Name        string   `default:"" flag:"name" info:"name of the flix"`
	Path        string   `default:"" flag:"path" info:"path to the flix file"`
}

var flags = flixFlags{}

var FlixCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "flix [<argument> <argument> ...]",
		Short:   "Execute flix operation with given id, name or path",
		Args:    cobra.MinimumNArgs(1),
		Example: `flow flix --id 123`,
	},
	Flags: &flags,
	RunS: func(args []string, _ command.GlobalFlags, _ output.Logger, flow flowkit.Services, state *flowkit.State) (result command.Result, err error) {
		flixService := flixkit.NewFlixService(&flixkit.Config{})
		ctx := context.Background()
		var template *flixkit.FlowInteractionTemplate

		if flags.ID != "" {
			template, err = flixService.GetFlixByID(ctx, flags.ID)
			if err != nil {
				return nil, fmt.Errorf("could not find flix with id %s", flags.ID)
			}
		} else if flags.Name != "" {
			template, err = flixService.GetFlix(ctx, flags.Name)
			if err != nil {
				return nil, fmt.Errorf("could not find flix with name %s", flags.Name)
			}
		} else if flags.Path != "" {
			file, err := os.ReadFile(flags.Path)
			if err != nil {
				return nil, fmt.Errorf("could not read file")
			}
			template, err = flixkit.ParseFlix(string(file))
		}

		if err != nil {
			return nil, fmt.Errorf("could not find or parse flix template")
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
			return scripts.SendScript([]byte(cadenceWithImportsReplaced), args, "", flow, scriptsFlags)
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
			return transactions.SendTransaction([]byte(cadenceWithImportsReplaced), args, "", flow, state, transactionFlags)
		}
	},
}
