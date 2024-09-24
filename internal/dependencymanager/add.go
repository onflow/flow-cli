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

	"github.com/onflow/flow-go/fvm/systemcontracts"

	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"

	flowGo "github.com/onflow/flow-go/model/flow"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

type addFlagsCollection struct {
	*Flags
	name string
}

var addFlags = addFlagsCollection{
	Flags: &Flags{},
}

var addCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:   "add <source string or core contract name>",
		Short: "Add a single contract and its dependencies.",
		Example: `flow dependencies add testnet://0afe396ebc8eee65.FlowToken
flow dependencies add FlowToken`,
		Args: cobra.ExactArgs(1),
	},
	RunS:  add,
	Flags: &struct{}{},
}

func init() {
	// Add common flags.
	addFlags.Flags.AddToCommand(addCommand.Cmd)
	// Add command-specific flags.
	addCommand.Cmd.Flags().StringVar(&addFlags.name, "name", "", "Name of the dependency")
}

func add(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	logger.Info(fmt.Sprintf("%s Installing dependencies for %s...", util.PrintEmoji("🔄"), args[0]))

	dep := args[0]

	installer, err := NewDependencyInstaller(logger, state, true, "", *addFlags.Flags)
	if err != nil {
		logger.Error(fmt.Sprintf("Error: %v", err))
		return nil, err
	}

	// First check if the dependency is a core contract.
	coreContractName := findCoreContractCaseInsensitive(dep)
	if coreContractName != "" {
		if err := installer.AddByCoreContractName(coreContractName, addFlags.name); err != nil {
			logger.Error(fmt.Sprintf("Error: %v", err))
			return nil, err
		}
		return nil, nil
	}

	// Otherwise, add the dependency by source string.
	if err := installer.AddBySourceString(dep, addFlags.name); err != nil {
		logger.Error(fmt.Sprintf("Error: %v", err))
		return nil, err
	}

	return nil, nil
}

func findCoreContractCaseInsensitive(name string) string {
	for _, contract := range systemcontracts.SystemContractsForChain(flowGo.Mainnet).All() {
		if name == contract.Name {
			return contract.Name
		}
	}
	return ""
}
