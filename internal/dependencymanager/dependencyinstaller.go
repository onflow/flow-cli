/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package dependencymanager

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"sync"

	"github.com/onflow/flow-cli/internal/util"

	"github.com/onflow/flow-cli/flowkit/gateway"

	"github.com/onflow/flow-cli/flowkit/project"

	flowsdk "github.com/onflow/flow-go-sdk"

	"github.com/onflow/flow-cli/flowkit/config"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/output"
)

type DependencyInstaller struct {
	Gateways map[string]gateway.Gateway
	Logger   output.Logger
	State    *flowkit.State
	Mutex    sync.Mutex
}

func NewDependencyInstaller(logger output.Logger, state *flowkit.State) *DependencyInstaller {
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

	return &DependencyInstaller{
		Gateways: gateways,
		Logger:   logger,
		State:    state,
	}
}

func (ci *DependencyInstaller) install() error {
	for _, dependency := range *ci.State.Dependencies() {
		if err := ci.processDependency(dependency); err != nil {
			ci.Logger.Error(fmt.Sprintf("Error processing dependency: %v", err))
			return err
		}
	}
	return nil
}

func (ci *DependencyInstaller) add(depRemoteSource, customName string) error {
	depNetwork, depAddress, depContractName, err := config.ParseRemoteSourceString(depRemoteSource)
	if err != nil {
		return fmt.Errorf("error parsing remote source: %w", err)
	}

	name := depContractName

	if customName != "" {
		name = customName
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

func (ci *DependencyInstaller) processDependency(dependency config.Dependency) error {
	depAddress := flowsdk.HexToAddress(dependency.RemoteSource.Address.String())
	return ci.fetchDependencies(dependency.RemoteSource.NetworkName, depAddress, dependency.Name, dependency.RemoteSource.ContractName)
}

func (ci *DependencyInstaller) fetchDependencies(networkName string, address flowsdk.Address, assignedName, contractName string) error {
	account, err := ci.Gateways[networkName].GetAccount(address)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}
	if account == nil {
		return fmt.Errorf("account is nil for address: %s", address)
	}

	if account.Contracts == nil {
		return fmt.Errorf("contracts are nil for account: %s", address)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(account.Contracts))

	for _, contract := range account.Contracts {

		program, err := project.NewProgram(contract, nil, "")
		if err != nil {
			return fmt.Errorf("failed to parse program: %w", err)
		}

		parsedContractName, err := program.Name()
		if err != nil {
			return fmt.Errorf("failed to parse contract name: %w", err)
		}

		if parsedContractName == contractName {

			if err := ci.handleFoundContract(networkName, address.String(), assignedName, parsedContractName, program); err != nil {
				return fmt.Errorf("failed to handle found contract: %w", err)
			}

			if program.HasAddressImports() {
				imports := program.AddressImportDeclarations()
				for _, imp := range imports {
					wg.Add(1)
					go func(importAddress flowsdk.Address, contractName string) {
						defer wg.Done()
						err := ci.fetchDependencies("testnet", importAddress, contractName, contractName)
						if err != nil {
							errCh <- err
						}
					}(flowsdk.HexToAddress(imp.Location.String()), imp.Identifiers[0].String())
				}
			}
		}
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func (ci *DependencyInstaller) handleFoundContract(networkName, contractAddr, assignedName, contractName string, program *project.Program) error {
	ci.Mutex.Lock()
	defer ci.Mutex.Unlock()

	hash := sha256.New()
	hash.Write(program.DevelopmentCode())
	originalContractDataHash := hex.EncodeToString(hash.Sum(nil))

	program.ConvertImports()
	contractData := string(program.DevelopmentCode())

	dependency := ci.State.Dependencies().ByName(assignedName)

	// If a dependency by this name already exists and its remote source network or address does not match, then give option to stop or continue
	if dependency != nil && (dependency.RemoteSource.NetworkName != networkName || dependency.RemoteSource.Address.String() != contractAddr) {
		ci.Logger.Info(fmt.Sprintf("🚫 A dependency named %s already exists with a different remote source. Please fix the conflict and retry.", assignedName))
		os.Exit(0)
		return nil
	}

	// Check if remote source version is different from local version
	// If it is, ask if they want to update
	// If no hash, ignore
	if dependency != nil && dependency.Version != "" && dependency.Version != originalContractDataHash {
		msg := fmt.Sprintf("The latest version of %s is different from the one you have locally. Do you want to update it?", contractName)
		if !util.GenericBoolPrompt(msg) {
			return nil
		}
	}

	if !contractFileExists(contractAddr, contractName) {
		if err := createContractFile(contractAddr, contractName, contractData); err != nil {
			return fmt.Errorf("failed to create contract file: %w", err)
		}

		ci.Logger.Info(fmt.Sprintf("Dependency Manager: %s from %s on %s installed", contractName, contractAddr, networkName))
	}

	err := ci.updateState(networkName, contractAddr, assignedName, contractName, originalContractDataHash)
	if err != nil {
		ci.Logger.Error(fmt.Sprintf("Error updating state: %v", err))
		return err
	}

	return nil
}

func (ci *DependencyInstaller) updateState(networkName, contractAddress, assignedName, contractName, contractHash string) error {
	dep := config.Dependency{
		Name: assignedName,
		RemoteSource: config.RemoteSource{
			NetworkName:  networkName,
			Address:      flowsdk.HexToAddress(contractAddress),
			ContractName: contractName,
		},
		Version: contractHash,
	}

	isNewDep := ci.State.Dependencies().ByName(dep.Name) == nil

	ci.State.Dependencies().AddOrUpdate(dep)
	ci.State.Contracts().AddDependencyAsContract(dep, networkName)
	err := ci.State.SaveDefault()
	if err != nil {
		return err
	}

	if isNewDep {
		ci.Logger.Info(fmt.Sprintf("Dependency Manager: %s added to flow.json", dep.Name))
	}

	return nil
}
