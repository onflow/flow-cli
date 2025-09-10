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
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go/fvm/systemcontracts"
	flowGo "github.com/onflow/flow-go/model/flow"
	flowkitConfig "github.com/onflow/flowkit/v2/config"
)

type ContractSection struct {
	Name         string
	Description  string
	Dependencies []flowkitConfig.Dependency
}

func GetAllContractSections() []ContractSection {
	return []ContractSection{
		getCoreContractsSection(),
		getDefiActionsSection(),
	}
}

func getCoreContractsSection() ContractSection {
	sc := systemcontracts.SystemContractsForChain(flowGo.Mainnet)
	var dependencies []flowkitConfig.Dependency

	for _, contract := range sc.All() {
		dependencies = append(dependencies, flowkitConfig.Dependency{
			Name: contract.Name,
			Source: flowkitConfig.Source{
				NetworkName:  flowkitConfig.MainnetNetwork.Name,
				Address:      flowsdk.HexToAddress(contract.Address.String()),
				ContractName: contract.Name,
			},
		})
	}

	return ContractSection{
		Name:         "Core Contracts",
		Description:  "Essential Flow blockchain system contracts",
		Dependencies: dependencies,
	}
}

func getDefiActionsSection() ContractSection {
	defiContracts := []struct {
		name           string
		mainnetAddress string
		testnetAddress string
	}{
		{"DeFiActions", "0x92195d814edf9cb0", "0x4c2ff9dd03ab442f"},
		{"DeFiActionsMathUtils", "0x92195d814edf9cb0", "0x4c2ff9dd03ab442f"},
		{"DeFiActionsUtils", "0x92195d814edf9cb0", "0x4c2ff9dd03ab442f"},
		{"FungibleTokenConnectors", "0x1d9a619393e9fb53", "0x5a7b9cee9aaf4e4e"},
		{"EVMNativeFLOWConnectors", "0xcc15a0c9c656b648", "0xb88ba0e976146cd1"},
		{"EVMTokenConnectors", "0xcc15a0c9c656b648", "0xb88ba0e976146cd1"},
		{"SwapConnectors", "0x0bce04a00aedf132", "0xaddd594cf410166a"},
		{"IncrementFiSwapConnectors", "0xefa9bd7d1b17f1ed", "0x49bae091e5ea16b5"},
		{"IncrementFiFlashloanConnectors", "0xefa9bd7d1b17f1ed", "0x49bae091e5ea16b5"},
		{"IncrementFiPoolLiquidityConnectors", "0xefa9bd7d1b17f1ed", "0x49bae091e5ea16b5"},
		{"IncrementFiStakingConnectors", "0xefa9bd7d1b17f1ed", "0x49bae091e5ea16b5"},
		{"BandOracleConnectors", "0xf627b5c89141ed99", "0x1a9f5d18d096cd7a"},
		{"UniswapV2Connectors", "0x0e5b1dececaca3a8", "0xfef8e4c5c16ccda5"},
	}

	var dependencies []flowkitConfig.Dependency

	for _, contract := range defiContracts {
		// Add mainnet version
		dependencies = append(dependencies, flowkitConfig.Dependency{
			Name: contract.name,
			Source: flowkitConfig.Source{
				NetworkName:  flowkitConfig.MainnetNetwork.Name,
				Address:      flowsdk.HexToAddress(contract.mainnetAddress),
				ContractName: contract.name,
			},
		})

		// Add testnet version as alias
		dependencies = append(dependencies, flowkitConfig.Dependency{
			Name: contract.name,
			Source: flowkitConfig.Source{
				NetworkName:  flowkitConfig.TestnetNetwork.Name,
				Address:      flowsdk.HexToAddress(contract.testnetAddress),
				ContractName: contract.name,
			},
		})
	}

	return ContractSection{
		Name:         "DeFi Actions",
		Description:  "DeFi protocol integration contracts for automated actions",
		Dependencies: dependencies,
	}
}
