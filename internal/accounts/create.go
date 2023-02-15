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
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"strings"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/onflow/flow-cli/pkg/flowkit/util"
)

type flagsCreate struct {
	Signer    string   `default:"emulator-account" flag:"signer" info:"Account name from configuration used to sign the transaction"`
	Keys      []string `flag:"key" info:"Public keys to attach to account"`
	Weights   []int    `flag:"key-weight" info:"Weight for the key"`
	SigAlgo   []string `default:"ECDSA_P256" flag:"sig-algo" info:"Signature algorithm used to generate the keys"`
	HashAlgo  []string `default:"SHA3_256" flag:"hash-algo" info:"Hash used for the digest"`
	Contracts []string `flag:"contract" info:"Contract to be deployed during account creation. <name:filename>"`
	Include   []string `default:"" flag:"include" info:"Fields to include in the output"`
}

var createFlags = flagsCreate{}

var CreateCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "create",
		Short:   "Create a new account on network",
		Example: `flow accounts create --key d651f1931a2...8745`,
	},
	Flags: &createFlags,
	RunS:  create,
}

func create(
	_ []string,
	loader flowkit.ReaderWriter,
	_ command.GlobalFlags,
	services *services.Services,
	state *flowkit.State,
) (command.Result, error) {
	// if user doesn't provide any flags go into interactive mode
	if len(createFlags.Keys) == 0 {
		err := createInteractive(state, loader)
		return nil, err
	}

	signer, err := state.Accounts().ByName(createFlags.Signer)
	if err != nil {
		return nil, err
	}

	if len(createFlags.SigAlgo) == 1 && len(createFlags.HashAlgo) == 1 {
		// Fill up depending on size of key input
		if len(createFlags.Keys) > 1 {
			for i := 1; i < len(createFlags.Keys); i++ {
				createFlags.SigAlgo = append(createFlags.SigAlgo, createFlags.SigAlgo[0])
				createFlags.HashAlgo = append(createFlags.HashAlgo, createFlags.HashAlgo[0])
			}
			// Deprecated usage message?
		}

	} else if len(createFlags.Keys) != len(createFlags.SigAlgo) || len(createFlags.SigAlgo) != len(createFlags.HashAlgo) { // double check matching array lengths on inputs
		return nil, fmt.Errorf("must provide a signature and hash algorithm for every key provided to --key: %d keys, %d signature algo, %d hash algo", len(createFlags.Keys), len(createFlags.SigAlgo), len(createFlags.HashAlgo))
	}

	keyWeights := createFlags.Weights

	sigAlgos, err := parseSignatureAlgorithms(createFlags.SigAlgo)
	if err != nil {
		return nil, err
	}
	hashAlgos, err := parseHashingAlgorithms(createFlags.HashAlgo)
	if err != nil {
		return nil, err
	}

	pubKeys, err := parsePublicKeys(createFlags.Keys, sigAlgos)
	if err != nil {
		return nil, err
	}

	account, err := services.Accounts.Create(
		signer,
		pubKeys,
		keyWeights,
		sigAlgos,
		hashAlgos,
		createFlags.Contracts,
	)

	if err != nil {
		return nil, err
	}

	return &AccountResult{
		Account: account,
		include: createFlags.Include,
	}, nil
}

func parseHashingAlgorithms(algorithms []string) ([]crypto.HashAlgorithm, error) {
	hashAlgos := make([]crypto.HashAlgorithm, 0, len(createFlags.HashAlgo))
	for _, hashAlgoStr := range createFlags.HashAlgo {
		hashAlgo := crypto.StringToHashAlgorithm(hashAlgoStr)
		if hashAlgo == crypto.UnknownHashAlgorithm {
			return nil, fmt.Errorf("invalid hash algorithm: %s", createFlags.HashAlgo)
		}
		hashAlgos = append(hashAlgos, hashAlgo)
	}
	return hashAlgos, nil
}

func parseSignatureAlgorithms(algorithms []string) ([]crypto.SignatureAlgorithm, error) {
	sigAlgos := make([]crypto.SignatureAlgorithm, 0, len(createFlags.SigAlgo))
	for _, sigAlgoStr := range algorithms {
		sigAlgo := crypto.StringToSignatureAlgorithm(sigAlgoStr)
		if sigAlgo == crypto.UnknownSignatureAlgorithm {
			return nil, fmt.Errorf("invalid signature algorithm: %s", createFlags.SigAlgo)
		}
		sigAlgos = append(sigAlgos, sigAlgo)
	}
	return sigAlgos, nil
}

func parsePublicKeys(publicKeys []string, sigAlgorithms []crypto.SignatureAlgorithm) ([]crypto.PublicKey, error) {
	pubKeys := make([]crypto.PublicKey, 0, len(createFlags.Keys))
	for i, k := range publicKeys {
		k = strings.TrimPrefix(k, "0x") // clear possible prefix
		key, err := crypto.DecodePublicKeyHex(sigAlgorithms[i], k)
		if err != nil {
			return nil, fmt.Errorf("failed decoding public key: %s with error: %w", key, err)
		}
		pubKeys = append(pubKeys, key)
	}
	return pubKeys, nil
}

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
		if err != nil {
			return err
		}
		log.StopProgress()

		log.Info(output.Italic("\nPlease note that the newly-created account will only be available while you keep the emulator service running. If you restart the emulator service, all accounts will be reset. If you want to persist accounts between restarts, please use the '--persist' flag when starting the flow emulator.\n"))

		address = account.Address
	} else {

		log.StopProgress()
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

func createFlowAccount(publicKey crypto.PublicKey, network config.Network) (flow.Address, error) {
	req := &lilicoAccountRequest{
		publicKey: publicKey.String(),
	}

	res, err := req.do(network.Name)
	if err != nil {
		return flow.EmptyAddress, err
	}

}

type lilicoAccountRequest struct {
	publicKey          string
	signatureAlgorithm string
	hashAlgorithm      string
	weight             int
}

type lilicoAccountResponse struct {
	txID      string
	succeeded bool
}

func newLilicoResponse(res []byte) *lilicoAccountResponse {
	return &lilicoAccountResponse{}
}

func (l *lilicoAccountRequest) do(network string) (*lilicoAccountResponse, error) {
	// fix to the defaults as we don't support other values
	l.hashAlgorithm = crypto.SHA3_256.String()
	l.signatureAlgorithm = crypto.ECDSA_P256.String()
	l.weight = flow.AccountKeyWeightThreshold

	data, err := json.Marshal(l)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	return newLilicoResponse(body), nil
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
