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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/onflow/flow-cli/pkg/flowkit/util"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"io"
	"net/http"
)

// createInteractive is used when user calls a default account create command without any provided values.
//
// This process takes the user through couple of steps with prompts asking for them to provide name and network,
// and it then uses account creation APIs to automatically create the account on the network as well as save it.
func createInteractive(state *flowkit.State, loader flowkit.ReaderWriter) error {
	log := output.NewStdoutLogger(output.InfoLog)

	name := output.AccountNamePrompt(state.Accounts()) // todo check for duplicate names
	networkName, selectedNetwork := output.CreateAccountNetworkPrompt()
	privateFile := output.Bold(fmt.Sprintf("%s.private.json", name))

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

	log.StartProgress("Creating an account")

	var address flow.Address
	if selectedNetwork == config.DefaultEmulatorNetwork() {
		signer, err := state.EmulatorServiceAccount()
		if err != nil {
			return err
		}
		account, err := service.Accounts.Create(
			signer,
			[]crypto.PublicKey{key.PublicKey()},
			[]int{flow.AccountKeyWeightThreshold},
			[]crypto.SignatureAlgorithm{crypto.ECDSA_P256},
			[]crypto.HashAlgorithm{crypto.SHA3_256},
			nil,
		)
		log.StopProgress()
		if err != nil {
			return err
		}

		log.Info(output.Italic("\nPlease note that the newly-created account will only be available while you keep the emulator service running. If you restart the emulator service, all accounts will be reset. If you want to persist accounts between restarts, please use the '--persist' flag when starting the flow emulator.\n"))

		address = account.Address
	} else {
		addr, err := createFlowAccount(service, key.PublicKey(), selectedNetwork)
		log.StopProgress()
		if err != nil {
			return err
		}
		address = *addr
	}

	onChainAccount, err := service.Accounts.Get(address)
	if err != nil {
		return err
	}

	account, err := flowkit.NewAccountFromOnChainAccount(name, onChainAccount, key)
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf(
		"%s New account created with address %s and name %s on %s network.\n",
		output.SuccessEmoji(),
		output.Bold(fmt.Sprintf("0x%s", account.Address().String())),
		output.Bold(name),
		output.Bold(networkName)),
	)

	err = saveAccount(loader, state, account, selectedNetwork)
	if err != nil {
		return err
	}

	items := []string{
		"Here’s a summary of all the actions that were taken",
		fmt.Sprintf("Added the new account to %s.", output.Bold("flow.json")),
	}
	if selectedNetwork != config.DefaultEmulatorNetwork() {
		items = append(items,
			fmt.Sprintf("Saved the private key to %s.", privateFile),
			fmt.Sprintf("Added %s to %s.", privateFile, output.Bold(".gitignore")),
		)
	}
	outputList(log, items, false)

	return nil
}

// createFlowAccount using the account creation API and return the newly created account address.
func createFlowAccount(
	services *services.Services,
	publicKey crypto.PublicKey,
	network config.Network,
) (*flow.Address, error) {
	account := &lilicoAccount{
		publicKey: publicKey.String(),
	}

	id, err := account.create(network.Name)
	if err != nil {
		return nil, err
	}

	_, result, err := services.Transactions.GetStatus(id, true)
	if err != nil {
		return nil, err
	}

	events := flowkit.EventsFromTransaction(result)
	address := events.GetCreatedAddresses()
	if len(address) == 0 {
		return nil, fmt.Errorf("account creation error")
	}

	return address[0], nil
}

// lilicoAccount contains all the data needed for interaction with lilico account creation API.
type lilicoAccount struct {
	publicKey          string
	signatureAlgorithm string
	hashAlgorithm      string
	weight             int
}

type lilicoResponse struct {
	data struct {
		txId string
	}
}

// create a new account using the lilico API and parsing the response, returning account creation transaction ID.
func (l *lilicoAccount) create(network string) (flow.Identifier, error) {
	// fix to the defaults as we don't support other values
	l.hashAlgorithm = crypto.SHA3_256.String()
	l.signatureAlgorithm = crypto.ECDSA_P256.String()
	l.weight = flow.AccountKeyWeightThreshold

	data, err := json.Marshal(l)
	if err != nil {
		return flow.EmptyID, err
	}

	if network == config.DefaultTestnetNetwork().Name {
		network = "/testnet"
	}
	request, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://openapi.lilico.org/v1/address%s", network),
		bytes.NewReader(data),
	)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	res, err := client.Do(request)
	if err != nil {
		return flow.EmptyID, fmt.Errorf("could not create an account: %w", err)
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	var lilicoRes lilicoResponse
	err = json.Unmarshal(body, &lilicoRes)
	if err != nil {
		return flow.EmptyID, fmt.Errorf("could not create an account: %w", err)
	}

	return flow.HexToID(lilicoRes.data.txId), nil
}

func saveAccount(
	loader flowkit.ReaderWriter,
	state *flowkit.State,
	account *flowkit.Account,
	network config.Network,
) error {
	state.Accounts().AddOrUpdate(account)

	// If not using emulator, save account private key private file for security.
	if network != config.DefaultEmulatorNetwork() {
		privateLocation := fmt.Sprintf("%s.private.json", account.Name())
		state.SetAccountFileLocation(*account, privateLocation)
		err := util.AddToGitIgnore(privateLocation, loader)
		if err != nil {
			return err
		}
	}

	return state.SaveDefault()
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
