package transactions

import (
	"fmt"

	"github.com/onflow/flow-cli/pkg/flowkit/templates"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/spf13/cobra"
)

type flagsTemplate struct {
	ArgsJSON string   `default:"" flag:"args-json" info:"arguments in JSON-Cadence format"`
	Signer   string   `default:"emulator-account" flag:"signer" info:"Account name from configuration used to sign the transaction"`
	GasLimit uint64   `default:"1000" flag:"gas-limit" info:"transaction gas limit"`
	Include  []string `default:"" flag:"include" info:"Fields to include in the output"`
	Exclude  []string `default:"" flag:"exclude" info:"Fields to exclude from the output (events)"`
}

var templateFlags = flagsTemplate{}

var SendTemplateCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "send-template <template name> [<argument> <argument> ...]",
		Short:   "Send a template transaction",
		Args:    cobra.MinimumNArgs(1),
		Example: `flow transactions send-template transfer 10 0x1`,
	},
	Flags: &templateFlags,
	RunS:  sendTemplate,
}

func sendTemplate(
	args []string,
	_ flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	services *services.Services,
	state *flowkit.State,
) (command.Result, error) {
	templateName := args[0]
	template, err := templates.ByName(templateName)
	if err != nil {
		return nil, err
	}

	source, err := template.Source(globalFlags.Network)
	if err != nil {
		return nil, err
	}

	signer, err := state.Accounts().ByName(templateFlags.Signer)
	if err != nil {
		return nil, err
	}

	var transactionArgs []cadence.Value
	if templateFlags.ArgsJSON != "" {
		transactionArgs, err = flowkit.ParseArgumentsJSON(templateFlags.ArgsJSON)
	} else {
		transactionArgs, err = flowkit.ParseArgumentsWithoutType("", source, args[1:])
	}

	if err != nil {
		return nil, fmt.Errorf("error parsing transaction arguments: %w", err)
	}

	tx, result, err := services.Transactions.Send(
		signer,
		source,
		"",
		templateFlags.GasLimit,
		transactionArgs,
		globalFlags.Network,
	)

	if err != nil {
		return nil, err
	}

	return &TransactionResult{
		result:  result,
		tx:      tx,
		include: templateFlags.Include,
		exclude: templateFlags.Exclude,
	}, nil
}
