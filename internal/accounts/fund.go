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

package accounts

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/onflow/cadence"
	flowsdk "github.com/onflow/flow-go-sdk"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/accounts"
	"github.com/onflow/flowkit/v2/output"
	"github.com/onflow/flowkit/v2/transactions"

	"github.com/onflow/flow-cli/common/branding"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/prompt"
	"github.com/onflow/flow-cli/internal/util"
)

type flagsFund struct {
	Include []string `default:"" flag:"include" info:"Fields to include in the output. Valid values: contracts."`
}

var fundFlags = flagsFund{}

var fundCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "fund [address|name]",
		Short:   "Funds an account by address or account name through the Testnet Faucet",
		Example: "flow accounts fund 8e94eaa81771313a\nflow accounts fund testnet-account\nflow accounts fund",
		Args:    cobra.MaximumNArgs(1),
	},
	Flags: &fundFlags,
	RunS:  fund,
}

func fund(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	var address flowsdk.Address

	if len(args) == 0 {
		// No address provided, prompt user to select from available accounts
		availableAccounts := util.GetAccountsByNetworks(state, []string{"testnet", "emulator"})
		if len(availableAccounts) == 0 {
			errorMsg := branding.ErrorStyle.Render("no accounts found in flow.json.")
			helpText := branding.GrayStyle.Render("Create an account first with:")
			suggestion := branding.GreenStyle.Render("flow accounts create")
			return nil, fmt.Errorf("%s\n%s %s", errorMsg, helpText, suggestion)
		}

		options := make([]string, len(availableAccounts))
		for i, account := range availableAccounts {
			network := "emulator"
			if util.IsAddressValidForNetwork(account.Address, "testnet") {
				network = "testnet"
			} else if util.IsAddressValidForNetwork(account.Address, "mainnet") {
				network = "mainnet"
			}
			options[i] = fmt.Sprintf("%s (%s) [%s]", account.Address.HexWithPrefix(), account.Name, network)
		}

		selected, err := prompt.RunSingleSelect(options, "Select an account to fund:")
		if err != nil {
			errorMsg := branding.ErrorStyle.Render("account selection cancelled")
			return nil, fmt.Errorf("%s: %w", errorMsg, err)
		}

		for i, option := range options {
			if option == selected {
				address = availableAccounts[i].Address
				break
			}
		}
	} else {
		var err error
		address, err = util.ResolveAddressOrAccountNameForNetworks(args[0], state, []string{"testnet", "emulator"})
		if err != nil {
			return nil, err
		}
	}

	if address.IsValid(flowsdk.Testnet) {
		return fundTestnetAccount(address, logger)
	} else if address.IsValid(flowsdk.Emulator) {
		return fundEmulatorAccount(address, logger, flow, state)
	} else {
		addressStr := branding.GrayStyle.Render(address.String())
		errorMsg := branding.ErrorStyle.Render("funding is only supported for testnet and emulator addresses")
		return nil, fmt.Errorf("unsupported address %s, %s", addressStr, errorMsg)
	}
}

// fundTestnetAccount funds a testnet account using the web faucet
func fundTestnetAccount(address flowsdk.Address, logger output.Logger) (command.Result, error) {
	addressStr := branding.PurpleStyle.Render(address.HexWithPrefix())
	linkStr := branding.GreenStyle.Render(testnetFaucetURL(address))

	logger.Info(
		fmt.Sprintf(
			"Opening the Testnet faucet to fund %s on your native browser."+
				"\n\nIf there is an issue, please use this link instead: %s",
			addressStr,
			linkStr,
		))
	// wait for the user to read the message
	time.Sleep(5 * time.Second)

	if err := browser.OpenURL(testnetFaucetURL(address)); err != nil {
		return nil, err
	}

	return nil, nil
}

// fundEmulatorAccount funds an emulator account by minting tokens directly
func fundEmulatorAccount(address flowsdk.Address, logger output.Logger, flow flowkit.Services, state *flowkit.State) (command.Result, error) {
	const defaultFundingAmount = "1000"

	addressStr := branding.PurpleStyle.Render(address.HexWithPrefix())
	logger.Info(fmt.Sprintf("Funding emulator account %s with %s FLOW tokens...", addressStr, defaultFundingAmount))

	fundingTx := `
import FlowToken from 0x0ae53cb6e3f42a79
import FungibleToken from 0xee82856bf20e2aa6

transaction(address: Address, amount: UFix64) {
    let tokenAdmin: &FlowToken.Administrator
    let tokenReceiver: &{FungibleToken.Receiver}

    prepare(signer: auth(BorrowValue) &Account) {
        self.tokenAdmin = signer.storage.borrow<&FlowToken.Administrator>(from: /storage/flowTokenAdmin)
            ?? panic("Signer is not the token admin")

        self.tokenReceiver = getAccount(address).capabilities.borrow<&{FungibleToken.Receiver}>(
                /public/flowTokenReceiver
            ) ?? panic("Could not borrow receiver reference to the recipient's Vault")
    }

    execute {
        let minter <- self.tokenAdmin.createNewMinter(allowedAmount: amount)
        let mintedVault <- minter.mintTokens(amount: amount)

        self.tokenReceiver.deposit(from: <-mintedVault)

        destroy minter
    }
}`

	// Get the emulator service account to sign the transaction
	serviceAccount, err := state.EmulatorServiceAccount()
	if err != nil {
		return nil, fmt.Errorf("failed to get emulator service account: %w", err)
	}

	amount, err := strconv.ParseFloat(defaultFundingAmount, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid funding amount: %w", err)
	}

	transactionArgs := []cadence.Value{
		cadence.NewAddress(address),
		cadence.UFix64(amount),
	}

	_, txResult, err := flow.SendTransaction(
		context.Background(),
		transactions.AccountRoles{
			Proposer:    *serviceAccount,
			Authorizers: []accounts.Account{*serviceAccount},
			Payer:       *serviceAccount,
		},
		flowkit.Script{
			Code: []byte(fundingTx),
			Args: transactionArgs,
		},
		1000,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to fund emulator account: %w", err)
	}

	if txResult.Error != nil {
		return nil, fmt.Errorf("funding transaction failed: %s", txResult.Error.Error())
	}

	successMsg := branding.GreenStyle.Render(fmt.Sprintf("✓ Successfully funded %s with %s FLOW tokens", addressStr, defaultFundingAmount))
	logger.Info(successMsg)

	return nil, nil
}
