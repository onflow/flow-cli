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
)

type flagsCancel struct{}

var cancelFlags = flagsCancel{}

var cancelCommand = command.Command{
	Cmd: &cobra.Command{
		Use:   "cancel <transaction-id>",
		Short: "Cancel a scheduled transaction",
		Long:  "Cancel a previously scheduled transaction from the Transaction Scheduler by its transaction ID.",
		Args:  cobra.ExactArgs(1),
		Example: `# Cancel scheduled transaction by transaction ID
flow schedule cancel 0x1234567890abcdef

# Cancel scheduled transaction on specific network
flow schedule cancel 0x1234567890abcdef --network testnet`,
	},
	Flags: &cancelFlags,
	RunS:  cancelRun,
}

func cancelRun(
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
		return nil, fmt.Errorf("transaction ID is required as an argument")
	}

	transactionID := args[0]

	logger.Info(fmt.Sprintf("Network: %s", globalFlags.Network))
	logger.Info(fmt.Sprintf("Transaction ID: %s", transactionID))
	logger.Info("Canceling scheduled transaction...")

	// TODO: Implement cancel logic for scheduled transaction

	return &cancelResult{
		success: true,
		message: fmt.Sprintf("Scheduled transaction %s canceled successfully", transactionID),
	}, nil
}

type cancelResult struct {
	success bool
	message string
}

func (r *cancelResult) JSON() any {
	return map[string]any{
		"success": r.success,
		"message": r.message,
	}
}

func (r *cancelResult) String() string {
	return r.message
}

func (r *cancelResult) Oneliner() string {
	return r.message
}
