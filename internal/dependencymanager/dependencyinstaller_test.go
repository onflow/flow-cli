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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/onflow/flowkit/v2/accounts"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/gateway"
	"github.com/onflow/flowkit/v2/gateway/mocks"
	"github.com/onflow/flowkit/v2/output"
	"github.com/onflow/flowkit/v2/tests"

	"github.com/onflow/flow-cli/internal/util"
)

// mockPrompter for testing
type mockPrompter struct {
	responses []bool // Queue of responses to return
	index     int    // Tracks number of prompts shown (and position in responses)
}

func (m *mockPrompter) GenericBoolPrompt(msg string) (bool, error) {
	if m.index >= len(m.responses) {
		return false, fmt.Errorf("no more mock responses available")
	}
	response := m.responses[m.index]
	m.index++
	return response, nil
}

func TestDependencyInstallerInstall(t *testing.T) {

	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	serviceAcc, _ := state.EmulatorServiceAccount()
	serviceAddress := serviceAcc.Address

	dep := config.Dependency{
		Name: "Hello",
		Source: config.Source{
			NetworkName:  "emulator",
			Address:      serviceAddress,
			ContractName: "Hello",
		},
	}

	state.Dependencies().AddOrUpdate(dep)

	t.Run("Success", func(t *testing.T) {
		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 100}}, nil)

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), serviceAcc.Address.String())
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				tests.ContractHelloString.Name: tests.ContractHelloString.Source,
			}

			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:           logger,
			State:            state,
			SaveState:        true,
			TargetDir:        "",
			SkipDeployments:  true,
			SkipAlias:        true,
			dependencies:     make(map[string]config.Dependency),
			blockHeightCache: make(map[string]uint64),
		}

		err := di.Install()
		assert.NoError(t, err, "Failed to install dependencies")

		filePath := fmt.Sprintf("imports/%s/%s.cdc", serviceAddress.String(), tests.ContractHelloString.Name)
		fileContent, err := state.ReaderWriter().ReadFile(filePath)
		assert.NoError(t, err, "Failed to read generated file")
		assert.NotNil(t, fileContent)
	})

	t.Run("Conflicting flags error", func(t *testing.T) {
		_, state, _ := util.TestMocks(t)

		// Try to create installer with both --update and --skip-update-prompts
		_, err := NewDependencyInstaller(
			logger,
			state,
			true,
			"",
			DependencyFlags{
				update:            true,
				skipUpdatePrompts: true,
			},
		)

		assert.Error(t, err, "Should fail when both flags are set")
		assert.Contains(t, err.Error(), "cannot use both", "Error should mention conflicting flags")
	})
}

func TestDependencyInstallerInstallFromFreshClone(t *testing.T) {
	// This test simulates cloning a repo that has dependencies in flow.json with matching hashes,
	// but no files in imports/ directory. It should create files without prompting.

	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	serviceAcc, _ := state.EmulatorServiceAccount()
	serviceAddress := serviceAcc.Address

	t.Run("First install, up-to-date hash", func(t *testing.T) {
		// Use the standard test contract
		contractCode := tests.ContractHelloString.Source

		// Calculate the hash for the contract (this is what would be in flow.json after a commit)
		hash := sha256.New()
		hash.Write(contractCode)
		contractHash := hex.EncodeToString(hash.Sum(nil))

		// Simulate a dependency that exists in flow.json with matching hash
		// (like what you'd have after cloning a repo - hash matches network but no local file)
		dep := config.Dependency{
			Name: "Hello",
			Source: config.Source{
				NetworkName:  "emulator",
				Address:      serviceAddress,
				ContractName: "Hello",
			},
			Hash: contractHash, // Hash matches what's on network
		}

		state.Dependencies().AddOrUpdate(dep)

		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 100}}, nil)

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), serviceAcc.Address.String())
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"Hello": contractCode,
			}

			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:            logger,
			State:             state,
			SaveState:         true,
			TargetDir:         "",
			SkipDeployments:   true,
			SkipAlias:         true,
			SkipUpdatePrompts: false,
			dependencies:      make(map[string]config.Dependency),
			accountAliases:    make(map[string]map[string]flow.Address),
			pendingPrompts:    make([]pendingPrompt, 0),
			blockHeightCache:  make(map[string]uint64),
			prompter:          &mockPrompter{responses: []bool{}},
		}

		err := di.Install()
		assert.NoError(t, err, "Failed to install dependencies")

		// Verify file was created
		filePath := fmt.Sprintf("imports/%s/%s.cdc", serviceAddress.String(), "Hello")
		fileContent, err := state.ReaderWriter().ReadFile(filePath)
		assert.NoError(t, err, "Failed to read generated file")
		assert.NotNil(t, fileContent)

		// Verify hash remained the same (no update needed)
		updatedDep := state.Dependencies().ByName("Hello")
		assert.NotNil(t, updatedDep)
		assert.Equal(t, contractHash, updatedDep.Hash, "Hash should remain unchanged")
	})

	t.Run("First install, outdated hash - user accepts", func(t *testing.T) {
		// Fresh state for this test
		_, state, _ := util.TestMocks(t)

		// Network has a newer version of the contract
		newContractCode := []byte(`access(all) contract Hello { access(all) fun sayHello(): String { return "Hello, World! v2" } }`)

		// Calculate the new hash
		newHash := sha256.New()
		newHash.Write(newContractCode)
		newContractHash := hex.EncodeToString(newHash.Sum(nil))

		// Simulate a dependency that exists in flow.json with an OLD hash
		dep := config.Dependency{
			Name: "Hello",
			Source: config.Source{
				NetworkName:  "emulator",
				Address:      serviceAddress,
				ContractName: "Hello",
			},
			Hash: "old_hash_from_previous_version",
		}

		state.Dependencies().AddOrUpdate(dep)

		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 100}}, nil)

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), serviceAddress.String())
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"Hello": newContractCode,
			}

			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		// Mock prompter that returns true (user says "yes")
		mockPrompter := &mockPrompter{responses: []bool{true}}

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:            logger,
			State:             state,
			SaveState:         true,
			TargetDir:         "",
			SkipDeployments:   true,
			SkipAlias:         true,
			SkipUpdatePrompts: false,
			dependencies:      make(map[string]config.Dependency),
			accountAliases:    make(map[string]map[string]flow.Address),
			pendingPrompts:    make([]pendingPrompt, 0),
			blockHeightCache:  make(map[string]uint64),
			prompter:          mockPrompter,
		}

		err := di.Install()
		assert.NoError(t, err, "Failed to install dependencies")

		// Verify file WAS created with new version
		filePath := fmt.Sprintf("imports/%s/%s.cdc", serviceAddress.String(), "Hello")
		fileContent, err := state.ReaderWriter().ReadFile(filePath)
		assert.NoError(t, err, "File should exist after accepting update")
		assert.NotNil(t, fileContent)
		assert.Contains(t, string(fileContent), "Hello, World! v2", "Should have the new contract version")

		// Verify hash WAS updated
		updatedDep := state.Dependencies().ByName("Hello")
		assert.NotNil(t, updatedDep)
		assert.Equal(t, newContractHash, updatedDep.Hash, "Hash should be updated to new version")
		assert.NotEqual(t, "old_hash_from_previous_version", updatedDep.Hash, "Should not have old hash")

		// Verify no warnings
		assert.Empty(t, di.logs.issues, "Should have no warnings when update is accepted")
	})

	t.Run("First install, outdated hash - user declines", func(t *testing.T) {
		// Fresh state for this test
		_, state, _ := util.TestMocks(t)

		// Network has a newer version of the contract
		newContractCode := []byte(`access(all) contract Hello { access(all) fun sayHello(): String { return "Hello, World! v2" } }`)

		// Simulate a dependency that exists in flow.json with an OLD hash
		dep := config.Dependency{
			Name: "Hello",
			Source: config.Source{
				NetworkName:  "emulator",
				Address:      serviceAddress,
				ContractName: "Hello",
			},
			Hash: "old_hash_from_previous_version",
		}

		state.Dependencies().AddOrUpdate(dep)

		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 100}}, nil)

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), serviceAddress.String())
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"Hello": newContractCode,
			}

			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		// Mock prompter that returns false (user says "no")
		mockPrompter := &mockPrompter{responses: []bool{false}}

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:            logger,
			State:             state,
			SaveState:         true,
			TargetDir:         "",
			SkipDeployments:   true,
			SkipAlias:         true,
			SkipUpdatePrompts: false,
			dependencies:      make(map[string]config.Dependency),
			accountAliases:    make(map[string]map[string]flow.Address),
			pendingPrompts:    make([]pendingPrompt, 0),
			blockHeightCache:  make(map[string]uint64),
			prompter:          mockPrompter,
		}

		err := di.Install()
		// Should FAIL because user declined and file doesn't exist (can't fetch old version)
		assert.Error(t, err, "Should fail when user declines update and file doesn't exist")
		assert.Contains(t, err.Error(), "Hello", "Error should mention the dependency name")
		assert.Contains(t, err.Error(), "does not exist locally", "Error should mention missing file")
		assert.Contains(t, err.Error(), "no way to fetch", "Error should explain can't fetch old version")
	})

	t.Run("First install, outdated hash - skip flag WITHOUT file", func(t *testing.T) {
		// Fresh state for this test
		_, state, _ := util.TestMocks(t)

		// Network has a newer version of the contract
		newContractCode := []byte(`access(all) contract Hello { access(all) fun sayHello(): String { return "Hello, World! v2" } }`)

		// Calculate the new hash
		newHash := sha256.New()
		newHash.Write(newContractCode)
		newContractHash := hex.EncodeToString(newHash.Sum(nil))

		// Simulate a dependency that exists in flow.json with an OLD hash
		// (like what you'd have after cloning a repo where the network has been updated)
		dep := config.Dependency{
			Name: "Hello",
			Source: config.Source{
				NetworkName:  "emulator",
				Address:      serviceAddress,
				ContractName: "Hello",
			},
			Hash: "old_hash_from_previous_version", // Old hash that doesn't match
		}

		state.Dependencies().AddOrUpdate(dep)

		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 100}}, nil)

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), serviceAddress.String())
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"Hello": newContractCode, // Network has new version
			}

			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:            logger,
			State:             state,
			SaveState:         true,
			TargetDir:         "",
			SkipDeployments:   true,
			SkipAlias:         true,
			SkipUpdatePrompts: true, // Skip prompts
			dependencies:      make(map[string]config.Dependency),
			accountAliases:    make(map[string]map[string]flow.Address),
			pendingPrompts:    make([]pendingPrompt, 0),
			prompter:          &mockPrompter{responses: []bool{}},
			blockHeightCache:  make(map[string]uint64),
		}

		err := di.Install()
		// Should SUCCEED - file doesn't exist, so just install from network
		assert.NoError(t, err, "Should succeed when file doesn't exist - just install from network")

		// Verify file WAS created with network version
		filePath := fmt.Sprintf("imports/%s/%s.cdc", serviceAddress.String(), "Hello")
		fileContent, err := state.ReaderWriter().ReadFile(filePath)
		assert.NoError(t, err, "File should exist after install")
		assert.Contains(t, string(fileContent), "Hello, World! v2", "Should have the new network version")

		// Verify hash WAS updated to match network
		updatedDep := state.Dependencies().ByName("Hello")
		assert.NotNil(t, updatedDep)
		assert.Equal(t, newContractHash, updatedDep.Hash, "Hash should be updated to network version")
	})

	t.Run("First install, outdated hash - skip flag with matching file", func(t *testing.T) {
		// Fresh state for this test
		_, state, _ := util.TestMocks(t)

		oldContractCode := []byte(`access(all) contract Hello { access(all) fun sayHello(): String { return "Hello, World! v1" } }`)
		newContractCode := []byte(`access(all) contract Hello { access(all) fun sayHello(): String { return "Hello, World! v2" } }`)

		// Calculate the old hash
		oldHash := sha256.New()
		oldHash.Write(oldContractCode)
		oldContractHash := hex.EncodeToString(oldHash.Sum(nil))

		// Simulate a dependency that exists in flow.json with an OLD hash
		dep := config.Dependency{
			Name: "Hello",
			Source: config.Source{
				NetworkName:  "emulator",
				Address:      serviceAddress,
				ContractName: "Hello",
			},
			Hash: oldContractHash,
		}

		state.Dependencies().AddOrUpdate(dep)

		// Create the OLD file that matches the stored hash
		filePath := fmt.Sprintf("imports/%s/Hello.cdc", serviceAddress.String())
		err := state.ReaderWriter().MkdirAll(filepath.Dir(filePath), 0755)
		assert.NoError(t, err)
		err = state.ReaderWriter().WriteFile(filePath, oldContractCode, 0644)
		assert.NoError(t, err)

		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 100}}, nil)

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), serviceAddress.String())
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"Hello": newContractCode, // Network has new version
			}

			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:            logger,
			State:             state,
			SaveState:         true,
			TargetDir:         "",
			SkipDeployments:   true,
			SkipAlias:         true,
			SkipUpdatePrompts: true, // Skip prompts flag set
			dependencies:      make(map[string]config.Dependency),
			accountAliases:    make(map[string]map[string]flow.Address),
			pendingPrompts:    make([]pendingPrompt, 0),
			prompter:          &mockPrompter{responses: []bool{}},
			blockHeightCache:  make(map[string]uint64),
		}

		err = di.Install()
		// Should SUCCEED because local file exists and matches stored hash (frozen deps are valid)
		assert.NoError(t, err, "Should succeed when file exists and matches stored hash with --skip-update-prompts")

		// Verify file was NOT changed (still has v1)
		fileContent, err := state.ReaderWriter().ReadFile(filePath)
		assert.NoError(t, err)
		assert.Contains(t, string(fileContent), "Hello, World! v1", "Should keep the old version")
		assert.NotContains(t, string(fileContent), "v2", "Should not have new version")

		// Verify hash was NOT updated in flow.json
		updatedDep := state.Dependencies().ByName("Hello")
		assert.NotNil(t, updatedDep)
		assert.Equal(t, oldContractHash, updatedDep.Hash, "Hash should remain at old version")
	})

	t.Run("First install, outdated hash - skip flag with modified file", func(t *testing.T) {
		// Fresh state for this test
		_, state, _ := util.TestMocks(t)

		oldContractCode := []byte(`access(all) contract Hello { access(all) fun sayHello(): String { return "Hello, World! v1" } }`)
		modifiedContractCode := []byte(`access(all) contract Hello { access(all) fun sayHello(): String { return "Hello, Modified!" } }`)
		newContractCode := []byte(`access(all) contract Hello { access(all) fun sayHello(): String { return "Hello, World! v2" } }`)

		// Calculate the old hash
		oldHash := sha256.New()
		oldHash.Write(oldContractCode)
		oldContractHash := hex.EncodeToString(oldHash.Sum(nil))

		// Simulate a dependency that exists in flow.json with an OLD hash
		dep := config.Dependency{
			Name: "Hello",
			Source: config.Source{
				NetworkName:  "emulator",
				Address:      serviceAddress,
				ContractName: "Hello",
			},
			Hash: oldContractHash,
		}

		state.Dependencies().AddOrUpdate(dep)

		// Create a MODIFIED file (different from both old and new versions)
		filePath := fmt.Sprintf("imports/%s/Hello.cdc", serviceAddress.String())
		err := state.ReaderWriter().MkdirAll(filepath.Dir(filePath), 0755)
		assert.NoError(t, err)
		err = state.ReaderWriter().WriteFile(filePath, modifiedContractCode, 0644)
		assert.NoError(t, err)

		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 100}}, nil)

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), serviceAddress.String())
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"Hello": newContractCode, // Network has new version
			}

			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:            logger,
			State:             state,
			SaveState:         true,
			TargetDir:         "",
			SkipDeployments:   true,
			SkipAlias:         true,
			SkipUpdatePrompts: true, // Skip prompts flag set
			dependencies:      make(map[string]config.Dependency),
			accountAliases:    make(map[string]map[string]flow.Address),
			pendingPrompts:    make([]pendingPrompt, 0),
			prompter:          &mockPrompter{responses: []bool{}},
			blockHeightCache:  make(map[string]uint64),
		}

		err = di.Install()
		// Should FAIL because local file has been modified (detected before checking network hash)
		assert.Error(t, err, "Should fail when local file is modified with --skip-update-prompts")
		assert.Contains(t, err.Error(), "local file has been modified", "Error should mention file modification")
		assert.Contains(t, err.Error(), "hash mismatch", "Error should mention hash mismatch")
		assert.Contains(t, err.Error(), "Hello", "Error should mention the dependency name")
	})

	t.Run("First install, outdated hash - update flag", func(t *testing.T) {
		// Fresh state for this test
		_, state, _ := util.TestMocks(t)

		// Network has a newer version of the contract
		newContractCode := []byte(`access(all) contract Hello { access(all) fun sayHello(): String { return "Hello, World! v2" } }`)

		// Calculate the new hash
		newHash := sha256.New()
		newHash.Write(newContractCode)
		newContractHash := hex.EncodeToString(newHash.Sum(nil))

		// Simulate a dependency that exists in flow.json with an OLD hash
		dep := config.Dependency{
			Name: "Hello",
			Source: config.Source{
				NetworkName:  "emulator",
				Address:      serviceAddress,
				ContractName: "Hello",
			},
			Hash: "old_hash_from_previous_version",
		}

		state.Dependencies().AddOrUpdate(dep)

		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 100}}, nil)

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), serviceAddress.String())
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"Hello": newContractCode,
			}

			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		// No prompter needed - --update auto-accepts
		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:            logger,
			State:             state,
			SaveState:         true,
			TargetDir:         "",
			SkipDeployments:   true,
			SkipAlias:         true,
			SkipUpdatePrompts: false,
			Update:            true, // Auto-accept updates
			dependencies:      make(map[string]config.Dependency),
			accountAliases:    make(map[string]map[string]flow.Address),
			pendingPrompts:    make([]pendingPrompt, 0),
			prompter:          &mockPrompter{responses: []bool{}},
			blockHeightCache:  make(map[string]uint64),
		}

		err := di.Install()
		assert.NoError(t, err, "Should succeed with --update flag")

		// Verify file WAS created with new version
		filePath := fmt.Sprintf("imports/%s/%s.cdc", serviceAddress.String(), "Hello")
		fileContent, err := state.ReaderWriter().ReadFile(filePath)
		assert.NoError(t, err, "File should exist after auto-update")
		assert.NotNil(t, fileContent)
		assert.Contains(t, string(fileContent), "Hello, World! v2", "Should have the new contract version")

		// Verify hash WAS updated
		updatedDep := state.Dependencies().ByName("Hello")
		assert.NotNil(t, updatedDep)
		assert.Equal(t, newContractHash, updatedDep.Hash, "Hash should be updated to new version")
		assert.NotEqual(t, "old_hash_from_previous_version", updatedDep.Hash, "Should not have old hash")

		// Verify success message logged
		assert.NotEmpty(t, di.logs.stateUpdates, "Should have state update messages")
		assert.Contains(t, di.logs.stateUpdates[0], "updated to latest version", "Should log update message")
	})

	t.Run("Already installed, up-to-date hash", func(t *testing.T) {
		// Fresh state for this test
		_, state, _ := util.TestMocks(t)

		contractCode := []byte(`access(all) contract Hello { access(all) fun sayHello(): String { return "Hello, World!" } }`)

		// Calculate the hash
		hash := sha256.New()
		hash.Write(contractCode)
		contractHash := hex.EncodeToString(hash.Sum(nil))

		// Simulate a dependency with matching hash
		dep := config.Dependency{
			Name: "Hello",
			Source: config.Source{
				NetworkName:  "emulator",
				Address:      serviceAddress,
				ContractName: "Hello",
			},
			Hash: contractHash, // Hash matches what's on network
		}

		state.Dependencies().AddOrUpdate(dep)

		// Create the file (already installed)
		filePath := fmt.Sprintf("imports/%s/Hello.cdc", serviceAddress.String())
		err := state.ReaderWriter().MkdirAll(filepath.Dir(filePath), 0755)
		assert.NoError(t, err)
		err = state.ReaderWriter().WriteFile(filePath, contractCode, 0644)
		assert.NoError(t, err)

		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 100}}, nil)

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), serviceAddress.String())
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"Hello": contractCode, // Same version on network
			}

			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		// Mock prompter - should NOT be called
		mockPrompter := &mockPrompter{responses: []bool{}}

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:            logger,
			State:             state,
			SaveState:         true,
			TargetDir:         "",
			SkipDeployments:   true,
			SkipAlias:         true,
			SkipUpdatePrompts: false,
			dependencies:      make(map[string]config.Dependency),
			accountAliases:    make(map[string]map[string]flow.Address),
			pendingPrompts:    make([]pendingPrompt, 0),
			blockHeightCache:  make(map[string]uint64),
			prompter:          mockPrompter,
		}

		err = di.Install()
		assert.NoError(t, err, "Failed to install dependencies")

		// Verify file still exists with same content
		fileContent, err := state.ReaderWriter().ReadFile(filePath)
		assert.NoError(t, err, "File should still exist")
		assert.NotNil(t, fileContent)
		assert.Equal(t, contractCode, fileContent, "File content should be unchanged")

		// Verify hash unchanged
		updatedDep := state.Dependencies().ByName("Hello")
		assert.NotNil(t, updatedDep)
		assert.Equal(t, contractHash, updatedDep.Hash, "Hash should remain the same")

		// Verify no prompts occurred (mockPrompter.index should be 0)
		assert.Equal(t, 0, mockPrompter.index, "No prompts should have been shown")

		// Verify no warnings or state updates
		assert.Empty(t, di.logs.issues, "Should have no warnings")
		assert.Empty(t, di.logs.stateUpdates, "Should have no state updates")
	})

	t.Run("Already installed, up-to-date hash BUT modified local file - user repairs", func(t *testing.T) {
		// Network hash matches flow.json hash, but local file has been tampered with
		// Should auto-repair WITHOUT prompting (flow.json is source of truth)
		_, state, _ := util.TestMocks(t)

		contractCode := []byte(`access(all) contract Hello { access(all) fun sayHello(): String { return "Hello, World!" } }`)
		modifiedContractCode := []byte(`access(all) contract Hello { access(all) fun sayHello(): String { return "Hello, HACKED!" } }`)

		// Calculate the hash of the correct contract
		hash := sha256.New()
		hash.Write(contractCode)
		contractHash := hex.EncodeToString(hash.Sum(nil))

		// Simulate a dependency with matching hash in flow.json
		dep := config.Dependency{
			Name: "Hello",
			Source: config.Source{
				NetworkName:  "emulator",
				Address:      serviceAddress,
				ContractName: "Hello",
			},
			Hash: contractHash, // Hash matches what's on network
		}

		state.Dependencies().AddOrUpdate(dep)

		// Create a MODIFIED file (different from what hash says it should be)
		filePath := fmt.Sprintf("imports/%s/Hello.cdc", serviceAddress.String())
		err := state.ReaderWriter().MkdirAll(filepath.Dir(filePath), 0755)
		assert.NoError(t, err)
		err = state.ReaderWriter().WriteFile(filePath, modifiedContractCode, 0644)
		assert.NoError(t, err)

		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 100}}, nil)

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), serviceAddress.String())
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"Hello": contractCode, // Network has the correct version
			}

			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		// No prompter needed - auto-repairs when network agrees with flow.json
		mockPrompter := &mockPrompter{responses: []bool{}}

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:            logger,
			State:             state,
			SaveState:         true,
			TargetDir:         "",
			SkipDeployments:   true,
			SkipAlias:         true,
			SkipUpdatePrompts: false,
			dependencies:      make(map[string]config.Dependency),
			accountAliases:    make(map[string]map[string]flow.Address),
			pendingPrompts:    make([]pendingPrompt, 0),
			blockHeightCache:  make(map[string]uint64),
			prompter:          mockPrompter,
		}

		err = di.Install()
		// Should SUCCEED - auto-repaired without prompting
		assert.NoError(t, err, "Should auto-repair when network agrees with flow.json")

		// Verify file WAS repaired
		fileContent, err := state.ReaderWriter().ReadFile(filePath)
		assert.NoError(t, err)
		assert.Contains(t, string(fileContent), "Hello, World!", "Should have correct version")
		assert.NotContains(t, string(fileContent), "HACKED", "Should not have hacked version")

		// Verify NO prompt was shown (auto-repair because network agrees with flow.json)
		assert.Equal(t, 0, mockPrompter.index, "Should not prompt when network agrees with flow.json")
	})

	t.Run("Already installed, up-to-date hash BUT modified local file - skip prompts mode", func(t *testing.T) {
		// Network hash matches flow.json hash, but local file has been tampered with
		// Should auto-repair even with --skip-update-prompts (no network change)
		_, state, _ := util.TestMocks(t)

		contractCode := []byte(`access(all) contract Hello { access(all) fun sayHello(): String { return "Hello, World!" } }`)
		modifiedContractCode := []byte(`access(all) contract Hello { access(all) fun sayHello(): String { return "Hello, HACKED!" } }`)

		// Calculate the hash of the correct contract
		hash := sha256.New()
		hash.Write(contractCode)
		contractHash := hex.EncodeToString(hash.Sum(nil))

		// Simulate a dependency with matching hash in flow.json
		dep := config.Dependency{
			Name: "Hello",
			Source: config.Source{
				NetworkName:  "emulator",
				Address:      serviceAddress,
				ContractName: "Hello",
			},
			Hash: contractHash, // Hash matches what's on network
		}

		state.Dependencies().AddOrUpdate(dep)

		// Create a MODIFIED file (different from what hash says it should be)
		filePath := fmt.Sprintf("imports/%s/Hello.cdc", serviceAddress.String())
		err := state.ReaderWriter().MkdirAll(filepath.Dir(filePath), 0755)
		assert.NoError(t, err)
		err = state.ReaderWriter().WriteFile(filePath, modifiedContractCode, 0644)
		assert.NoError(t, err)

		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 100}}, nil)

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), serviceAddress.String())
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"Hello": contractCode, // Network has the correct version
			}

			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		// No prompter needed - auto-repairs regardless of flags
		mockPrompter := &mockPrompter{responses: []bool{}}

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:            logger,
			State:             state,
			SaveState:         true,
			TargetDir:         "",
			SkipDeployments:   true,
			SkipAlias:         true,
			SkipUpdatePrompts: true, // Should still auto-repair (no network change)
			dependencies:      make(map[string]config.Dependency),
			accountAliases:    make(map[string]map[string]flow.Address),
			pendingPrompts:    make([]pendingPrompt, 0),
			prompter:          mockPrompter,
			blockHeightCache:  make(map[string]uint64),
		}

		err = di.Install()
		// Should SUCCEED - auto-repaired even with --skip-update-prompts
		assert.NoError(t, err, "Should succeed even with --skip-update-prompts (no network change)")

		// Verify file WAS repaired
		fileContent, err := state.ReaderWriter().ReadFile(filePath)
		assert.NoError(t, err)
		assert.Contains(t, string(fileContent), "Hello, World!", "Should have correct version")
		assert.NotContains(t, string(fileContent), "HACKED", "Should not have hacked version")

		// Verify no prompts (auto-repair because network agrees with flow.json)
		assert.Equal(t, 0, mockPrompter.index, "Should not prompt when network agrees with flow.json")
	})

	t.Run("Already installed, outdated hash - user accepts", func(t *testing.T) {
		// Fresh state for this test
		_, state, _ := util.TestMocks(t)

		oldContractCode := []byte(`access(all) contract Hello { access(all) fun sayHello(): String { return "Hello, World! v1" } }`)
		newContractCode := []byte(`access(all) contract Hello { access(all) fun sayHello(): String { return "Hello, World! v2" } }`)

		// Calculate the old hash
		oldHash := sha256.New()
		oldHash.Write(oldContractCode)
		oldContractHash := hex.EncodeToString(oldHash.Sum(nil))

		// Calculate the new hash
		newHash := sha256.New()
		newHash.Write(newContractCode)
		newContractHash := hex.EncodeToString(newHash.Sum(nil))

		// Simulate a dependency that exists in flow.json with an OLD hash
		dep := config.Dependency{
			Name: "Hello",
			Source: config.Source{
				NetworkName:  "emulator",
				Address:      serviceAddress,
				ContractName: "Hello",
			},
			Hash: oldContractHash,
		}

		state.Dependencies().AddOrUpdate(dep)

		// Create the old file first
		filePath := fmt.Sprintf("imports/%s/Hello.cdc", serviceAddress.String())
		err := state.ReaderWriter().MkdirAll(filepath.Dir(filePath), 0755)
		assert.NoError(t, err)
		err = state.ReaderWriter().WriteFile(filePath, oldContractCode, 0644)
		assert.NoError(t, err)

		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 100}}, nil)

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), serviceAddress.String())
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"Hello": newContractCode,
			}

			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		// Mock prompter that returns true (user says "yes")
		mockPrompter := &mockPrompter{responses: []bool{true}}

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:            logger,
			State:             state,
			SaveState:         true,
			TargetDir:         "",
			SkipDeployments:   true,
			SkipAlias:         true,
			SkipUpdatePrompts: false,
			dependencies:      make(map[string]config.Dependency),
			accountAliases:    make(map[string]map[string]flow.Address),
			pendingPrompts:    make([]pendingPrompt, 0),
			blockHeightCache:  make(map[string]uint64),
			prompter:          mockPrompter,
		}

		err = di.Install()
		assert.NoError(t, err, "Failed to install dependencies")

		// Verify file WAS overwritten with new version
		fileContent, err := state.ReaderWriter().ReadFile(filePath)
		assert.NoError(t, err, "File should exist after accepting update")
		assert.NotNil(t, fileContent)
		assert.Contains(t, string(fileContent), "Hello, World! v2", "Should have the new contract version")
		assert.NotContains(t, string(fileContent), "v1", "Should not have old version")

		// Verify hash WAS updated
		updatedDep := state.Dependencies().ByName("Hello")
		assert.NotNil(t, updatedDep)
		assert.Equal(t, newContractHash, updatedDep.Hash, "Hash should be updated to new version")

		// Verify no warnings
		assert.Empty(t, di.logs.issues, "Should have no warnings when update is accepted")
	})

	t.Run("Already installed, outdated hash - user declines", func(t *testing.T) {
		// Fresh state for this test
		_, state, _ := util.TestMocks(t)

		oldContractCode := []byte(`access(all) contract Hello { access(all) fun sayHello(): String { return "Hello, World! v1" } }`)
		newContractCode := []byte(`access(all) contract Hello { access(all) fun sayHello(): String { return "Hello, World! v2" } }`)

		// Calculate the old hash
		oldHash := sha256.New()
		oldHash.Write(oldContractCode)
		oldContractHash := hex.EncodeToString(oldHash.Sum(nil))

		// Simulate a dependency that exists in flow.json with an OLD hash
		dep := config.Dependency{
			Name: "Hello",
			Source: config.Source{
				NetworkName:  "emulator",
				Address:      serviceAddress,
				ContractName: "Hello",
			},
			Hash: oldContractHash,
		}

		state.Dependencies().AddOrUpdate(dep)

		// Create the old file first
		filePath := fmt.Sprintf("imports/%s/Hello.cdc", serviceAddress.String())
		err := state.ReaderWriter().MkdirAll(filepath.Dir(filePath), 0755)
		assert.NoError(t, err)
		err = state.ReaderWriter().WriteFile(filePath, oldContractCode, 0644)
		assert.NoError(t, err)

		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 100}}, nil)

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), serviceAddress.String())
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"Hello": newContractCode,
			}

			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		// Mock prompter that returns false (user says "no")
		mockPrompter := &mockPrompter{responses: []bool{false}}

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:            logger,
			State:             state,
			SaveState:         true,
			TargetDir:         "",
			SkipDeployments:   true,
			SkipAlias:         true,
			SkipUpdatePrompts: false,
			dependencies:      make(map[string]config.Dependency),
			accountAliases:    make(map[string]map[string]flow.Address),
			pendingPrompts:    make([]pendingPrompt, 0),
			blockHeightCache:  make(map[string]uint64),
			prompter:          mockPrompter,
		}

		err = di.Install()
		assert.NoError(t, err, "Failed to install dependencies")

		// Verify file was NOT changed (still has v1)
		fileContent, err := state.ReaderWriter().ReadFile(filePath)
		assert.NoError(t, err, "File should still exist")
		assert.NotNil(t, fileContent)
		assert.Contains(t, string(fileContent), "Hello, World! v1", "Should still have the old version")
		assert.NotContains(t, string(fileContent), "v2", "Should not have new version")

		// Verify hash was NOT updated
		updatedDep := state.Dependencies().ByName("Hello")
		assert.NotNil(t, updatedDep)
		assert.Equal(t, oldContractHash, updatedDep.Hash, "Hash should remain at old version")

		// Verify no warnings (file exists, so no incomplete state)
		assert.Empty(t, di.logs.issues, "Should have no warnings when file exists")
	})

	t.Run("Already installed, outdated hash - user declines with modified file", func(t *testing.T) {
		// Fresh state for this test
		_, state, _ := util.TestMocks(t)

		oldContractCode := []byte(`access(all) contract Hello { access(all) fun sayHello(): String { return "Hello, World! v1" } }`)
		modifiedContractCode := []byte(`access(all) contract Hello { access(all) fun sayHello(): String { return "Hello, Modified!" } }`)
		newContractCode := []byte(`access(all) contract Hello { access(all) fun sayHello(): String { return "Hello, World! v2" } }`)

		// Calculate the old hash (of the converted contract)
		oldHash := sha256.New()
		oldHash.Write(oldContractCode)
		oldContractHash := hex.EncodeToString(oldHash.Sum(nil))

		// Simulate a dependency that exists in flow.json with an OLD hash
		dep := config.Dependency{
			Name: "Hello",
			Source: config.Source{
				NetworkName:  "emulator",
				Address:      serviceAddress,
				ContractName: "Hello",
			},
			Hash: oldContractHash,
		}

		state.Dependencies().AddOrUpdate(dep)

		// Create a MODIFIED file (different from both old and new versions)
		filePath := fmt.Sprintf("imports/%s/Hello.cdc", serviceAddress.String())
		err := state.ReaderWriter().MkdirAll(filepath.Dir(filePath), 0755)
		assert.NoError(t, err)
		err = state.ReaderWriter().WriteFile(filePath, modifiedContractCode, 0644)
		assert.NoError(t, err)

		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 100}}, nil)

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), serviceAddress.String())
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"Hello": newContractCode,
			}

			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		// Mock prompter that returns false (user says "no")
		mockPrompter := &mockPrompter{responses: []bool{false}}

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:            logger,
			State:             state,
			SaveState:         true,
			TargetDir:         "",
			SkipDeployments:   true,
			SkipAlias:         true,
			SkipUpdatePrompts: false,
			dependencies:      make(map[string]config.Dependency),
			accountAliases:    make(map[string]map[string]flow.Address),
			pendingPrompts:    make([]pendingPrompt, 0),
			blockHeightCache:  make(map[string]uint64),
			prompter:          mockPrompter,
		}

		err = di.Install()
		assert.Error(t, err, "Should fail when file hash doesn't match stored hash")
		assert.Contains(t, err.Error(), "local file has been modified", "Error should mention file modification")
		assert.Contains(t, err.Error(), "hash mismatch", "Error should mention hash mismatch")
	})

	t.Run("Already installed, outdated hash - update flag", func(t *testing.T) {
		// Fresh state for this test
		_, state, _ := util.TestMocks(t)

		oldContractCode := []byte(`access(all) contract Hello { access(all) fun sayHello(): String { return "Hello, World! v1" } }`)
		newContractCode := []byte(`access(all) contract Hello { access(all) fun sayHello(): String { return "Hello, World! v2" } }`)

		// Calculate the old hash
		oldHash := sha256.New()
		oldHash.Write(oldContractCode)
		oldContractHash := hex.EncodeToString(oldHash.Sum(nil))

		// Calculate the new hash
		newHash := sha256.New()
		newHash.Write(newContractCode)
		newContractHash := hex.EncodeToString(newHash.Sum(nil))

		// Simulate a dependency that exists in flow.json with an OLD hash
		dep := config.Dependency{
			Name: "Hello",
			Source: config.Source{
				NetworkName:  "emulator",
				Address:      serviceAddress,
				ContractName: "Hello",
			},
			Hash: oldContractHash,
		}

		state.Dependencies().AddOrUpdate(dep)

		// Create the old file first
		filePath := fmt.Sprintf("imports/%s/Hello.cdc", serviceAddress.String())
		err := state.ReaderWriter().MkdirAll(filepath.Dir(filePath), 0755)
		assert.NoError(t, err)
		err = state.ReaderWriter().WriteFile(filePath, oldContractCode, 0644)
		assert.NoError(t, err)

		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 100}}, nil)

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), serviceAddress.String())
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"Hello": newContractCode,
			}

			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		// No prompter needed - --update auto-accepts
		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:            logger,
			State:             state,
			SaveState:         true,
			TargetDir:         "",
			SkipDeployments:   true,
			SkipAlias:         true,
			SkipUpdatePrompts: false,
			Update:            true, // Auto-accept updates
			dependencies:      make(map[string]config.Dependency),
			accountAliases:    make(map[string]map[string]flow.Address),
			pendingPrompts:    make([]pendingPrompt, 0),
			prompter:          &mockPrompter{responses: []bool{}},
			blockHeightCache:  make(map[string]uint64),
		}

		err = di.Install()
		assert.NoError(t, err, "Should succeed with --update flag")

		// Verify file WAS overwritten with new version
		fileContent, err := state.ReaderWriter().ReadFile(filePath)
		assert.NoError(t, err, "File should exist after auto-update")
		assert.NotNil(t, fileContent)
		assert.Contains(t, string(fileContent), "Hello, World! v2", "Should have the new contract version")
		assert.NotContains(t, string(fileContent), "v1", "Should not have old version")

		// Verify hash WAS updated
		updatedDep := state.Dependencies().ByName("Hello")
		assert.NotNil(t, updatedDep)
		assert.Equal(t, newContractHash, updatedDep.Hash, "Hash should be updated to new version")

		// Verify success message logged
		assert.NotEmpty(t, di.logs.stateUpdates, "Should have state update messages")
		assert.Contains(t, di.logs.stateUpdates[0], "updated to latest version", "Should log update message")
	})
}

func TestDependencyInstallerAdd(t *testing.T) {

	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	serviceAcc, _ := state.EmulatorServiceAccount()
	serviceAddress := serviceAcc.Address

	t.Run("Success", func(t *testing.T) {
		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 100}}, nil)

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), serviceAcc.Address.String())
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				tests.ContractHelloString.Name: tests.ContractHelloString.Source,
			}

			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:           logger,
			State:            state,
			SaveState:        true,
			TargetDir:        "",
			SkipDeployments:  true,
			SkipAlias:        true,
			dependencies:     make(map[string]config.Dependency),
			blockHeightCache: make(map[string]uint64),
		}

		sourceStr := fmt.Sprintf("emulator://%s.%s", serviceAddress.String(), tests.ContractHelloString.Name)
		err := di.AddBySourceString(sourceStr)
		assert.NoError(t, err, "Failed to install dependencies")

		filePath := fmt.Sprintf("imports/%s/%s.cdc", serviceAddress.String(), tests.ContractHelloString.Name)
		fileContent, err := state.ReaderWriter().ReadFile(filePath)
		assert.NoError(t, err, "Failed to read generated file")
		assert.NotNil(t, fileContent)
	})

	t.Run("Success", func(t *testing.T) {
		gw := mocks.DefaultMockGateway()

		setupAccountMocks := func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), serviceAcc.Address.String())
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				tests.ContractHelloString.Name: tests.ContractHelloString.Source,
			}
			gw.GetAccountAtBlockHeight.Return(acc, nil)
		}

		gw.GetAccount.Run(setupAccountMocks)
		gw.GetAccountAtBlockHeight.Run(setupAccountMocks)

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:           logger,
			State:            state,
			SaveState:        true,
			TargetDir:        "",
			SkipDeployments:  true,
			SkipAlias:        true,
			dependencies:     make(map[string]config.Dependency),
			blockHeightCache: make(map[string]uint64),
		}

		dep := config.Dependency{
			Name: tests.ContractHelloString.Name,
			Source: config.Source{
				NetworkName:  "emulator",
				Address:      flow.HexToAddress(serviceAddress.String()),
				ContractName: tests.ContractHelloString.Name,
			},
		}
		err := di.Add(dep)
		assert.NoError(t, err, "Failed to install dependencies")

		filePath := fmt.Sprintf("imports/%s/%s.cdc", serviceAddress.String(), tests.ContractHelloString.Name)
		fileContent, err := state.ReaderWriter().ReadFile(filePath)
		assert.NoError(t, err, "Failed to read generated file")
		assert.NotNil(t, fileContent)
	})

	t.Run("Add by core contract name", func(t *testing.T) {
		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 100}}, nil)

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), "1654653399040a61")
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"FlowToken": []byte("access(all) contract FlowToken {}"),
			}

			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:           logger,
			State:            state,
			SaveState:        true,
			TargetDir:        "",
			SkipDeployments:  true,
			SkipAlias:        true,
			dependencies:     make(map[string]config.Dependency),
			blockHeightCache: make(map[string]uint64),
		}

		err := di.AddByCoreContractName("FlowToken")
		assert.NoError(t, err, "Failed to install dependencies")

		filePath := fmt.Sprintf("imports/%s/%s.cdc", "1654653399040a61", "FlowToken")
		fileContent, err := state.ReaderWriter().ReadFile(filePath)
		assert.NoError(t, err, "Failed to read generated file")
		assert.NotNil(t, fileContent)
		assert.Contains(t, string(fileContent), "contract FlowToken")
	})
}

func TestDependencyInstallerAddMany(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	serviceAcc, _ := state.EmulatorServiceAccount()
	serviceAddress := serviceAcc.Address.String()

	dependencies := []config.Dependency{
		{
			Name: "ContractOne",
			Source: config.Source{
				NetworkName:  "emulator",
				Address:      flow.HexToAddress(serviceAddress),
				ContractName: "ContractOne",
			},
		},
		{
			Name: "ContractTwo",
			Source: config.Source{
				NetworkName:  "emulator",
				Address:      flow.HexToAddress(serviceAddress),
				ContractName: "ContractTwo",
			},
		},
	}

	t.Run("AddMultipleDependencies", func(t *testing.T) {
		gw := mocks.DefaultMockGateway()
		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), serviceAddress)
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"ContractOne": []byte("access(all) contract ContractOne {}"),
				"ContractTwo": []byte("access(all) contract ContractTwo {}"),
			}
			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:           logger,
			State:            state,
			SaveState:        true,
			TargetDir:        "",
			SkipDeployments:  true,
			SkipAlias:        true,
			dependencies:     make(map[string]config.Dependency),
			blockHeightCache: make(map[string]uint64),
		}

		err := di.AddMany(dependencies)
		assert.NoError(t, err, "Failed to add multiple dependencies")

		for _, dep := range dependencies {
			filePath := fmt.Sprintf("imports/%s/%s.cdc", dep.Source.Address.String(), dep.Name)
			_, err := state.ReaderWriter().ReadFile(filePath)
			assert.NoError(t, err, fmt.Sprintf("Failed to read generated file for %s", dep.Name))
		}
	})
}

func TestTransitiveConflictAllowedWithMatchingAlias(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	// Pre-install Foo as a mainnet dependency and add alias for testnet to match incoming
	state.Dependencies().AddOrUpdate(config.Dependency{
		Name: "Foo",
		Source: config.Source{
			NetworkName:  config.MainnetNetwork.Name,
			Address:      flow.HexToAddress("0x0a"),
			ContractName: "Foo",
		},
	})
	// Ensure contract entry exists and add alias for testnet address 0x0b
	state.Contracts().AddDependencyAsContract(config.Dependency{
		Name: "Foo",
		Source: config.Source{
			NetworkName:  config.MainnetNetwork.Name,
			Address:      flow.HexToAddress("0x0a"),
			ContractName: "Foo",
		},
	}, config.MainnetNetwork.Name)
	c, _ := state.Contracts().ByName("Foo")
	c.Aliases.Add(config.TestnetNetwork.Name, flow.HexToAddress("0x0b"))

	// Gateways per network
	gwTestnet := mocks.DefaultMockGateway()
	gwTestnet.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 100}}, nil)
	gwMainnet := mocks.DefaultMockGateway()
	gwMainnet.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 100}}, nil)
	gwEmulator := mocks.DefaultMockGateway()
	gwEmulator.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 100}}, nil)

	// Addresses
	barAddr := flow.HexToAddress("0x0c")     // testnet address hosting Bar
	fooTestAddr := flow.HexToAddress("0x0b") // testnet Foo address (transitive)

	// Testnet GetAccountAtBlockHeight returns Bar at barAddr and Foo at fooTestAddr
	gwTestnet.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
		addr := args.Get(1).(flow.Address)
		switch addr.String() {
		case barAddr.String():
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"Bar": []byte("import Foo from 0x0b\naccess(all) contract Bar {}"),
			}
			gwTestnet.GetAccountAtBlockHeight.Return(acc, nil)
		case fooTestAddr.String():
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"Foo": []byte("access(all) contract Foo {}"),
			}
			gwTestnet.GetAccountAtBlockHeight.Return(acc, nil)
		default:
			gwTestnet.GetAccountAtBlockHeight.Return(nil, fmt.Errorf("not found"))
		}
	})

	// Mainnet/emulator not used for these addresses
	gwMainnet.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
		gwMainnet.GetAccountAtBlockHeight.Return(nil, fmt.Errorf("not found"))
	})
	gwEmulator.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
		gwEmulator.GetAccountAtBlockHeight.Return(nil, fmt.Errorf("not found"))
	})

	di := &DependencyInstaller{
		Gateways: map[string]gateway.Gateway{
			config.EmulatorNetwork.Name: gwEmulator.Mock,
			config.TestnetNetwork.Name:  gwTestnet.Mock,
			config.MainnetNetwork.Name:  gwMainnet.Mock,
		},
		Logger:           logger,
		State:            state,
		SaveState:        true,
		TargetDir:        "",
		SkipDeployments:  true,
		SkipAlias:        true,
		dependencies:     make(map[string]config.Dependency),
		prompter:         &mockPrompter{responses: []bool{}},
		blockHeightCache: make(map[string]uint64),
	}

	// Attempt to install Bar from testnet, which imports Foo from testnet transitively
	// With matching alias, this should be allowed (no error)
	err := di.AddBySourceString(fmt.Sprintf("%s://%s.%s", config.TestnetNetwork.Name, barAddr.String(), "Bar"))
	assert.NoError(t, err)
}

func TestTransitiveConflictErrorsWithoutAlias(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	// Pre-install Foo as a mainnet dependency WITHOUT an alias for testnet
	state.Dependencies().AddOrUpdate(config.Dependency{
		Name: "Foo",
		Source: config.Source{
			NetworkName:  config.MainnetNetwork.Name,
			Address:      flow.HexToAddress("0x0a"),
			ContractName: "Foo",
		},
	})
	state.Contracts().AddDependencyAsContract(config.Dependency{
		Name: "Foo",
		Source: config.Source{
			NetworkName:  config.MainnetNetwork.Name,
			Address:      flow.HexToAddress("0x0a"),
			ContractName: "Foo",
		},
	}, config.MainnetNetwork.Name)
	// NOTE: No alias added - this will cause a conflict

	// Gateways
	gwTestnet := mocks.DefaultMockGateway()
	gwTestnet.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 100}}, nil)

	// Addresses
	barAddr := flow.HexToAddress("0x0c")     // testnet address hosting Bar
	fooTestAddr := flow.HexToAddress("0x0b") // testnet Foo address (different from mainnet 0x0a)

	gwTestnet.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
		addr := args.Get(1).(flow.Address)
		switch addr.String() {
		case barAddr.String():
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"Bar": []byte("import Foo from 0x0b\naccess(all) contract Bar {}"),
			}
			gwTestnet.GetAccountAtBlockHeight.Return(acc, nil)
		case fooTestAddr.String():
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"Foo": []byte("access(all) contract Foo {}"),
			}
			gwTestnet.GetAccountAtBlockHeight.Return(acc, nil)
		default:
			gwTestnet.GetAccountAtBlockHeight.Return(nil, fmt.Errorf("not found"))
		}
	})

	di := &DependencyInstaller{
		Gateways: map[string]gateway.Gateway{
			config.TestnetNetwork.Name: gwTestnet.Mock,
		},
		Logger:           logger,
		State:            state,
		SkipDeployments:  true,
		SkipAlias:        true,
		dependencies:     make(map[string]config.Dependency),
		prompter:         &mockPrompter{responses: []bool{}},
		blockHeightCache: make(map[string]uint64),
	}

	// Attempt to install Bar from testnet, which imports Foo from testnet transitively
	// Without a matching alias, this should ERROR (naming conflict)
	err := di.AddBySourceString(fmt.Sprintf("%s://%s.%s", config.TestnetNetwork.Name, barAddr.String(), "Bar"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists with a different source")
	assert.Contains(t, err.Error(), "naming conflict")
}

func TestDependencyInstallerAliasTracking(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	serviceAcc, _ := state.EmulatorServiceAccount()
	serviceAddress := serviceAcc.Address

	t.Run("AutoApplyAliasForSameAccount", func(t *testing.T) {
		gw := mocks.DefaultMockGateway()

		// Mock the same account for both contracts
		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), serviceAcc.Address.String())
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"ContractOne": []byte("access(all) contract ContractOne {}"),
				"ContractTwo": []byte("access(all) contract ContractTwo {}"),
			}

			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:          logger,
			State:           state,
			SaveState:       true,
			TargetDir:       "",
			SkipDeployments: true,
			SkipAlias:       false,
			dependencies:    make(map[string]config.Dependency),
			accountAliases:  make(map[string]map[string]flow.Address),
			prompter:        &mockPrompter{responses: []bool{}},
		}

		dep1 := config.Dependency{
			Name: "ContractOne",
			Source: config.Source{
				NetworkName:  "mainnet",
				Address:      flow.HexToAddress(serviceAddress.String()),
				ContractName: "ContractOne",
			},
		}
		di.dependencies["mainnet://"+serviceAddress.String()+".ContractOne"] = dep1

		aliasAddress := flow.HexToAddress("0x1234567890abcdef")
		di.setAccountAlias(serviceAddress.String(), "testnet", aliasAddress)

		// Add second contract - this should automatically use the same alias
		dep2 := config.Dependency{
			Name: "ContractTwo",
			Source: config.Source{
				NetworkName:  "mainnet",
				Address:      flow.HexToAddress(serviceAddress.String()),
				ContractName: "ContractTwo",
			},
		}
		di.dependencies["mainnet://"+serviceAddress.String()+".ContractTwo"] = dep2

		existingAlias, exists := di.getAccountAlias(serviceAddress.String(), "testnet")
		assert.True(t, exists, "Alias should exist for the account")
		assert.Equal(t, aliasAddress, existingAlias, "Alias should match the stored value")

		accountAddr := di.getCurrentContractAccountAddress("ContractOne", "mainnet")
		assert.Equal(t, serviceAddress.String(), accountAddr, "Should return correct account address")
	})
}

func TestDependencyFlagsDeploymentAccount(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	testAccount, _ := accounts.NewEmulatorAccount(state.ReaderWriter(), crypto.ECDSA_P256, crypto.SHA3_256, "")
	testAccount.Name = "test-account"
	testAccount.Address = flow.HexToAddress("0x1234567890abcdef")
	state.Accounts().AddOrUpdate(testAccount)

	t.Run("Valid deployment account - skips prompt", func(t *testing.T) {
		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 100}}, nil)

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				tests.ContractHelloString.Name: tests.ContractHelloString.Source,
			}
			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:            logger,
			State:             state,
			SaveState:         true,
			TargetDir:         "",
			SkipDeployments:   false,
			SkipAlias:         true,
			DeploymentAccount: "test-account",
			dependencies:      make(map[string]config.Dependency),
			accountAliases:    make(map[string]map[string]flow.Address),
		}

		err := di.updateDependencyDeployment("TestContract")
		assert.NoError(t, err, "Should succeed with valid deployment account")

		deployment := state.Deployments().ByAccountAndNetwork("test-account", "emulator")
		assert.NotNil(t, deployment, "Deployment should be created with specified account")
		assert.Equal(t, "test-account", deployment.Account, "Deployment should use specified account")
		assert.Equal(t, "emulator", deployment.Network, "Deployment should use emulator network")

		found := false
		for _, contract := range deployment.Contracts {
			if contract.Name == "TestContract" {
				found = true
				break
			}
		}
		assert.True(t, found, "TestContract should be added to deployment")
	})

	t.Run("Valid deployment account with forced network", func(t *testing.T) {
		gw := mocks.DefaultMockGateway()

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:            logger,
			State:             state,
			SaveState:         true,
			TargetDir:         "",
			SkipDeployments:   false,
			SkipAlias:         true,
			DeploymentAccount: "test-account",
			dependencies:      make(map[string]config.Dependency),
			accountAliases:    make(map[string]map[string]flow.Address),
		}

		err := di.updateDependencyDeployment("DeFiActions", "emulator")
		assert.NoError(t, err, "Should succeed with valid deployment account and forced network")

		deployment := state.Deployments().ByAccountAndNetwork("test-account", "emulator")
		assert.NotNil(t, deployment, "Deployment should be created with specified account")
		assert.Equal(t, "test-account", deployment.Account, "Deployment should use specified account")
		assert.Equal(t, "emulator", deployment.Network, "Deployment should use forced network")

		found := false
		for _, contract := range deployment.Contracts {
			if contract.Name == "DeFiActions" {
				found = true
				break
			}
		}
		assert.True(t, found, "DeFiActions should be added to deployment")
	})

	t.Run("Invalid deployment account - returns error", func(t *testing.T) {
		gw := mocks.DefaultMockGateway()

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:            logger,
			State:             state,
			SaveState:         true,
			TargetDir:         "",
			SkipDeployments:   false,
			SkipAlias:         true,
			DeploymentAccount: "non-existent-account",
			dependencies:      make(map[string]config.Dependency),
			accountAliases:    make(map[string]map[string]flow.Address),
		}

		err := di.updateDependencyDeployment("TestContract")
		assert.Error(t, err, "Should fail with invalid deployment account")
		assert.Contains(t, err.Error(), "deployment account 'non-existent-account' not found in flow.json accounts")
	})

	t.Run("Empty deployment account - uses prompt behavior", func(t *testing.T) {
		gw := mocks.DefaultMockGateway()

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:            logger,
			State:             state,
			SaveState:         true,
			TargetDir:         "",
			SkipDeployments:   false,
			SkipAlias:         true,
			DeploymentAccount: "", // Empty - should use prompt behavior
			dependencies:      make(map[string]config.Dependency),
			accountAliases:    make(map[string]map[string]flow.Address),
		}

		// This test would normally call the prompt, but since we can't test interactive prompts easily,
		// we'll just verify that it doesn't error due to account validation
		// The prompt would return nil in non-interactive mode, which is handled gracefully
		err := di.updateDependencyDeployment("TestContract")
		assert.NoError(t, err, "Should not error when using prompt behavior")
	})
}

func TestDependencyFlagsIntegration(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	testAccount, _ := accounts.NewEmulatorAccount(state.ReaderWriter(), crypto.ECDSA_P256, crypto.SHA3_256, "")
	testAccount.Name = "test-account"
	testAccount.Address = flow.HexToAddress("0x1234567890abcdef")
	state.Accounts().AddOrUpdate(testAccount)

	t.Run("NewDependencyInstaller with deployment account flag", func(t *testing.T) {
		flags := DependencyFlags{
			skipDeployments:   false,
			skipAlias:         true,
			deploymentAccount: "test-account",
		}

		di, err := NewDependencyInstaller(logger, state, true, "", flags)
		assert.NoError(t, err, "Should create installer successfully")
		assert.Equal(t, "test-account", di.DeploymentAccount, "Should set deployment account from flags")
		assert.False(t, di.SkipDeployments, "Should set skip deployments from flags")
		assert.True(t, di.SkipAlias, "Should set skip alias from flags")
	})

	t.Run("DependencyFlags struct validation", func(t *testing.T) {
		flags := DependencyFlags{
			skipDeployments:   true,
			skipAlias:         false,
			deploymentAccount: "my-special-account",
		}

		assert.True(t, flags.skipDeployments, "Should handle skipDeployments flag")
		assert.False(t, flags.skipAlias, "Should handle skipAlias flag")
		assert.Equal(t, "my-special-account", flags.deploymentAccount, "Should handle deploymentAccount flag")
	})

	t.Run("DeFi Actions contracts deploy only on emulator", func(t *testing.T) {
		gw := mocks.DefaultMockGateway()

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:            logger,
			State:             state,
			SaveState:         true,
			TargetDir:         "",
			SkipDeployments:   false,
			SkipAlias:         true,
			DeploymentAccount: "test-account",
			dependencies:      make(map[string]config.Dependency),
			accountAliases:    make(map[string]map[string]flow.Address),
		}

		// Test updateDependencyDeployment with forced emulator network
		err := di.updateDependencyDeployment("DeFiActions", "emulator")
		assert.NoError(t, err, "Should succeed for DeFi Actions contract")

		// Verify deployment was added only on emulator
		deployment := state.Deployments().ByAccountAndNetwork("test-account", "emulator")
		assert.NotNil(t, deployment, "Deployment should be created on emulator network")
		assert.Equal(t, "emulator", deployment.Network, "Deployment should be on emulator network only")

		found := false
		for _, contract := range deployment.Contracts {
			if contract.Name == "DeFiActions" {
				found = true
				break
			}
		}
		assert.True(t, found, "DeFiActions should be added to emulator deployment")

		testnetDeployment := state.Deployments().ByAccountAndNetwork("test-account", "testnet")
		mainnetDeployment := state.Deployments().ByAccountAndNetwork("test-account", "mainnet")
		assert.Nil(t, testnetDeployment, "Should not create deployment on testnet")
		assert.Nil(t, mainnetDeployment, "Should not create deployment on mainnet")
	})
}

func TestAliasedImportHandling(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	gw := mocks.DefaultMockGateway()

	barAddr := flow.HexToAddress("0x0c")     // testnet address hosting Bar
	fooTestAddr := flow.HexToAddress("0x0b") // testnet Foo address

	t.Run("AliasedImportCreatesCanonicalMapping", func(t *testing.T) {
		// Testnet GetAccount returns Bar at barAddr and Foo at fooTestAddr
		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			switch addr.String() {
			case barAddr.String():
				acc := tests.NewAccountWithAddress(addr.String())
				// Bar imports Foo with an alias: import Foo as FooAlias from 0x0b
				acc.Contracts = map[string][]byte{
					"Bar": []byte("import Foo as FooAlias from 0x0b\naccess(all) contract Bar {}"),
				}
				gw.GetAccountAtBlockHeight.Return(acc, nil)
			case fooTestAddr.String():
				acc := tests.NewAccountWithAddress(addr.String())
				acc.Contracts = map[string][]byte{
					"Foo": []byte("access(all) contract Foo {}"),
				}
				gw.GetAccountAtBlockHeight.Return(acc, nil)
			default:
				gw.GetAccount.Return(nil, fmt.Errorf("not found"))
			}
		})

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger:           logger,
			State:            state,
			SaveState:        true,
			TargetDir:        "",
			SkipDeployments:  true,
			SkipAlias:        true,
			dependencies:     make(map[string]config.Dependency),
			blockHeightCache: make(map[string]uint64),
		}

		err := di.AddBySourceString(fmt.Sprintf("%s://%s.%s", config.TestnetNetwork.Name, barAddr.String(), "Bar"))
		assert.NoError(t, err)

		barDep := state.Dependencies().ByName("Bar")
		assert.NotNil(t, barDep, "Bar dependency should exist")

		fooAliasDep := state.Dependencies().ByName("FooAlias")
		assert.NotNil(t, fooAliasDep, "FooAlias dependency should exist")
		assert.Equal(t, "Foo", fooAliasDep.Source.ContractName, "Source ContractName should be the actual contract name (Foo)")

		fooAliasContract, err := state.Contracts().ByName("FooAlias")
		assert.NoError(t, err, "FooAlias contract should exist")
		assert.Equal(t, "Foo", fooAliasContract.Canonical, "Canonical should be set to Foo")

		filePath := fmt.Sprintf("imports/%s/Foo.cdc", fooTestAddr.String())
		fileContent, err := state.ReaderWriter().ReadFile(filePath)
		assert.NoError(t, err, "Contract file should exist at imports/address/Foo.cdc")
		assert.NotNil(t, fileContent)
	})
}

func TestDependencyInstallerWithAlias(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	serviceAcc, _ := state.EmulatorServiceAccount()
	serviceAddress := serviceAcc.Address

	t.Run("AddBySourceStringWithName", func(t *testing.T) {
		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 100}}, nil)

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), serviceAddress.String())
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"NumberFormatter": []byte("access(all) contract NumberFormatter {}"),
			}
			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
			},
			Logger:           logger,
			State:            state,
			SaveState:        true,
			TargetDir:        "",
			SkipDeployments:  true,
			SkipAlias:        true,
			Name:             "NumberFormatterCustom",
			dependencies:     make(map[string]config.Dependency),
			blockHeightCache: make(map[string]uint64),
		}

		err := di.AddBySourceString(fmt.Sprintf("%s://%s.%s", config.EmulatorNetwork.Name, serviceAddress.String(), "NumberFormatter"))
		assert.NoError(t, err, "Failed to add dependency with import alias")

		// Check that the dependency was added with the import alias name
		dep := state.Dependencies().ByName("NumberFormatterCustom")
		assert.NotNil(t, dep, "Dependency should exist with import alias name")
		assert.Equal(t, "NumberFormatter", dep.Source.ContractName, "Source ContractName should be the actual contract name")
		assert.Equal(t, "NumberFormatter", dep.Canonical, "Canonical should be set to the actual contract name for import aliasing")

		// Check that the contract was added with canonical field for Cadence import aliasing
		contract, err := state.Contracts().ByName("NumberFormatterCustom")
		assert.NoError(t, err, "Contract should exist")
		assert.Equal(t, "NumberFormatter", contract.Canonical, "Contract Canonical should be set for import aliasing")

		// Check that the file was created with the actual contract name
		filePath := fmt.Sprintf("imports/%s/NumberFormatter.cdc", serviceAddress.String())
		fileContent, err := state.ReaderWriter().ReadFile(filePath)
		assert.NoError(t, err, "Contract file should exist at imports/address/NumberFormatter.cdc")
		assert.NotNil(t, fileContent)
	})

	t.Run("AddByCoreContractNameWithName", func(t *testing.T) {
		// Mock the gateway to return FlowToken contract
		gw := mocks.DefaultMockGateway()
		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"FlowToken": []byte("access(all) contract FlowToken {}"),
			}
			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.MainnetNetwork.Name: gw.Mock,
			},
			Logger:           logger,
			State:            state,
			SaveState:        true,
			TargetDir:        "",
			SkipDeployments:  true,
			SkipAlias:        true,
			Name:             "FlowTokenCustom",
			dependencies:     make(map[string]config.Dependency),
			blockHeightCache: make(map[string]uint64),
		}

		err := di.AddByCoreContractName("FlowToken")
		assert.NoError(t, err, "Failed to add core contract with import alias")

		// Check that the dependency was added with the import alias name
		dep := state.Dependencies().ByName("FlowTokenCustom")
		assert.NotNil(t, dep, "Dependency should exist with import alias name")
		assert.Equal(t, "FlowToken", dep.Source.ContractName, "Source ContractName should be FlowToken")
		assert.Equal(t, "FlowToken", dep.Canonical, "Canonical should be set to FlowToken for import aliasing")
	})

	t.Run("AddAllByNetworkAddressWithNameError", func(t *testing.T) {
		// This test doesn't need gateways since it returns an error before making any gateway calls
		di := &DependencyInstaller{
			Logger:    logger,
			State:     state,
			SaveState: true,
			TargetDir: "",
			Name:      "SomeName",
		}

		err := di.AddAllByNetworkAddress(fmt.Sprintf("%s://%s", config.EmulatorNetwork.Name, serviceAddress.String()))
		assert.Error(t, err, "Should error when using --name with network://address format")
		assert.Contains(t, err.Error(), "--name flag is not supported when installing all contracts", "Error message should mention name flag limitation")
	})
}

func TestBlockHeightPinning(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	serviceAddress := flow.HexToAddress("f8d6e0586b0a20c7")

	t.Run("NewDependencyGetsPinnedToCurrentBlockHeight", func(t *testing.T) {
		_, state, _ := util.TestMocks(t)

		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{
			BlockHeader: flow.BlockHeader{Height: 12345},
		}, nil)

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			acc := tests.NewAccountWithAddress(args.Get(1).(flow.Address).String())
			acc.Contracts = map[string][]byte{
				"MyContract": []byte("access(all) contract MyContract {}"),
			}
			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		di := &DependencyInstaller{
			Gateways:         map[string]gateway.Gateway{config.EmulatorNetwork.Name: gw.Mock},
			Logger:           logger,
			State:            state,
			SkipDeployments:  true,
			SkipAlias:        true,
			dependencies:     make(map[string]config.Dependency),
			logs:             categorizedLogs{},
			prompter:         &mockPrompter{responses: []bool{}},
			blockHeightCache: make(map[string]uint64),
		}

		dep := config.Dependency{
			Name: "MyContract",
			Source: config.Source{
				NetworkName:  config.EmulatorNetwork.Name,
				Address:      serviceAddress,
				ContractName: "MyContract",
			},
		}

		err := di.Add(dep)
		assert.NoError(t, err)

		savedDep := state.Dependencies().ByName("MyContract")
		assert.Equal(t, uint64(12345), savedDep.BlockHeight)
	})

	t.Run("OldFormatDependencyAutoMigrates", func(t *testing.T) {
		_, state, _ := util.TestMocks(t)

		contractCode := []byte("access(all) contract LegacyContract {}")
		oldDep := config.Dependency{
			Name:        "LegacyContract",
			BlockHeight: 0,
			Hash:        computeHash(contractCode),
			Source: config.Source{
				NetworkName:  config.EmulatorNetwork.Name,
				Address:      serviceAddress,
				ContractName: "LegacyContract",
			},
		}
		state.Dependencies().AddOrUpdate(oldDep)

		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{
			BlockHeader: flow.BlockHeader{Height: 55555},
		}, nil)

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			acc := tests.NewAccountWithAddress(serviceAddress.String())
			acc.Contracts = map[string][]byte{
				"LegacyContract": contractCode,
			}
			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		di := &DependencyInstaller{
			Gateways:         map[string]gateway.Gateway{config.EmulatorNetwork.Name: gw.Mock},
			Logger:           logger,
			State:            state,
			SkipDeployments:  true,
			SkipAlias:        true,
			dependencies:     make(map[string]config.Dependency),
			logs:             categorizedLogs{},
			prompter:         &mockPrompter{responses: []bool{}},
			blockHeightCache: make(map[string]uint64),
		}

		err := di.Add(oldDep)
		assert.NoError(t, err)

		migratedDep := state.Dependencies().ByName("LegacyContract")
		assert.NotNil(t, migratedDep)
		assert.Equal(t, uint64(55555), migratedDep.BlockHeight)
	})

	t.Run("FrozenDependencyUsesHistoricalBlockHeight", func(t *testing.T) {
		_, state, _ := util.TestMocks(t)

		contractCode := []byte("access(all) contract OldVersion {}")
		frozenDep := config.Dependency{
			Name:        "FrozenContract",
			BlockHeight: 10000,
			Hash:        computeHash(contractCode),
			Source: config.Source{
				NetworkName:  config.EmulatorNetwork.Name,
				Address:      serviceAddress,
				ContractName: "FrozenContract",
			},
		}
		state.Dependencies().AddOrUpdate(frozenDep)

		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{
			BlockHeader: flow.BlockHeader{Height: 99999},
		}, nil)

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			assert.Equal(t, uint64(10000), args.Get(2).(uint64))
			acc := tests.NewAccountWithAddress(serviceAddress.String())
			acc.Contracts = map[string][]byte{"FrozenContract": contractCode}
			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		di := &DependencyInstaller{
			Gateways:         map[string]gateway.Gateway{config.EmulatorNetwork.Name: gw.Mock},
			Logger:           logger,
			State:            state,
			SkipDeployments:  true,
			SkipAlias:        true,
			dependencies:     make(map[string]config.Dependency),
			logs:             categorizedLogs{},
			prompter:         &mockPrompter{responses: []bool{}},
			blockHeightCache: make(map[string]uint64),
		}

		err := di.Add(frozenDep)
		assert.NoError(t, err)

		gw.Mock.AssertCalled(t, "GetAccountAtBlockHeight", mock.Anything, serviceAddress, uint64(10000))
	})

	t.Run("ChangedContractPromptsUpdate", func(t *testing.T) {
		_, state, _ := util.TestMocks(t)

		oldCode := []byte("access(all) contract OldVersion {}")
		newCode := []byte("access(all) contract NewVersion {}")

		existingDep := config.Dependency{
			Name:        "UpdatableContract",
			BlockHeight: 0,
			Hash:        computeHash(oldCode),
			Source: config.Source{
				NetworkName:  config.EmulatorNetwork.Name,
				Address:      serviceAddress,
				ContractName: "UpdatableContract",
			},
		}
		state.Dependencies().AddOrUpdate(existingDep)

		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{
			BlockHeader: flow.BlockHeader{Height: 50000},
		}, nil)

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			acc := tests.NewAccountWithAddress(serviceAddress.String())
			acc.Contracts = map[string][]byte{"UpdatableContract": newCode}
			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		prompter := &mockPrompter{responses: []bool{false}}

		di := &DependencyInstaller{
			Gateways:         map[string]gateway.Gateway{config.EmulatorNetwork.Name: gw.Mock},
			Logger:           logger,
			State:            state,
			SkipDeployments:  true,
			SkipAlias:        true,
			dependencies:     make(map[string]config.Dependency),
			logs:             categorizedLogs{},
			accountAliases:   make(map[string]map[string]flow.Address),
			pendingPrompts:   make([]pendingPrompt, 0),
			prompter:         prompter,
			blockHeightCache: make(map[string]uint64),
		}

		err := di.Install()
		if err != nil {
			assert.Contains(t, err.Error(), "file does not exist")
		}

		assert.Equal(t, 1, prompter.index)
	})

	t.Run("PinnedDependencyWithUpdateFlagAutoUpdates", func(t *testing.T) {
		_, state, _ := util.TestMocks(t)

		oldCode := []byte("access(all) contract OldVersion {}")
		newCode := []byte("access(all) contract NewVersion {}")

		pinnedDep := config.Dependency{
			Name:        "AutoUpdateContract",
			BlockHeight: 10000,
			Hash:        computeHash(oldCode),
			Source: config.Source{
				NetworkName:  config.EmulatorNetwork.Name,
				Address:      serviceAddress,
				ContractName: "AutoUpdateContract",
			},
		}
		state.Dependencies().AddOrUpdate(pinnedDep)

		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{
			BlockHeader: flow.BlockHeader{Height: 99999},
		}, nil)

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			acc := tests.NewAccountWithAddress(serviceAddress.String())
			acc.Contracts = map[string][]byte{"AutoUpdateContract": newCode}
			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		di := &DependencyInstaller{
			Gateways:         map[string]gateway.Gateway{config.EmulatorNetwork.Name: gw.Mock},
			Logger:           logger,
			State:            state,
			SkipDeployments:  true,
			SkipAlias:        true,
			Update:           true,
			dependencies:     make(map[string]config.Dependency),
			logs:             categorizedLogs{},
			prompter:         &mockPrompter{responses: []bool{}},
			blockHeightCache: make(map[string]uint64),
		}

		err := di.Add(pinnedDep)
		assert.NoError(t, err)

		updatedDep := state.Dependencies().ByName("AutoUpdateContract")
		assert.Equal(t, uint64(99999), updatedDep.BlockHeight)
		assert.Equal(t, computeHash(newCode), updatedDep.Hash)
	})

	t.Run("OutdatedPinWithoutLocalFilePromptsAndUpdatesBlockHeight", func(t *testing.T) {
		_, state, _ := util.TestMocks(t)

		oldCode := []byte("access(all) contract OldVersion {}")
		newCode := []byte("access(all) contract NewVersion {}")

		// flow.json has outdated pin (no local file yet, e.g., after git clone)
		outdatedDep := config.Dependency{
			Name:        "OutdatedContract",
			BlockHeight: 10000,
			Hash:        computeHash(oldCode),
			Source: config.Source{
				NetworkName:  config.EmulatorNetwork.Name,
				Address:      serviceAddress,
				ContractName: "OutdatedContract",
			},
		}
		state.Dependencies().AddOrUpdate(outdatedDep)

		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{
			BlockHeader: flow.BlockHeader{Height: 99999},
		}, nil)

		// At pinned block 10000, contract has changed
		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			acc := tests.NewAccountWithAddress(serviceAddress.String())
			acc.Contracts = map[string][]byte{"OutdatedContract": newCode}
			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		// User accepts update
		prompter := &mockPrompter{responses: []bool{true}}

		di := &DependencyInstaller{
			Gateways:         map[string]gateway.Gateway{config.EmulatorNetwork.Name: gw.Mock},
			Logger:           logger,
			State:            state,
			SkipDeployments:  true,
			SkipAlias:        true,
			dependencies:     make(map[string]config.Dependency),
			logs:             categorizedLogs{},
			accountAliases:   make(map[string]map[string]flow.Address),
			pendingPrompts:   make([]pendingPrompt, 0),
			prompter:         prompter,
			blockHeightCache: make(map[string]uint64),
		}

		err := di.Install()
		assert.NoError(t, err)

		// Should have prompted
		assert.Equal(t, 1, prompter.index)

		// Should update to latest block height and new hash
		updatedDep := state.Dependencies().ByName("OutdatedContract")
		assert.Equal(t, uint64(99999), updatedDep.BlockHeight)
		assert.Equal(t, computeHash(newCode), updatedDep.Hash)
	})

	t.Run("SkipUpdatePromptsWithoutFileInstallsOnChainVersion", func(t *testing.T) {
		_, state, _ := util.TestMocks(t)

		onChainCode := []byte("access(all) contract ChangedContract {}")
		pinnedDep := config.Dependency{
			Name:        "ChangedContract",
			BlockHeight: 10000,
			Hash:        "old_hash_different",
			Source: config.Source{
				NetworkName:  config.EmulatorNetwork.Name,
				Address:      serviceAddress,
				ContractName: "ChangedContract",
			},
		}
		state.Dependencies().AddOrUpdate(pinnedDep)

		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{
			BlockHeader: flow.BlockHeader{Height: 50000},
		}, nil)

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			acc := tests.NewAccountWithAddress(serviceAddress.String())
			acc.Contracts = map[string][]byte{"ChangedContract": onChainCode}
			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		di := &DependencyInstaller{
			Gateways:          map[string]gateway.Gateway{config.EmulatorNetwork.Name: gw.Mock},
			Logger:            logger,
			State:             state,
			SkipDeployments:   true,
			SkipAlias:         true,
			SkipUpdatePrompts: true,
			dependencies:      make(map[string]config.Dependency),
			logs:              categorizedLogs{},
			prompter:          &mockPrompter{responses: []bool{}},
			blockHeightCache:  make(map[string]uint64),
		}

		err := di.Add(pinnedDep)
		assert.NoError(t, err)

		savedDep := state.Dependencies().ByName("ChangedContract")
		assert.Equal(t, computeHash(onChainCode), savedDep.Hash)
	})

	t.Run("BlockHeightFetchFailureReturnsError", func(t *testing.T) {
		_, state, _ := util.TestMocks(t)

		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(nil, fmt.Errorf("network error"))

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			acc := tests.NewAccountWithAddress(serviceAddress.String())
			acc.Contracts = map[string][]byte{
				"TestContract": []byte("access(all) contract TestContract {}"),
			}
			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		di := &DependencyInstaller{
			Gateways:         map[string]gateway.Gateway{config.EmulatorNetwork.Name: gw.Mock},
			Logger:           logger,
			State:            state,
			SkipDeployments:  true,
			SkipAlias:        true,
			dependencies:     make(map[string]config.Dependency),
			logs:             categorizedLogs{},
			prompter:         &mockPrompter{responses: []bool{}},
			blockHeightCache: make(map[string]uint64),
		}

		dep := config.Dependency{
			Name: "TestContract",
			Source: config.Source{
				NetworkName:  config.EmulatorNetwork.Name,
				Address:      serviceAddress,
				ContractName: "TestContract",
			},
		}

		err := di.Add(dep)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get latest block height")
	})

	t.Run("BlockHeightCachedAcrossMultipleDependencies", func(t *testing.T) {
		_, state, _ := util.TestMocks(t)

		callCount := 0
		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Run(func(args mock.Arguments) {
			callCount++
			// Simulate blockchain progressing: each call returns a higher block
			block := &flow.Block{BlockHeader: flow.BlockHeader{Height: uint64(10000 + callCount*10)}}
			gw.GetLatestBlock.Return(block, nil)
		})

		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"ContractA": []byte("access(all) contract ContractA {}"),
				"ContractB": []byte("access(all) contract ContractB {}"),
			}
			gw.GetAccountAtBlockHeight.Return(acc, nil)
		})

		di := &DependencyInstaller{
			Gateways:         map[string]gateway.Gateway{config.EmulatorNetwork.Name: gw.Mock},
			Logger:           logger,
			State:            state,
			SkipDeployments:  true,
			SkipAlias:        true,
			dependencies:     make(map[string]config.Dependency),
			logs:             categorizedLogs{},
			prompter:         &mockPrompter{responses: []bool{}},
			blockHeightCache: make(map[string]uint64),
		}

		depA := config.Dependency{
			Name: "ContractA",
			Source: config.Source{
				NetworkName:  config.EmulatorNetwork.Name,
				Address:      serviceAddress,
				ContractName: "ContractA",
			},
		}

		depB := config.Dependency{
			Name: "ContractB",
			Source: config.Source{
				NetworkName:  config.EmulatorNetwork.Name,
				Address:      serviceAddress,
				ContractName: "ContractB",
			},
		}

		err := di.Add(depA)
		assert.NoError(t, err)

		err = di.Add(depB)
		assert.NoError(t, err)

		// Verify GetLatestBlock was called only ONCE (cached for second dependency)
		assert.Equal(t, 1, callCount, "GetLatestBlock should be called only once per network")

		savedDepA := state.Dependencies().ByName("ContractA")
		savedDepB := state.Dependencies().ByName("ContractB")

		assert.NotNil(t, savedDepA)
		assert.NotNil(t, savedDepB)

		// Both deps should have THE SAME block height (10010 from first call)
		assert.Equal(t, uint64(10010), savedDepA.BlockHeight, "ContractA should be pinned to first fetch")
		assert.Equal(t, uint64(10010), savedDepB.BlockHeight, "ContractB should reuse cached block height")
		assert.Equal(t, savedDepA.BlockHeight, savedDepB.BlockHeight, "All deps in same install should have same block height")
	})

	t.Run("PreSporkBlockHeightWithMatchingHashUpdatesMetadata", func(t *testing.T) {
		_, state, _ := util.TestMocks(t)

		contractCode := []byte("access(all) contract TestContract { access(all) let name: String; init() { self.name = \"Test\" } }")

		// Add an existing dependency with a pre-spork block height but correct hash
		existingDep := config.Dependency{
			Name: "TestContract",
			Source: config.Source{
				NetworkName:  config.EmulatorNetwork.Name,
				Address:      serviceAddress,
				ContractName: "TestContract",
			},
			BlockHeight: 138158854,                 // Pre-spork block height
			Hash:        computeHash(contractCode), // Hash matches current on-chain code
		}
		state.Dependencies().AddOrUpdate(existingDep)

		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 280224020}}, nil)

		// Simulate spork error for old block height, success for current block height
		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			requestedHeight := args.Get(2).(uint64) // arg 0 = ctx, arg 1 = address, arg 2 = blockHeight
			if requestedHeight == 138158854 {
				// Old pre-spork block → error
				gw.GetAccountAtBlockHeight.Return(nil, fmt.Errorf("not found: block height 138158854 is less than the spork root block height 280224020"))
			} else if requestedHeight == 280224020 {
				// Current block → success
				acc := tests.NewAccountWithAddress(serviceAddress.String())
				acc.Contracts = map[string][]byte{
					"TestContract": contractCode,
				}
				gw.GetAccountAtBlockHeight.Return(acc, nil)
			}
		})

		di := &DependencyInstaller{
			Gateways:         map[string]gateway.Gateway{config.EmulatorNetwork.Name: gw.Mock},
			Logger:           logger,
			State:            state,
			SkipDeployments:  true,
			SkipAlias:        true,
			Update:           false, // NO update flag - but should succeed because hash matches
			dependencies:     make(map[string]config.Dependency),
			logs:             categorizedLogs{},
			prompter:         &mockPrompter{responses: []bool{}},
			blockHeightCache: make(map[string]uint64),
		}

		dep := config.Dependency{
			Name: "TestContract",
			Source: config.Source{
				NetworkName:  config.EmulatorNetwork.Name,
				Address:      serviceAddress,
				ContractName: "TestContract",
			},
		}

		err := di.Add(dep)
		assert.NoError(t, err)

		// Verify the block height was updated (metadata fix)
		savedDep := state.Dependencies().ByName("TestContract")
		assert.NotNil(t, savedDep)
		assert.Equal(t, uint64(280224020), savedDep.BlockHeight, "Block height should be updated")
		assert.Equal(t, computeHash(contractCode), savedDep.Hash, "Hash should remain the same")
	})

	t.Run("PreSporkBlockHeightWithMismatchedHashAndSkipUpdatePromptsErrors", func(t *testing.T) {
		_, state, _ := util.TestMocks(t)

		oldCode := []byte("access(all) contract TestContract { access(all) let name: String; init() { self.name = \"OldVersion\" } }")
		newCode := []byte("access(all) contract TestContract { access(all) let name: String; init() { self.name = \"NewVersion\" } }")

		// Add an existing dependency with a pre-spork block height and old hash
		existingDep := config.Dependency{
			Name: "TestContract",
			Source: config.Source{
				NetworkName:  config.EmulatorNetwork.Name,
				Address:      serviceAddress,
				ContractName: "TestContract",
			},
			BlockHeight: 138158854, // Pre-spork block height
			Hash:        computeHash(oldCode),
		}
		state.Dependencies().AddOrUpdate(existingDep)

		// Create the old file matching the stored hash
		filePath := fmt.Sprintf("imports/%s/TestContract.cdc", serviceAddress.String())
		err := state.ReaderWriter().MkdirAll(filepath.Dir(filePath), 0755)
		assert.NoError(t, err)
		err = state.ReaderWriter().WriteFile(filePath, oldCode, 0644)
		assert.NoError(t, err)

		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 280224020}}, nil)

		// Simulate pre-spork error then success at current block with NEW hash
		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			requestedHeight := args.Get(2).(uint64)
			if requestedHeight == 138158854 {
				// Old pre-spork block → error
				gw.GetAccountAtBlockHeight.Return(nil, fmt.Errorf("not found: block height 138158854 is less than the spork root block height 280224020"))
			} else if requestedHeight == 280224020 {
				// Current block → success with NEW code
				acc := tests.NewAccountWithAddress(serviceAddress.String())
				acc.Contracts = map[string][]byte{
					"TestContract": newCode,
				}
				gw.GetAccountAtBlockHeight.Return(acc, nil)
			}
		})

		di := &DependencyInstaller{
			Gateways:          map[string]gateway.Gateway{config.EmulatorNetwork.Name: gw.Mock},
			Logger:            logger,
			State:             state,
			SkipDeployments:   true,
			SkipAlias:         true,
			SkipUpdatePrompts: true, // Want to keep frozen, but can't!
			dependencies:      make(map[string]config.Dependency),
			logs:              categorizedLogs{},
			prompter:          &mockPrompter{responses: []bool{}},
			blockHeightCache:  make(map[string]uint64),
		}

		dep := config.Dependency{
			Name: "TestContract",
			Source: config.Source{
				NetworkName:  config.EmulatorNetwork.Name,
				Address:      serviceAddress,
				ContractName: "TestContract",
			},
		}

		err = di.Add(dep)
		// Should ERROR: pre-spork block not accessible, network has different hash, can't keep frozen
		assert.Error(t, err, "Should error when trying to keep frozen with pre-spork block and hash mismatch")
		assert.Contains(t, err.Error(), "cannot keep frozen", "Error should mention inability to freeze")
		assert.Contains(t, err.Error(), "138158854", "Error should mention the old block height")
		assert.Contains(t, err.Error(), "280224020", "Error should mention the new block height")
		assert.Contains(t, err.Error(), "no longer accessible", "Error should explain block is not accessible")
	})

	t.Run("PreSporkBlockHeightWithMismatchedHashRequiresUpdateFlag", func(t *testing.T) {
		_, state, _ := util.TestMocks(t)

		// Add an existing dependency with a pre-spork block height
		existingDep := config.Dependency{
			Name: "TestContract",
			Source: config.Source{
				NetworkName:  config.EmulatorNetwork.Name,
				Address:      serviceAddress,
				ContractName: "TestContract",
			},
			BlockHeight: 138158854, // Pre-spork block height
			Hash:        "oldhash",
		}
		state.Dependencies().AddOrUpdate(existingDep)

		contractCode := []byte("access(all) contract TestContract { access(all) let name: String; init() { self.name = \"Test\" } }")

		gw := mocks.DefaultMockGateway()
		gw.GetLatestBlock.Return(&flow.Block{BlockHeader: flow.BlockHeader{Height: 280224020}}, nil)

		// Track calls to GetAccountAtBlockHeight
		callCount := 0
		gw.GetAccountAtBlockHeight.Run(func(args mock.Arguments) {
			callCount++
			requestedHeight := args.Get(2).(uint64) // arg 0 = ctx, arg 1 = address, arg 2 = blockHeight
			if requestedHeight == 138158854 {
				// Old pre-spork block → error
				gw.GetAccountAtBlockHeight.Return(nil, fmt.Errorf("not found: block height 138158854 is less than the spork root block height 280224020"))
			} else if requestedHeight == 280224020 {
				// Current block → success
				acc := tests.NewAccountWithAddress(serviceAddress.String())
				acc.Contracts = map[string][]byte{
					"TestContract": contractCode,
				}
				gw.GetAccountAtBlockHeight.Return(acc, nil)
			}
		})

		di := &DependencyInstaller{
			Gateways:         map[string]gateway.Gateway{config.EmulatorNetwork.Name: gw.Mock},
			Logger:           logger,
			State:            state,
			SkipDeployments:  true,
			SkipAlias:        true,
			Update:           true, // WITH update flag - should succeed
			dependencies:     make(map[string]config.Dependency),
			logs:             categorizedLogs{},
			prompter:         &mockPrompter{responses: []bool{}},
			blockHeightCache: make(map[string]uint64),
		}

		dep := config.Dependency{
			Name: "TestContract",
			Source: config.Source{
				NetworkName:  config.EmulatorNetwork.Name,
				Address:      serviceAddress,
				ContractName: "TestContract",
			},
		}

		err := di.Add(dep)
		assert.NoError(t, err)

		// Verify that GetAccountAtBlockHeight was called only once
		// With --update flag, we skip trying the old block and go straight to latest
		assert.Equal(t, 1, callCount, "GetAccountAtBlockHeight should be called once (--update skips old block, goes directly to latest)")

		// Verify the dependency was updated with latest version
		savedDep := state.Dependencies().ByName("TestContract")
		assert.NotNil(t, savedDep)
		assert.Equal(t, uint64(280224020), savedDep.BlockHeight, "Should be updated to current block height")
		assert.NotEqual(t, "oldhash", savedDep.Hash, "Hash should be updated")
		assert.Equal(t, computeHash(contractCode), savedDep.Hash, "Hash should match the new contract code")
	})

	t.Run("AliasedContractSkipsRediscovery", func(t *testing.T) {
		_, state, _ := util.TestMocks(t)

		mainnetAddr := flow.HexToAddress("0xf233dcee88fe0abe")
		testnetAddr := flow.HexToAddress("0x9a0766d93b6608b7")

		// Add Burner as mainnet dependency
		existingBurner := config.Dependency{
			Name: "Burner",
			Source: config.Source{
				NetworkName:  config.MainnetNetwork.Name,
				Address:      mainnetAddr,
				ContractName: "Burner",
			},
			BlockHeight: 95000000,
			Hash:        "existinghash",
		}
		state.Dependencies().AddOrUpdate(existingBurner)

		// Add the contract entry with aliases
		state.Contracts().AddDependencyAsContract(existingBurner, config.MainnetNetwork.Name)
		c, _ := state.Contracts().ByName("Burner")
		c.Aliases.Add(config.TestnetNetwork.Name, testnetAddr)

		di := &DependencyInstaller{
			Gateways:         map[string]gateway.Gateway{},
			Logger:           logger,
			State:            state,
			SkipDeployments:  true,
			SkipAlias:        true,
			dependencies:     make(map[string]config.Dependency),
			logs:             categorizedLogs{},
			prompter:         &mockPrompter{responses: []bool{}},
			blockHeightCache: make(map[string]uint64),
		}

		// Discover Burner via testnet alias (transitive import scenario)
		depViaTestnet := config.Dependency{
			Name: "Burner",
			Source: config.Source{
				NetworkName:  config.TestnetNetwork.Name,
				Address:      testnetAddr,
				ContractName: "Burner",
			},
		}

		err := di.Add(depViaTestnet)
		assert.NoError(t, err)

		// Verify: Burner should remain unchanged (alias rediscovery just skips)
		savedDep := state.Dependencies().ByName("Burner")
		assert.NotNil(t, savedDep)
		assert.Equal(t, config.MainnetNetwork.Name, savedDep.Source.NetworkName, "Source should remain mainnet")
		assert.Equal(t, mainnetAddr, savedDep.Source.Address, "Address should remain mainnet")
		assert.Equal(t, uint64(95000000), savedDep.BlockHeight, "Block height should remain unchanged")
		assert.Equal(t, "existinghash", savedDep.Hash, "Hash should remain unchanged")
	})
}

func computeHash(code []byte) string {
	h := sha256.Sum256(code)
	return hex.EncodeToString(h[:])
}
