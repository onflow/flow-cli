/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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

package project

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/onflow/flow-go-sdk/templates"

	"github.com/onflow/flow-cli/pkg/flowcli"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"

	"github.com/onflow/flow-cli/pkg/flowcli/util"
)

type signerRole string

const (
	defaultGasLimit = 1000
)

func NewTransaction() *Transaction {
	return &Transaction{
		tx: flow.NewTransaction(),
	}
}

func NewTransactionFromPayload(filename string) (*Transaction, error) {
	partialTxHex, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read partial transaction from %s: %v", filename, err)
	}

	partialTxBytes, err := hex.DecodeString(string(partialTxHex))
	if err != nil {
		return nil, fmt.Errorf("failed to decode partial transaction from %s: %v", filename, err)
	}

	decodedTx, err := flow.DecodeTransaction(partialTxBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode transaction from %s: %v", filename, err)
	}

	tx := &Transaction{
		tx: decodedTx,
	}

	return tx, nil
}

func NewUpdateAccountContractTransaction(signer *Account, name string, source string) (*Transaction, error) {
	contract := templates.Contract{
		Name:   name,
		Source: source,
	}

	tx := &Transaction{
		tx: templates.UpdateAccountContract(signer.Address(), contract),
	}

	err := tx.SetSigner(signer)
	if err != nil {
		return nil, err
	}
	tx.SetPayer(signer.Address())

	return tx, nil
}

func NewAddAccountContractTransaction(signer *Account, name string, source string) (*Transaction, error) {
	contract := templates.Contract{
		Name:   name,
		Source: source,
	}

	tx := &Transaction{
		tx: templates.AddAccountContract(signer.Address(), contract),
	}

	err := tx.SetSigner(signer)
	if err != nil {
		return nil, err
	}
	tx.SetPayer(signer.Address())

	return tx, nil
}

func NewRemoveAccountContractTransaction(signer *Account, name string) (*Transaction, error) {
	tx := &Transaction{
		tx: templates.RemoveAccountContract(signer.Address(), name),
	}

	err := tx.SetSigner(signer)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func NewCreateAccountTransaction(
	signer *Account,
	keys []*flow.AccountKey,
	contractArgs []string,
) (*Transaction, error) {

	contracts := make([]templates.Contract, 0)
	for _, contract := range contractArgs {
		contractFlagContent := strings.SplitN(contract, ":", 2)
		if len(contractFlagContent) != 2 {
			return nil, fmt.Errorf("wrong format for contract. Correct format is name:path, but got: %s", contract)
		}

		contractSource, err := util.LoadFile(contractFlagContent[1])
		if err != nil {
			return nil, err
		}

		contracts = append(contracts, templates.Contract{
			Name:   contractFlagContent[0],
			Source: string(contractSource),
		})
	}

	tx := &Transaction{
		tx: templates.CreateAccount(keys, contracts, signer.Address()),
	}

	err := tx.SetSigner(signer)
	if err != nil {
		return nil, err
	}
	tx.SetPayer(signer.Address())

	return tx, nil
}

type Transaction struct {
	signer     *Account
	signerRole signerRole
	proposer   *flow.Account
	tx         *flow.Transaction
}

func (t *Transaction) Signer() *Account {
	return t.signer
}

func (t *Transaction) Proposer() *flow.Account {
	return t.proposer
}

func (t *Transaction) FlowTransaction() *flow.Transaction {
	return t.tx
}

func (t *Transaction) SetScriptWithArgsFromFile(filepath string, args []string, argsJSON string) error {
	script, err := util.LoadFile(filepath)
	if err != nil {
		return err
	}

	return t.SetScriptWithArgs(script, args, argsJSON)
}

func (t *Transaction) SetScriptWithArgs(script []byte, args []string, argsJSON string) error {
	t.tx.SetScript(script)
	return t.AddRawArguments(args, argsJSON)
}

func (t *Transaction) SetSigner(account *Account) error {
	err := account.DefaultKey().Validate()
	if err != nil {
		return err
	}

	t.signer = account
	return nil
}

func (t *Transaction) SetProposer(proposer *flow.Account, keyIndex int) *Transaction {
	t.proposer = proposer
	proposerKey := proposer.Keys[keyIndex]

	t.tx.SetProposalKey(
		proposer.Address,
		proposerKey.Index,
		proposerKey.SequenceNumber,
	)

	return t
}

func (t *Transaction) SetPayer(address flow.Address) *Transaction {
	t.tx.SetPayer(address)
	return t
}

func (t *Transaction) SetBlockReference(block *flow.Block) *Transaction {
	t.tx.SetReferenceBlockID(block.ID)
	return t
}

func (t *Transaction) SetDefaultGasLimit() *Transaction {
	t.tx.SetGasLimit(defaultGasLimit)
	return t
}

func (t *Transaction) AddRawArguments(args []string, argsJSON string) error {
	txArguments, err := flowcli.ParseArguments(args, argsJSON)
	if err != nil {
		return err
	}

	return t.AddArguments(txArguments)
}

func (t *Transaction) AddArguments(args []cadence.Value) error {
	for _, arg := range args {
		err := t.AddArgument(arg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Transaction) AddArgument(arg cadence.Value) error {
	return t.tx.AddArgument(arg)
}

func (t *Transaction) AddAuthorizers(authorizers []flow.Address) *Transaction {
	for _, authorizer := range authorizers {
		t.tx.AddAuthorizer(authorizer)
	}

	return t
}

func (t *Transaction) Sign() (*Transaction, error) {
	keyIndex := t.signer.DefaultKey().Index()
	signer, err := t.signer.DefaultKey().Signer(context.Background())
	if err != nil {
		return nil, err
	}

	if t.shouldSignEnvelope() {
		err = t.tx.SignEnvelope(t.signer.address, keyIndex, signer)
		if err != nil {
			return nil, fmt.Errorf("failed to sign transaction: %s", err)
		}
	} else {
		err = t.tx.SignPayload(t.signer.address, keyIndex, signer)
		if err != nil {
			return nil, fmt.Errorf("failed to sign transaction: %s", err)
		}
	}

	return t, nil
}

func (t *Transaction) shouldSignEnvelope() bool {
	// if signer is payer
	if t.signer.address == t.tx.Payer {
		return true
		/*
			// and either authorizer or proposer - special case
			if len(t.tx.Authorizers) == 1 && t.tx.Authorizers[0] == t.signer.address {
				return true
			} else if t.signer.address == t.tx.ProposalKey.Address {
				return true
			} else {
				// ?
			}
		*/
	}

	return false
}
