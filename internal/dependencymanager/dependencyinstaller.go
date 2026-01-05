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
	"strings"

	"github.com/psiemens/sconfig"

	"github.com/onflow/flow-cli/common/branding"
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

// pendingPrompt represents a dependency that needs interactive prompts after tree display
type pendingPrompt struct {
	contractName    string
	networkName     string
	contractAddr    string
	contractData    string
	needsDeployment bool
	needsAlias      bool
	needsUpdate     bool
	updateHash      string
}

func (pp *pendingPrompt) matches(name, network string) bool {
	return pp.contractName == name && pp.networkName == network
}

func (di *DependencyInstaller) logFileSystemAction(message string) {
	msg := util.MessageWithEmojiPrefix("✅", message)
	di.logs.fileSystemActions = append(di.logs.fileSystemActions, msg)
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

type DependencyFlags struct {
	skipDeployments   bool   `default:"false" flag:"skip-deployments" info:"Skip adding the dependency to deployments"`
	skipAlias         bool   `default:"false" flag:"skip-alias" info:"Skip prompting for an alias"`
	skipUpdatePrompts bool   `default:"false" flag:"skip-update-prompts" info:"Skip prompting to update existing dependencies"`
	update            bool   `default:"false" flag:"update" info:"Automatically accept all dependency updates"`
	deploymentAccount string `default:"" flag:"deployment-account,d" info:"Account name to use for deployments (skips deployment account prompt)"`
	name              string `default:"" flag:"name" info:"Import alias name for the dependency (sets canonical field for Cadence import aliasing)"`
}

func (f *DependencyFlags) AddToCommand(cmd *cobra.Command) {
	err := sconfig.New(f).
		FromEnvironment(util.EnvPrefix).
		BindFlags(cmd.Flags()).
		Parse()

	if err != nil {
		panic(err)
	}
}

type DependencyInstaller struct {
	Gateways          map[string]gateway.Gateway
	Logger            output.Logger
	State             *flowkit.State
	SaveState         bool
	TargetDir         string
	SkipDeployments   bool
	SkipAlias         bool
	SkipUpdatePrompts bool
	Update            bool
	DeploymentAccount string
	Name              string
	logs              categorizedLogs
	dependencies      map[string]config.Dependency
	accountAliases    map[string]map[string]flowsdk.Address // network -> account -> alias
	installCount      int                                   // Track number of dependencies installed
	pendingPrompts    []pendingPrompt                       // Dependencies that need prompts after tree display
	prompter          Prompter                              // Optional: for testing. If nil, uses real prompts
	blockHeightCache  map[string]uint64                     // Cache of latest block heights per network for consistent pinning
}

type Prompter interface {
	GenericBoolPrompt(msg string) (bool, error)
}

type prompter struct{}

func (prompter) GenericBoolPrompt(msg string) (bool, error) {
	return prompt.GenericBoolPrompt(msg)
}

// NewDependencyInstaller creates a new instance of DependencyInstaller
func NewDependencyInstaller(logger output.Logger, state *flowkit.State, saveState bool, targetDir string, flags DependencyFlags) (*DependencyInstaller, error) {
	// Validate flags: --update and --skip-update-prompts are mutually exclusive
	if flags.update && flags.skipUpdatePrompts {
		return nil, fmt.Errorf("cannot use both --update and --skip-update-prompts flags together")
	}

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
		Gateways:          gateways,
		Logger:            logger,
		State:             state,
		SaveState:         saveState,
		TargetDir:         targetDir,
		SkipDeployments:   flags.skipDeployments,
		SkipAlias:         flags.skipAlias,
		SkipUpdatePrompts: flags.skipUpdatePrompts,
		Update:            flags.update,
		DeploymentAccount: flags.deploymentAccount,
		Name:              flags.name,
		dependencies:      make(map[string]config.Dependency),
		logs:              categorizedLogs{},
		accountAliases:    make(map[string]map[string]flowsdk.Address),
		pendingPrompts:    make([]pendingPrompt, 0),
		prompter:          prompter{},
		blockHeightCache:  make(map[string]uint64),
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
	// Phase 1: Process all dependencies and display tree (no prompts)
	for _, dependency := range *di.State.Dependencies() {
		if err := di.processDependency(dependency); err != nil {
			return err
		}
	}

	// Phase 2: Handle all collected prompts after tree is complete
	if err := di.processPendingPrompts(); err != nil {
		return err
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

	// If a name is provided, use it as the import alias and set canonical for Cadence import aliasing
	// This enables "import OriginalContract as AliasName from address" syntax
	if di.Name != "" {
		dep.Name = di.Name
		dep.Canonical = depContractName
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

	// Log installation with detailed information and branding colors
	contractNameStyled := branding.PurpleStyle.Render(coreContractName)
	shortAddress := "0x..." + depAddress[len(depAddress)-4:]
	addressStyled := branding.GreenStyle.Render(shortAddress)
	networkStyled := branding.GrayStyle.Render(depNetwork)
	di.Logger.Info(fmt.Sprintf("%s @ %s (%s)", contractNameStyled, addressStyled, networkStyled))

	dep := config.Dependency{
		Name: depContractName,
		Source: config.Source{
			NetworkName:  depNetwork,
			Address:      flowsdk.HexToAddress(depAddress),
			ContractName: depContractName,
		},
	}

	// If a name is provided, use it as the import alias and set canonical for Cadence import aliasing
	// This enables "import OriginalContract as AliasName from address" syntax
	if di.Name != "" {
		dep.Name = di.Name
		dep.Canonical = depContractName
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

	// If a custom name is provided, use it as the dependency name and set canonical
	if di.Name != "" {
		targetDep.Name = di.Name
		targetDep.Canonical = defiContractName
	}

	return di.Add(*targetDep)
}

func isDefiActionsContract(contractName string) bool {
	defiActionsSection := getDefiActionsSection()
	for _, dep := range defiActionsSection.Dependencies {
		if dep.Name == contractName {
			return true
		}
	}
	return false
}

// Add processes a single dependency and installs it and any dependencies it has, as well as adding it to the state
func (di *DependencyInstaller) Add(dep config.Dependency) error {
	// Phase 1: Process dependency and display tree (no prompts)
	if err := di.processDependency(dep); err != nil {
		return fmt.Errorf("error processing dependency: %w", err)
	}

	// Phase 2: Handle all collected prompts after tree is complete
	if err := di.processPendingPrompts(); err != nil {
		return err
	}

	di.checkForConflictingContracts()

	if err := di.saveState(); err != nil {
		return err
	}

	return nil
}

// AddMany processes multiple dependencies and installs them as well as adding them to the state
func (di *DependencyInstaller) AddMany(dependencies []config.Dependency) error {
	// Phase 1: Process all dependencies and display tree (no prompts)
	for _, dep := range dependencies {
		if err := di.processDependency(dep); err != nil {
			return fmt.Errorf("error processing dependency: %w", err)
		}
	}

	// Phase 2: Handle all collected prompts after tree is complete
	if err := di.processPendingPrompts(); err != nil {
		return err
	}

	di.checkForConflictingContracts()

	if err := di.saveState(); err != nil {
		return err
	}

	return nil
}

func (di *DependencyInstaller) AddAllByNetworkAddress(sourceStr string) error {
	// Check if name flag is set - not supported when installing all contracts at an address
	if di.Name != "" {
		return fmt.Errorf("--name flag is not supported when installing all contracts at an address (network://address). Please specify a specific contract using network://address.ContractName format")
	}

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
	return di.processDependencies(dependency)
}

// getLatestBlockHeight returns the current block height for a given network.
// Results are cached per network to ensure all dependencies in a single install
// operation get pinned to the same block height for consistency.
func (di *DependencyInstaller) getLatestBlockHeight(network string) (uint64, error) {
	// Check cache first
	if height, ok := di.blockHeightCache[network]; ok {
		return height, nil
	}

	gw, ok := di.Gateways[network]
	if !ok {
		return 0, fmt.Errorf("gateway for network %s not found", network)
	}

	ctx := context.Background()
	latestBlock, err := gw.GetLatestBlock(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest block: %w", err)
	}

	// Cache the result
	di.blockHeightCache[network] = latestBlock.Height
	return latestBlock.Height, nil
}

func (di *DependencyInstaller) getContracts(network string, address flowsdk.Address) (map[string][]byte, error) {
	return di.getContractsAtBlockHeight(network, address, 0)
}

// getContractsAtBlockHeight retrieves contracts at a specific block height.
// If blockHeight is 0, it fetches the latest version.
// Uses GetAccountAtBlockHeight from flowkit Gateway interface for historical queries.
func (di *DependencyInstaller) getContractsAtBlockHeight(network string, address flowsdk.Address, blockHeight uint64) (map[string][]byte, error) {
	gw, ok := di.Gateways[network]
	if !ok {
		return nil, fmt.Errorf("gateway for network %s not found", network)
	}

	ctx := context.Background()
	var acct *flowsdk.Account
	var err error

	if blockHeight > 0 {
		// Query at specific block height (historical)
		acct, err = gw.GetAccountAtBlockHeight(ctx, address, blockHeight)
		if err != nil {
			return nil, fmt.Errorf("failed to get account at block height %d on %s: %w", blockHeight, network, err)
		}
	} else {
		// Query latest version
		acct, err = gw.GetAccount(ctx, address)
		if err != nil {
			return nil, fmt.Errorf("failed to get account at %s on %s: %w", address, network, err)
		}
	}

	if acct == nil {
		return nil, fmt.Errorf("no account found at address %s on network %s", address, network)
	}

	if len(acct.Contracts) == 0 {
		return nil, fmt.Errorf("no contracts found at address %s on network %s", address, network)
	}

	return acct.Contracts, nil
}

func (di *DependencyInstaller) processDependencies(dependency config.Dependency) error {
	return di.fetchDependenciesWithDepth(dependency, 0)
}

func (di *DependencyInstaller) fetchDependenciesWithDepth(dependency config.Dependency, depth int) error {
	networkName := dependency.Source.NetworkName
	address := dependency.Source.Address
	contractName := dependency.Source.ContractName
	// Safety limit to prevent excessive recursion
	const maxDepth = 10
	if depth > maxDepth {
		di.Logger.Info(fmt.Sprintf("⚠️  Skipping dependency %s: maximum depth (%d) exceeded", contractName, maxDepth))
		return nil
	}

	sourceString := fmt.Sprintf("%s://%s.%s", networkName, address.String(), contractName)

	if _, exists := di.dependencies[sourceString]; exists {
		return nil // Skip already processed dependencies
	}

	// Log installation with visual hierarchy and branding colors
	indent := ""
	prefix := ""

	if depth > 0 {
		// Create indentation with proper tree characters
		for i := 0; i < depth; i++ {
			indent += "  "
		}
		prefix = "├─ "

		// Add depth limit warning for very deep chains
		if depth >= 5 {
			di.Logger.Info(fmt.Sprintf("%s⚠️  Deep dependency chain (depth %d)", indent, depth))
		}
	}

	contractNameStyled := branding.PurpleStyle.Render(contractName)
	fullAddress := address.String()
	shortAddress := "0x..." + fullAddress[len(fullAddress)-4:]
	addressStyled := branding.GreenStyle.Render(shortAddress)
	networkStyled := branding.GrayStyle.Render(networkName)
	di.Logger.Info(fmt.Sprintf("%s%s%s @ %s (%s)", indent, prefix, contractNameStyled, addressStyled, networkStyled))
	di.installCount++

	err := di.addDependency(dependency)
	if err != nil {
		return fmt.Errorf("error adding dependency: %w", err)
	}

	// Check if this dependency already exists and we're discovering it via an alias
	// If so, handle it early using the source network (not the incoming network)
	existingDependency := di.State.Dependencies().ByName(dependency.Name)
	if existingDependency != nil && (existingDependency.Source.NetworkName != networkName || existingDependency.Source.Address.String() != address.String()) {
		if !di.existingAliasMatches(dependency.Name, networkName, address.String()) {
			di.Logger.Info(fmt.Sprintf("%s A dependency named %s already exists with a different remote source. Please fix the conflict and retry.", util.PrintEmoji("🚫"), dependency.Name))
			os.Exit(0)
			return nil
		}
		// Alias matched - contract already in flow.json, discovered via different network.
		// Only update block height if --update flag is set (to respect frozen dependencies)
		if di.Update {
			latestHeight, err := di.getLatestBlockHeight(existingDependency.Source.NetworkName)
			if err != nil {
				return fmt.Errorf("failed to get latest block height: %w", err)
			}
			// Update block height (--update flag triggers this)
			blockHeight := di.shouldUpdateBlockHeight(existingDependency.Name, latestHeight, false, false)
			dep := *existingDependency
			dep.BlockHeight = blockHeight
			if err := di.saveDependencyState(dep); err != nil {
				return fmt.Errorf("error updating dependency: %w", err)
			}
		}
		return nil
	}

	// Determine which block height to use for querying
	// If --update flag is set, always use latest (even for pinned dependencies)
	// Otherwise, use existing block height for frozen dependencies
	var blockHeight uint64
	hadSporkRecovery := false // Track if we had to do spork recovery

	if di.Update || existingDependency == nil || existingDependency.BlockHeight == 0 {
		// Use latest block height for:
		// 1. --update flag (force update to latest)
		// 2. New dependencies
		// 3. Dependencies without pinned block height
		latestHeight, err := di.getLatestBlockHeight(networkName)
		if err != nil {
			return fmt.Errorf("failed to get latest block height: %w", err)
		}
		blockHeight = latestHeight
	} else {
		// Use pinned block height for frozen dependencies
		blockHeight = existingDependency.BlockHeight
	}

	accountContracts, err := di.getContractsAtBlockHeight(networkName, address, blockHeight)
	if err != nil {
		// If we get a spork-related error (block height too old), fall back to latest
		// This happens when flow.json has old block heights from before the current spork
		// We'll check the hash later - if it matches, we just update metadata; if not, normal update flow applies
		if strings.Contains(err.Error(), "spork root block height") || strings.Contains(err.Error(), "key not found") {
			di.Logger.Info(fmt.Sprintf("  %s Block height %d is from before current spork, fetching latest version", util.PrintEmoji("⚠️"), blockHeight))
			hadSporkRecovery = true
			// Get the current block height (will be cached from above for new deps)
			latestHeight, err := di.getLatestBlockHeight(networkName)
			if err != nil {
				return fmt.Errorf("failed to get latest block height: %w", err)
			}
			// Fetch at that specific block height
			accountContracts, err = di.getContractsAtBlockHeight(networkName, address, latestHeight)
			if err != nil {
				return fmt.Errorf("error fetching contracts: %w", err)
			}
			// Update blockHeight so it's used consistently for this dependency
			blockHeight = latestHeight
		} else {
			return fmt.Errorf("error fetching contracts: %w", err)
		}
	}

	contract, ok := accountContracts[contractName]
	if !ok {
		return fmt.Errorf("contract %s not found at address %s", contractName, address.String())
	}

	program, err := project.NewProgram(contract, nil, "")
	if err != nil {
		return fmt.Errorf("failed to parse program: %w", err)
	}

	if err := di.handleFoundContract(dependency, program, blockHeight, hadSporkRecovery); err != nil {
		return fmt.Errorf("failed to handle found contract: %w", err)
	}

	if program.HasAddressImports() {
		imports := program.AddressImportDeclarations()
		for _, imp := range imports {

			actualContractName := imp.Imports[0].Identifier.Identifier
			importAddress := flowsdk.HexToAddress(imp.Location.String())

			// Check if this import has an alias (e.g., "import FUSD as FUSD1 from 0xaddress")
			// If aliased, use the alias as the dependency name so "import FUSD1" resolves correctly
			dependencyName := actualContractName
			if imp.Imports[0].Alias.Identifier != "" {
				dependencyName = imp.Imports[0].Alias.Identifier
			}

			// Create a dependency for the import
			// Name is the alias (or actual name if not aliased) - this is what gets resolved in imports
			// ContractName is the actual contract name on chain - this is what gets fetched
			importDependency := config.Dependency{
				Name: dependencyName,
				Source: config.Source{
					NetworkName:  networkName,
					Address:      importAddress,
					ContractName: actualContractName,
				},
			}

			err := di.fetchDependenciesWithDepth(importDependency, depth+1)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (di *DependencyInstaller) getContractFilePath(address, contractName string) string {
	fileName := fmt.Sprintf("%s.cdc", contractName)
	return filepath.Join("imports", address, fileName)
}

func (di *DependencyInstaller) contractFileExists(address, contractName string) bool {
	path := di.getContractFilePath(address, contractName)
	_, err := di.State.ReaderWriter().Stat(path)
	return err == nil
}

func (di *DependencyInstaller) createContractFile(address, contractName, data string) error {
	relativePath := di.getContractFilePath(address, contractName)
	path := filepath.Join(di.TargetDir, relativePath)
	dir := filepath.Dir(path)

	if err := di.State.ReaderWriter().MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directories: %w", err)
	}

	if err := di.State.ReaderWriter().WriteFile(path, []byte(data), 0644); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

// verifyLocalFileIntegrity checks if the local file matches the expected hash in flow.json
func (di *DependencyInstaller) verifyLocalFileIntegrity(contractAddr, contractName, expectedHash string) error {
	if !di.contractFileExists(contractAddr, contractName) {
		return nil // File doesn't exist, nothing to verify
	}

	filePath := di.getContractFilePath(contractAddr, contractName)
	fileContent, err := di.State.ReaderWriter().ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file for integrity check: %w", err)
	}

	// Calculate hash of existing file
	fileHash := sha256.New()
	fileHash.Write(fileContent)
	existingFileHash := hex.EncodeToString(fileHash.Sum(nil))

	// Compare hashes
	if expectedHash != existingFileHash {
		return fmt.Errorf(
			"dependency %s: local file has been modified (hash mismatch). Expected hash %s but file has %s. The file content does not match what is recorded in flow.json. Run 'flow dependencies install --update' to sync with the network version, or restore the file to match the stored hash",
			contractName,
			expectedHash,
			existingFileHash,
		)
	}

	return nil
}

func (di *DependencyInstaller) handleFoundContract(dependency config.Dependency, program *project.Program, fetchedBlockHeight uint64, hadSporkRecovery bool) error {
	networkName := dependency.Source.NetworkName
	contractAddr := dependency.Source.Address.String()
	contractName := dependency.Source.ContractName

	program.ConvertAddressImports()
	contractData := string(program.CodeWithUnprocessedImports())

	// Calculate hash of converted contract (what gets written to disk)
	// This is what we store in flow.json so we can verify file integrity later
	// Imported contracts are still checked for consistency by traversing the dependency tree.
	hash := sha256.New()
	hash.Write([]byte(contractData))
	contractDataHash := hex.EncodeToString(hash.Sum(nil))

	existingDependency := di.State.Dependencies().ByName(dependency.Name)

	// Check if remote source version is different from local version
	// Decide what to do: defer prompt, skip (frozen), or auto-update
	hashMismatch := existingDependency != nil && existingDependency.Hash != "" && existingDependency.Hash != contractDataHash

	if hashMismatch {
		// If skip update prompts flag is set, check if we can keep frozen dependencies
		if di.SkipUpdatePrompts && di.contractFileExists(contractAddr, contractName) {
			// File exists - verify it matches stored hash
			if err := di.verifyLocalFileIntegrity(contractAddr, contractName, existingDependency.Hash); err != nil {
				// Local file was modified - FAIL
				return fmt.Errorf("cannot install with --skip-update-prompts flag when local files have been modified. %w", err)
			}

			// File exists and matches stored hash - but network version changed
			// Check if we fetched at a different block height (e.g., pre-spork recovery)
			// Only error if the dependency was actually pinned (BlockHeight > 0)
			if existingDependency.BlockHeight > 0 && fetchedBlockHeight != existingDependency.BlockHeight {
				// Pre-spork recovery scenario: we couldn't fetch the old block, had to use latest
				// But the hash on the network differs from what we have frozen
				// This means we can't truly keep it frozen - ERROR OUT
				return fmt.Errorf(
					"dependency %s: cannot keep frozen with --skip-update-prompts. "+
						"The stored block height (%d) is no longer accessible (likely pre-spork), "+
						"and the contract on-chain at the current block height (%d) has a different hash. "+
						"Run 'flow dependencies install --update' to fetch the latest version, "+
						"or remove --skip-update-prompts to be prompted for updates",
					dependency.Name,
					existingDependency.BlockHeight,
					fetchedBlockHeight,
				)
			}

			// File exists, matches stored hash, and we fetched at the stored block height
			// This is truly frozen - keep it as is
			return nil
		}

		// If --update flag is set, auto-accept the update (fall through to install)
		// If --skip-update-prompts with no file, install from network (fall through to install)
		// Otherwise (normal mode), defer prompt until after tree display
		if !di.Update && !di.SkipUpdatePrompts {
			found := false
			for i := range di.pendingPrompts {
				if di.pendingPrompts[i].matches(dependency.Name, networkName) {
					di.pendingPrompts[i].needsUpdate = true
					di.pendingPrompts[i].updateHash = contractDataHash
					di.pendingPrompts[i].contractAddr = contractAddr
					di.pendingPrompts[i].contractData = contractData
					found = true
					break
				}
			}
			if !found {
				di.pendingPrompts = append(di.pendingPrompts, pendingPrompt{
					contractName: dependency.Name,
					networkName:  networkName,
					contractAddr: contractAddr,
					contractData: contractData,
					needsUpdate:  true,
					updateHash:   contractDataHash,
				})
			}
			return nil
		}
		// Fall through: --update or --skip-update-prompts without file → install from network
	}

	// Check if file exists and needs repair (out of sync with current network version)
	fileExists := di.contractFileExists(contractAddr, contractName)
	fileModified := false
	if fileExists {
		// Check if the file matches what we just fetched from the network
		if err := di.verifyLocalFileIntegrity(contractAddr, contractName, contractDataHash); err != nil {
			fileModified = true
		}
	}

	// Install or update: new deps, out-of-sync files, or network updates with --update/--skip-update-prompts
	isNewDep := di.State.Dependencies().ByName(dependency.Name) == nil

	// Determine final block height and save
	blockHeight := di.shouldUpdateBlockHeight(dependency.Name, fetchedBlockHeight, hashMismatch, hadSporkRecovery)
	dep := dependency
	dep.Hash = contractDataHash
	dep.BlockHeight = blockHeight

	if err := di.saveDependencyState(dep); err != nil {
		return fmt.Errorf("error updating state: %w", err)
	}

	// Log if this was an auto-update (with --update flag) or file repair
	if (hashMismatch || fileModified) && di.Update {
		msg := util.MessageWithEmojiPrefix("✅", fmt.Sprintf("%s updated to latest version", dependency.Name))
		di.logs.stateUpdates = append(di.logs.stateUpdates, msg)
	} else if fileModified {
		// File repair without --update flag (common after git clone)
		msg := util.MessageWithEmojiPrefix("✅", fmt.Sprintf("%s synced", dependency.Name))
		di.logs.stateUpdates = append(di.logs.stateUpdates, msg)
	}

	// Handle additional tasks for new dependencies or when contract file doesn't exist
	if isNewDep || !fileExists {
		err := di.handleAdditionalDependencyTasks(networkName, dependency.Name)
		if err != nil {
			di.Logger.Error(fmt.Sprintf("Error handling additional dependency tasks: %v", err))
			return err
		}
	}

	// Create or overwrite file
	shouldWrite := !fileExists || fileModified || (hashMismatch && di.Update)
	if !shouldWrite {
		return nil
	}

	if err := di.createContractFile(contractAddr, contractName, contractData); err != nil {
		return fmt.Errorf("error creating contract file: %w", err)
	}

	if !fileExists {
		di.logFileSystemAction(fmt.Sprintf("Contract %s from %s on %s installed", contractName, contractAddr, networkName))
	}

	return nil
}

// existingAliasMatches returns true if an existing contract with the given name has an alias
// for the provided network that matches the specified address.
func (di *DependencyInstaller) existingAliasMatches(contractName, networkName, contractAddr string) bool {
	if di.State == nil || di.State.Contracts() == nil {
		return false
	}
	contract, err := di.State.Contracts().ByName(contractName)
	if err != nil || contract == nil {
		return false
	}
	alias := contract.Aliases.ByNetwork(networkName)
	if alias == nil {
		return false
	}
	return alias.Address.String() == contractAddr
}

func (di *DependencyInstaller) handleAdditionalDependencyTasks(networkName, contractName string) error {
	// If the contract is not a core contract and the user does not want to skip deployments, then collect for prompting later
	needsDeployment := !di.SkipDeployments && !util.IsCoreContract(contractName)

	// For DeFi Actions contracts, only allow deployment on emulator (handle immediately since no prompt needed)
	if needsDeployment && isDefiActionsContract(contractName) {
		err := di.updateDependencyDeployment(contractName, "emulator")
		if err != nil {
			di.Logger.Error(fmt.Sprintf("Error updating deployment: %v", err))
			return err
		}
		msg := util.MessageWithEmojiPrefix("✅", fmt.Sprintf("%s added to emulator deployments (DeFi Actions contracts only supported on emulator)", contractName))
		di.logs.stateUpdates = append(di.logs.stateUpdates, msg)
		needsDeployment = false // Already handled
	}

	// If the contract is not a core contract and the user does not want to skip aliasing, then collect for prompting later
	needsAlias := !di.SkipAlias && !util.IsCoreContract(contractName) && !isDefiActionsContract(contractName)

	// Only add/update pending prompts if we need to prompt for something
	if needsDeployment || needsAlias {
		di.pendingPrompts = append(di.pendingPrompts, pendingPrompt{
			contractName:    contractName,
			networkName:     networkName,
			needsDeployment: needsDeployment,
			needsAlias:      needsAlias,
		})
	}

	return nil
}

func (di *DependencyInstaller) updateDependencyDeployment(contractName string, forceNetwork ...string) error {
	var raw *prompt.DeploymentData
	network := "emulator"

	// If a forced network is specified, use it
	if len(forceNetwork) > 0 {
		network = forceNetwork[0]
	}

	// If deployment account is specified via flag, use it; otherwise prompt
	if di.DeploymentAccount != "" {
		account, err := di.State.Accounts().ByName(di.DeploymentAccount)
		if err != nil || account == nil {
			return fmt.Errorf("deployment account '%s' not found in flow.json accounts", di.DeploymentAccount)
		}

		raw = &prompt.DeploymentData{
			Network:   network,
			Account:   di.DeploymentAccount,
			Contracts: []string{contractName},
		}
	} else {
		raw = prompt.AddContractToDeploymentPrompt(network, *di.State.Accounts(), contractName)
	}

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

// shouldUpdateBlockHeight determines if we should update to a new block height or keep the existing one
func (di *DependencyInstaller) shouldUpdateBlockHeight(depName string, newHeight uint64, hashChanged bool, hadSporkRecovery bool) uint64 {
	existing := di.State.Dependencies().ByName(depName)

	// Always use new height if: new dep, hash changed, spork recovery, old format, or --update flag
	if existing == nil || hashChanged || hadSporkRecovery || existing.BlockHeight == 0 || di.Update {
		return newHeight
	}

	// Otherwise keep existing (frozen dependency)
	return existing.BlockHeight
}

// saveDependencyState saves the dependency to state and logs changes
func (di *DependencyInstaller) saveDependencyState(dep config.Dependency) error {
	existing := di.State.Dependencies().ByName(dep.Name)
	isNew := existing == nil

	// Save to state
	di.State.Dependencies().AddOrUpdate(dep)
	di.State.Contracts().AddDependencyAsContract(dep, dep.Source.NetworkName)

	// Handle aliased imports (enables "import X as Y from address" syntax)
	if dep.Name != dep.Source.ContractName {
		if contract, err := di.State.Contracts().ByName(dep.Name); err == nil && contract != nil {
			contract.Canonical = dep.Source.ContractName
		}
	}

	// Log changes
	if isNew {
		msg := util.MessageWithEmojiPrefix("✅", fmt.Sprintf("%s added to flow.json", dep.Name))
		di.logs.stateUpdates = append(di.logs.stateUpdates, msg)
	} else if existing.BlockHeight > 0 && existing.BlockHeight != dep.BlockHeight {
		msg := util.MessageWithEmojiPrefix("🔄", fmt.Sprintf("%s block height updated (%d → %d)",
			dep.Name, existing.BlockHeight, dep.BlockHeight))
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

// processPendingPrompts handles all collected prompts after the dependency tree is displayed
func (di *DependencyInstaller) processPendingPrompts() error {
	if len(di.pendingPrompts) == 0 {
		return nil
	}

	di.Logger.Info("") // Add spacing after tree display

	// Check if we have any dependencies that need deployments
	hasDeployments := false
	for _, pending := range di.pendingPrompts {
		if pending.needsDeployment {
			hasDeployments = true
			break
		}
	}

	// Check if we have any dependencies that need aliases
	hasAliases := false
	for _, pending := range di.pendingPrompts {
		if pending.needsAlias {
			hasAliases = true
			break
		}
	}

	setupDeployments := false
	if hasDeployments {
		result, err := di.prompter.GenericBoolPrompt("Do you want to set up deployments for these dependencies?")
		if err != nil {
			return err
		}
		setupDeployments = result
	}

	setupAliases := false
	if hasAliases {
		result, err := di.prompter.GenericBoolPrompt("Do you want to set up aliases for these dependencies?")
		if err != nil {
			return err
		}
		setupAliases = result
	}

	// Process prompts based on user choices
	for _, pending := range di.pendingPrompts {
		if pending.needsUpdate {
			msg := fmt.Sprintf("The latest version of %s is different from the one you have locally. Do you want to update it?", pending.contractName)
			shouldUpdate, err := di.prompter.GenericBoolPrompt(msg)
			if err != nil {
				return err
			}
			if shouldUpdate {
				dependency := di.State.Dependencies().ByName(pending.contractName)
				if dependency != nil {
					// Get latest block height for the update
					latestHeight, err := di.getLatestBlockHeight(dependency.Source.NetworkName)
					if err != nil {
						return fmt.Errorf("failed to get latest block height: %w", err)
					}

					// User accepted update - hash changed, so use new block height
					blockHeight := di.shouldUpdateBlockHeight(dependency.Name, latestHeight, true, false)
					dep := *dependency
					dep.Hash = pending.updateHash
					dep.BlockHeight = blockHeight

					if err := di.saveDependencyState(dep); err != nil {
						return fmt.Errorf("error updating dependency: %w", err)
					}

					// Write the updated contract file (force overwrite)
					if err := di.createContractFile(pending.contractAddr, pending.contractName, pending.contractData); err != nil {
						return fmt.Errorf("failed to update contract file: %w", err)
					}

					msg := util.MessageWithEmojiPrefix("✅", fmt.Sprintf("%s updated to latest version", pending.contractName))
					di.logs.stateUpdates = append(di.logs.stateUpdates, msg)
				}
			} else {
				// User chose not to update
				// If file doesn't exist, we MUST fail - can't guarantee frozen deps (no way to fetch old version)
				if !di.contractFileExists(pending.contractAddr, pending.contractName) {
					return fmt.Errorf("dependency %s has changed on-chain but file does not exist locally. Cannot keep at current version because we have no way to fetch the old version from the blockchain. Either accept the update or manually add the contract file", pending.contractName)
				}

				// Get the stored hash from flow.json
				dependency := di.State.Dependencies().ByName(pending.contractName)
				if dependency == nil {
					return fmt.Errorf("dependency %s not found in state", pending.contractName)
				}

				// Verify the existing file's hash matches what's in flow.json to ensure integrity
				if err := di.verifyLocalFileIntegrity(pending.contractAddr, pending.contractName, dependency.Hash); err != nil {
					return err
				}

				// File exists and hash matches - keep it at current version
				msg := util.MessageWithEmojiPrefix("⏸️", fmt.Sprintf("%s kept at current version", pending.contractName))
				di.logs.stateUpdates = append(di.logs.stateUpdates, msg)
			}
		}
	}

	for _, pending := range di.pendingPrompts {
		if pending.needsDeployment && setupDeployments {
			err := di.updateDependencyDeployment(pending.contractName)
			if err != nil {
				di.Logger.Error(fmt.Sprintf("Error updating deployment: %v", err))
				return err
			}
			msg := util.MessageWithEmojiPrefix("✅", fmt.Sprintf("%s added to emulator deployments", pending.contractName))
			di.logs.stateUpdates = append(di.logs.stateUpdates, msg)
		}

		if pending.needsAlias && setupAliases {
			err := di.updateDependencyAlias(pending.contractName, pending.networkName)
			if err != nil {
				di.Logger.Error(fmt.Sprintf("Error updating alias: %v", err))
				return err
			}
			msg := util.MessageWithEmojiPrefix("✅", fmt.Sprintf("Alias added for %s on %s", pending.contractName, pending.networkName))
			di.logs.stateUpdates = append(di.logs.stateUpdates, msg)
		}
	}

	// Clear pending prompts after processing
	di.pendingPrompts = make([]pendingPrompt, 0)

	return nil
}

// GetInstallCount returns the number of dependencies installed
func (di *DependencyInstaller) GetInstallCount() int {
	return di.installCount
}
