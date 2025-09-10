/*
 * Flow CLI
 *
 * Copyright Flow Foundation
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

	"github.com/onflow/flow-cli/internal/prompt"
	"github.com/onflow/flow-cli/internal/util"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-go/fvm/systemcontracts"
	flowGo "github.com/onflow/flow-go/model/flow"

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
	issues            []string
}

func (cl *categorizedLogs) LogAll(logger output.Logger) {
	logger.Info(util.MessageWithEmojiPrefix("📝", "Dependency Manager Actions Summary"))
	logger.Info("") // Add a line break after the section

	if len(cl.fileSystemActions) > 0 {
		logger.Info(util.MessageWithEmojiPrefix("🗃️", "File System Actions:"))
		for _, msg := range cl.fileSystemActions {
			logger.Info(msg)
		}
		logger.Info("") // Add a line break after the section
	}

	if len(cl.stateUpdates) > 0 {
		logger.Info(util.MessageWithEmojiPrefix("💾", "State Updates:"))
		for _, msg := range cl.stateUpdates {
			logger.Info(msg)
		}
		logger.Info("") // Add a line break after the section
	}

	if len(cl.issues) > 0 {
		logger.Info(util.MessageWithEmojiPrefix("⚠️", "Issues:"))
		for _, msg := range cl.issues {
			logger.Info(msg)
		}
		logger.Info("")
	}

	if len(cl.fileSystemActions) == 0 && len(cl.stateUpdates) == 0 {
		logger.Info(util.MessageWithEmojiPrefix("👍", "Zero changes were made. Everything looks good."))
	}
}

type Flags struct {
	skipDeployments bool `default:"false" flag:"skip-deployments" info:"Skip adding the dependency to deployments"`
	skipAlias       bool `default:"false" flag:"skip-alias" info:"Skip prompting for an alias"`
}

func (f *Flags) AddToCommand(cmd *cobra.Command) {
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
	SaveState       bool
	TargetDir       string
	SkipDeployments bool
	SkipAlias       bool
	logs            categorizedLogs
	dependencies    map[string]config.Dependency
	accountAliases  map[string]map[string]flowsdk.Address // network -> account -> alias
}

// NewDependencyInstaller creates a new instance of DependencyInstaller
func NewDependencyInstaller(logger output.Logger, state *flowkit.State, saveState bool, targetDir string, flags Flags) (*DependencyInstaller, error) {
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
		Gateways:        gateways,
		Logger:          logger,
		State:           state,
		SaveState:       saveState,
		TargetDir:       targetDir,
		SkipDeployments: flags.skipDeployments,
		SkipAlias:       flags.skipAlias,
		dependencies:    make(map[string]config.Dependency),
		logs:            categorizedLogs{},
		accountAliases:  make(map[string]map[string]flowsdk.Address),
	}, nil
}

// saveState checks the SaveState flag and saves the state if set to true.
func (di *DependencyInstaller) saveState() error {
	if di.SaveState {
		statePath := filepath.Join(di.TargetDir, "flow.json")
		if err := di.State.Save(statePath); err != nil {
			return fmt.Errorf("error saving state: %w", err)
		}
	}
	return nil
}

// Install processes all the dependencies in the state and installs them and any dependencies they have
func (di *DependencyInstaller) Install() error {
	for _, dependency := range *di.State.Dependencies() {
		if err := di.processDependency(dependency); err != nil {
			di.Logger.Error(fmt.Sprintf("Error processing dependency: %v", err))
			return err
		}
	}

	di.checkForConflictingContracts()

	if err := di.saveState(); err != nil {
		return fmt.Errorf("error saving state: %w", err)
	}

	return nil
}

// AddBySourceString processes a single dependency and installs it and any dependencies it has, as well as adding it to the state
func (di *DependencyInstaller) AddBySourceString(depSource string) error {
	depNetwork, depAddress, depContractName, err := config.ParseSourceString(depSource)
	if err != nil {
		return fmt.Errorf("error parsing source: %w", err)
	}

	dep := config.Dependency{
		Name: depContractName,
		Source: config.Source{
			NetworkName:  depNetwork,
			Address:      flowsdk.HexToAddress(depAddress),
			ContractName: depContractName,
		},
	}

	return di.Add(dep)
}

func (di *DependencyInstaller) AddByCoreContractName(coreContractName string) error {
	var depNetwork, depAddress, depContractName string
	sc := systemcontracts.SystemContractsForChain(flowGo.Mainnet)
	for _, coreContract := range sc.All() {
		if coreContract.Name == coreContractName {
			depAddress = coreContract.Address.String()
			depNetwork = config.MainnetNetwork.Name
			depContractName = coreContractName
			break
		}
	}

	if depAddress == "" {
		return fmt.Errorf("contract %s not found in core contracts", coreContractName)
	}

	dep := config.Dependency{
		Name: depContractName,
		Source: config.Source{
			NetworkName:  depNetwork,
			Address:      flowsdk.HexToAddress(depAddress),
			ContractName: depContractName,
		},
	}

	return di.Add(dep)
}

func (di *DependencyInstaller) AddByDefiContractName(defiContractName string) error {
	defiActionsSection := getDefiActionsSection()
	var targetDep *config.Dependency
	
	for _, dep := range defiActionsSection.Dependencies {
		if dep.Name == defiContractName && dep.Source.NetworkName == config.MainnetNetwork.Name {
			targetDep = &dep
			break
		}
	}

	if targetDep == nil {
		return fmt.Errorf("contract %s not found in DeFi actions contracts", defiContractName)
	}

	return di.Add(*targetDep)
}

// Add processes a single dependency and installs it and any dependencies it has, as well as adding it to the state
func (di *DependencyInstaller) Add(dep config.Dependency) error {
	if err := di.processDependency(dep); err != nil {
		return fmt.Errorf("error processing dependency: %w", err)
	}

	di.checkForConflictingContracts()

	if err := di.saveState(); err != nil {
		return err
	}

	return nil
}

// AddMany processes multiple dependencies and installs them as well as adding them to the state
func (di *DependencyInstaller) AddMany(dependencies []config.Dependency) error {
	for _, dep := range dependencies {
		if err := di.processDependency(dep); err != nil {
			return fmt.Errorf("error processing dependency: %w", err)
		}
	}

	di.checkForConflictingContracts()

	if err := di.saveState(); err != nil {
		return err
	}

	return nil
}

func (di *DependencyInstaller) AddAllByNetworkAddress(sourceStr string) error {
	network, address := ParseNetworkAddressString(sourceStr)

	accountContracts, err := di.getContracts(network, flowsdk.HexToAddress(address))
	if err != nil {
		return fmt.Errorf("failed to fetch account contracts: %w", err)
	}

	var dependencies []config.Dependency

	for _, contract := range accountContracts {
		program, err := project.NewProgram(contract, nil, "")
		if err != nil {
			return fmt.Errorf("failed to parse program: %w", err)
		}

		contractName, err := program.Name()
		if err != nil {
			return fmt.Errorf("failed to parse contract name: %w", err)
		}

		dep := config.Dependency{
			Name: contractName,
			Source: config.Source{
				NetworkName:  network,
				Address:      flowsdk.HexToAddress(address),
				ContractName: contractName,
			},
		}

		dependencies = append(dependencies, dep)
	}

	if err := di.AddMany(dependencies); err != nil {
		return err
	}

	return nil
}

func (di *DependencyInstaller) addDependency(dep config.Dependency) error {
	sourceString := fmt.Sprintf("%s://%s.%s", dep.Source.NetworkName, dep.Source.Address.String(), dep.Source.ContractName)

	if _, exists := di.dependencies[sourceString]; exists {
		return nil
	}

	di.dependencies[sourceString] = dep

	return nil
}

// checkForConflictingContracts checks if any of the dependencies conflict with contracts already in the state
func (di *DependencyInstaller) checkForConflictingContracts() {
	for _, dependency := range di.dependencies {
		foundContract, _ := di.State.Contracts().ByName(dependency.Name)
		if foundContract != nil && !foundContract.IsDependency {
			msg := util.MessageWithEmojiPrefix("❌", fmt.Sprintf("Contract named %s already exists in flow.json", dependency.Name))
			di.logs.issues = append(di.logs.issues, msg)
		}
	}
}

func (di *DependencyInstaller) processDependency(dependency config.Dependency) error {
	depAddress := flowsdk.HexToAddress(dependency.Source.Address.String())
	return di.fetchDependencies(dependency.Source.NetworkName, depAddress, dependency.Source.ContractName)
}

func (di *DependencyInstaller) getContracts(network string, address flowsdk.Address) (map[string][]byte, error) {
	gw, ok := di.Gateways[network]
	if !ok {
		return nil, fmt.Errorf("gateway for network %s not found", network)
	}

	ctx := context.Background()
	acct, err := gw.GetAccount(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to get account at %s on %s: %w", address, network, err)
	}

	if acct == nil {
		return nil, fmt.Errorf("no account found at address %s on network %s", address, network)
	}

	if len(acct.Contracts) == 0 {
		return nil, fmt.Errorf("no contracts found at address %s on network %s", address, network)
	}

	return acct.Contracts, nil
}

func (di *DependencyInstaller) fetchDependencies(networkName string, address flowsdk.Address, contractName string) error {
	sourceString := fmt.Sprintf("%s://%s.%s", networkName, address.String(), contractName)

	if _, exists := di.dependencies[sourceString]; exists {
		return nil // Skip already processed dependencies
	}

	err := di.addDependency(config.Dependency{
		Name: contractName,
		Source: config.Source{
			NetworkName:  networkName,
			Address:      address,
			ContractName: contractName,
		},
	})
	if err != nil {
		return fmt.Errorf("error adding dependency: %w", err)
	}

	accountContracts, err := di.getContracts(networkName, address)
	if err != nil {
		return fmt.Errorf("error fetching contracts: %w", err)
	}

	contract, ok := accountContracts[contractName]
	if !ok {
		return fmt.Errorf("contract %s not found at address %s", contractName, address.String())
	}

	program, err := project.NewProgram(contract, nil, "")
	if err != nil {
		return fmt.Errorf("failed to parse program: %w", err)
	}

	if err := di.handleFoundContract(networkName, address.String(), contractName, program); err != nil {
		return fmt.Errorf("failed to handle found contract: %w", err)
	}

	if program.HasAddressImports() {
		imports := program.AddressImportDeclarations()
		for _, imp := range imports {
			contractName := imp.Identifiers[0].String()
			err := di.fetchDependencies(networkName, flowsdk.HexToAddress(imp.Location.String()), contractName)
			if err != nil {
				return err
			}
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
	path := filepath.Join(di.TargetDir, "imports", address, fileName)
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

		msg := util.MessageWithEmojiPrefix("✅️", fmt.Sprintf("Contract %s from %s on %s installed", contractName, contractAddr, networkName))
		di.logs.fileSystemActions = append(di.logs.fileSystemActions, msg)
	}

	return nil
}

func (di *DependencyInstaller) handleFoundContract(networkName, contractAddr, contractName string, program *project.Program) error {
	hash := sha256.New()
	hash.Write(program.CodeWithUnprocessedImports())
	originalContractDataHash := hex.EncodeToString(hash.Sum(nil))

	program.ConvertAddressImports()
	contractData := string(program.CodeWithUnprocessedImports())

	dependency := di.State.Dependencies().ByName(contractName)

	// If a dependency by this name already exists and its remote source network or address does not match, then give option to stop or continue
	if dependency != nil && (dependency.Source.NetworkName != networkName || dependency.Source.Address.String() != contractAddr) {
		di.Logger.Info(fmt.Sprintf("%s A dependency named %s already exists with a different remote source. Please fix the conflict and retry.", util.PrintEmoji("🚫"), contractName))
		os.Exit(0)
		return nil
	}

	// Check if remote source version is different from local version
	// If it is, ask if they want to update
	// If no hash, ignore
	if dependency != nil && dependency.Hash != "" && dependency.Hash != originalContractDataHash {
		msg := fmt.Sprintf("The latest version of %s is different from the one you have locally. Do you want to update it?", contractName)
		if !prompt.GenericBoolPrompt(msg) {
			return nil
		}
	}

	err := di.updateDependencyState(networkName, contractAddr, contractName, originalContractDataHash)
	if err != nil {
		di.Logger.Error(fmt.Sprintf("Error updating state: %v", err))
		return err
	}

	// Needs to happen before handleFileSystem
	if !di.contractFileExists(contractAddr, contractName) {
		err := di.handleAdditionalDependencyTasks(networkName, contractName)
		if err != nil {
			di.Logger.Error(fmt.Sprintf("Error handling additional dependency tasks: %v", err))
			return err
		}
	}

	err = di.handleFileSystem(contractAddr, contractName, contractData, networkName)
	if err != nil {
		return fmt.Errorf("error handling file system: %w", err)
	}

	return nil
}

func (di *DependencyInstaller) handleAdditionalDependencyTasks(networkName, contractName string) error {
	// If the contract is not a core contract and the user does not want to skip deployments, then prompt for a deployment
	if !di.SkipDeployments && !util.IsCoreContract(contractName) {
		err := di.updateDependencyDeployment(contractName)
		if err != nil {
			di.Logger.Error(fmt.Sprintf("Error updating deployment: %v", err))
			return err
		}

		msg := util.MessageWithEmojiPrefix("✅", fmt.Sprintf("%s added to emulator deployments", contractName))
		di.logs.stateUpdates = append(di.logs.stateUpdates, msg)
	}

	// If the contract is not a core contract and the user does not want to skip aliasing, then prompt for an alias
	if !di.SkipAlias && !util.IsCoreContract(contractName) {
		err := di.updateDependencyAlias(contractName, networkName)
		if err != nil {
			di.Logger.Error(fmt.Sprintf("Error updating alias: %v", err))
			return err
		}

		msg := util.MessageWithEmojiPrefix("✅", fmt.Sprintf("Alias added for %s on %s", contractName, networkName))
		di.logs.stateUpdates = append(di.logs.stateUpdates, msg)
	}

	return nil
}

func (di *DependencyInstaller) updateDependencyDeployment(contractName string) error {
	// Add to deployments
	// If a deployment already exists for that account, contract, and network, then ignore
	raw := prompt.AddContractToDeploymentPrompt("emulator", *di.State.Accounts(), contractName)

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
	var missingNetworks []string

	switch aliasNetwork {
	case config.MainnetNetwork.Name:
		missingNetworks = []string{config.TestnetNetwork.Name}
	case config.TestnetNetwork.Name:
		missingNetworks = []string{config.MainnetNetwork.Name}
	}

	for _, missingNetwork := range missingNetworks {
		// Check if we already have an alias for this account on this network
		accountAddress := di.getCurrentContractAccountAddress(contractName, aliasNetwork)
		if accountAddress != "" {
			if existingAlias, exists := di.getAccountAlias(accountAddress, missingNetwork); exists {
				// Automatically apply the existing alias
				contract, err := di.State.Contracts().ByName(contractName)
				if err != nil {
					return err
				}
				contract.Aliases.Add(missingNetwork, existingAlias)
				di.Logger.Info(fmt.Sprintf("%s Automatically applied alias %s for %s on %s (from same account)",
					util.PrintEmoji("🔄"), existingAlias.String(), contractName, missingNetwork))
				continue
			}
		}

		label := fmt.Sprintf("Enter an alias address for %s on %s if you have one, otherwise leave blank", contractName, missingNetwork)
		raw := prompt.AddressPromptOrEmpty(label, "Invalid alias address")

		if raw != "" {
			aliasAddress := flowsdk.HexToAddress(raw)

			if accountAddress != "" {
				di.setAccountAlias(accountAddress, missingNetwork, aliasAddress)
			}

			contract, err := di.State.Contracts().ByName(contractName)
			if err != nil {
				return err
			}

			contract.Aliases.Add(missingNetwork, aliasAddress)
		}
	}

	return nil
}

func (di *DependencyInstaller) updateDependencyState(networkName, contractAddress, contractName, contractHash string) error {
	dep := config.Dependency{
		Name: contractName,
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
		msg := util.MessageWithEmojiPrefix("✅", fmt.Sprintf("%s added to flow.json", dep.Name))
		di.logs.stateUpdates = append(di.logs.stateUpdates, msg)
	}

	return nil
}

// getCurrentContractAccountAddress returns the account address for the current contract being processed
func (di *DependencyInstaller) getCurrentContractAccountAddress(contractName, networkName string) string {
	for _, dep := range di.dependencies {
		if dep.Name == contractName && dep.Source.NetworkName == networkName {
			return dep.Source.Address.String()
		}
	}
	return ""
}

// getAccountAlias returns the stored alias for an account on a specific network
func (di *DependencyInstaller) getAccountAlias(accountAddress, networkName string) (flowsdk.Address, bool) {
	if networkAliases, exists := di.accountAliases[networkName]; exists {
		if alias, exists := networkAliases[accountAddress]; exists {
			return alias, true
		}
	}
	return flowsdk.Address{}, false
}

// setAccountAlias stores an alias for an account on a specific network
func (di *DependencyInstaller) setAccountAlias(accountAddress, networkName string, alias flowsdk.Address) {
	if di.accountAliases[networkName] == nil {
		di.accountAliases[networkName] = make(map[string]flowsdk.Address)
	}
	di.accountAliases[networkName][accountAddress] = alias
}
