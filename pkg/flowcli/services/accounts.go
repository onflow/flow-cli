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

	"github.com/onflow/flow-cli/pkg/flowcli/config"

	"github.com/onflow/flow-cli/pkg/flowcli"
	"github.com/onflow/flow-cli/pkg/flowcli/gateway"
	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-cli/pkg/flowcli/util"

	"github.com/onflow/cadence"
	tmpl "github.com/onflow/flow-core-contracts/lib/go/templates"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

// Accounts is a service that handles all account-related interactions.
type Accounts struct {
	gateway gateway.Gateway
	project *project.Project
	logger  output.Logger
}

// NewAccounts returns a new accounts service.
func NewAccounts(
	gateway gateway.Gateway,
	project *project.Project,
	logger output.Logger,
) *Accounts {
	return &Accounts{
		gateway: gateway,
		project: project,
		logger:  logger,
	}
}

// Get returns an account by on address.
func (a *Accounts) Get(address string) (*flow.Account, error) {
	a.logger.StartProgress(fmt.Sprintf("Loading %s...", address))

	flowAddress := flow.HexToAddress(address)

	account, err := a.gateway.GetAccount(flowAddress)
	a.logger.StopProgress()

	return account, err
}

// StakingInfo returns the staking information for an account.
func (a *Accounts) StakingInfo(accountAddress string) (*cadence.Value, *cadence.Value, error) {
	a.logger.StartProgress(fmt.Sprintf("Fetching info for %s...", accountAddress))
	defer a.logger.StopProgress()

	address := flow.HexToAddress(accountAddress)

	cadenceAddress := []cadence.Value{cadence.NewAddress(address)}

	chain, err := util.GetAddressNetwork(address)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to determine network from address, check the address and network",
		)
	}

	if chain == flow.Emulator {
		return nil, nil, fmt.Errorf("emulator chain not supported")
	}

	env := util.EnvFromNetwork(chain)

	stakingInfoScript := tmpl.GenerateGetLockedStakerInfoScript(env)
	delegationInfoScript := tmpl.GenerateGetLockedDelegatorInfoScript(env)

	stakingValue, err := a.gateway.ExecuteScript(stakingInfoScript, cadenceAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting staking info: %s", err.Error())
	}

	delegationValue, err := a.gateway.ExecuteScript(delegationInfoScript, cadenceAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting delegation info: %s", err.Error())
	}

	a.logger.StopProgress()

	return &stakingValue, &delegationValue, nil
}

// Create creates and returns a new account.
//
// The new account is created with the given public keys and contracts.
//
// The account creation transaction is signed by the specified signer.
func (a *Accounts) Create(
	signerName string,
	keys []string,
	keyWeights []int,
	signatureAlgorithm string,
	hashingAlgorithm string,
	contracts []string,
) (*flow.Account, error) {
	if a.project == nil {
		return nil, config.ErrDoesNotExist
	}

	// if more than one key is provided and at least one weight is specified, make sure there isn't a missmatch
	if len(keys) > 1 && len(keyWeights) > 0 && len(keys) != len(keyWeights) {
		return nil, fmt.Errorf(
			"number of keys and weights provided must match, number of provided keys: %d, number of provided key weights: %d",
			len(keys),
			len(keyWeights),
		)
	}

	signer := a.project.AccountByName(signerName)
	if signer == nil {
		return nil, fmt.Errorf("signer account: [%s] doesn't exists in configuration", signerName)
	}

	accountKeys := make([]*flow.AccountKey, len(keys))

	sigAlgo, hashAlgo, err := util.ConvertSigAndHashAlgo(signatureAlgorithm, hashingAlgorithm)
	if err != nil {
		return nil, err
	}

	for i, publicKeyHex := range keys {
		publicKey, err := crypto.DecodePublicKeyHex(
			sigAlgo,
			strings.ReplaceAll(publicKeyHex, "0x", ""),
		)
		if err != nil {
			return nil, fmt.Errorf(
				"could not decode public key for key: %s, with signature algorith: %s",
				publicKeyHex,
				sigAlgo,
			)
		}

		weight := flow.AccountKeyWeightThreshold
		if len(keyWeights) > i {
			weight = keyWeights[i]

			if weight > flow.AccountKeyWeightThreshold || weight <= 0 {
				return nil, fmt.Errorf("invalid key weight, valid range (0 - 1000)")
			}
		}

		accountKeys[i] = &flow.AccountKey{
			PublicKey: publicKey,
			SigAlgo:   sigAlgo,
			HashAlgo:  hashAlgo,
			Weight:    weight,
		}
	}

	tx, err := project.NewCreateAccountTransaction(signer, accountKeys, contracts)
	if err != nil {
		return nil, err
	}

	tx, err = a.prepareTransaction(tx, signer)
	if err != nil {
		return nil, err
	}

	a.logger.Info(fmt.Sprintf("Transaction ID: %s", tx.FlowTransaction().ID()))
	a.logger.StartProgress("Creating account...")
	defer a.logger.StopProgress()

	sentTx, err := a.gateway.SendSignedTransaction(tx)
	if err != nil {
		return nil, err
	}

	a.logger.StartProgress("Waiting for transaction to be sealed...")

	result, err := a.gateway.GetTransactionResult(sentTx, true)
	if err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, result.Error
	}

	events := flowcli.EventsFromTransaction(result)
	newAccountAddress := events.GetAddress()

	if newAccountAddress == nil {
		return nil, fmt.Errorf("new account address couldn't be fetched")
	}

	a.logger.StopProgress()

	return a.gateway.GetAccount(*newAccountAddress)
}

// AddContract adds a new contract to an account and returns the updated account.
func (a *Accounts) AddContract(
	accountName string,
	contractName string,
	contractFilename string,
	updateExisting bool,
) (*flow.Account, error) {
	if a.project == nil {
		return nil, config.ErrDoesNotExist
	}

	account := a.project.AccountByName(accountName)
	if account == nil {
		return nil, fmt.Errorf("account: [%s] doesn't exists in configuration", accountName)
	}

	contractSource, err := util.LoadFile(contractFilename)
	if err != nil {
		return nil, err
	}

	return a.addContract(account, contractName, contractSource, updateExisting)
}

// AddContractForAddress adds a new contract to an address using private key specified
func (a *Accounts) AddContractForAddress(
	accountAddress string,
	accountPrivateKey string,
	contractName string,
	contractFilename string,
	updateExisting bool,
) (*flow.Account, error) {
	account, err := accountFromAddressAndKey(accountAddress, accountPrivateKey)
	if err != nil {
		return nil, err
	}

	contractSource, err := util.LoadFile(contractFilename)
	if err != nil {
		return nil, err
	}

	return a.addContract(account, contractName, contractSource, updateExisting)
}

// AddContractForAddressWithCode adds a new contract to an address using private key and code specified
func (a *Accounts) AddContractForAddressWithCode(
	accountAddress string,
	accountPrivateKey string,
	contractName string,
	contractCode []byte,
	updateExisting bool,
) (*flow.Account, error) {
	account, err := accountFromAddressAndKey(accountAddress, accountPrivateKey)
	if err != nil {
		return nil, err
	}

	return a.addContract(account, contractName, contractCode, updateExisting)
}

func (a *Accounts) addContract(
	account *project.Account,
	contractName string,
	contractSource []byte,
	updateExisting bool,
) (*flow.Account, error) {
	tx, err := project.NewAddAccountContractTransaction(
		account,
		contractName,
		string(contractSource),
		[]cadence.Value{}, // todo add support for args on account add-contract
	)
	if err != nil {
		return nil, err
	}

	// if we are updating contract
	if updateExisting {
		tx, err = project.NewUpdateAccountContractTransaction(
			account,
			contractName,
			string(contractSource),
		)
		if err != nil {
			return nil, err
		}
	}

	tx, err = a.prepareTransaction(tx, account)
	if err != nil {
		return nil, err
	}

	a.logger.Info(fmt.Sprintf("Transaction ID: %s", tx.FlowTransaction().ID()))

	status := "Adding contract '%s' to account '%s'..."
	if updateExisting {
		status = "Updating contract '%s' on account '%s'..."
	}

	a.logger.StartProgress(
		fmt.Sprintf(
			status,
			contractName,
			account.Address(),
		),
	)
	defer a.logger.StopProgress()

	// send transaction with contract
	sentTx, err := a.gateway.SendSignedTransaction(tx)
	if err != nil {
		return nil, err
	}

	// we wait for transaction to be sealed
	trx, err := a.gateway.GetTransactionResult(sentTx, true)
	if err != nil {
		return nil, err
	}

	if trx.Error != nil {
		a.logger.Error("Failed to deploy contract")
		return nil, trx.Error
	}

	update, err := a.gateway.GetAccount(account.Address())

	a.logger.StopProgress()

	if updateExisting {
		a.logger.Info(fmt.Sprintf(
			"Contract '%s' updated on the account '%s'.",
			contractName,
			account.Address(),
		))
	} else {
		a.logger.Info(fmt.Sprintf(
			"Contract '%s' deployed to the account '%s'.",
			contractName,
			account.Address(),
		))
	}

	return update, err
}

// RemoveContracts removes a contract from an account and returns the updated account.
func (a *Accounts) RemoveContract(
	contractName string,
	accountName string,
) (*flow.Account, error) {
	account := a.project.AccountByName(accountName)
	if account == nil {
		return nil, fmt.Errorf("account: [%s] doesn't exists in configuration", accountName)
	}

	tx, err := project.NewRemoveAccountContractTransaction(account, contractName)
	if err != nil {
		return nil, err
	}

	tx, err = a.prepareTransaction(tx, account)
	if err != nil {
		return nil, err
	}

	a.logger.Info(fmt.Sprintf("Transaction ID: %s", tx.FlowTransaction().ID().String()))
	a.logger.StartProgress(
		fmt.Sprintf("Removing contract %s from %s...", contractName, account.Address()),
	)
	defer a.logger.StopProgress()

	sentTx, err := a.gateway.SendSignedTransaction(tx)
	if err != nil {
		return nil, err
	}

	txr, err := a.gateway.GetTransactionResult(sentTx, true)
	if err != nil {
		return nil, err
	}
	if txr != nil && txr.Error != nil {
		a.logger.Error("Removing contract failed")
		return nil, txr.Error
	}

	a.logger.StopProgress()
	a.logger.Info(fmt.Sprintf(
		"Contract %s removed from account %s.",
		contractName,
		account.Address(),
	))

	return a.gateway.GetAccount(account.Address())
}

// prepareTransaction prepares transaction for sending with data from network
func (a *Accounts) prepareTransaction(
	tx *project.Transaction,
	account *project.Account,
) (*project.Transaction, error) {

	block, err := a.gateway.GetLatestBlock()
	if err != nil {
		return nil, err
	}

	proposer, err := a.gateway.GetAccount(account.Address())
	if err != nil {
		return nil, err
	}

	tx.SetBlockReference(block).
		SetProposer(proposer, account.Key().Index())

	tx, err = tx.Sign()
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// AccountFromAddressAndKey get account from address and private key
func accountFromAddressAndKey(address string, accountPrivateKey string) (*project.Account, error) {
	privateKey, err := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, accountPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("private key is not correct")
	}

	return project.AccountFromAddressAndKey(
		flow.HexToAddress(address),
		privateKey,
	), nil
}
