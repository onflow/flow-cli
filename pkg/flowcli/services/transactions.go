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

	"github.com/onflow/cadence"

	"github.com/onflow/flow-cli/pkg/flowcli/contracts"

	"github.com/onflow/flow-go-sdk"

	"github.com/onflow/flow-cli/pkg/flowcli/gateway"
	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
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

	result, err := t.gateway.GetTransactionResult(tx, waitSeal)

	t.logger.StopProgress()

	return tx, result, err
}

// Build builds a transaction with specified payer, proposer and authorizer
func (t *Transactions) Build(
	proposer flow.Address,
	authorizers []flow.Address,
	payer flow.Address,
	proposerKeyIndex int,
	code []byte,
	codeFilename string,
	gasLimit uint64,
	args []cadence.Value,
	network string,
) (*project.Transaction, error) {

	latestBlock, err := t.gateway.GetLatestBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest sealed block: %w", err)
	}

	proposerAccount, err := t.gateway.GetAccount(proposer)
	if err != nil {
		return nil, err
	}

	tx := project.NewTransaction().
		SetPayer(payer).
		SetProposer(proposerAccount, proposerKeyIndex).
		AddAuthorizers(authorizers).
		SetGasLimit(gasLimit).
		SetBlockReference(latestBlock)

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

		contractsNetwork, err := t.project.DeploymentContractsByNetwork(network)
		if err != nil {
			return nil, err
		}

		code, err = resolver.ResolveImports(
			codeFilename,
			contractsNetwork,
			t.project.AliasesForNetwork(network),
		)
		if err != nil {
			return nil, err
		}
	}

	err = tx.SetScriptWithArgs(code, args)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// Sign transaction
func (t *Transactions) Sign(
	payload []byte,
	signer project.Account,
	approveSigning bool,
) (*project.Transaction, error) {
	if t.project == nil {
		return nil, fmt.Errorf("missing configuration, initialize it: flow project init")
	}

	tx, err := project.NewTransactionFromPayload(payload)
	if err != nil {
		return nil, err
	}

	err = tx.SetSigner(signer)
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
	payload []byte,
) (*flow.Transaction, *flow.TransactionResult, error) {
	tx, err := project.NewTransactionFromPayload(payload)
	if err != nil {
		return nil, nil, err
	}

	t.logger.StartProgress(fmt.Sprintf("Sending Transaction with ID: %s", tx.FlowTransaction().ID()))
	defer t.logger.StopProgress()

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
	code []byte,
	signer project.Account,
	codeFilename string,
	gasLimit uint64,
	args []cadence.Value,
	network string,
) (*flow.Transaction, *flow.TransactionResult, error) {
	if t.project == nil {
		return nil, nil, fmt.Errorf("missing configuration, initialize it: flow project init")
	}

	signerKeyIndex := signerAccount.Key().Index()

	tx, err := t.Build(
		signer.Address(),
		[]flow.Address{signer.Address()},
		signer.Address(),
		signerKeyIndex,
		code,
		codeFilename,
		gasLimit,
		args,
		network,
	)
	if err != nil {
		return nil, nil, err
	}

	err = tx.SetSigner(signer)
	if err != nil {
		return nil, nil, err
	}

	signed, err := tx.Sign()
	if err != nil {
		return nil, nil, err
	}

	t.logger.Info(fmt.Sprintf("Transaction ID: %s", signed.FlowTransaction().ID()))
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
