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

	logger.Info(fmt.Sprintf("🌐 Network: %s", networkStr))
	logger.Info(fmt.Sprintf("🔍 Transaction ID: %s", txIDStr))
	logger.Info("")
	logger.Info("⏳ Retrieving scheduled transaction details...")

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

	txData, err := parseTransactionData(value)
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

	return txData, nil
}

// parseTransactionData parses the cadence.Value returned from the script
func parseTransactionData(value cadence.Value) (*getResult, error) {
	// Check if result is nil (optional return)
	optional, ok := value.(cadence.Optional)
	if !ok {
		return nil, fmt.Errorf("expected optional value, got %T", value)
	}

	if optional.Value == nil {
		return nil, nil // Transaction not found
	}

	structValue, ok := optional.Value.(cadence.Struct)
	if !ok {
		return nil, fmt.Errorf("expected struct value, got %T", optional.Value)
	}

	fields := cadence.FieldsMappedByName(structValue)

	result := &getResult{}

	if id, ok := fields["id"].(cadence.UInt64); ok {
		result.id = uint64(id)
	}

	if priority, ok := fields["priority"].(cadence.Enum); ok {
		priorityFields := cadence.FieldsMappedByName(priority)
		if rawValue, ok := priorityFields["rawValue"].(cadence.UInt8); ok {
			result.priority = uint8(rawValue)
		}
	}

	if effort, ok := fields["executionEffort"].(cadence.UInt64); ok {
		result.executionEffort = uint64(effort)
	}

	if status, ok := fields["status"].(cadence.Enum); ok {
		statusFields := cadence.FieldsMappedByName(status)
		if rawValue, ok := statusFields["rawValue"].(cadence.UInt8); ok {
			result.status = uint8(rawValue)
		}
	}

	if fees, ok := fields["fees"].(cadence.UFix64); ok {
		result.fees = fees.String()
	}

	if timestamp, ok := fields["scheduledTimestamp"].(cadence.UFix64); ok {
		result.scheduledTimestamp = timestamp.String()
	}

	if handlerType, ok := fields["handlerTypeIdentifier"].(cadence.String); ok {
		result.handlerTypeIdentifier = string(handlerType)
	}

	if handlerAddr, ok := fields["handlerAddress"].(cadence.Address); ok {
		result.handlerAddress = handlerAddr.String()
	}

	return result, nil
}

type getResult struct {
	id                    uint64
	priority              uint8
	executionEffort       uint64
	status                uint8
	fees                  string
	scheduledTimestamp    string
	handlerTypeIdentifier string
	handlerAddress        string
}

func (r *getResult) JSON() any {
	return map[string]any{
		"id":                      r.id,
		"priority":                r.priority,
		"execution_effort":        r.executionEffort,
		"status":                  r.status,
		"fees":                    r.fees,
		"scheduled_timestamp":     r.scheduledTimestamp,
		"handler_type_identifier": r.handlerTypeIdentifier,
		"handler_address":         r.handlerAddress,
	}
}

func (r *getResult) String() string {
	var output string

	// ID
	idLabel := branding.GrayStyle.Render("   ID:")
	idValue := branding.PurpleStyle.Render(fmt.Sprintf("%d", r.id))
	output += fmt.Sprintf("%s %s\n", idLabel, idValue)

	// Status
	statusLabel := branding.GrayStyle.Render("   Status:")
	statusValue := branding.GreenStyle.Render(getStatusString(r.status))
	output += fmt.Sprintf("%s %s\n", statusLabel, statusValue)

	// Priority
	priorityLabel := branding.GrayStyle.Render("   Priority:")
	priorityValue := branding.PurpleStyle.Render(getPriorityString(r.priority))
	output += fmt.Sprintf("%s %s\n", priorityLabel, priorityValue)

	// Execution Effort
	effortLabel := branding.GrayStyle.Render("   Execution Effort:")
	effortValue := branding.PurpleStyle.Render(fmt.Sprintf("%d", r.executionEffort))
	output += fmt.Sprintf("%s %s\n", effortLabel, effortValue)

	// Fees
	feesLabel := branding.GrayStyle.Render("   Fees:")
	feesValue := branding.PurpleStyle.Render(fmt.Sprintf("%s FLOW", r.fees))
	output += fmt.Sprintf("%s %s\n", feesLabel, feesValue)

	// Scheduled Timestamp
	timestampLabel := branding.GrayStyle.Render("   Scheduled Timestamp:")
	timestampValue := branding.PurpleStyle.Render(r.scheduledTimestamp)
	output += fmt.Sprintf("%s %s\n", timestampLabel, timestampValue)

	// Handler Type
	handlerTypeLabel := branding.GrayStyle.Render("   Handler Type:")
	handlerTypeValue := branding.PurpleStyle.Render(r.handlerTypeIdentifier)
	output += fmt.Sprintf("%s %s\n", handlerTypeLabel, handlerTypeValue)

	// Handler Address
	handlerAddrLabel := branding.GrayStyle.Render("   Handler Address:")
	handlerAddrValue := branding.PurpleStyle.Render(r.handlerAddress)
	output += fmt.Sprintf("%s %s\n", handlerAddrLabel, handlerAddrValue)

	return output
}

func (r *getResult) Oneliner() string {
	statusStr := getStatusString(r.status)
	return fmt.Sprintf("Transaction %d - Status: %s", r.id, statusStr)
}

// getStatusString converts status code to readable string
func getStatusString(status uint8) string {
	switch status {
	case 0:
		return "Pending"
	case 1:
		return "Scheduled"
	case 2:
		return "Executing"
	case 3:
		return "Executed"
	case 4:
		return "Failed"
	case 5:
		return "Cancelled"
	default:
		return fmt.Sprintf("Unknown(%d)", status)
	}
}

// getPriorityString converts priority code to readable string
func getPriorityString(priority uint8) string {
	switch priority {
	case 0:
		return "Low"
	case 1:
		return "Medium"
	case 2:
		return "High"
	default:
		return fmt.Sprintf("Unknown(%d)", priority)
	}
}
