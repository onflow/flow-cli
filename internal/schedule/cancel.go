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
	"github.com/onflow/flowkit/v2/accounts"
	"github.com/onflow/flowkit/v2/output"
	"github.com/onflow/flowkit/v2/transactions"

	"github.com/onflow/flow-cli/common/branding"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

type flagsCancel struct {
	Signer string `default:"emulator-account" flag:"signer" info:"account to use for canceling the scheduled transaction"`
}

var cancelFlags = flagsCancel{}

var cancelCommand = command.Command{
	Cmd: &cobra.Command{
		Use:   "cancel <transaction-id>",
		Short: "Cancel a scheduled transaction",
		Long:  "Cancel a previously scheduled transaction from the Transaction Scheduler by its transaction ID.",
		Args:  cobra.ExactArgs(1),
		Example: `# Cancel scheduled transaction by transaction ID
flow schedule cancel 0x1234567890abcdef --signer my-account

# Cancel scheduled transaction on specific network
flow schedule cancel 0x1234567890abcdef --signer my-account --network testnet

# Cancel using default signer (emulator-account)
flow schedule cancel 0x1234567890abcdef`,
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

	transactionIDStr := args[0]

	// Parse transaction ID as UInt64
	transactionID, err := strconv.ParseUint(transactionIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction ID: %w", err)
	}

	signer, err := util.GetSignerAccount(state, cancelFlags.Signer)
	if err != nil {
		return nil, err
	}

	chainID, err := util.NetworkToChainID(globalFlags.Network)
	if err != nil {
		return nil, err
	}

	if chainID == flowsdk.Mainnet {
		return nil, fmt.Errorf("transaction scheduling is not yet supported on mainnet")
	}

	schedulerUtilsAddress, err := getContractAddress(FlowTransactionSchedulerUtils, chainID)
	if err != nil {
		return nil, err
	}

	flowTokenAddress, err := getContractAddress(FlowToken, chainID)
	if err != nil {
		return nil, err
	}

	fungibleTokenAddress, err := getContractAddress(FungibleToken, chainID)
	if err != nil {
		return nil, err
	}

	networkStr := branding.GrayStyle.Render(globalFlags.Network)
	addressStr := branding.PurpleStyle.Render(signer.Address.HexWithPrefix())
	signerStr := branding.GrayStyle.Render(cancelFlags.Signer)
	txIDStr := branding.PurpleStyle.Render(transactionIDStr)

	logger.Info("Canceling scheduled transaction...")
	logger.Info("")
	logger.Info(fmt.Sprintf("🌐 Network: %s", networkStr))
	logger.Info(fmt.Sprintf("📝 Signer: %s (%s)", signerStr, addressStr))
	logger.Info(fmt.Sprintf("🔍 Transaction ID: %s", txIDStr))
	logger.Info("")

	// Build transaction code
	cancelTx := fmt.Sprintf(`import FlowTransactionSchedulerUtils from %s
import FlowToken from %s
import FungibleToken from %s

transaction(transactionId: UInt64) {
    let manager: auth(FlowTransactionSchedulerUtils.Owner) &{FlowTransactionSchedulerUtils.Manager}
    let tokenReceiver: &{FungibleToken.Receiver}

    prepare(signer: auth(BorrowValue) &Account) {
        // 1. Borrow Manager reference
        self.manager = signer.storage.borrow<auth(FlowTransactionSchedulerUtils.Owner) &{FlowTransactionSchedulerUtils.Manager}>(
            from: FlowTransactionSchedulerUtils.managerStoragePath
        ) ?? panic("Could not borrow Manager. Please ensure you have a Manager set up.")

         // Verify transaction exists in manager
        assert(
            self.manager.getTransactionIDs().contains(transactionId),
            message: "Transaction with ID ".concat(transactionId.toString()).concat(" not found in manager")
        )

        // 2. Get FlowToken receiver to deposit refunds
        self.tokenReceiver = signer.capabilities.get<&{FungibleToken.Receiver}>(/public/flowTokenReceiver)
            .borrow()
            ?? panic("Could not borrow FlowToken receiver")
    }

    execute {
        // Cancel the transaction and receive refunded fees
        let refundVault <- self.manager.cancel(id: transactionId)

        // Deposit refunded fees back to the account
        self.tokenReceiver.deposit(from: <-refundVault)
    }
}`, schedulerUtilsAddress, flowTokenAddress, fungibleTokenAddress)

	_, txResult, err := flow.SendTransaction(
		context.Background(),
		transactions.AccountRoles{
			Proposer:    *signer,
			Authorizers: []accounts.Account{*signer},
			Payer:       *signer,
		},
		flowkit.Script{
			Code: []byte(cancelTx),
			Args: []cadence.Value{cadence.NewUInt64(transactionID)},
		},
		1000,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to cancel scheduled transaction: %w", err)
	}

	if txResult.Error != nil {
		return nil, fmt.Errorf("cancel transaction failed: %s", txResult.Error.Error())
	}

	logger.Info("")
	successIcon := branding.GreenStyle.Render("✅")
	successMsg := branding.GreenStyle.Render("Scheduled transaction canceled successfully")
	logger.Info(fmt.Sprintf("%s %s", successIcon, successMsg))

	return &cancelResult{
		success:       true,
		transactionID: transactionIDStr,
	}, nil
}

type cancelResult struct {
	success       bool
	transactionID string
}

func (r *cancelResult) JSON() any {
	return map[string]any{
		"success":       r.success,
		"transactionID": r.transactionID,
		"message":       "Scheduled transaction canceled successfully",
	}
}

func (r *cancelResult) String() string {
	return ""
}

func (r *cancelResult) Oneliner() string {
	return fmt.Sprintf("Scheduled transaction %s canceled successfully", r.transactionID)
}
