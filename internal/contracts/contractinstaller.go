package contracts

import (
	"context"
	"fmt"

	"github.com/onflow/flow-cli/flowkit/project"

	"github.com/onflow/flow-cli/flowkit/config"
	flowsdk "github.com/onflow/flow-go-sdk"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/output"
)

type ContractInstaller struct {
	FlowService flowkit.Services
	Logger      output.Logger
	State       *flowkit.State
}

func NewContractInstaller(flow flowkit.Services, logger output.Logger, state *flowkit.State) *ContractInstaller {
	return &ContractInstaller{
		FlowService: flow,
		Logger:      logger,
		State:       state,
	}
}

func (ci *ContractInstaller) install() error {
	for _, dependency := range *ci.State.Dependencies() {
		if err := ci.processDependency(dependency); err != nil {
			ci.Logger.Error(fmt.Sprintf("Error processing dependency: %v", err))
			return err
		}
	}
	return nil
}

func (ci *ContractInstaller) processDependency(dependency config.Dependency) error {
	depAddress := flowsdk.HexToAddress(dependency.RemoteSource.Address.String())
	return ci.fetchDependencies(depAddress, dependency.RemoteSource.ContractName)
}

func (ci *ContractInstaller) fetchDependencies(address flowsdk.Address, contractName string) error {
	ci.Logger.Info(fmt.Sprintf("Fetching dependencies for %s at %s", contractName, address))
	account, err := ci.FlowService.GetAccount(context.Background(), address)
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

		program, err := project.NewProgram(contract, nil, "")
		if err != nil {
			return fmt.Errorf("failed to parse program: %v", err)
		}

		parsedContractName, err := program.Name()
		if err != nil {
			return fmt.Errorf("failed to parse contract name: %v", err)
		}

		if parsedContractName == contractName {
			program.ConvertImports()

			if err := ci.handleFoundContract(address.String(), parsedContractName, string(program.DevelopmentCode())); err != nil {
				return fmt.Errorf("failed to handle found contract: %v", err)
			}

			if program.HasImports() {
				imports := program.Imports()
				for _, imp := range imports {
					importAddress := flowsdk.HexToAddress(imp)
					err := ci.fetchDependencies(importAddress, imp)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (ci *ContractInstaller) handleFoundContract(contractAddr, contractName, contractData string) error {
	if !contractFileExists(contractAddr, contractName) {
		if err := createContractFile(contractAddr, contractName, contractData); err != nil {
			return fmt.Errorf("failed to create contract file: %v", err)
		}
	}

	err := ci.updateState(contractAddr, contractName)
	if err != nil {
		ci.Logger.Error(fmt.Sprintf("Error updating state: %v", err))
		return err
	}

	return nil
}

func (ci *ContractInstaller) updateState(contractAddress, contractName string) error {
	dep := config.Dependency{
		Name: contractName,
		RemoteSource: config.RemoteSource{
			NetworkName:  ci.FlowService.Network().Name,
			Address:      flowsdk.HexToAddress(contractAddress),
			ContractName: contractName,
		},
	}
	ci.State.Dependencies().AddOrUpdate(dep)
	err := ci.State.SaveDefault()
	if err != nil {
		return err
	}

	return nil
}
