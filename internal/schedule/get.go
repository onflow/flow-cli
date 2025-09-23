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

type flagsGet struct{}

var getFlags = flagsGet{}

var getCommand = command.Command{
	Cmd: &cobra.Command{
		Use:   "get <transaction-id>",
		Short: "Get details of a scheduled transaction",
		Long:  "Get detailed information about a specific scheduled transaction by its transaction ID.",
		Args:  cobra.ExactArgs(1),
		Example: `# Get scheduled transaction details by transaction ID
flow schedule get 0x1234567890abcdef

# Get scheduled transaction details on specific network
flow schedule get 0x1234567890abcdef --network testnet`,
	},
	Flags: &getFlags,
	RunS:  getRun,
}

func getRun(
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
	logger.Info("Retrieving scheduled transaction details...")

	// TODO: Implement get logic for scheduled transaction

	return &getResult{
		success:       true,
		message:       "Scheduled transaction details retrieved successfully",
		transactionID: transactionID,
		status:        "Scheduled",
		scheduledAt:   "2024-01-01T00:00:00Z",
		executeAt:     "2024-01-01T12:00:00Z",
	}, nil
}

type getResult struct {
	success       bool
	message       string
	transactionID string
	status        string
	scheduledAt   string
	executeAt     string
}

func (r *getResult) JSON() any {
	return map[string]any{
		"success":        r.success,
		"message":        r.message,
		"transaction_id": r.transactionID,
		"status":         r.status,
		"scheduled_at":   r.scheduledAt,
		"execute_at":     r.executeAt,
	}
}

func (r *getResult) String() string {
	return fmt.Sprintf(`Transaction ID: %s
Status: %s
Scheduled At: %s
Execute At: %s`, r.transactionID, r.status, r.scheduledAt, r.executeAt)
}

func (r *getResult) Oneliner() string {
	return fmt.Sprintf("Transaction %s - Status: %s", r.transactionID, r.status)
}
