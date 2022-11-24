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

package services

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"io/ioutil"
	"net/http"

	"github.com/onflow/flow-cli/pkg/flowkit/project"

	"github.com/onflow/flow-go-sdk"

	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
)

// Transactions is a service that handles all transaction-related interactions.
type Transactions struct {
	gateway gateway.Gateway
	state   *flowkit.State
	logger  output.Logger
}

// NewTransactions returns a new transactions service.
func NewTransactions(
	gateway gateway.Gateway,
	state *flowkit.State,
	logger output.Logger,
) *Transactions {
	return &Transactions{
		gateway: gateway,
		state:   state,
		logger:  logger,
	}
}

// GetStatus of transaction.
func (t *Transactions) GetStatus(
	id flow.Identifier,
	waitSeal bool,
) (*flow.Transaction, *flow.TransactionResult, error) {
	t.logger.StartProgress("Fetching Transaction...")

	tx, err := t.gateway.GetTransaction(id)
	if err != nil {
		return nil, nil, err
	}

	if waitSeal {
		t.logger.StartProgress("Waiting for transaction to be sealed...")
	}

	result, err := t.gateway.GetTransactionResult(id, waitSeal)
	t.logger.StopProgress()

	return tx, result, err
}

// NewTransactionAccountRoles defines transaction roles by accounts.
//
// You can read more about roles here: https://developers.flow.com/learn/concepts/accounts-and-keys
func NewTransactionAccountRoles(
	proposer *flowkit.Account,
	payer *flowkit.Account,
	authorizers []*flowkit.Account,
) (*transactionAccountRoles, error) {
	if proposer == nil || payer == nil {
		return nil, fmt.Errorf("must provide both proposer and payer")
	}

	return &transactionAccountRoles{
		proposer:    proposer,
		authorizers: authorizers,
		payer:       payer,
	}, nil
}

// NewSingleTransactionAccount creates transaction accounts from a single provided
// account fulfilling all the roles (proposer, payer, authorizer).
func NewSingleTransactionAccount(account *flowkit.Account) *transactionAccountRoles {
	return &transactionAccountRoles{
		proposer:    account,
		authorizers: []*flowkit.Account{account},
		payer:       account,
	}
}

// transactionAccountRoles define all the accounts for different transaction roles.
type transactionAccountRoles struct {
	proposer    *flowkit.Account
	authorizers []*flowkit.Account
	payer       *flowkit.Account
}

func (t *transactionAccountRoles) toAddresses() *transactionAddresses {
	auths := make([]flow.Address, len(t.authorizers))
	for i, a := range t.authorizers {
		auths[i] = a.Address()
	}

	return &transactionAddresses{
		proposer:    t.proposer.Address(),
		authorizers: auths,
		payer:       t.payer.Address(),
	}
}

// getSigners for signing the transaction, detect if all accounts are same so only return the one account.
func (t *transactionAccountRoles) getSigners() []*flowkit.Account {
	// build only unique accounts to sign, it's important payer account is last
	sigs := make([]*flowkit.Account, 0)
	addLastIfUnique := func(signer *flowkit.Account) {
		for _, sig := range sigs {
			if sig.Address() == signer.Address() {
				return
			}
		}
		sigs = append(sigs, signer)
	}

	addLastIfUnique(t.proposer)
	for _, auth := range t.authorizers {
		addLastIfUnique(auth)
	}
	addLastIfUnique(t.payer)

	return sigs
}

// NewTransactionAddresses defines transaction roles by account addresses.
//
// You can read more about roles here: https://developers.flow.com/learn/concepts/accounts-and-keys
func NewTransactionAddresses(
	proposer flow.Address,
	payer flow.Address,
	authorizers []flow.Address,
) *transactionAddresses {
	return &transactionAddresses{
		proposer:    proposer,
		authorizers: authorizers,
		payer:       payer,
	}
}

type transactionAddresses struct {
	proposer    flow.Address
	authorizers []flow.Address
	payer       flow.Address
}

// Build builds a transaction with specified payer, proposer and authorizer.
func (t *Transactions) Build(
	addresses *transactionAddresses,
	proposerKeyIndex int,
	script *flowkit.Script,
	gasLimit uint64,
	network string,
) (*flowkit.Transaction, error) {
	if t.state == nil {
		return nil, fmt.Errorf("missing configuration, initialize it: flow state init")
	}

	latestBlock, err := t.gateway.GetLatestBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest sealed block: %w", err)
	}

	proposerAccount, err := t.gateway.GetAccount(addresses.proposer)
	if err != nil {
		return nil, err
	}

	tx := flowkit.NewTransaction().
		SetPayer(addresses.payer).
		SetGasLimit(gasLimit).
		SetBlockReference(latestBlock)

	program, err := flowkit.NewProgram(script)
	if err != nil {
		return nil, err
	}

	if program.HasImports() {
		if network == "" {
			return nil, fmt.Errorf("missing network, specify which network to use to resolve imports in transaction code")
		}
		if script.Location == "" { // when used as lib with code we don't support imports
			return nil, fmt.Errorf("resolving imports in transactions not supported")
		}

		contracts, err := t.state.DeploymentContractsByNetwork(network)
		if err != nil {
			return nil, err
		}

		importReplacer := project.NewFileImports(
			contracts,
			t.state.AliasesForNetwork(network),
		)

		program, err = importReplacer.Replace(program)
		if err != nil {
			return nil, fmt.Errorf("error resolving imports: %w", err)
		}
	}

	if err := tx.SetProposer(proposerAccount, proposerKeyIndex); err != nil {
		return nil, err
	}

	tx, err = tx.AddAuthorizers(addresses.authorizers)
	if err != nil {
		return nil, err
	}

	if err := tx.SetScriptWithArgs(program.Code(), script.Args); err != nil {
		return nil, err
	}

	return tx, nil
}

// Sign transaction payload using the signer account.
func (t *Transactions) Sign(
	signer *flowkit.Account,
	payload []byte,
) (*flowkit.Transaction, error) {
	if t.state == nil {
		return nil, fmt.Errorf("missing configuration, initialize it: flow state init")
	}

	tx, err := flowkit.NewTransactionFromPayload(payload)
	if err != nil {
		return nil, err
	}

	err = tx.SetSigner(signer)
	if err != nil {
		return nil, err
	}

	return tx.Sign()
}

// SendSigned sends the transaction that is already signed.
func (t *Transactions) SendSigned(tx *flowkit.Transaction) (*flow.Transaction, *flow.TransactionResult, error) {
	t.logger.StartProgress(fmt.Sprintf("Sending transaction with ID: %s", tx.FlowTransaction().ID()))
	defer t.logger.StopProgress()

	sentTx, err := t.gateway.SendSignedTransaction(tx)
	if err != nil {
		return nil, nil, err
	}

	res, err := t.gateway.GetTransactionResult(sentTx.ID(), true)
	if err != nil {
		return nil, nil, err
	}

	return sentTx, res, nil
}

// Send a transaction code using the signer account and arguments for the specified network.
func (t *Transactions) Send(
	accounts *transactionAccountRoles,
	script *flowkit.Script,
	gasLimit uint64,
	network string,
) (*flow.Transaction, *flow.TransactionResult, error) {
	if t.state == nil {
		return nil, nil, fmt.Errorf("missing configuration, initialize it: flow state init")
	}

	tx, err := t.Build(
		accounts.toAddresses(),
		accounts.proposer.Key().Index(),
		script,
		gasLimit,
		network,
	)
	if err != nil {
		return nil, nil, err
	}

	for _, signer := range accounts.getSigners() {
		fmt.Println("signer ", signer.Address(), signer.Name())

		err = tx.SetSigner(signer)
		if err != nil {
			return nil, nil, err
		}

		tx, err = tx.Sign()
		if err != nil {
			return nil, nil, err
		}
	}

	t.logger.Info(fmt.Sprintf("Transaction ID: %s", tx.FlowTransaction().ID()))
	t.logger.StartProgress("Sending transaction...")

	sentTx, err := t.gateway.SendSignedTransaction(tx)
	if err != nil {
		return nil, nil, err
	}

	t.logger.StopProgress()
	t.logger.StartProgress("Waiting for transaction to be sealed...")
	defer t.logger.StopProgress()

	res, err := t.gateway.GetTransactionResult(sentTx.ID(), true)

	return sentTx, res, err
}

func (t *Transactions) GetRLP(rlpUrl string) ([]byte, error) {

	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	resp, err := client.Get(rlpUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error downloading RLP identifier")
	}

	return ioutil.ReadAll(resp.Body)
}

func (t *Transactions) PostRLP(rlpUrl string, tx *flow.Transaction) error {
	signedRlp := hex.EncodeToString(tx.Encode())
	resp, err := http.Post(rlpUrl, "application/text", bytes.NewBufferString(signedRlp))

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error posting signed RLP")
	}

	return nil
}
