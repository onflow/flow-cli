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

	"github.com/onflow/flow-cli/pkg/flowcli/contracts"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flow-cli/pkg/flowcli/gateway"
	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-cli/pkg/flowcli/util"
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

// Build builds a transaction with specified payer, proposer and authorizer
func (t *Transactions) Build(
	proposer string,
	authorizer []string,
	payer string,
	proposerKeyIndex int,
	codeFilename string,
	args []string,
	argsJSON string,
	network string,
) (*project.Transaction, error) {
	proposerAddress, err := getAddressFromStringOrConfig(proposer, t.project)
	if err != nil {
		return nil, err
	}

	payerAddress, err := getAddressFromStringOrConfig(payer, t.project)
	if err != nil {
		return nil, err
	}

	authorizerAddresses := make([]flow.Address, 0)
	for _, a := range authorizer {
		authorizerAddress, err := getAddressFromStringOrConfig(a, t.project)
		if err != nil {
			return nil, err
		}
		authorizerAddresses = append(authorizerAddresses, authorizerAddress)
	}

	latestBlock, err := t.gateway.GetLatestBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest sealed block: %w", err)
	}

	proposerAccount, err := t.gateway.GetAccount(proposerAddress)
	if err != nil {
		return nil, err
	}

	tx := project.NewTransaction().
		SetPayer(payerAddress).
		SetProposer(proposerAccount, proposerKeyIndex).
		AddAuthorizers(authorizerAddresses).
		SetDefaultGasLimit().
		SetBlockReference(latestBlock)

	code, err := util.LoadFile(codeFilename)
	if err != nil {
		return nil, err
	}

	resolver, err := contracts.NewResolver(code)
	if err != nil {
		return nil, err
	}

	if resolver.HasFileImports() {
		if network == "" {
			return nil, fmt.Errorf("missing network, specify which network to use to resolve imports in transaction code")
		}
		if codeFilename == "" { // when used as lib with code we don't support imports
			return nil, fmt.Errorf("resolving imports in transactions not supported")
		}

		code, err = resolver.ResolveImports(
			codeFilename,
			t.project.ContractsByNetwork(network),
			t.project.AliasesForNetwork(network),
		)
		if err != nil {
			return nil, err
		}
	}

	err = tx.SetScriptWithArgs(code, args, argsJSON)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// Sign transaction
func (t *Transactions) Sign(
	payloadFilename string,
	signerName string,
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

	tx, err := project.NewTransactionFromPayload(payloadFilename)
	if err != nil {
		return nil, err
	}

	err = tx.SetSigner(signerAccount)
	if err != nil {
		return nil, err
	}

	if approveSigning {
		return tx.Sign()
	}

	if !output.ApproveTransactionPrompt(tx) {
		return nil, fmt.Errorf("transaction was not approved for signing")
	}

	return tx.Sign()
}

// SendSigned sends the transaction that is already signed
func (t *Transactions) SendSigned(
	signedFilename string,
) (*flow.Transaction, *flow.TransactionResult, error) {
	tx, err := project.NewTransactionFromPayload(signedFilename)
	if err != nil {
		return nil, nil, err
	}

	sentTx, err := t.gateway.SendSignedTransaction(tx)
	if err != nil {
		return nil, nil, err
	}

	res, err := t.gateway.GetTransactionResult(sentTx, true)
	if err != nil {
		return nil, nil, err
	}

	return sentTx, res, nil
}

// Send sends a transaction from a file.
func (t *Transactions) Send(
	transactionFilename string,
	signerName string,
	args []string,
	argsJSON string,
	network string,
) (*flow.Transaction, *flow.TransactionResult, error) {
	if t.project == nil {
		return nil, nil, fmt.Errorf("missing configuration, initialize it: flow project init")
	}

	signerAccount := t.project.AccountByName(signerName)
	if signerAccount == nil {
		return nil, nil, fmt.Errorf("signer account: [%s] doesn't exists in configuration", signerName)
	}

	tx, err := t.Build(
		signerName,
		[]string{signerName},
		signerName,
		0, // default 0 key
		transactionFilename,
		args,
		argsJSON,
		network,
	)
	if err != nil {
		return nil, nil, err
	}

	err = tx.SetSigner(signerAccount)
	if err != nil {
		return nil, nil, err
	}

	signed, err := tx.Sign()
	if err != nil {
		return nil, nil, err
	}

	t.logger.StartProgress("Sending Transaction...")
	defer t.logger.StopProgress()

	sentTx, err := t.gateway.SendSignedTransaction(signed)
	if err != nil {
		return nil, nil, err
	}

	t.logger.StartProgress("Waiting for transaction to be sealed...")

	res, err := t.gateway.GetTransactionResult(sentTx, true)

	t.logger.StopProgress()

	return sentTx, res, err
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
	defer t.logger.StopProgress()

	sentTx, err := t.gateway.SendSignedTransaction(tx)
	if err != nil {
		return nil, nil, err
	}

	t.logger.StartProgress("Waiting for transaction to be sealed...")

	res, err := t.gateway.GetTransactionResult(sentTx, true)

	t.logger.StopProgress()
	return sentTx, res, err
}

// getAddressFromStringOrConfig try to parse value as address or as an account from the config
func getAddressFromStringOrConfig(value string, project *project.Project) (flow.Address, error) {
	if util.ValidAddress(value) {
		return flow.HexToAddress(value), nil
	} else if project != nil {
		account := project.AccountByName(value)
		if account == nil {
			return flow.EmptyAddress, fmt.Errorf("account could not be found")
		}
		return account.Address(), nil
	} else {
		return flow.EmptyAddress, fmt.Errorf("could not parse address or account name from config, missing configuration")
	}
}
