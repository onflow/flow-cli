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
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/onflow/flow-cli/internal/prompt"

	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/accounts"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/gateway"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/util"
)

// createInteractive is used when user calls a default account create command without any provided values.
//
// This process takes the user through couple of steps with prompts asking for them to provide name and network,
// and it then uses account creation APIs to automatically create the account on the network as well as save it.
func createInteractive(state *flowkit.State) (*accountResult, error) {
	log := output.NewStdoutLogger(output.InfoLog)
	name := prompt.AccountNamePrompt(state.Accounts().Names())
	networkName, selectedNetwork := prompt.CreateAccountNetworkPrompt()
	privateFile := accounts.PrivateKeyFile(name, "")

	// create new gateway based on chosen network
	gw, err := gateway.NewGrpcGateway(selectedNetwork)
	if err != nil {
		return nil, err
	}
	flow := flowkit.NewFlowkit(state, selectedNetwork, gw, output.NewStdoutLogger(output.NoneLog))

	key, err := flow.GenerateKey(context.Background(), defaultSignAlgo, "")
	if err != nil {
		return nil, err
	}

	log.StartProgress(fmt.Sprintf("Creating account %s on %s...", name, networkName))
	defer log.StopProgress()

	var account *accounts.Account
	if selectedNetwork == config.EmulatorNetwork {
		account, err = createEmulatorAccount(state, flow, name, key)
		log.StopProgress()
		log.Info(output.Italic("\nPlease note that the newly-created account will only be available while you keep the emulator service running. If you restart the emulator service, all accounts will be reset. If you want to persist accounts between restarts, please use the '--persist' flag when starting the flow emulator.\n"))
	} else {
		account, err = createNetworkAccount(state, flow, name, key, privateFile, selectedNetwork)
		log.StopProgress()
	}
	if err != nil {
		return nil, err
	}

	log.Info(fmt.Sprintf(
		"%s New account created with address %s and name %s on %s network.\n",
		output.SuccessEmoji(),
		output.Bold(fmt.Sprintf("0x%s", account.Address.String())),
		output.Bold(name),
		output.Bold(networkName)),
	)

	state.Accounts().AddOrUpdate(account)
	err = state.SaveDefault()
	if err != nil {
		return nil, err
	}

	items := []string{
		"Here’s a summary of all the actions that were taken",
		fmt.Sprintf("Added the new account to %s.", output.Bold("flow.json")),
	}
	if selectedNetwork != config.EmulatorNetwork {
		items = append(items,
			fmt.Sprintf("Saved the private key to %s.", output.Bold(privateFile)),
			fmt.Sprintf("Added %s to %s.", output.Bold(privateFile), output.Bold(".gitignore")),
		)
	}
	outputList(log, items, false)

	return &accountResult{
		Account: &flowsdk.Account{
			Address: account.Address,
			Balance: 0,
			Keys:    []*flowsdk.AccountKey{flowsdk.NewAccountKey().FromPrivateKey(key)},
		},
		include: nil,
	}, nil
}

// createNetworkAccount using the account creation API and return the newly created account address.
func createNetworkAccount(
	state *flowkit.State,
	flow flowkit.Services,
	name string,
	key crypto.PrivateKey,
	privateFile string,
	network config.Network,
) (*accounts.Account, error) {
	networkAccount := &lilicoAccount{
		PublicKey: strings.TrimPrefix(key.PublicKey().String(), "0x"),
	}

	id, err := networkAccount.create(network.Name)
	if err != nil {
		return nil, err
	}

	result, err := getAccountCreationResult(flow, id)
	if err != nil {
		return nil, err
	}

	events := flowkit.EventsFromTransaction(result)
	address := events.GetCreatedAddresses()
	if len(address) == 0 {
		return nil, fmt.Errorf("account creation error")
	}

	err = util.AddToGitIgnore(privateFile, state.ReaderWriter())
	if err != nil {
		return nil, err
	}

	err = state.ReaderWriter().WriteFile(privateFile, []byte(key.String()), os.FileMode(0644))
	if err != nil {
		return nil, fmt.Errorf("failed saving private key: %w", err)
	}

	return &accounts.Account{
		Name:    name,
		Address: *address[0],
		Key:     accounts.NewFileKey(privateFile, 0, defaultSignAlgo, defaultHashAlgo, state.ReaderWriter()),
	}, nil
}

func createEmulatorAccount(
	state *flowkit.State,
	flow flowkit.Services,
	name string,
	key crypto.PrivateKey,
) (*accounts.Account, error) {
	signer, err := state.EmulatorServiceAccount()
	if err != nil {
		return nil, err
	}

	networkAccount, _, err := flow.CreateAccount(
		context.Background(),
		signer,
		[]accounts.PublicKey{{
			Public:   key.PublicKey(),
			Weight:   flowsdk.AccountKeyWeightThreshold,
			SigAlgo:  defaultSignAlgo,
			HashAlgo: defaultHashAlgo,
		}},
	)
	if err != nil {
		return nil, err
	}

	return &accounts.Account{
		Name:    name,
		Address: networkAccount.Address,
		Key:     accounts.NewHexKeyFromPrivateKey(0, defaultHashAlgo, key),
	}, nil
}

func getAccountCreationResult(flow flowkit.Services, id flowsdk.Identifier) (*flowsdk.TransactionResult, error) {
	_, result, err := flow.GetTransactionByID(context.Background(), id, true)
	if err != nil {
		if status.Code(err) == codes.NotFound { // if transaction not yet propagated, wait for it
			time.Sleep(1 * time.Second)
			return getAccountCreationResult(flow, id)
		}
		return nil, err
	}

	return result, nil
}

// lilicoAccount contains all the data needed for interaction with lilico account creation API.
type lilicoAccount struct {
	PublicKey          string `json:"publicKey"`
	SignatureAlgorithm string `json:"signatureAlgorithm"`
	HashAlgorithm      string `json:"hashAlgorithm"`
	Weight             int    `json:"weight"`
}

type lilicoResponse struct {
	Data struct {
		TxId string `json:"txId"`
	} `json:"data"`
}

var accountToken = ""

const defaultHashAlgo = crypto.SHA3_256

const defaultSignAlgo = crypto.ECDSA_P256

// create a new account using the lilico API and parsing the response, returning account creation transaction ID.
func (l *lilicoAccount) create(network string) (flowsdk.Identifier, error) {
	// fix to the defaults as we don't support other values
	l.HashAlgorithm = defaultHashAlgo.String()
	l.SignatureAlgorithm = defaultSignAlgo.String()
	l.Weight = flowsdk.AccountKeyWeightThreshold

	data, err := json.Marshal(l)
	if err != nil {
		return flowsdk.EmptyID, err
	}

	apiNetwork := ""
	if network == config.TestnetNetwork.Name {
		apiNetwork = "/testnet"
	}

	if network == config.PreviewnetNetwork.Name {
		apiNetwork = "/previewnet"
	}

	request, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://openapi.lilico.org/v1/address%s", apiNetwork),
		bytes.NewReader(data),
	)
	if err != nil {
		return flowsdk.EmptyID, fmt.Errorf("could not create an account: %w", err)
	}

	request.Header.Add("Content-Type", "application/json; charset=UTF-8")
	request.Header.Add("Authorization", accountToken)

	client := &http.Client{
		Timeout: time.Second * 20,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // lilico api doesn't yet have a valid cert, todo reevaluate
		},
	}
	res, err := client.Do(request)
	if err != nil {
		return flowsdk.EmptyID, fmt.Errorf("could not create an account: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(res.Body)
		return flowsdk.EmptyID, fmt.Errorf("account creation failed with status %d: %s", res.StatusCode, string(bodyBytes))
	}

	body, _ := io.ReadAll(res.Body)
	var lilicoRes lilicoResponse

	err = json.Unmarshal(body, &lilicoRes)
	if err != nil {
		return flowsdk.EmptyID, fmt.Errorf("could not create an account: %w", err)
	}
	return flowsdk.HexToID(lilicoRes.Data.TxId), nil
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
