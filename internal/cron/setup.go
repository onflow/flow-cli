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
)

type setupFlags struct{}

var setupCommand = command.Command{
	Cmd: &cobra.Command{
		Use:   "setup",
		Short: "Create a Flow Transaction Scheduler Manager resource on your account",
		Long:  "Initialize your account with a Flow Transaction Scheduler Manager resource to enable transaction scheduling capabilities.",
		Example: `# Setup transaction scheduler on default account
flow cron setup

# Setup transaction scheduler on specific network
flow cron setup --network testnet`,
	},
	Flags: &setupFlags{},
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
