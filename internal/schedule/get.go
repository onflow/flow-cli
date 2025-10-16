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
	"context"
	"fmt"
	"strconv"

	"github.com/onflow/cadence"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/common/branding"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
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

	transactionIDStr := args[0]

	// Parse transaction ID as UInt64
	transactionID, err := strconv.ParseUint(transactionIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction ID: %w", err)
	}

	chainID, err := util.NetworkToChainID(globalFlags.Network)
	if err != nil {
		return nil, err
	}

	// Check if network is supported
	if chainID == flowsdk.Mainnet {
		return nil, fmt.Errorf("transaction scheduling is not yet supported on mainnet")
	}

	contractAddress, err := getContractAddress(FlowTransactionScheduler, chainID)
	if err != nil {
		return nil, err
	}

	networkStr := branding.GrayStyle.Render(globalFlags.Network)
	txIDStr := branding.PurpleStyle.Render(transactionIDStr)

	logger.Info("Getting scheduled transaction details")
	logger.Info("")
	logger.Info(fmt.Sprintf("🌐 Network: %s", networkStr))
	logger.Info(fmt.Sprintf("🔍 Transaction ID: %s", txIDStr))

	script := fmt.Sprintf(`import FlowTransactionScheduler from %s

access(all) fun main(transactionID: UInt64): FlowTransactionScheduler.TransactionData? {
    // Get the transaction data directly from the FlowTransactionScheduler contract
    return FlowTransactionScheduler.getTransactionData(id: transactionID)
}`, contractAddress)

	value, err := flow.ExecuteScript(
		context.Background(),
		flowkit.Script{
			Code: []byte(script),
			Args: []cadence.Value{cadence.NewUInt64(transactionID)},
		},
		flowkit.LatestScriptQuery,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute script: %w", err)
	}

	txData, err := ParseTransactionData(value)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction data: %w", err)
	}

	if txData == nil {
		return nil, fmt.Errorf("scheduled transaction not found")
	}

	// Log success
	logger.Info("")
	successIcon := branding.GreenStyle.Render("✅")
	successMsg := branding.GreenStyle.Render("Transaction data retrieved successfully")
	logger.Info(fmt.Sprintf("%s %s", successIcon, successMsg))

	return &getResult{data: txData}, nil
}

type getResult struct {
	data *TransactionData
}

func (r *getResult) JSON() any {
	return map[string]any{
		"id":                      r.data.ID,
		"priority":                r.data.Priority,
		"execution_effort":        r.data.ExecutionEffort,
		"status":                  r.data.Status,
		"fees":                    r.data.Fees,
		"scheduled_timestamp":     r.data.ScheduledTimestamp,
		"handler_type_identifier": r.data.HandlerTypeIdentifier,
		"handler_address":         r.data.HandlerAddress,
	}
}

func (r *getResult) String() string {
	var output string

	// ID
	idLabel := branding.GrayStyle.Render("   ID:")
	idValue := branding.PurpleStyle.Render(fmt.Sprintf("%d", r.data.ID))
	output += fmt.Sprintf("%s %s\n", idLabel, idValue)

	// Status
	statusLabel := branding.GrayStyle.Render("   Status:")
	statusValue := branding.GreenStyle.Render(GetStatusString(r.data.Status))
	output += fmt.Sprintf("%s %s\n", statusLabel, statusValue)

	// Priority
	priorityLabel := branding.GrayStyle.Render("   Priority:")
	priorityValue := branding.PurpleStyle.Render(GetPriorityString(r.data.Priority))
	output += fmt.Sprintf("%s %s\n", priorityLabel, priorityValue)

	// Execution Effort
	effortLabel := branding.GrayStyle.Render("   Execution Effort:")
	effortValue := branding.PurpleStyle.Render(fmt.Sprintf("%d", r.data.ExecutionEffort))
	output += fmt.Sprintf("%s %s\n", effortLabel, effortValue)

	// Fees
	feesLabel := branding.GrayStyle.Render("   Fees:")
	feesValue := branding.PurpleStyle.Render(fmt.Sprintf("%s FLOW", r.data.Fees))
	output += fmt.Sprintf("%s %s\n", feesLabel, feesValue)

	// Scheduled Timestamp
	timestampLabel := branding.GrayStyle.Render("   Scheduled Timestamp:")
	timestampValue := branding.PurpleStyle.Render(r.data.ScheduledTimestamp)
	output += fmt.Sprintf("%s %s\n", timestampLabel, timestampValue)

	// Handler Type
	handlerTypeLabel := branding.GrayStyle.Render("   Handler Type:")
	handlerTypeValue := branding.PurpleStyle.Render(r.data.HandlerTypeIdentifier)
	output += fmt.Sprintf("%s %s\n", handlerTypeLabel, handlerTypeValue)

	// Handler Address
	handlerAddrLabel := branding.GrayStyle.Render("   Handler Address:")
	handlerAddrValue := branding.PurpleStyle.Render(r.data.HandlerAddress)
	output += fmt.Sprintf("%s %s\n", handlerAddrLabel, handlerAddrValue)

	return output
}

func (r *getResult) Oneliner() string {
	statusStr := GetStatusString(r.data.Status)
	return fmt.Sprintf("Transaction %d - Status: %s", r.data.ID, statusStr)
}
