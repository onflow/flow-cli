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

type flagsSetup struct {
	Signer string `default:"emulator-account" flag:"signer" info:"account to setup the transaction scheduler on"`
}

var setupFlags = flagsSetup{}

var setupCommand = command.Command{
	Cmd: &cobra.Command{
		Use:   "setup",
		Short: "Create a Flow Transaction Scheduler Manager resource on your account",
		Long:  "Initialize your account with a Flow Transaction Scheduler Manager resource to enable transaction scheduling capabilities.",
		Args:  cobra.NoArgs,
		Example: `# Setup transaction scheduler using account name from flow.json
flow schedule setup --signer my-account

# Setup transaction scheduler on specific network
flow schedule setup --signer my-account --network testnet

# Setup transaction scheduler using default signer (emulator-account)
flow schedule setup`,
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

	signer, err := util.GetSignerAccount(state, setupFlags.Signer)
	if err != nil {
		return nil, err
	}

	address := signer.Address

	chainID, err := util.NetworkToChainID(globalFlags.Network)
	if err != nil {
		return nil, err
	}

	// Check if network is supported
	if chainID == flowsdk.Mainnet {
		return nil, fmt.Errorf("transaction scheduling is not yet supported on mainnet")
	}

	contractAddress, err := getContractAddress(FlowTransactionSchedulerUtils, chainID)
	if err != nil {
		return nil, err
	}

	// Log setup information with styled output
	networkStr := branding.GrayStyle.Render(globalFlags.Network)
	addressStr := branding.PurpleStyle.Render(address.HexWithPrefix())
	signerStr := branding.GrayStyle.Render(setupFlags.Signer)

	logger.Info(fmt.Sprintf("🌐 Network: %s", networkStr))
	logger.Info(fmt.Sprintf("📝 Signer: %s (%s)", signerStr, addressStr))
	logger.Info("")
	logger.Info("⚡ Setting up Transaction Scheduler Manager...")

	setupTx := fmt.Sprintf(`import FlowTransactionSchedulerUtils from %s

transaction() {
    prepare(signer: auth(BorrowValue, SaveValue) &Account) {
        // Check if Manager already exists
        if signer.storage.borrow<&{FlowTransactionSchedulerUtils.Manager}>(from: FlowTransactionSchedulerUtils.managerStoragePath) == nil {
            // Create and save Manager
            signer.storage.save(
                <-FlowTransactionSchedulerUtils.createManager(),
                to: FlowTransactionSchedulerUtils.managerStoragePath
            )
        }
    }
}`, contractAddress)

	_, txResult, err := flow.SendTransaction(
		context.Background(),
		transactions.AccountRoles{
			Proposer:    *signer,
			Authorizers: []accounts.Account{*signer},
			Payer:       *signer,
		},
		flowkit.Script{
			Code: []byte(setupTx),
			Args: nil,
		},
		1000,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to setup transaction scheduler: %w", err)
	}

	if txResult.Error != nil {
		return nil, fmt.Errorf("setup transaction failed: %s", txResult.Error.Error())
	}

	// Log success with styled output
	logger.Info("")
	successIcon := branding.GreenStyle.Render("✅")
	successMsg := branding.GreenStyle.Render("Transaction Scheduler Manager setup successfully!")
	logger.Info(fmt.Sprintf("%s %s", successIcon, successMsg))

	if txResult.TransactionID.String() != "" {
		txIDLabel := branding.GrayStyle.Render("Transaction ID:")
		txID := branding.PurpleStyle.Render(txResult.TransactionID.String())
		logger.Info(fmt.Sprintf("   %s %s", txIDLabel, txID))
	}

	return &setupResult{
		success:       true,
		transactionID: txResult.TransactionID.String(),
	}, nil
}

type setupResult struct {
	success       bool
	transactionID string
}

func (r *setupResult) JSON() any {
	return map[string]any{
		"success":        r.success,
		"transactionID":  r.transactionID,
		"message":        "Transaction Scheduler Manager setup successfully",
	}
}

func (r *setupResult) String() string {
	// Return empty string since we already logged everything in the command
	return ""
}

func (r *setupResult) Oneliner() string {
	return "Transaction Scheduler Manager setup successfully"
}
