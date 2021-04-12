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

package services

import (
	"fmt"
	"strings"

	"github.com/onflow/flow-cli/pkg/flowcli"
	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"

	"github.com/onflow/flow-cli/pkg/flowcli/config"
	"github.com/onflow/flow-cli/pkg/flowcli/gateway"
	"github.com/onflow/flow-cli/pkg/flowcli/util"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

// Transactions is a service that handles all transaction-related interactions.
type Transactions struct {
	gateway gateway.Gateway
	project *project.Project
	logger  output.Logger
}

// NewTransactions returns a new transactions service.
func NewTransactions(
	gateway gateway.Gateway,
	project *project.Project,
	logger output.Logger,
) *Transactions {
	return &Transactions{
		gateway: gateway,
		project: project,
		logger:  logger,
	}
}

// GetStatus of transaction
func (t *Transactions) GetStatus(
	transactionID string,
	waitSeal bool,
) (*flow.Transaction, *flow.TransactionResult, error) {
	txID := flow.HexToID(
		strings.ReplaceAll(transactionID, "0x", ""),
	)

	t.logger.StartProgress("Fetching Transaction...")

	tx, err := t.gateway.GetTransaction(txID)
	if err != nil {
		return nil, nil, err
	}

	if waitSeal {
		t.logger.StartProgress("Waiting for transaction to be sealed...")
	}

	result, err := t.gateway.GetTransactionResult(tx, waitSeal)

	t.logger.StopProgress()

	return tx, result, err
}

// Sign transaction
func (t *Transactions) Sign(
	signerName string,
	proposerName string,
	payerAddress string,
	additionalAuthorizers []string,
	role string,
	scriptFilename string,
	payloadFilename string,
	args []string,
	argsJSON string,
	approveSigning bool,
) (*project.Transaction, error) {
	if t.project == nil {
		return nil, fmt.Errorf("missing configuration, initialize it: flow project init")
	}

	// get the signer account
	signerAccount := t.project.AccountByName(signerName)
	if signerAccount == nil {
		return nil, fmt.Errorf("signer account: [%s] doesn't exists in configuration", signerName)
	}

	if payloadFilename != "" && scriptFilename != "" {
		return nil, fmt.Errorf("can not use both a transaction payload and Cadence code file")
	} else if payloadFilename == "" && scriptFilename == "" {
		return nil, fmt.Errorf("provide either a transaction payload or Cadence code file")
	}

	// if we received already created transaction payload, create from payload and return signed
	if payloadFilename != "" {
		tx, err := project.NewTransactionFromPayload(signerAccount, payloadFilename, role)
		if err != nil {
			return nil, err
		}

		if approveSigning {
			return tx.Sign()
		}

		if output.ApproveTransactionPrompt(tx) {
			return tx.Sign()
		} else {
			return nil, fmt.Errorf("transaction was not approved for signing")
		}
	}

	// we are creating a new transaction
	tx := project.NewTransaction()

	err := tx.SetSigner(signerAccount)
	if err != nil {
		return nil, err
	}

	err = tx.SetSignerRole(role)
	if err != nil {
		return nil, err
	}

	// if proposer is specified and exists assign it
	if proposerName != "" {
		proposerAccount := t.project.AccountByName(proposerName)
		if proposerAccount == nil {
			return nil, fmt.Errorf("proposer account: [%s] doesn't exists in configuration", signerName)
		}

		err = tx.SetProposer(proposerAccount)
		if err != nil {
			return nil, err
		}
	}

	// set payer if specified, else set current signer as payer if tx doesn't have one yet associated
	if payerAddress != "" {
		tx.SetPayer(flow.HexToAddress(payerAddress))
	} else if !tx.HasPayer() {
		tx.SetPayer(signerAccount.Address())
	}

	err = tx.SetScriptWithArgsFromFile(scriptFilename, args, argsJSON)
	if err != nil {
		return nil, err
	}

	err = tx.AddAuthorizers(additionalAuthorizers)
	if err != nil {
		return nil, err
	}

	tx, err = t.gateway.PrepareTransactionPayload(tx)
	if err != nil {
		return nil, err
	}

	return tx.Sign()
}

// Send sends a transaction from a file.
func (t *Transactions) Send(
	transactionFilename string,
	payloadFilename string,
	signerName string,
	args []string,
	argsJSON string,
) (*flow.Transaction, *flow.TransactionResult, error) {

	signed, err := t.Sign(
		signerName,
		"",
		"",
		[]string{},
		"",
		transactionFilename,
		payloadFilename,
		args,
		argsJSON,
		true,
	)
	if err != nil {
		return nil, nil, err
	}

	t.logger.StartProgress("Sending Transaction...")

	tx, err := t.gateway.SendSignedTransaction(signed)
	if err != nil {
		return nil, nil, err
	}

	t.logger.StartProgress("Waiting for transaction to be sealed...")

	res, err := t.gateway.GetTransactionResult(tx, true)

	t.logger.StopProgress()

	return tx, res, err
}

// SendForAddressWithCode send transaction for address and private key specified with code
func (t *Transactions) SendForAddressWithCode(
	code []byte,
	signerAddress string,
	signerPrivateKey string,
	args []string,
	argsJSON string,
) (*flow.Transaction, *flow.TransactionResult, error) {
	address := flow.HexToAddress(signerAddress)

	privateKey, err := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, signerPrivateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("private key is not correct")
	}

	signer := project.AccountFromAddressAndKey(address, privateKey)

	tx := project.NewTransaction()
	err = tx.SetSigner(signer)
	if err != nil {
		return nil, nil, err
	}

	err = tx.SetScriptWithArgs(code, args, argsJSON)
	if err != nil {
		return nil, nil, err
	}

	tx, err = tx.Sign()
	if err != nil {
		return nil, nil, err
	}

	t.logger.StartProgress("Sending transaction...")

	sentTx, err := t.gateway.SendSignedTransaction(tx)
	if err != nil {
		return nil, nil, err
	}

	t.logger.StartProgress("Waiting for transaction to be sealed...")

	res, err := t.gateway.GetTransactionResult(sentTx, true)

	t.logger.StopProgress()
	return sentTx, res, err
}
