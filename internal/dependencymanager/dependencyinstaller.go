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
	"path/filepath"
	"sync"

	"github.com/onflow/flow-cli/internal/util"

	"github.com/onflow/flowkit/gateway"

	"github.com/onflow/flowkit/project"

	flowsdk "github.com/onflow/flow-go-sdk"

	"github.com/onflow/flowkit/config"

	"github.com/onflow/flowkit"
	"github.com/onflow/flowkit/output"
)

type DependencyInstaller struct {
	Gateways map[string]gateway.Gateway
	Logger   output.Logger
	State    *flowkit.State
	Mutex    sync.Mutex
}

// NewDependencyInstaller creates a new instance of DependencyInstaller
func NewDependencyInstaller(logger output.Logger, state *flowkit.State) (*DependencyInstaller, error) {
	emulatorGateway, err := gateway.NewGrpcGateway(config.EmulatorNetwork)
	if err != nil {
		return nil, fmt.Errorf("error creating emulator gateway: %v", err)
	}

	testnetGateway, err := gateway.NewGrpcGateway(config.TestnetNetwork)
	if err != nil {
		return nil, fmt.Errorf("error creating testnet gateway: %v", err)
	}

	mainnetGateway, err := gateway.NewGrpcGateway(config.MainnetNetwork)
	if err != nil {
		return nil, fmt.Errorf("error creating mainnet gateway: %v", err)
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
	}, nil
}

// Install processes all the dependencies in the state and installs them and any dependencies they have
func (di *DependencyInstaller) Install() error {
	for _, dependency := range *di.State.Dependencies() {
		if err := di.processDependency(dependency); err != nil {
			di.Logger.Error(fmt.Sprintf("Error processing dependency: %v", err))
			return err
		}
	}
	return nil
}

// Add processes a single dependency and installs it and any dependencies it has, as well as adding it to the state
func (di *DependencyInstaller) Add(depSource, customName string) error {
	depNetwork, depAddress, depContractName, err := config.ParseSourceString(depSource)
	if err != nil {
		return fmt.Errorf("error parsing source: %w", err)
	}

	name := depContractName

	if customName != "" {
		name = customName
	}

	dep := config.Dependency{
		Name: name,
		Source: config.Source{
			NetworkName:  depNetwork,
			Address:      flowsdk.HexToAddress(depAddress),
			ContractName: depContractName,
		},
	}

	if err := di.processDependency(dep); err != nil {
		return fmt.Errorf("error processing dependency: %w", err)
	}

	return nil
}

func (di *DependencyInstaller) processDependency(dependency config.Dependency) error {
	depAddress := flowsdk.HexToAddress(dependency.Source.Address.String())
	return di.fetchDependencies(dependency.Source.NetworkName, depAddress, dependency.Name, dependency.Source.ContractName)
}

func (di *DependencyInstaller) fetchDependencies(networkName string, address flowsdk.Address, assignedName, contractName string) error {
	account, err := di.Gateways[networkName].GetAccount(address)
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

	// Create a max number of goroutines so that we don't rate limit the access node
	maxGoroutines := 5
	semaphore := make(chan struct{}, maxGoroutines)

	found := false

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
			found = true

			if err := di.handleFoundContract(networkName, address.String(), assignedName, parsedContractName, program); err != nil {
				return fmt.Errorf("failed to handle found contract: %w", err)
			}

			if program.HasAddressImports() {
				imports := program.AddressImportDeclarations()
				for _, imp := range imports {
					wg.Add(1)
					go func(importAddress flowsdk.Address, contractName string) {
						semaphore <- struct{}{}
						defer func() {
							<-semaphore
							wg.Done()
						}()
						err := di.fetchDependencies(networkName, importAddress, contractName, contractName)
						if err != nil {
							errCh <- err
						}
					}(flowsdk.HexToAddress(imp.Location.String()), imp.Identifiers[0].String())
				}
			}
		}
	}

	if !found {
		errMsg := fmt.Sprintf("contract %s not found for account %s on network %s", contractName, address, networkName)
		di.Logger.Error(errMsg)
	}

	wg.Wait()
	close(errCh)
	close(semaphore)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func (di *DependencyInstaller) contractFileExists(address, contractName string) bool {
	fileName := fmt.Sprintf("%s.cdc", contractName)
	path := filepath.Join("imports", address, fileName)

	_, err := di.State.ReaderWriter().Stat(path)

	return err == nil
}

func (di *DependencyInstaller) createContractFile(address, contractName, data string) error {
	fileName := fmt.Sprintf("%s.cdc", contractName)
	path := filepath.Join("imports", address, fileName)
	dir := filepath.Dir(path)

	if err := di.State.ReaderWriter().MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directories: %w", err)
	}

	if err := di.State.ReaderWriter().WriteFile(path, []byte(data), 0644); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

func (di *DependencyInstaller) handleFileSystem(contractAddr, contractName, contractData, networkName string) error {
	di.Mutex.Lock()
	defer di.Mutex.Unlock()

	if !di.contractFileExists(contractAddr, contractName) {
		if err := di.createContractFile(contractAddr, contractName, contractData); err != nil {
			return fmt.Errorf("failed to create contract file: %w", err)
		}

		di.Logger.Info(fmt.Sprintf("Dependency Manager: %s from %s on %s installed", contractName, contractAddr, networkName))
	}

	return nil
}

func (di *DependencyInstaller) handleFoundContract(networkName, contractAddr, assignedName, contractName string, program *project.Program) error {
	hash := sha256.New()
	hash.Write(program.CodeWithUnprocessedImports())
	originalContractDataHash := hex.EncodeToString(hash.Sum(nil))

	program.ConvertAddressImports()
	contractData := string(program.CodeWithUnprocessedImports())

	dependency := di.State.Dependencies().ByName(assignedName)

	// If a dependency by this name already exists and its remote source network or address does not match, then give option to stop or continue
	if dependency != nil && (dependency.Source.NetworkName != networkName || dependency.Source.Address.String() != contractAddr) {
		di.Logger.Info(fmt.Sprintf("🚫 A dependency named %s already exists with a different remote source. Please fix the conflict and retry.", assignedName))
		os.Exit(0)
		return nil
	}

	// Check if remote source version is different from local version
	// If it is, ask if they want to update
	// If no hash, ignore
	if dependency != nil && dependency.Hash != "" && dependency.Hash != originalContractDataHash {
		msg := fmt.Sprintf("The latest version of %s is different from the one you have locally. Do you want to update it?", contractName)
		if !util.GenericBoolPrompt(msg) {
			return nil
		}
	}

	err := di.handleFileSystem(contractAddr, contractName, contractData, networkName)
	if err != nil {
		return fmt.Errorf("error handling file system: %w", err)
	}

	err = di.updateState(networkName, contractAddr, assignedName, contractName, originalContractDataHash)
	if err != nil {
		di.Logger.Error(fmt.Sprintf("Error updating state: %v", err))
		return err
	}

	return nil
}

func (di *DependencyInstaller) updateState(networkName, contractAddress, assignedName, contractName, contractHash string) error {
	dep := config.Dependency{
		Name: assignedName,
		Source: config.Source{
			NetworkName:  networkName,
			Address:      flowsdk.HexToAddress(contractAddress),
			ContractName: contractName,
		},
		Hash: contractHash,
	}

	isNewDep := di.State.Dependencies().ByName(dep.Name) == nil

	di.State.Dependencies().AddOrUpdate(dep)
	di.State.Contracts().AddDependencyAsContract(dep, networkName)
	err := di.State.SaveDefault()
	if err != nil {
		return err
	}

	if isNewDep {
		di.Logger.Info(fmt.Sprintf("Dependency Manager: %s added to flow.json", dep.Name))
	}

	return nil
}
