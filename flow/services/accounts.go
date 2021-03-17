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

	"github.com/onflow/flow-cli/flow/lib"

	"github.com/onflow/flow-cli/flow/gateway"
	"github.com/onflow/flow-cli/flow/util"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/templates"
)

// Accounts service handles all interactions for accounts
type Accounts struct {
	gateway gateway.Gateway
	project *lib.Project
	logger  util.Logger
}

// NewAccounts create new account service
func NewAccounts(
	gateway gateway.Gateway,
	project *lib.Project,
	logger util.Logger,
) *Accounts {
	return &Accounts{
		gateway: gateway,
		project: project,
		logger:  logger,
	}
}

// Get gets an account based on address
func (a *Accounts) Get(address string) (*flow.Account, error) {
	flowAddress := flow.HexToAddress(
		strings.ReplaceAll(address, "0x", ""),
	)

	return a.gateway.GetAccount(flowAddress)
}

// Create creates an account with signer name, keys, algorithms, contracts and returns the new account
func (a *Accounts) Create(
	signerName string,
	keys []string,
	signatureAlgorithm string,
	hashingAlgorithm string,
	contracts []string,
) (*flow.Account, error) {

	signer := a.project.GetAccountByName(signerName)
	if signer == nil {
		return nil, fmt.Errorf("signer account: [%s] doesn't exists in configuration", signerName)
	}

	accountKeys := make([]*flow.AccountKey, len(keys))

	sigAlgo := crypto.StringToSignatureAlgorithm(signatureAlgorithm)
	if sigAlgo == crypto.UnknownSignatureAlgorithm {
		return nil, fmt.Errorf("failed to determine signature algorithm from %s", sigAlgo)
	}
	hashAlgo := crypto.StringToHashAlgorithm(hashingAlgorithm)
	if hashAlgo == crypto.UnknownHashAlgorithm {
		return nil, fmt.Errorf("failed to determine hash algorithm from %s", hashAlgo)
	}

	for i, publicKeyHex := range keys {
		publicKey, err := crypto.DecodePublicKeyHex(
			sigAlgo,
			strings.ReplaceAll(publicKeyHex, "0x", ""),
		)
		if err != nil {
			return nil, fmt.Errorf("could not decode public key for key: %s, with signature algorith: %s", publicKeyHex, sigAlgo)
		}

		accountKeys[i] = &flow.AccountKey{
			PublicKey: publicKey,
			SigAlgo:   sigAlgo,
			HashAlgo:  hashAlgo,
			Weight:    flow.AccountKeyWeightThreshold,
		}
	}

	var contractTemplates []templates.Contract

	for _, contract := range contracts {
		contractFlagContent := strings.SplitN(contract, ":", 2)
		if len(contractFlagContent) != 2 {
			return nil, fmt.Errorf("wrong format for contract flag. Correct format is name:path, but got: %s", contract)
		}
		contractName := contractFlagContent[0]
		contractPath := contractFlagContent[1]

		contractSource, err := util.LoadFile(contractPath)
		if err != nil {
			return nil, err
		}

		contractTemplates = append(contractTemplates,
			templates.Contract{
				Name:   contractName,
				Source: string(contractSource),
			},
		)
	}

	tx := templates.CreateAccount(accountKeys, contractTemplates, signer.Address())
	tx, err := a.gateway.SendTransaction(tx, signer)
	if err != nil {
		return nil, err
	}

	a.logger.StartProgress("Waiting for transaction to be sealed...")

	result, err := a.gateway.GetTransactionResult(tx)
	if err != nil {
		return nil, err
	}

	a.logger.StopProgress("")

	events := lib.EventsFromTransaction(result)
	newAccountAddress := events.GetAddress()

	if newAccountAddress == nil {
		return nil, fmt.Errorf("new account address couldn't be fetched")
	}

	return a.gateway.GetAccount(*newAccountAddress)
}

// AddContract adds new contract to the account and returns the updated account
func (a *Accounts) AddContract(
	accountName string,
	contractName string,
	contractFilename string,
	updateExisting bool,
) (*flow.Account, error) {

	account := a.project.GetAccountByName(accountName)
	if account == nil {
		return nil, fmt.Errorf("account: [%s] doesn't exists in configuration", accountName)
	}

	return a.addContract(account, contractName, contractFilename, updateExisting)
}

// AddContractForAddress adds new contract to the address using private key specified
func (a *Accounts) AddContractForAddress(
	accountAddress string,
	accountPrivateKey string,
	contractName string,
	contractFilename string,
	updateExisting bool,
) (*flow.Account, error) {
	account, err := util.AccountFromAddressAndKey(accountAddress, accountPrivateKey)
	if err != nil {
		return nil, err
	}

	return a.addContract(account, contractName, contractFilename, updateExisting)
}

func (a *Accounts) addContract(
	account *lib.Account,
	contractName string,
	contractFilename string,
	updateExisting bool,
) (*flow.Account, error) {
	contractSource, err := util.LoadFile(contractFilename)
	if err != nil {
		return nil, err
	}

	tx := templates.AddAccountContract(
		account.Address(),
		templates.Contract{
			Name:   contractName,
			Source: string(contractSource),
		},
	)

	// if we are updating contract
	if updateExisting {
		tx = templates.UpdateAccountContract(
			account.Address(),
			templates.Contract{
				Name:   contractName,
				Source: string(contractSource),
			},
		)
	}

	// send transaction with contract
	tx, err = a.gateway.SendTransaction(tx, account)
	if err != nil {
		return nil, err
	}

	// we wait for transaction to be sealed
	_, err = a.gateway.GetTransactionResult(tx)
	if err != nil {
		return nil, err
	}

	return a.gateway.GetAccount(account.Address())
}

// RemoveContracts removes a contract from the account
func (a *Accounts) RemoveContract(
	contractName string,
	accountName string,
) (*flow.Account, error) {
	account := a.project.GetAccountByName(accountName)
	if account == nil {
		return nil, fmt.Errorf("account: [%s] doesn't exists in configuration", accountName)
	}

	return a.removeContract(contractName, account)
}

// RemoveContractForAddress removes contract from address using private key
func (a *Accounts) RemoveContractForAddress(
	contractName string,
	accountAddress string,
	accountPrivateKey string,
) (*flow.Account, error) {
	account, err := util.AccountFromAddressAndKey(accountAddress, accountPrivateKey)
	if err != nil {
		return nil, err
	}

	return a.removeContract(contractName, account)
}

func (a *Accounts) removeContract(
	contractName string,
	account *lib.Account,
) (*flow.Account, error) {
	tx := templates.RemoveAccountContract(account.Address(), contractName)
	tx, err := a.gateway.SendTransaction(tx, account)
	if err != nil {
		return nil, err
	}

	return a.gateway.GetAccount(account.Address())
}
