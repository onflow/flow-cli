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
	SignerRoleAuthorizer signerRole = "authorizer"
	SignerRoleProposer   signerRole = "proposer"
	SignerRolePayer      signerRole = "payer"
)

func NewTransaction() *Transaction {
	return &Transaction{
		tx: flow.NewTransaction(),
	}
}

func NewTransactionFromPayload(signer *Account, filename string, role string) (*Transaction, error) {
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

	err = tx.SetSigner(signer)
	if err != nil {
		return nil, err
	}
	// we need to set the role here for signing purpose, so we know what to sign envelope or payload
	tx.signerRole = signerRole(role)

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
		contractName := contractFlagContent[0]
		contractPath := contractFlagContent[1]

		contractSource, err := util.LoadFile(contractPath)
		if err != nil {
			return nil, err
		}

		contracts = append(contracts, templates.Contract{
			Name:   contractName,
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
	proposer   *Account
	tx         *flow.Transaction
}

func (t *Transaction) Signer() *Account {
	return t.signer
}

func (t *Transaction) Proposer() *Account {
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

func (t *Transaction) SetProposer(account *Account) error {
	err := account.DefaultKey().Validate()
	if err != nil {
		return err
	}

	t.proposer = account
	return nil
}

func (t *Transaction) SetPayer(address flow.Address) {
	t.tx.SetPayer(address)
}

func (t *Transaction) HasPayer() bool {
	return t.tx.Payer != flow.EmptyAddress
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

func (t *Transaction) AddAuthorizers(addresses []string) error {
	for _, address := range addresses {
		err := t.AddAuthorizer(address)
		if err != nil { // return error even if one breaks
			return err
		}
	}

	return nil
}

func (t *Transaction) AddAuthorizer(address string) error {
	authorizerAddress := flow.HexToAddress(address)
	if authorizerAddress == flow.EmptyAddress {
		return fmt.Errorf("invalid authorizer address provided %s", address)
	}

	t.tx.AddAuthorizer(authorizerAddress)
	return nil
}

func (t *Transaction) SetSignerRole(role string) error {
	t.signerRole = signerRole(role)

	if t.signerRole == SignerRoleAuthorizer {
		err := t.AddAuthorizer(t.signer.Address().String())
		if err != nil {
			return err
		}
	}
	if t.signerRole == SignerRolePayer && t.tx.Payer != t.signer.Address() {
		return fmt.Errorf("role specified as Payer, but Payer address also provided, and different: %s != %s", t.tx.Payer, t.signer.Address())
	}

	return nil
}

func (t *Transaction) Sign() (*Transaction, error) {
	keyIndex := t.signer.DefaultKey().Index()
	signerAddress := t.signer.Address()
	signer, err := t.signer.DefaultKey().Signer(context.Background())
	if err != nil {
		return nil, err
	}

	if t.signerRole == SignerRoleAuthorizer || t.signerRole == SignerRoleProposer {
		err := t.tx.SignPayload(signerAddress, keyIndex, signer)
		if err != nil {
			return nil, fmt.Errorf("failed to sign transaction: %s", err)
		}
	} else {
		// make sure we have at least signer as authorizer else add self
		if len(t.tx.Authorizers) == 0 {
			t.tx.AddAuthorizer(t.signer.Address())
		}

		err := t.tx.SignEnvelope(signerAddress, keyIndex, signer)
		if err != nil {
			return nil, fmt.Errorf("failed to sign transaction: %s", err)
		}
	}

	return t, nil
}
