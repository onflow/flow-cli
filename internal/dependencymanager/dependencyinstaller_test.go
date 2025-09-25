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
	"fmt"
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

		gw.GetAccount.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), serviceAcc.Address.String())
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				tests.ContractHelloString.Name: tests.ContractHelloString.Source,
			}

			gw.GetAccount.Return(acc, nil)
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
			SkipAlias:       true,
			dependencies:    make(map[string]config.Dependency),
		}

		err := di.Install()
		assert.NoError(t, err, "Failed to install dependencies")

		filePath := fmt.Sprintf("imports/%s/%s.cdc", serviceAddress.String(), tests.ContractHelloString.Name)
		fileContent, err := state.ReaderWriter().ReadFile(filePath)
		assert.NoError(t, err, "Failed to read generated file")
		assert.NotNil(t, fileContent)
	})
}

func TestDependencyInstallerAdd(t *testing.T) {

	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	serviceAcc, _ := state.EmulatorServiceAccount()
	serviceAddress := serviceAcc.Address

	t.Run("Success", func(t *testing.T) {
		gw := mocks.DefaultMockGateway()

		gw.GetAccount.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), serviceAcc.Address.String())
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				tests.ContractHelloString.Name: tests.ContractHelloString.Source,
			}

			gw.GetAccount.Return(acc, nil)
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
			SkipAlias:       true,
			dependencies:    make(map[string]config.Dependency),
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

		gw.GetAccount.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), serviceAcc.Address.String())
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				tests.ContractHelloString.Name: tests.ContractHelloString.Source,
			}

			gw.GetAccount.Return(acc, nil)
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
			SkipAlias:       true,
			dependencies:    make(map[string]config.Dependency),
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

		gw.GetAccount.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), "1654653399040a61")
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"FlowToken": []byte("access(all) contract FlowToken {}"),
			}

			gw.GetAccount.Return(acc, nil)
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
			SkipAlias:       true,
			dependencies:    make(map[string]config.Dependency),
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
		gw.GetAccount.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), serviceAddress)
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"ContractOne": []byte("access(all) contract ContractOne {}"),
				"ContractTwo": []byte("access(all) contract ContractTwo {}"),
			}
			gw.GetAccount.Return(acc, nil)
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
			SkipAlias:       true,
			dependencies:    make(map[string]config.Dependency),
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

func TestDependencyInstallerAliasTracking(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	serviceAcc, _ := state.EmulatorServiceAccount()
	serviceAddress := serviceAcc.Address

	t.Run("AutoApplyAliasForSameAccount", func(t *testing.T) {
		gw := mocks.DefaultMockGateway()

		// Mock the same account for both contracts
		gw.GetAccount.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, addr.String(), serviceAcc.Address.String())
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				"ContractOne": []byte("access(all) contract ContractOne {}"),
				"ContractTwo": []byte("access(all) contract ContractTwo {}"),
			}

			gw.GetAccount.Return(acc, nil)
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

		gw.GetAccount.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			acc := tests.NewAccountWithAddress(addr.String())
			acc.Contracts = map[string][]byte{
				tests.ContractHelloString.Name: tests.ContractHelloString.Source,
			}
			gw.GetAccount.Return(acc, nil)
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
