package contracts

import (
	"context"
	"fmt"

	"github.com/onflow/cadence/runtime/parser"

	flowsdk "github.com/onflow/flow-go-sdk"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/spf13/cobra"
)

type flagsCollection struct{}

var installFlags = flagsCollection{}

var installCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:   "install",
		Short: "Install contract and dependencies.",
	},
	Flags: &installFlags,
	RunS:  install,
}

func install(
	_ []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {

	for _, dependency := range *state.Dependencies() {
		fmt.Println("dependency: ", dependency.Name)
		fmt.Println("dependency remote source address: ", dependency.RemoteSource.Address.String())
		fmt.Println("dependency remote source contract name: ", dependency.RemoteSource.ContractName)

		depAddress := flowsdk.HexToAddress(dependency.RemoteSource.Address.String())
		logger.Info(fmt.Sprintf("Fetching contract and dependencies for %s", depAddress))
		account, err := flow.GetAccount(context.Background(), depAddress)
		if err != nil {
			return nil, err
		}

		for _, contract := range account.Contracts {
			parsedProgram, err := parser.ParseProgram(nil, contract, parser.Config{})
			if err != nil {
				return nil, err
			}

			fmt.Println("Contract Name: ", parsedProgram.SoleContractDeclaration().Identifier)
			fmt.Println("Imports: ", parsedProgram.ImportDeclarations())
		}
	}

	return nil, err
}
