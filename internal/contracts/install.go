package contracts

import (
	"context"
	"fmt"

	"github.com/onflow/flow-cli/flowkit/config"

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
		if err := processDependency(flow, logger, dependency); err != nil {
			fmt.Println("Error:", err)
			return nil, err
		}
	}
	return nil, nil
}

func processDependency(flow flowkit.Services, logger output.Logger, dependency config.Dependency) error {
	depAddress := flowsdk.HexToAddress(dependency.RemoteSource.Address.String())
	return fetchDependencies(flow, logger, depAddress, dependency.RemoteSource.ContractName)
}

func handleFoundContract(contractAddr, contractName, contractData string) error {
	if !contractFileExists(contractAddr, contractName) {
		if err := createContractFile(contractAddr, contractName, contractData); err != nil {
			return fmt.Errorf("failed to create contract file: %v", err)
		}
	}

	return nil
}

func fetchDependencies(flow flowkit.Services, logger output.Logger, address flowsdk.Address, contractName string) error {
	logger.Info(fmt.Sprintf("Fetching dependencies for %s at %s", contractName, address))
	account, err := flow.GetAccount(context.Background(), address)
	if err != nil {
		return fmt.Errorf("failed to get account: %v", err)
	}
	if account == nil {
		return fmt.Errorf("account is nil for address: %s", address)
	}

	if account.Contracts == nil {
		return fmt.Errorf("contracts are nil for account: %s", address)
	}

	for _, contract := range account.Contracts {
		parsedProgram, err := parser.ParseProgram(nil, contract, parser.Config{})
		if err != nil {
			return fmt.Errorf("failed to parse program: %v", err)
		}

		if parsedProgram == nil {
			return fmt.Errorf("parsed program is nil")
		}

		var parsedContractName string

		if contractDeclaration := parsedProgram.SoleContractDeclaration(); contractDeclaration != nil {
			parsedContractName = contractDeclaration.Identifier.String()
		} else if contractInterfaceDeclaration := parsedProgram.SoleContractInterfaceDeclaration(); contractInterfaceDeclaration != nil {
			parsedContractName = contractInterfaceDeclaration.Identifier.String()
		} else {
			continue
		}

		if parsedContractName == contractName {
			if err := handleFoundContract(address.String(), parsedContractName, string(contract)); err != nil {
				return fmt.Errorf("failed to handle found contract: %v", err)
			}

			for _, importDeclaration := range parsedProgram.ImportDeclarations() {
				importName := importDeclaration.Identifiers[0].String()
				importAddress := flowsdk.HexToAddress(importDeclaration.Location.String())

				err := fetchDependencies(flow, logger, importAddress, importName)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
