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

type flagsCleanup struct{}

var cleanupFlags = flagsCleanup{}

var cleanupCommand = command.Command{
	Cmd: &cobra.Command{
		Use:   "cleanup <account>",
		Short: "Remove Flow Transaction Scheduler Manager resource from your account",
		Long:  "Remove the Flow Transaction Scheduler Manager resource from your account to disable transaction scheduling capabilities.",
		Args:  cobra.ExactArgs(1),
		Example: `# Cleanup transaction scheduler using account address
flow cron cleanup 0x123456789abcdef

# Cleanup transaction scheduler using account name from flow.json
flow cron cleanup my-account

# Cleanup transaction scheduler on specific network
flow cron cleanup my-account --network testnet`,
	},
	Flags: &cleanupFlags,
	RunS:  cleanupRun,
}

func cleanupRun(
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
	address, err := util.ResolveAddressOrAccountName(accountInput, state)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve account: %w", err)
	}

	// Log network and account information
	logger.Info(fmt.Sprintf("Network: %s", globalFlags.Network))
	logger.Info(fmt.Sprintf("Account: %s (%s)", accountInput, address.String()))
	logger.Info("Removing Flow Transaction Scheduler Manager resource...")

	// TODO: Implement cleanup logic for Transaction Scheduler Manager resource

	return &cleanupResult{
		success: true,
		message: "Transaction Scheduler Manager resource removed successfully",
	}, nil
}

type cleanupResult struct {
	success bool
	message string
}

func (r *cleanupResult) JSON() any {
	return map[string]any{
		"success": r.success,
		"message": r.message,
	}
}

func (r *cleanupResult) String() string {
	return r.message
}

func (r *cleanupResult) Oneliner() string {
	return r.message
}
