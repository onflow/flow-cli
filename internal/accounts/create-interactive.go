/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
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
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/onflow/flow-cli/pkg/flowkit/util"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"os"
)

// createInteractive is used when user calls a default account create command without any provided values.
//
// This process takes the user through couple of steps with prompts asking for them to provide name and network,
// and it then uses account creation APIs to automatically create the account on the network as well as save it.
func createInteractive(state *flowkit.State) error {
	log := output.NewStdoutLogger(output.InfoLog)
	name := output.AccountNamePrompt(state.Accounts()) // todo check for duplicate names
	networkName, selectedNetwork := output.CreateAccountNetworkPrompt()
	privateFile := fmt.Sprintf("%s.pkey", name)

	// create new gateway based on chosen network
	gw, err := gateway.NewGrpcGateway(selectedNetwork.Host)
	if err != nil {
		return err
	}
	service := services.NewServices(gw, state, output.NewStdoutLogger(output.NoneLog))

	key, err := service.Keys.Generate("", crypto.ECDSA_P256)
	if err != nil {
		return err
	}

	log.StartProgress(fmt.Sprintf("Creating account %s on %s...", name, networkName))

	account, err := createAccount(name, key, selectedNetwork, state, service)

	state.Accounts().AddOrUpdate(account)
	err = state.SaveDefault()
	if err != nil {
		return err
	}

	log.StopProgress()

	items := []string{
		"Here’s a summary of all the actions that were taken",
		fmt.Sprintf("Added the new account to %s.", output.Bold("flow.json")),
	}
	if selectedNetwork != config.DefaultEmulatorNetwork() {
		items = append(items,
			fmt.Sprintf("Saved the private key to %s.", output.Bold(privateFile)),
			fmt.Sprintf("Added %s to %s.", output.Bold(privateFile), output.Bold(".gitignore")),
		)
	} else {
		log.Info(output.Italic("\nPlease note that the newly-created account will only be available while you keep the emulator service running. If you restart the emulator service, all accounts will be reset. If you want to persist accounts between restarts, please use the '--persist' flag when starting the flow emulator.\n"))
	}
	log.Info(fmt.Sprintf(
		"%s New account created with address %s and name %s on %s network.\n",
		output.SuccessEmoji(),
		output.Bold(fmt.Sprintf("0x%s", account.Address().String())),
		output.Bold(name),
		output.Bold(networkName)),
	)

	outputList(log, items, false)

	return nil
}

const (
	testAddress = ""
	mainAddress = ""
	testKey     = ""
	mainKey     = ""
)

// createAccount on the network using the available signers
func createAccount(
	name string,
	key crypto.PrivateKey,
	network config.Network,
	state *flowkit.State,
	service *services.Services,
) (*flowkit.Account, error) {
	privateFile := fmt.Sprintf("%s.pkey", name)

	var (
		accountKey flowkit.AccountKey
		signer     *flowkit.Account
		err        error
	)

	rawKey := map[string]string{
		config.DefaultTestnetNetwork().Name: testKey,
		config.DefaultMainnetNetwork().Name: mainKey,
	}[network.Name]

	rawAddr := map[string]string{
		config.DefaultTestnetNetwork().Name: testAddress,
		config.DefaultMainnetNetwork().Name: mainAddress,
	}[network.Name]

	if network == config.DefaultEmulatorNetwork() {
		signer, err = state.EmulatorServiceAccount()
		if err != nil {
			return nil, err
		}
		accountKey = flowkit.NewHexAccountKeyFromPrivateKey(0, crypto.SHA3_256, key)

	} else {
		pk, _ := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, rawKey)
		signer = flowkit.NewAccount("signer").
			SetAddress(flow.HexToAddress(rawAddr)).
			SetKey(flowkit.NewHexAccountKeyFromPrivateKey(0, crypto.SHA3_256, pk))

		accountKey = flowkit.NewFileAccountKey(privateFile, 0, crypto.ECDSA_P256, crypto.SHA3_256)

		err = util.AddToGitIgnore(privateFile, state.ReaderWriter())
		if err != nil {
			return nil, err
		}
		// create the private key file
		err = state.ReaderWriter().WriteFile(privateFile, []byte(key.String()), os.FileMode(0644))
		if err != nil {
			return nil, fmt.Errorf("failed saving private key: %w", err)
		}
	}
	if err != nil {
		return nil, err
	}

	networkAccount, err := service.Accounts.Create(
		signer,
		[]crypto.PublicKey{key.PublicKey()},
		[]int{flow.AccountKeyWeightThreshold},
		[]crypto.SignatureAlgorithm{crypto.ECDSA_P256},
		[]crypto.HashAlgorithm{crypto.SHA3_256},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return flowkit.NewAccount(name).SetAddress(networkAccount.Address).SetKey(accountKey), nil
}

func fundAccount() {

}

// outputList helper for printing lists
func outputList(log *output.StdoutLogger, items []string, numbered bool) {
	log.Info(fmt.Sprintf("%s:", items[0]))
	items = items[1:]
	for n, item := range items {
		sep := " -"
		if numbered {
			sep = fmt.Sprintf(" %d.", n+1)
		}
		log.Info(fmt.Sprintf("%s %s", sep, item))
	}
	log.Info("")
}
