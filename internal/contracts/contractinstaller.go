package contracts

import (
	"fmt"

	"github.com/onflow/flow-cli/flowkit/gateway"

	"github.com/onflow/flow-cli/flowkit/project"

	"github.com/onflow/flow-cli/flowkit/config"
	flowsdk "github.com/onflow/flow-go-sdk"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/output"
)

type ContractInstaller struct {
	Gateways map[string]gateway.Gateway
	Logger   output.Logger
	State    *flowkit.State
}

func NewContractInstaller(logger output.Logger, state *flowkit.State) *ContractInstaller {
	emulatorGateway, err := gateway.NewGrpcGateway(config.EmulatorNetwork)
	if err != nil {
		logger.Error(fmt.Sprintf("Error creating emulator gateway: %v", err))
	}

	testnetGateway, err := gateway.NewGrpcGateway(config.TestnetNetwork)
	if err != nil {
		logger.Error(fmt.Sprintf("Error creating testnet gateway: %v", err))
	}

	mainnetGateway, err := gateway.NewGrpcGateway(config.MainnetNetwork)
	if err != nil {
		logger.Error(fmt.Sprintf("Error creating mainnet gateway: %v", err))
	}

	gateways := map[string]gateway.Gateway{
		config.EmulatorNetwork.Name: emulatorGateway,
		config.TestnetNetwork.Name:  testnetGateway,
		config.MainnetNetwork.Name:  mainnetGateway,
	}

	return &ContractInstaller{
		Gateways: gateways,
		Logger:   logger,
		State:    state,
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

func (ci *ContractInstaller) add(depRemoteSource, customName string) error {
	depNetwork, depAddress, depContractName, err := config.ParseRemoteSourceString(depRemoteSource)
	if err != nil {
		return fmt.Errorf("error parsing remote source: %w", err)
	}

	var name string

	if customName != "" {
		fmt.Printf("Using custom name: %s\n", customName)
		name = customName
	} else {
		name = depContractName
	}

	dep := config.Dependency{
		Name: name,
		RemoteSource: config.RemoteSource{
			NetworkName:  depNetwork,
			Address:      flowsdk.HexToAddress(depAddress),
			ContractName: depContractName,
		},
	}

	if err := ci.processDependency(dep); err != nil {
		return fmt.Errorf("error processing dependency: %w", err)
	}

	return nil
}

func (ci *ContractInstaller) processDependency(dependency config.Dependency) error {
	depAddress := flowsdk.HexToAddress(dependency.RemoteSource.Address.String())
	return ci.fetchDependencies(dependency.RemoteSource.NetworkName, depAddress, dependency.Name, dependency.RemoteSource.ContractName)
}

func (ci *ContractInstaller) fetchDependencies(networkName string, address flowsdk.Address, assignedName, contractName string) error {
	ci.Logger.Info(fmt.Sprintf("Fetching dependencies for %s at %s", contractName, address))
	account, err := ci.Gateways[networkName].GetAccount(address)
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

			if err := ci.handleFoundContract(networkName, address.String(), assignedName, parsedContractName, string(program.DevelopmentCode())); err != nil {
				return fmt.Errorf("failed to handle found contract: %v", err)
			}

			if program.HasAddressImports() {
				imports := program.AddressImportDeclarations()
				for _, imp := range imports {
					importAddress := flowsdk.HexToAddress(imp.Location.String())
					contractName := imp.Identifiers[0].String()
					err := ci.fetchDependencies("testnet", importAddress, contractName, contractName)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (ci *ContractInstaller) handleFoundContract(networkName, contractAddr, assignedName, contractName, contractData string) error {
	if !contractFileExists(contractAddr, contractName) {
		if err := createContractFile(contractAddr, contractName, contractData); err != nil {
			return fmt.Errorf("failed to create contract file: %v", err)
		}
	}

	err := ci.updateState(networkName, contractAddr, assignedName, contractName)
	if err != nil {
		ci.Logger.Error(fmt.Sprintf("Error updating state: %v", err))
		return err
	}

	return nil
}

func (ci *ContractInstaller) updateState(networkName, contractAddress, assignedName, contractName string) error {
	dep := config.Dependency{
		Name: assignedName,
		RemoteSource: config.RemoteSource{
			NetworkName:  networkName,
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
