package contracts

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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

func contractFileExists(address, contractName string) bool {
	path := filepath.Join("imports", address, contractName)
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func createContractFile(address, contractName, data string) error {
	path := filepath.Join("imports", address, contractName)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(data), 0644)
}

func handleFoundContract(contractAddr, contractName, contractData string) error {
	if !contractFileExists(contractAddr, contractName) {
		fmt.Println("Create file!")
		if err := createContractFile(contractAddr, contractName, contractData); err != nil {
			return fmt.Errorf("failed to create contract file: %v", err)
		}
		fmt.Println("File created!")
	}

	return nil
}

func fetchDependencies(flow flowkit.Services, address flowsdk.Address, contractName string) error {
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

		fmt.Println("Parsed Contract Name: ", parsedContractName)

		if parsedContractName == contractName {
			fmt.Println("Contract found!")
			if err := handleFoundContract(address.String(), parsedContractName, string(contract)); err != nil {
				return fmt.Errorf("failed to handle found contract: %v", err)
			}

			for _, importDeclaration := range parsedProgram.ImportDeclarations() {
				importName := importDeclaration.Identifiers[0].String()
				importAddress := flowsdk.HexToAddress(importDeclaration.Location.String())

				err := fetchDependencies(flow, importAddress, importName)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
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

		err := fetchDependencies(flow, depAddress, dependency.RemoteSource.ContractName)
		if err != nil {
			fmt.Println("Error:", err)
			return nil, err
		}
	}

	return nil, err
}
