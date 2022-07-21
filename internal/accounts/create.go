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
	"strings"
	"time"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/cobra"

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
		_, err := createInteractive(state, loader)
		if err != nil {
			return nil, err
		}
		return nil, nil
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

	// read all signature algorithms
	sigAlgos := make([]crypto.SignatureAlgorithm, 0, len(createFlags.SigAlgo))
	for _, sigAlgoStr := range createFlags.SigAlgo {
		sigAlgo := crypto.StringToSignatureAlgorithm(sigAlgoStr)
		if sigAlgo == crypto.UnknownSignatureAlgorithm {
			return nil, fmt.Errorf("invalid signature algorithm: %s", createFlags.SigAlgo)
		}
		sigAlgos = append(sigAlgos, sigAlgo)
	}

	// read all hash algorithms
	hashAlgos := make([]crypto.HashAlgorithm, 0, len(createFlags.HashAlgo))
	for _, hashAlgoStr := range createFlags.HashAlgo {

		hashAlgo := crypto.StringToHashAlgorithm(hashAlgoStr)
		if hashAlgo == crypto.UnknownHashAlgorithm {
			return nil, fmt.Errorf("invalid hash algorithm: %s", createFlags.HashAlgo)
		}
		hashAlgos = append(hashAlgos, hashAlgo)
	}

	keyWeights := createFlags.Weights

	// decode public keys
	pubKeys := make([]crypto.PublicKey, 0, len(createFlags.Keys))
	for i, k := range createFlags.Keys {
		k = strings.TrimPrefix(k, "0x") // clear possible prefix
		key, err := crypto.DecodePublicKeyHex(sigAlgos[i], k)
		if err != nil {
			return nil, fmt.Errorf("failed decoding public key: %s with error: %w", key, err)
		}
		pubKeys = append(pubKeys, key)
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

func createInteractive(state *flowkit.State, loader flowkit.ReaderWriter) (*flow.Account, error) {
	log := output.NewStdoutLogger(output.InfoLog)

	name := output.AccountNamePrompt() // todo check for duplicate names
	networkName, selectedNetwork := output.CreateAccountNetworkPrompt()

	// create new gateway based on chosen network
	gw, err := gateway.NewGrpcGateway(selectedNetwork.Host)
	if err != nil {
		return nil, err
	}

	privateFile := output.Bold(fmt.Sprintf("%s.private.json", name))

	items := []string{
		fmt.Sprintf("%sThis command will perform the following", output.WarningEmoji()),
		"Generate a new ECDSA P-256 public and private key pair.",
	}
	if selectedNetwork != config.DefaultEmulatorNetwork() {
		items = append(items, fmt.Sprintf("Save the private key to %s.", privateFile))
	}
	items = append(items,
		fmt.Sprintf("Create a new account on %s paired with the public key.", output.Bold(networkName)),
		fmt.Sprintf("Save the newly-created account to %s.\n", output.Bold("flow.json")),
	)
	outputList(log, items, false)

	if !output.WantToContinue() {
		return nil, fmt.Errorf("process terminated")
	}

	service := services.NewServices(gw, state, output.NewStdoutLogger(output.NoneLog))

	key, err := service.Keys.Generate("", crypto.ECDSA_P256)
	if err != nil {
		return nil, err
	}

	startHeight, err := service.Blocks.GetLatestBlockHeight()
	if err != nil {
		return nil, err
	}

	var address flow.Address

	if selectedNetwork == config.DefaultEmulatorNetwork() {
		signer, err := state.EmulatorServiceAccount()
		if err != nil {
			return nil, err
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
			return nil, err
		}
		log.StopProgress()

		log.Info(output.Italic("\nPlease note, that the newly created account will only be available until you keep the emulator service up and running, if you restart the emulator service all accounts will be reset. If you want to persist accounts between restarts you must use the '--persist' flag when starting the flow emulator."))

		address = account.Address
	} else {
		var link string
		switch selectedNetwork {
		case config.DefaultTestnetNetwork():
			outputList(log, []string{
				"Please complete the following steps in the browser",
				"Complete the captcha challenge.",
				"Click the 'Create Account' button.",
				"Return to this window.",
			}, true)
			link = util.TestnetFaucetURL(key.PublicKey().String(), crypto.ECDSA_P256)

		case config.DefaultMainnetNetwork():
			outputList(log, []string{
				"Please complete the following steps in the browser",
				"Click on 'Submit' button.",
				"Connect existing Blocto account or create new.",
				"Click on confirm and approve transaction.",
			}, true)
			link = util.MainnetFlowPortURL(key.PublicKey().String())
		}

		output.ConfirmOpenBrowser()

		log.StartProgress("Waiting for your account to be created, please finish all the steps in the browser...\n")
		_ = util.OpenBrowserWindow(link)
		log.Info(output.Italic(fmt.Sprintf("You can also navigate to the link manually: %s\n", link)))

		addr, err := getAccountCreatedAddressWithPubKey(service, key.PublicKey(), startHeight)
		if err != nil {
			return nil, err
		}
		address = *addr

		log.StopProgress()
	}

	onChainAccount, err := service.Accounts.Get(address)
	if err != nil {
		return nil, err
	}

	account, err := flowkit.NewAccountFromOnChainAccount(name, onChainAccount, key)
	if err != nil {
		return nil, err
	}

	log.Info(fmt.Sprintf(
		"%s New account created with address 0x%s and name %s",
		output.SuccessEmoji(),
		output.Bold(account.Address().String()),
		output.Bold(name)),
	)

	err = saveAccount(loader, state, account, selectedNetwork)
	if err != nil {
		return nil, err
	}

	items = []string{
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

	return onChainAccount, nil
}

// getAccountCreatedAddressWithPubKey monitors the network for account creation events, if the event
// contains the public key we are interested in then it extracts the newly created address from the event payload.
func getAccountCreatedAddressWithPubKey(
	service *services.Services,
	pubKey crypto.PublicKey,
	startHeight uint64,
) (*flow.Address, error) {
	lastHeight, err := service.Blocks.GetLatestBlockHeight()
	if err != nil {
		return nil, err
	}

	flowEvents, err := service.Events.Get([]string{flow.EventAccountKeyAdded}, startHeight, lastHeight, 20, 1)
	if err != nil {
		return nil, err
	}

	var address *flow.Address
	for _, block := range flowEvents {
		events := flowkit.NewEvents(block.Events)
		address = events.GetAddressForKeyAdded(pubKey)
		if address != nil {
			break
		}
	}

	if address == nil {
		if lastHeight-startHeight > 400 { // if something goes wrong don't keep waiting forever to avoid spamming network
			return nil, fmt.Errorf("failed to get the account address due to time out")
		}

		time.Sleep(time.Second * 2)
		address, err = getAccountCreatedAddressWithPubKey(service, pubKey, startHeight)
		if err != nil {
			return nil, err
		}

		return address, nil
	}

	return address, nil
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
