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
	"strings"

	"github.com/onflow/flow-go/fvm/systemcontracts"
	flowGo "github.com/onflow/flow-go/model/flow"

	"github.com/onflow/flow-cli/internal/util"

	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
)

var installFlags = Flags{}

var installCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:   "install [dependencies...]",
		Short: "Install contracts and their dependencies.",
		Long: `Install Flow contracts and their dependencies.

By default, this command will install any dependencies listed in the flow.json file at the root of your project.
You can also specify one or more dependencies directly on the command line, using any of the following formats:

  • network://address
  • network://address.ContractName
  • core contract name (e.g., FlowToken, NonFungibleToken)
  • DeFi Actions contract name (e.g., BandOracleConnectors, SwapConnectors)

Examples:
  1. Install dependencies listed in flow.json:
     flow dependencies install

  2. Install a specific core contract by name:
     flow dependencies install FlowToken

  3. Install a specific DeFi actions contract by name:
     flow dependencies install BandOracleConnectors

  4. Install a single contract by network and address (all contracts at that address):
     flow dependencies install testnet://0x1234abcd

  5. Install a specific contract by network, address, and contract name:
     flow dependencies install testnet://0x1234abcd.MyContract

  6. Install multiple dependencies:
     flow dependencies install FungibleToken NonFungibleToken BandOracleConnectors

Note:
• Using 'network://address' will attempt to install all contracts deployed at that address.
• Using 'network://address.ContractName' will install only the specified contract.
• Specifying a known core contract (e.g., FlowToken) will install it from the official system contracts
  address on Mainnet or Testnet (depending on your project's default network).
• Specifying a known DeFi actions contract (e.g., BandOracleConnectors) will install it from the
  official DeFi actions address on Mainnet.
`,
		Example: `flow dependencies install
flow dependencies install testnet://0x7e60df042a9c0868.FlowToken
flow dependencies install FlowToken
flow dependencies install BandOracleConnectors
flow dependencies install FlowToken NonFungibleToken BandOracleConnectors`,
		Args: cobra.ArbitraryArgs,
	},
	Flags: &installFlags,
	RunS:  install,
}

func install(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	installer, err := NewDependencyInstaller(logger, state, true, "", installFlags)
	if err != nil {
		logger.Error(fmt.Sprintf("Error initializing dependency installer: %v", err))
		return nil, err
	}

	if len(args) > 0 {
		for _, dep := range args {
			logger.Info(fmt.Sprintf("%s Processing dependency %s...", util.PrintEmoji("🔄"), dep))

			// Check if the dependency is a core contract
			coreContractName := findCoreContractCaseInsensitive(dep)
			if coreContractName != "" {
				if err := installer.AddByCoreContractName(coreContractName); err != nil {
					logger.Error(fmt.Sprintf("Error adding core contract %s: %v", coreContractName, err))
					return nil, err
				}
				continue
			}

			// Check if the dependency is a DeFi actions contract
			defiContractName := findDefiActionsContractCaseInsensitive(dep)
			if defiContractName != "" {
				if err := installer.AddByDefiContractName(defiContractName); err != nil {
					logger.Error(fmt.Sprintf("Error adding DeFi actions contract %s: %v", defiContractName, err))
					return nil, err
				}
				continue
			}

			// Check if the dependency is in the "network://address" format (address only)
			hasContract, err := hasContractName(dep)
			if err != nil {
				return nil, fmt.Errorf("invalid dependency format")
			}

			if !hasContract {
				if err := installer.AddAllByNetworkAddress(dep); err != nil {
					logger.Error(fmt.Sprintf("Error adding contracts by address: %v", err))
					return nil, err
				}
			} else {
				if err := installer.AddBySourceString(dep); err != nil {
					if strings.Contains(err.Error(), "invalid dependency source format") {
						logger.Error(fmt.Sprintf("Error: '%s' is neither a core contract, DeFi actions contract, nor a valid dependency source format.\nPlease provide a valid dependency source in the format 'network://address.ContractName', e.g., 'testnet://0x1234567890abcdef.MyContract', or use a valid core contract name such as 'FlowToken', or a valid DeFi actions contract name such as 'BandOracleConnectors'.", dep))
					} else {
						logger.Error(fmt.Sprintf("Error adding dependency %s: %v", dep, err))
					}
					return nil, err
				}
			}
		}

		logger.Info(util.MessageWithEmojiPrefix("🔄", "Installing added dependencies..."))

		if err := installer.Install(); err != nil {
			logger.Error(fmt.Sprintf("Error installing dependencies: %v", err))
			return nil, err
		}

		installer.logs.LogAll(logger)

		return nil, nil
	}

	logger.Info(util.MessageWithEmojiPrefix("🔄", "Installing dependencies from flow.json..."))

	if err := installer.Install(); err != nil {
		logger.Error(fmt.Sprintf("Error installing dependencies: %v", err))
		return nil, err
	}

	installer.logs.LogAll(logger)

	return nil, nil
}

func findCoreContractCaseInsensitive(name string) string {
	for _, contract := range systemcontracts.SystemContractsForChain(flowGo.Mainnet).All() {
		if strings.EqualFold(contract.Name, name) {
			return contract.Name
		}
	}
	return ""
}

func findDefiActionsContractCaseInsensitive(name string) string {
	defiActionsSection := getDefiActionsSection()
	for _, dep := range defiActionsSection.Dependencies {
		if strings.EqualFold(dep.Name, name) {
			return dep.Name
		}
	}
	return ""
}

// Check if the input is in "network://address" or "network://address.contract" format
func hasContractName(dep string) (bool, error) {
	parts := strings.SplitN(dep, "://", 2)
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid format: missing '://'")
	}

	return strings.Contains(parts[1], "."), nil
}

func ParseNetworkAddressString(sourceStr string) (network, address string) {
	parts := strings.Split(sourceStr, "://")
	return parts[0], parts[1]
}
