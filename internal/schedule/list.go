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
	"strings"

	"github.com/onflow/cadence"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/common/branding"
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

	address, err := util.ResolveAddressOrAccountNameForNetworks(accountInput, state, []string{"mainnet", "testnet", "emulator"})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve account: %w", err)
	}

	chainID, err := util.NetworkToChainID(globalFlags.Network)
	if err != nil {
		return nil, err
	}

	schedulerAddress, err := getContractAddress(FlowTransactionScheduler, chainID)
	if err != nil {
		return nil, err
	}

	schedulerUtilsAddress, err := getContractAddress(FlowTransactionSchedulerUtils, chainID)
	if err != nil {
		return nil, err
	}

	networkStr := branding.GrayStyle.Render(globalFlags.Network)
	accountStr := branding.PurpleStyle.Render(accountInput)
	addressStr := branding.GrayStyle.Render(address.String())

	logger.Info("Listing scheduled transactions...")
	logger.Info("")
	logger.Info(fmt.Sprintf("🌐 Network: %s", networkStr))
	logger.Info(fmt.Sprintf("📝 Account: %s (%s)", accountStr, addressStr))

	script := fmt.Sprintf(`import FlowTransactionScheduler from %s
import FlowTransactionSchedulerUtils from %s

access(all) fun main(managerAddress: Address): [FlowTransactionScheduler.TransactionData] {
    // Use the helper function to borrow the Manager
    let manager = FlowTransactionSchedulerUtils.borrowManager(at: managerAddress)
        ?? panic("Could not borrow Manager from account")

    let transactionIds = manager.getTransactionIDs()
    var transactions: [FlowTransactionScheduler.TransactionData] = []

    // Get transaction data through the Manager instead of directly from FlowTransactionScheduler
    for id in transactionIds {
        if let txData = manager.getTransactionData(id) {
            transactions.append(txData)
        }
    }

    return transactions
}`, schedulerAddress, schedulerUtilsAddress)

	value, err := flow.ExecuteScript(
		context.Background(),
		flowkit.Script{
			Code: []byte(script),
			Args: []cadence.Value{cadence.NewAddress(address)},
		},
		flowkit.LatestScriptQuery,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute script: %w", err)
	}

	transactions, err := parseTransactionList(value)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction list: %w", err)
	}

	logger.Info("")
	if len(transactions) == 0 {
		warningIcon := branding.GrayStyle.Render("ℹ")
		warningMsg := branding.GrayStyle.Render("No scheduled transactions found")
		logger.Info(fmt.Sprintf("%s  %s", warningIcon, warningMsg))
	} else {
		successIcon := branding.GreenStyle.Render("✅")
		successMsg := branding.GreenStyle.Render(fmt.Sprintf("Found %d scheduled transaction(s)", len(transactions)))
		logger.Info(fmt.Sprintf("%s %s", successIcon, successMsg))
	}

	return &listResult{
		transactions: transactions,
	}, nil
}

func parseTransactionList(value cadence.Value) ([]*TransactionData, error) {
	array, ok := value.(cadence.Array)
	if !ok {
		return nil, fmt.Errorf("expected array value, got %T", value)
	}

	var transactions []*TransactionData
	for _, item := range array.Values {
		optional := cadence.NewOptional(item)
		txData, err := ParseTransactionData(optional)
		if err != nil {
			return nil, fmt.Errorf("failed to parse transaction: %w", err)
		}
		if txData != nil {
			transactions = append(transactions, txData)
		}
	}

	return transactions, nil
}

type listResult struct {
	transactions []*TransactionData
}

func (r *listResult) JSON() any {
	var txList []map[string]any
	for _, tx := range r.transactions {
		txList = append(txList, map[string]any{
			"id":                      tx.ID,
			"priority":                tx.Priority,
			"execution_effort":        tx.ExecutionEffort,
			"status":                  tx.Status,
			"fees":                    tx.Fees,
			"scheduled_timestamp":     tx.ScheduledTimestamp,
			"handler_type_identifier": tx.HandlerTypeIdentifier,
			"handler_address":         tx.HandlerAddress,
		})
	}
	return map[string]any{
		"transactions": txList,
		"count":        len(r.transactions),
	}
}

func (r *listResult) String() string {
	if len(r.transactions) == 0 {
		return ""
	}

	var output strings.Builder

	// Display each transaction with details
	for i, tx := range r.transactions {
		if i > 0 {
			output.WriteString("\n")
		}

		// Transaction header line
		txLabel := branding.GrayStyle.Render("Transaction")
		txID := branding.PurpleStyle.Render(fmt.Sprintf("%d", tx.ID))
		output.WriteString(fmt.Sprintf("%s %s\n", txLabel, txID))

		// Transaction details using shared formatting
		output.WriteString(FormatTransactionDetails(tx))
	}

	return output.String()
}

func (r *listResult) Oneliner() string {
	if len(r.transactions) == 0 {
		return "No scheduled transactions found"
	}
	return fmt.Sprintf("Found %d scheduled transaction(s)", len(r.transactions))
}
