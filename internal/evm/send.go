package evm

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/transactions"
)

//go:embed send.cdc
var sendCode []byte

type flagsSend struct {
	Signer string `default:"" flag:"signer" info:"Account name from configuration used to sign the transaction as proposer, payer and suthorizer"`
}

var sendFlags = flagsDeploy{}

var sendCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "send <signed rlp encoded transaction file>",
		Short:   "Send a signed transaction to the EVM",
		Args:    cobra.ExactArgs(1),
		Example: "flow evm send ./tx",
	},
	Flags: &sendFlags,
	RunS:  send,
}

// todo only for demo, super hacky now

func send(
	args []string,
	_ command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	filename := args[0]
	signedTx, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("can't open tx file: %w", err)
	}

	_, err = sendSignedTx(flow, state, signedTx)

	return nil, err
}

func sendSignedTx(flow flowkit.Services, state *flowkit.State, rawTx []byte) (*flow.TransactionResult, error) {
	encodedTx := cadenceByteArrayString(rawTx)

	result, err := transactions.SendTransaction(
		sendCode,
		[]string{encodedTx},
		"",
		flow,
		state,
		transactions.Flags{
			Signer: deployFlags.Signer,
		},
	)
	if err != nil {
		return nil, err
	}

	txResult := result.(*transactions.TransactionResult)
	if txResult.Result.Error != nil {
		return nil, txResult.Result.Error
	}

	return txResult.Result, nil
}
