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
	"fmt"
	"time"

	flowsdk "github.com/onflow/flow-go-sdk"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/accounts"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/common/branding"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/prompt"
)

type flagsFund struct {
	Include []string `default:"" flag:"include" info:"Fields to include in the output. Valid values: contracts."`
}

var fundFlags = flagsFund{}

// getTestnetAccounts returns all accounts that have testnet-valid addresses
func getTestnetAccounts(state *flowkit.State) []accounts.Account {
	var testnetAccounts []accounts.Account

	allAccounts := *state.Accounts()
	for _, account := range allAccounts {
		if account.Address.IsValid(flowsdk.Testnet) {
			testnetAccounts = append(testnetAccounts, account)
		}
	}

	return testnetAccounts
}

// resolveAddressOrAccountName resolves a string that could be either an address or account name
func resolveAddressOrAccountName(input string, state *flowkit.State) (flowsdk.Address, error) {
	address := flowsdk.HexToAddress(input)

	if address.IsValid(flowsdk.Mainnet) || address.IsValid(flowsdk.Testnet) || address.IsValid(flowsdk.Emulator) {
		// For direct addresses, we'll let the caller handle testnet validation
		return address, nil
	}

	account, err := state.Accounts().ByName(input)
	if err != nil {
		accountName := branding.GrayStyle.Render(input)
		return flowsdk.EmptyAddress, fmt.Errorf("could not find account with name %s", accountName)
	}

	if !account.Address.IsValid(flowsdk.Testnet) {
		accountName := branding.PurpleStyle.Render(input)
		addressStr := branding.GrayStyle.Render(account.Address.String())
		errorMsg := branding.ErrorStyle.Render("The faucet can only fund testnet addresses")
		return flowsdk.EmptyAddress, fmt.Errorf("account %s has address %s which is not valid for testnet. %s", accountName, addressStr, errorMsg)
	}

	return account.Address, nil
}

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
		// No address provided, prompt user to select from testnet accounts
		testnetAccounts := getTestnetAccounts(state)
		if len(testnetAccounts) == 0 {
			errorMsg := branding.ErrorStyle.Render("no testnet accounts found in flow.json.")
			helpText := branding.GrayStyle.Render("Create a testnet account first with:")
			suggestion := branding.GreenStyle.Render("flow accounts create --network testnet")
			return nil, fmt.Errorf("%s\n%s %s", errorMsg, helpText, suggestion)
		}

		options := make([]string, len(testnetAccounts))
		for i, account := range testnetAccounts {
			options[i] = fmt.Sprintf("%s (%s)", account.Address.HexWithPrefix(), account.Name)
		}

		selected, err := prompt.RunSingleSelect(options, "Select a testnet account to fund:")
		if err != nil {
			errorMsg := branding.ErrorStyle.Render("account selection cancelled")
			return nil, fmt.Errorf("%s: %w", errorMsg, err)
		}

		for i, option := range options {
			if option == selected {
				address = testnetAccounts[i].Address
				break
			}
		}
	} else {
		var err error
		address, err = resolveAddressOrAccountName(args[0], state)
		if err != nil {
			return nil, err
		}
	}

	if !address.IsValid(flowsdk.Testnet) {
		addressStr := branding.GrayStyle.Render(address.String())
		errorMsg := branding.ErrorStyle.Render("faucet can only work for valid Testnet addresses")
		return nil, fmt.Errorf("unsupported address %s, %s", addressStr, errorMsg)
	}

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
