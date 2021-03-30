package transactions

import (
	"fmt"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/spf13/cobra"
)

type flagsSign struct {
	ArgsJSON              string   `default:"" flag:"args-json" info:"arguments in JSON-Cadence format"`
	Args                  []string `default:"" flag:"arg" info:"argument in Type:Value format"`
	Signer                string   `default:"emulator-account" flag:"signer"`
	Payload               string   `flag:"payload" info:"path to the transaction payload file"`
	Code                  string   `default:"" flag:"code" info:"⚠️  DEPRECATED: use filename argument"`
	Results               bool     `default:"" flag:"results" info:"⚠️  DEPRECATED: all transactions will provide result"`
	Proposer              string   `default:"" flag:"proposer"`
	Role                  string   `default:"authorizer" flag:"role"`
	AdditionalAuthorizers []string `flag:"additional-authorizers" info:"Additional authorizer addresses to add to the transaction"`
	PayerAddress          string   `flag:"payer-address" info:"Specify payer of the transaction. Defaults to current signer."`
	Encoding              string   `default:"hexrlp" flag:"encoding" info:"Encoding to use for transactio (rlp)"`
}

var signFlags = flagsSign{}

var SignCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "sign",
		Short:   "Sign a transaction",
		Example: `flow transactions sign`,
	},
	Flags: &signFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		if sendFlags.Code != "" {
			return nil, fmt.Errorf("⚠️  DEPRECATED: use filename argument")
		}

		if sendFlags.Results {
			return nil, fmt.Errorf("⚠️  DEPRECATED: all transactions will provide results")
		}

		signed, err := services.Transactions.Sign(
			signFlags.Signer,
			signFlags.Proposer,
			signFlags.PayerAddress,
			signFlags.AdditionalAuthorizers,
			signFlags.Role,
			signFlags.Code,
			signFlags.Payload,
			signFlags.Args,
			signFlags.ArgsJSON,
		)
		if err != nil {
			return nil, err
		}

		return &TransactionResult{
			result: nil,
			tx:     signed.FlowTransaction(),
		}, nil
	},
}
