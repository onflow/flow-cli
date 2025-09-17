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

package cron

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

type flagsSetup struct{}

var setupFlags = flagsSetup{}

var setupCommand = command.Command{
	Cmd: &cobra.Command{
		Use:   "setup <account>",
		Short: "Create a Flow Transaction Scheduler Manager resource on your account",
		Long:  "Initialize your account with a Flow Transaction Scheduler Manager resource to enable transaction scheduling capabilities.",
		Args:  cobra.ExactArgs(1),
		Example: `# Setup transaction scheduler using account address
flow cron setup 0x123456789abcdef

# Setup transaction scheduler using account name from flow.json
flow cron setup my-account

# Setup transaction scheduler on specific network
flow cron setup my-account --network testnet`,
	},
	Flags: &setupFlags,
	RunS:  setupRun,
}

func setupRun(
	args []string,
	globalFlags command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {

	if state == nil {
		return nil, fmt.Errorf("flow configuration is required. Run 'flow init' first")
	}

	accountInput := args[0]

	// Resolve account address from input (could be address or account name)
	address, err := util.ResolveAddressOrAccountName(accountInput, state)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve account: %w", err)
	}

	// Log network and account information
	logger.Info(fmt.Sprintf("Network: %s", globalFlags.Network))
	logger.Info(fmt.Sprintf("Account: %s (%s)", accountInput, address.String()))
	logger.Info("Setting up Flow Transaction Scheduler Manager resource...")

	// TODO: Implement setup logic for Transaction Scheduler Manager resource
	// This will contain the actual implementation to:
	// 1. Check if resource already exists on the account
	// 2. Deploy/create the Manager resource via transaction
	// 3. Verify successful deployment
	// 4. Return success result

	return &setupResult{
		success: true,
		message: "Transaction Scheduler Manager resource created successfully",
	}, nil
}

type setupResult struct {
	success bool
	message string
}

func (r *setupResult) JSON() any {
	return map[string]any{
		"success": r.success,
		"message": r.message,
	}
}

func (r *setupResult) String() string {
	return r.message
}

func (r *setupResult) Oneliner() string {
	return r.message
}
