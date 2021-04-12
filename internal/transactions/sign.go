package transactions

import (
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
)

type flagsSign struct {
	ArgsJSON              string   `default:"" flag:"args-json" info:"arguments in JSON-Cadence format"`
	Args                  []string `default:"" flag:"arg" info:"argument in Type:Value format"`
	Signer                string   `default:"emulator-account" flag:"signer" info:"name of the account used to sign"`
	Payload               string   `flag:"payload" info:"path to the transaction payload file"`
	Proposer              string   `default:"" flag:"proposer" info:"name of the account that is proposing the transaction"`
	Role                  string   `default:"authorizer" flag:"role" info:"Specify a role of the signer, values: proposer, payer, authorizer"`
	AdditionalAuthorizers []string `flag:"add-authorizer" info:"Additional authorizer addresses to add to the transaction"`
	PayerAddress          string   `flag:"payer-address" info:"Specify payer of the transaction. Defaults to first signer."`
}

var signFlags = flagsSign{}

var SignCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "sign <optional code filename>",
		Short:   "Sign a transaction",
		Example: "flow transactions sign",
		Args:    cobra.MaximumNArgs(1),
	},
	Flags: &signFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {

		codeFilename := ""
		if len(args) > 0 {
			codeFilename = args[0]
		}

		signed, err := services.Transactions.Sign(
			signFlags.Signer,
			signFlags.Proposer,
			signFlags.PayerAddress,
			signFlags.AdditionalAuthorizers,
			signFlags.Role,
			codeFilename,
			signFlags.Payload,
			signFlags.Args,
			signFlags.ArgsJSON,
			globalFlags.Yes,
		)
		if err != nil {
			return nil, err
		}

		return &TransactionResult{
			tx: signed.FlowTransaction(),
		}, nil
	},
}
