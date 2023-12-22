package evm

import (
	_ "embed"
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/transactions"
)

//go:embed fund.cdc
var fundCode []byte

type flagsFund struct {
	Signer string `default:"" flag:"signer" info:"Account name from configuration used to sign the transaction as proposer, payer and suthorizer"`
}

var fundFlags = flagsFund{}

var fundCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "fund <EVM address> <amount>",
		Short:   "Fund an EVM address with the provided funding amount",
		Args:    cobra.ExactArgs(2),
		Example: "flow evm fund 522b3294e6d06aa25ad0f1b8891242e335d3b459 1.0",
	},
	Flags: &fundFlags,
	RunS:  fund,
}

// todo only for demo, super hacky now

func fund(
	args []string,
	g command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	address := args[0]
	amount := args[1]

	a, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}

	addressBytes := cadenceByteArrayString(a)

	res, err := transactions.SendTransaction(
		fundCode,
		[]string{"", addressBytes, amount},
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

	txRes := res.(*transactions.TransactionResult)
	if txRes.Result.Error != nil {
		return nil, txRes.Result.Error
	}

	fmt.Printf("\n🔥🔥🔥🔥🔥🔥🔥 EVM Address Funded 🔥🔥🔥🔥🔥🔥🔥")
	fmt.Printf("\n-------------------------------------------------------------\n\n")

	fmt.Println("Address:       ", address)
	fmt.Println("Funded with:   ", amount)

	return nil, nil
}
