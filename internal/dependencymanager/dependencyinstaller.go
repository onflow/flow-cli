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
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/psiemens/sconfig"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-go/fvm/systemcontracts"
	flowGo "github.com/onflow/flow-go/model/flow"

	"github.com/onflow/flow-cli/internal/util"

	"github.com/onflow/flowkit/v2/gateway"

	"github.com/onflow/flowkit/v2/project"

	flowsdk "github.com/onflow/flow-go-sdk"

	"github.com/onflow/flowkit/v2/config"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"
)

type categorizedLogs struct {
	fileSystemActions []string
	stateUpdates      []string
}

func (cl *categorizedLogs) LogAll(logger output.Logger) {
	logger.Info(util.MessageWithEmojiPrefix("📝", "Dependency Manager Actions Summary"))
	logger.Info("") // Add a line break after the section

	if len(cl.fileSystemActions) > 0 {
		logger.Info("🗃️  File System Actions:")
		for _, msg := range cl.fileSystemActions {
			logger.Info(util.MessageWithEmojiPrefix("✅", msg))
		}
		logger.Info("") // Add a line break after the section
	}

	if len(cl.stateUpdates) > 0 {
		logger.Info("💾 State Updates:")
		for _, msg := range cl.stateUpdates {
			logger.Info(util.MessageWithEmojiPrefix("✅", msg))
		}
		logger.Info("") // Add a line break after the section
	}

	if len(cl.fileSystemActions) == 0 && len(cl.stateUpdates) == 0 {
		logger.Info(util.MessageWithEmojiPrefix("👍", "Zero changes were made. Everything looks good."))
	}
}

type dependencyManagerFlagsCollection struct {
	skipDeployments bool `default:"false" flag:"skip-deployments" info:"Skip adding the dependency to deployments"`
	skipAlias       bool `default:"false" flag:"skip-alias" info:"Skip prompting for an alias"`
}

func (f *dependencyManagerFlagsCollection) AddToCommand(cmd *cobra.Command) {
	err := sconfig.New(f).
		FromEnvironment(util.EnvPrefix).
		BindFlags(cmd.Flags()).
		Parse()

	if err != nil {
		panic(err)
	}
}

type DependencyInstaller struct {
	Gateways        map[string]gateway.Gateway
	Logger          output.Logger
	State           *flowkit.State
	SkipDeployments bool
	SkipAlias       bool
	logs            categorizedLogs
}

// NewDependencyInstaller creates a new instance of DependencyInstaller
func NewDependencyInstaller(logger output.Logger, state *flowkit.State, flags dependencyManagerFlagsCollection) (*DependencyInstaller, error) {
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

	previewnetGateway, err := gateway.NewGrpcGateway(config.PreviewnetNetwork)
	if err != nil {
		return nil, fmt.Errorf("error creating previewnet gateway: %v", err)
	}

	gateways := map[string]gateway.Gateway{
		config.EmulatorNetwork.Name:   emulatorGateway,
		config.TestnetNetwork.Name:    testnetGateway,
		config.MainnetNetwork.Name:    mainnetGateway,
		config.PreviewnetNetwork.Name: previewnetGateway,
	}

	return &DependencyInstaller{
		Gateways:        gateways,
		Logger:          logger,
		State:           state,
		SkipDeployments: flags.skipDeployments,
		SkipAlias:       flags.skipAlias,
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

	err := di.State.SaveDefault()
	if err != nil {
		return fmt.Errorf("error saving state: %w", err)
	}

	di.logs.LogAll(di.Logger)

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

	err = di.State.SaveDefault()
	if err != nil {
		return fmt.Errorf("error saving state: %w", err)
	}

	di.logs.LogAll(di.Logger)

	return nil
}

func (di *DependencyInstaller) processDependency(dependency config.Dependency) error {
	depAddress := flowsdk.HexToAddress(dependency.Source.Address.String())
	return di.fetchDependencies(dependency.Source.NetworkName, depAddress, dependency.Name, dependency.Source.ContractName)
}

func (di *DependencyInstaller) fetchDependencies(networkName string, address flowsdk.Address, assignedName, contractName string) error {
	ctx := context.Background()
	account, err := di.Gateways[networkName].GetAccount(ctx, address)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}
	if account == nil {
		return fmt.Errorf("account is nil for address: %s", address)
	}

	if account.Contracts == nil {
		return fmt.Errorf("contracts are nil for account: %s", address)
	}

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
					contractName := imp.Identifiers[0].String()
					err := di.fetchDependencies(networkName, flowsdk.HexToAddress(imp.Location.String()), contractName, contractName)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	if !found {
		errMsg := fmt.Sprintf("contract %s not found for account %s on network %s", contractName, address, networkName)
		di.Logger.Error(errMsg)
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
	if !di.contractFileExists(contractAddr, contractName) {
		if err := di.createContractFile(contractAddr, contractName, contractData); err != nil {
			return fmt.Errorf("failed to create contract file: %w", err)
		}

		di.logs.fileSystemActions = append(di.logs.fileSystemActions, fmt.Sprintf("%s from %s on %s installed", contractName, contractAddr, networkName))
	}

	return nil
}

func isCoreContract(contractName string) bool {
	sc := systemcontracts.SystemContractsForChain(flowGo.Emulator)

	for _, coreContract := range sc.All() {
		if coreContract.Name == contractName {
			return true
		}
	}
	return false
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
		di.Logger.Info(fmt.Sprintf("%s A dependency named %s already exists with a different remote source. Please fix the conflict and retry.", util.PrintEmoji("🚫"), assignedName))
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

	err = di.updateDependencyState(networkName, contractAddr, assignedName, contractName, originalContractDataHash)
	if err != nil {
		di.Logger.Error(fmt.Sprintf("Error updating state: %v", err))
		return err
	}

	// If the contract is not a core contract and the user does not want to skip deployments, then prompt for a deployment
	if !di.SkipDeployments && !isCoreContract(contractName) {
		err = di.updateDependencyDeployment(contractName)
		if err != nil {
			di.Logger.Error(fmt.Sprintf("Error updating deployment: %v", err))
			return err
		}

		di.logs.stateUpdates = append(di.logs.stateUpdates, fmt.Sprintf("%s added to emulator deployments", contractName))
	}

	// If the contract is not a core contract and the user does not want to skip aliasing, then prompt for an alias
	if !di.SkipAlias && !isCoreContract(contractName) {
		err = di.updateDependencyAlias(contractName, networkName)
		if err != nil {
			di.Logger.Error(fmt.Sprintf("Error updating alias: %v", err))
			return err
		}

		di.logs.stateUpdates = append(di.logs.stateUpdates, fmt.Sprintf("Alias added for %s on %s", contractName, networkName))
	}

	return nil
}

func (di *DependencyInstaller) updateDependencyDeployment(contractName string) error {
	// Add to deployments
	// If a deployment already exists for that account, contract, and network, then ignore
	raw := util.AddContractToDeploymentPrompt("emulator", *di.State.Accounts(), contractName)

	if raw != nil {
		deployment := di.State.Deployments().ByAccountAndNetwork(raw.Account, raw.Network)
		if deployment == nil {
			di.State.Deployments().AddOrUpdate(config.Deployment{
				Network: raw.Network,
				Account: raw.Account,
			})
			deployment = di.State.Deployments().ByAccountAndNetwork(raw.Account, raw.Network)
		}

		for _, c := range raw.Contracts {
			deployment.AddContract(config.ContractDeployment{Name: c})
		}
	}

	return nil
}

func (di *DependencyInstaller) updateDependencyAlias(contractName, aliasNetwork string) error {
	var missingNetwork string

	if aliasNetwork == config.TestnetNetwork.Name {
		missingNetwork = config.MainnetNetwork.Name
	} else {
		missingNetwork = config.TestnetNetwork.Name
	}

	label := fmt.Sprintf("Enter an alias address for %s on %s if you have one, otherwise leave blank", contractName, missingNetwork)
	raw := util.AddressPromptOrEmpty(label, "Invalid alias address")

	if raw != "" {
		contract, err := di.State.Contracts().ByName(contractName)
		if err != nil {
			return err
		}

		contract.Aliases.Add(missingNetwork, flowsdk.HexToAddress(raw))
	}

	return nil
}

func (di *DependencyInstaller) updateDependencyState(networkName, contractAddress, assignedName, contractName, contractHash string) error {
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

	if isNewDep {
		di.logs.stateUpdates = append(di.logs.stateUpdates, fmt.Sprintf("%s added to flow.json", dep.Name))
	}

	return nil
}
