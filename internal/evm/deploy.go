package evm

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/onflow/flow-go/fvm/evm/types"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/transactions"
)

//go:embed deploy.cdc
var deployCode []byte

type flagsDeploy struct {
	Signer string `default:"" flag:"signer" info:"Account name from configuration used to sign the transaction as proposer, payer and suthorizer"`
}

var deployFlags = flagsDeploy{}

var deployCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "deploy <evm bytecode file>",
		Short:   "Deploy compiled bytecode to the Flow EVM",
		Args:    cobra.ExactArgs(1),
		Example: "flow evm deploy ./hello",
	},
	Flags: &deployFlags,
	RunS:  deploy,
}

// todo only for demo, super hacky now

func deploy(
	args []string,
	_ command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	filename := args[0]

	// read file containing hex-encoded evm bytecode
	evmCode, err := state.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error loading transaction file: %w", err)
	}

	result, err := transactions.SendTransaction(
		deployCode,
		[]string{string(evmCode)},
		filename,
		flow,
		state,
		transactions.Flags{
			Signer: deployFlags.Signer,
		},
	)
	if err != nil {
		return nil, err
	}

	printDeployResult(result)
	return nil, nil
}

func getDeployedAddress(event flowkit.Event) string {
	addr, ok := event.Values["deployedContractAddress"]
	if !ok || addr.String() == "\"0000000000000000000000000000000000000000\"" {
		return ""
	}

	return strings.ReplaceAll(addr.String(), "\"", "")
}

func getGasConsumd(event flowkit.Event) uint64 {
	gas, ok := event.Values["gasConsumed"]
	if !ok {
		return 0
	}
	return gas.ToGoValue().(uint64)
}

func getLatestHeight(event flowkit.Event) uint64 {
	h, ok := event.Values["height"]
	if !ok {
		return 0
	}
	return h.ToGoValue().(uint64)
}

func printDeployResult(result command.Result) {
	fmt.Printf("\n🔥🔥🔥🔥🔥🔥🔥 EVM Contract Deployment Summary 🔥🔥🔥🔥🔥🔥🔥")
	fmt.Printf("\n-------------------------------------------------------------\n\n")

	txResult := result.(*transactions.TransactionResult)
	events := flowkit.EventsFromTransaction(txResult.Result)
	var (
		gasConsumed     uint64
		deployedAddress string
		latestHeight    uint64
	)

	for _, e := range events {
		if e.Type == fmt.Sprintf("flow.%s", types.EventTypeTransactionExecuted) {
			if address := getDeployedAddress(e); address != "" {
				deployedAddress = address
			}
		}
		gasConsumed += getGasConsumd(e)
		latestHeight = getLatestHeight(e)
	}

	fmt.Println("Contract Address:      ", deployedAddress)
	fmt.Println("Gas Consumed:          ", gasConsumed)
	fmt.Println("Gas Price:              TBD")
	fmt.Println("Latest Block Height:   ", latestHeight)

	fmt.Printf("\n\n\nFlow Transaction Details\n\n")
	fmt.Println(result)
}
