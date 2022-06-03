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

package flowkit

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/cmd"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/templates"
)

const maxGasLimit uint64 = 9999

// NewTransaction create new instance of transaction.
func NewTransaction() *Transaction {
	return &Transaction{
		tx: flow.NewTransaction(),
	}
}

// NewTransactionFromPayload build transaction from payload.
func NewTransactionFromPayload(payload []byte) (*Transaction, error) {
	partialTxBytes, err := hex.DecodeString(string(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to decode partial transaction from %s: %v", payload, err)
	}

	decodedTx, err := flow.DecodeTransaction(partialTxBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode transaction from %s: %v", payload, err)
	}

	tx := &Transaction{
		tx: decodedTx,
	}

	return tx, nil
}

// NewUpdateAccountContractTransaction update account contract.
func NewUpdateAccountContractTransaction(signer *Account, name string, source string) (*Transaction, error) {
	contract := templates.Contract{
		Name:   name,
		Source: source,
	}

	return newTransactionFromTemplate(
		templates.UpdateAccountContract(signer.Address(), contract),
		signer,
	)
}

// NewAddAccountContractTransaction add new contract to the account.
func NewAddAccountContractTransaction(
	signer *Account,
	name string,
	source string,
	args []cadence.Value,
) (*Transaction, error) {
	return addAccountContractWithArgs(signer, templates.Contract{
		Name:   name,
		Source: source,
	}, args)
}

// NewRemoveAccountContractTransaction creates new transaction to remove contract.
func NewRemoveAccountContractTransaction(signer *Account, name string) (*Transaction, error) {
	return newTransactionFromTemplate(
		templates.RemoveAccountContract(signer.Address(), name),
		signer,
	)
}

func addAccountContractWithArgs(
	signer *Account,
	contract templates.Contract,
	args []cadence.Value,
) (*Transaction, error) {
	const addAccountContractTemplate = `
	transaction(name: String, code: String %s) {
		prepare(signer: AuthAccount) {
			signer.contracts.add(name: name, code: code.decodeHex() %s)
		}
	}`

	cadenceName := cadence.String(contract.Name)
	cadenceCode := cadence.String(contract.SourceHex())

	tx := flow.NewTransaction().
		AddRawArgument(jsoncdc.MustEncode(cadenceName)).
		AddRawArgument(jsoncdc.MustEncode(cadenceCode)).
		AddAuthorizer(signer.Address())

	for _, arg := range args {
		arg.Type().ID()
		tx.AddRawArgument(jsoncdc.MustEncode(arg))
	}

	txArgs, addArgs := "", ""
	for i, arg := range args {
		txArgs += fmt.Sprintf(",arg%d:%s", i, arg.Type().ID())
		addArgs += fmt.Sprintf(",arg%d", i)
	}

	script := fmt.Sprintf(addAccountContractTemplate, txArgs, addArgs)
	tx.SetScript([]byte(script))
	tx.SetGasLimit(maxGasLimit)

	t := &Transaction{tx: tx}
	err := t.SetSigner(signer)
	if err != nil {
		return nil, err
	}
	t.SetPayer(signer.Address())

	return t, nil
}

// NewCreateAccountTransaction creates new transaction for account.
func NewCreateAccountTransaction(
	signer *Account,
	keys []*flow.AccountKey,
	contracts []templates.Contract,
) (*Transaction, error) {
	template, err := templates.CreateAccount(keys, contracts, signer.Address())
	if err != nil {
		return nil, err
	}

	return newTransactionFromTemplate(template, signer)
}

func newTransactionFromTemplate(templateTx *flow.Transaction, signer *Account) (*Transaction, error) {
	tx := &Transaction{tx: templateTx}

	err := tx.SetSigner(signer)
	if err != nil {
		return nil, err
	}
	tx.SetPayer(signer.Address())
	tx.SetGasLimit(maxGasLimit) // todo change this to calculated limit

	return tx, nil
}

// Transaction builder of flow transactions.
type Transaction struct {
	signer   *Account
	proposer *flow.Account
	tx       *flow.Transaction
}

// Signer get signer.
func (t *Transaction) Signer() *Account {
	return t.signer
}

// Proposer get proposer.
func (t *Transaction) Proposer() *flow.Account {
	return t.proposer
}

// FlowTransaction get flow transaction.
func (t *Transaction) FlowTransaction() *flow.Transaction {
	return t.tx
}

func (t *Transaction) SetScriptWithArgs(script []byte, args []cadence.Value) error {
	t.tx.SetScript(script)
	return t.AddArguments(args)
}

// SetSigner sets the signer for transaction.
func (t *Transaction) SetSigner(account *Account) error {
	err := account.Key().Validate()
	if err != nil {
		return err
	}

	if !t.validSigner(account.Address()) {
		return fmt.Errorf(
			"not a valid signer %s, proposer: %s, payer: %s, authorizers: %s",
			account.Address(),
			t.tx.ProposalKey.Address,
			t.tx.Payer,
			t.tx.Authorizers,
		)
	}

	t.signer = account
	return nil
}

// validSigner checks whether the signer is valid for transaction
func (t *Transaction) validSigner(s flow.Address) bool {
	return t.tx.ProposalKey.Address == s ||
		t.tx.Payer == s ||
		t.authorizersContains(s)
}

// authorizersContains checks whether address is in the authorizer list
func (t *Transaction) authorizersContains(address flow.Address) bool {
	for _, a := range t.tx.Authorizers {
		if address == a {
			return true
		}
	}

	return false
}

// SetProposer sets the proposer for transaction.
func (t *Transaction) SetProposer(proposer *flow.Account, keyIndex int) error {
	if len(proposer.Keys) <= keyIndex {
		return fmt.Errorf("failed to retrieve proposer key at index %d", keyIndex)
	}

	t.proposer = proposer
	proposerKey := proposer.Keys[keyIndex]

	t.tx.SetProposalKey(
		proposer.Address,
		proposerKey.Index,
		proposerKey.SequenceNumber,
	)

	return nil
}

// SetPayer sets the payer for transaction.
func (t *Transaction) SetPayer(address flow.Address) *Transaction {
	t.tx.SetPayer(address)
	return t
}

// SetBlockReference sets block reference for transaction.
func (t *Transaction) SetBlockReference(block *flow.Block) *Transaction {
	t.tx.SetReferenceBlockID(block.ID)
	return t
}

// SetGasLimit sets the gas limit for transaction.
func (t *Transaction) SetGasLimit(gasLimit uint64) *Transaction {
	t.tx.SetGasLimit(gasLimit)
	return t
}

// AddArguments add array of cadence arguments.
func (t *Transaction) AddArguments(args []cadence.Value) error {
	for _, arg := range args {
		err := t.AddArgument(arg)
		if err != nil {
			return err
		}
	}

	return nil
}

// AddArgument add cadence typed argument.
func (t *Transaction) AddArgument(arg cadence.Value) error {
	return t.tx.AddArgument(arg)
}

// AddAuthorizers add group of authorizers.
func (t *Transaction) AddAuthorizers(authorizers []flow.Address) (*Transaction, error) {
	program, _ := cmd.PrepareProgram(
		string(t.tx.Script),
		common.StringLocation(""),
		map[common.LocationID]string{},
	)

	// get authorizers param list if exists
	if len(program.TransactionDeclarations()) == 1 {
		declaration := program.TransactionDeclarations()[0]
		requiredAuths := make([]*ast.Parameter, 0)

		// if prepare block is missing set default authorizers to empty
		if declaration.Prepare == nil {
			authorizers = nil
		} else { // if prepare block is present get authorizers
			requiredAuths = declaration.
				Prepare.
				FunctionDeclaration.
				ParameterList.
				Parameters
		}

		if len(requiredAuths) != len(authorizers) {
			return nil, fmt.Errorf(
				"provided authorizers length mismatch, required authorizers %d, but provided %d",
				len(requiredAuths),
				len(authorizers),
			)
		}
	} else {
		return nil, fmt.Errorf("can only support one transaction declaration per file")
	}

	for _, authorizer := range authorizers {
		t.tx.AddAuthorizer(authorizer)
	}

	return t, nil
}

// Sign signs transaction using signer account.
func (t *Transaction) Sign() (*Transaction, error) {
	keyIndex := t.signer.Key().Index()
	signer, err := t.signer.Key().Signer(context.Background())
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

// shouldSignEnvelope checks if signer should sign envelope or payload
func (t *Transaction) shouldSignEnvelope() bool {
	return t.signer.address == t.tx.Payer
}
