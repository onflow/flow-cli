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

package schedule

import (
	"fmt"

	flowsdk "github.com/onflow/flow-go-sdk"
)

// ContractName represents a scheduler-related contract name
type ContractName string

const (
	// FlowTransactionSchedulerUtils contract provides utility functions for transaction scheduling
	FlowTransactionSchedulerUtils ContractName = "FlowTransactionSchedulerUtils"
	// FlowTransactionScheduler contract handles the core transaction scheduling logic
	FlowTransactionScheduler ContractName = "FlowTransactionScheduler"
)

// contractAddresses maps contract names to their addresses on different networks
var contractAddresses = map[ContractName]map[flowsdk.ChainID]string{
	FlowTransactionSchedulerUtils: {
		flowsdk.Emulator: "0xf8d6e0586b0a20c7",
		flowsdk.Testnet:  "0x8c5303eaa26202d6",
	},
	FlowTransactionScheduler: {
		flowsdk.Emulator: "0xf8d6e0586b0a20c7",
		flowsdk.Testnet:  "0x8c5303eaa26202d6",
	},
}

// getContractAddress returns the contract address for the given contract name and network
func getContractAddress(contract ContractName, chainID flowsdk.ChainID) (string, error) {
	// Check if mainnet
	if chainID == flowsdk.Mainnet {
		return "", fmt.Errorf("transaction scheduling is not yet supported on mainnet")
	}

	// Look up the contract address
	networkAddresses, contractExists := contractAddresses[contract]
	if !contractExists {
		return "", fmt.Errorf("unknown contract: %s", contract)
	}

	contractAddress, networkSupported := networkAddresses[chainID]
	if !networkSupported {
		return "", fmt.Errorf("contract %s is not available on network %s", contract, chainID)
	}

	return contractAddress, nil
}
