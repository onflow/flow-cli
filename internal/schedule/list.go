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

	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

type flagsList struct{}

var listFlags = flagsList{}

var listCommand = command.Command{
	Cmd: &cobra.Command{
		Use:   "list <account>",
		Short: "List scheduled transactions for an account",
		Long:  "List all scheduled transactions for a given account address or account name from flow.json.",
		Args:  cobra.ExactArgs(1),
		Example: `# List scheduled transactions using account address
flow schedule list 0x123456789abcdef

# List scheduled transactions using account name from flow.json
flow schedule list my-account

# List scheduled transactions on specific network
flow schedule list my-account --network testnet`,
	},
	Flags: &listFlags,
	RunS:  listRun,
}

func listRun(
	args []string,
	globalFlags command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {

	if state == nil {
		return nil, fmt.Errorf("flow configuration is required. Run 'flow init' first")
	}

	if len(args) == 0 {
		return nil, fmt.Errorf("account is required as an argument")
	}

	accountInput := args[0]

	// Resolve account address from input (could be address or account name)
	address, err := util.ResolveAddressOrAccountNameForNetworks(accountInput, state, []string{"mainnet", "testnet", "emulator"})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve account: %w", err)
	}

	logger.Info(fmt.Sprintf("Network: %s", globalFlags.Network))
	logger.Info(fmt.Sprintf("Account: %s (%s)", accountInput, address.String()))
	logger.Info("Retrieving scheduled transactions...")

	// TODO: Implement list logic for scheduled transactions

	return &listResult{
		success:      true,
		message:      "Scheduled transactions retrieved successfully",
		transactions: []string{"0x1234567890abcdef", "0xfedcba0987654321"}, // Mock data for now
	}, nil
}

type listResult struct {
	success      bool
	message      string
	transactions []string
}

func (r *listResult) JSON() any {
	return map[string]any{
		"success":      r.success,
		"message":      r.message,
		"transactions": r.transactions,
	}
}

func (r *listResult) String() string {
	if len(r.transactions) == 0 {
		return "No scheduled transactions found"
	}

	result := fmt.Sprintf("Found %d scheduled transaction(s):\n", len(r.transactions))
	for i, txID := range r.transactions {
		result += fmt.Sprintf("  %d. %s\n", i+1, txID)
	}
	return result
}

func (r *listResult) Oneliner() string {
	if len(r.transactions) == 0 {
		return "No scheduled transactions found"
	}
	return fmt.Sprintf("Found %d scheduled transactions", len(r.transactions))
}
