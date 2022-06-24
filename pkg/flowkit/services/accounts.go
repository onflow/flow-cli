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
	"fmt"
	"strings"

	"github.com/onflow/flow-go-sdk/templates"

	"github.com/onflow/flow-cli/pkg/flowkit/config"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/util"

	"github.com/onflow/cadence"
	tmpl "github.com/onflow/flow-core-contracts/lib/go/templates"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

// Accounts is a service that handles all account-related interactions.
type Accounts struct {
	gateway gateway.Gateway
	state   *flowkit.State
	logger  output.Logger
}

// NewAccounts returns a new accounts service.
func NewAccounts(
	gateway gateway.Gateway,
	state *flowkit.State,
	logger output.Logger,
) *Accounts {
	return &Accounts{
		gateway: gateway,
		state:   state,
		logger:  logger,
	}
}

// Get returns an account by on address.
func (a *Accounts) Get(address flow.Address) (*flow.Account, error) {
	a.logger.StartProgress(fmt.Sprintf("Loading %s...", address))

	account, err := a.gateway.GetAccount(address)
	a.logger.StopProgress()

	return account, err
}

// StakingInfo returns the staking and delegation information for an account.
func (a *Accounts) StakingInfo(address flow.Address) ([]map[string]interface{}, []map[string]interface{}, error) {
	a.logger.StartProgress(fmt.Sprintf("Fetching info for %s...", address.String()))
	defer a.logger.StopProgress()

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

	stakingInfoScript := tmpl.GenerateCollectionGetAllNodeInfoScript(env)
	delegationInfoScript := tmpl.GenerateCollectionGetAllDelegatorInfoScript(env)

	stakingValue, err := a.gateway.ExecuteScript(stakingInfoScript, cadenceAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting staking info: %s", err.Error())
	}

	delegationValue, err := a.gateway.ExecuteScript(delegationInfoScript, cadenceAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting delegation info: %s", err.Error())
	}

	// get staking infos and delegation infos
	stakingInfos, err := flowkit.NewStakingInfoFromValue(stakingValue)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing staking info: %s", err.Error())
	}
	delegationInfos, err := flowkit.NewStakingInfoFromValue(delegationValue)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing delegation info: %s", err.Error())
	}

	// get a set of node ids from all staking infos
	nodeStakes := make(map[string]cadence.Value)
	for _, stakingInfo := range stakingInfos {
		nodeID, ok := stakingInfo["id"]
		if ok {
			nodeStakes[nodeIDToString(nodeID)] = nil
		}
	}
	totalCommitmentScript := tmpl.GenerateGetTotalCommitmentBalanceScript(env)

	// foreach node id, get the node total stake
	for nodeID := range nodeStakes {
		stake, err := a.gateway.ExecuteScript(totalCommitmentScript, []cadence.Value{cadence.String(nodeID)})
		if err != nil {
			return nil, nil, fmt.Errorf("error getting total stake for node: %s", err.Error())
		}

		nodeStakes[nodeID] = stake
	}

	// foreach staking info, add the node total stake
	for _, stakingInfo := range stakingInfos {
		nodeID, ok := stakingInfo["id"]
		if ok {
			stakingInfo["nodeTotalStake"] = nodeStakes[nodeIDToString(nodeID)].(cadence.UFix64)
		}
	}

	a.logger.StopProgress()

	return stakingInfos, delegationInfos, nil
}

func nodeIDToString(value interface{}) string {
	return value.(cadence.String).ToGoValue().(string)
}

// NodeTotalStake returns the total stake including delegations of a node.
func (a *Accounts) NodeTotalStake(nodeId string, chain flow.ChainID) (*cadence.Value, error) {
	a.logger.StartProgress(fmt.Sprintf("Fetching total stake for node id %s...", nodeId))
	defer a.logger.StopProgress()

	if chain == flow.Emulator {
		return nil, fmt.Errorf("emulator chain not supported")
	}

	env := util.EnvFromNetwork(chain)

	stakingInfoScript := tmpl.GenerateGetTotalCommitmentBalanceScript(env)
	stakingValue, err := a.gateway.ExecuteScript(stakingInfoScript, []cadence.Value{cadence.String(nodeId)})
	if err != nil {
		return nil, fmt.Errorf("error getting total stake for node: %s", err.Error())
	}

	a.logger.StopProgress()

	return &stakingValue, nil
}

// Create creates and returns a new account.
//
// The new account is created with the given public keys and contracts.
//
// The account creation transaction is signed by the specified signer.
func (a *Accounts) Create(
	signer *flowkit.Account,
	pubKeys []crypto.PublicKey,
	keyWeights []int,
	sigAlgo []crypto.SignatureAlgorithm,
	hashAlgo []crypto.HashAlgorithm,
	contractArgs []string,
) (*flow.Account, error) {
	if a.state == nil {
		return nil, config.ErrDoesNotExist
	}

	// if more than one key is provided and at least one weight is specified, make sure there isn't a mismatch
	if len(keyWeights) > 0 && len(pubKeys) != len(keyWeights) {
		return nil, fmt.Errorf(
			"number of keys and weights provided must match, number of provided keys: %d, number of provided key weights: %d",
			len(pubKeys),
			len(keyWeights),
		)
	}

	var accKeys []*flow.AccountKey
	for i, pubKey := range pubKeys {
		weight := flow.AccountKeyWeightThreshold
		if len(keyWeights) > i { // if key weight is specified
			weight = keyWeights[i]
		}

		accKey := &flow.AccountKey{
			PublicKey: pubKey,
			SigAlgo:   sigAlgo[i],
			HashAlgo:  hashAlgo[i],
			Weight:    weight,
		}

		err := accKey.Validate()
		if err != nil {
			return nil, fmt.Errorf("invalid account key: %w", err)
		}

		accKeys = append(accKeys, accKey)
	}

	contracts := make([]templates.Contract, 0)
	for _, contract := range contractArgs {
		contractFlagContent := strings.SplitN(contract, ":", 2)
		if len(contractFlagContent) != 2 {
			return nil, fmt.Errorf("wrong format for contract. Correct format is name:path, but got: %s", contract)
		}

		contractSource, err := a.state.ReadFile(contractFlagContent[1])
		if err != nil {
			return nil, err
		}

		contracts = append(contracts, templates.Contract{
			Name:   contractFlagContent[0],
			Source: string(contractSource),
		})
	}

	tx, err := flowkit.NewCreateAccountTransaction(signer, accKeys, contracts)
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

	events := flowkit.EventsFromTransaction(result)
	newAccountAddress := events.GetAddress()

	if newAccountAddress == nil {
		return nil, fmt.Errorf("new account address couldn't be fetched")
	}

	a.logger.StopProgress()

	return a.gateway.GetAccount(*newAccountAddress)
}

// AddContract deploys a contract code to the account provided with possible update flag.
func (a *Accounts) AddContract(
	account *flowkit.Account,
	contractName string,
	contractSource []byte,
	updateExisting bool,
	contractArgs []cadence.Value,
) (*flow.Account, error) {
	tx, err := flowkit.NewAddAccountContractTransaction(
		account,
		contractName,
		string(contractSource),
		contractArgs, // TODO(sideninja) add support for args on account add-contract
	)
	if err != nil {
		return nil, err
	}

	// if we are updating contract
	if updateExisting {
		tx, err = flowkit.NewUpdateAccountContractTransaction(
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

// RemoveContract removes a contract from an account and returns the updated account.
func (a *Accounts) RemoveContract(
	account *flowkit.Account,
	contractName string,
) (*flow.Account, error) {
	tx, err := flowkit.NewRemoveAccountContractTransaction(account, contractName)
	if err != nil {
		return nil, err
	}

	tx, err = a.prepareTransaction(tx, account)
	if err != nil {
		return nil, err
	}

	a.logger.Info(fmt.Sprintf("Transaction ID: %s", tx.FlowTransaction().ID().String()))
	a.logger.StartProgress(
		fmt.Sprintf("Removing Contract %s from %s...", contractName, account.Address()),
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
	tx *flowkit.Transaction,
	account *flowkit.Account,
) (*flowkit.Transaction, error) {

	block, err := a.gateway.GetLatestBlock()
	if err != nil {
		return nil, err
	}

	proposer, err := a.gateway.GetAccount(account.Address())
	if err != nil {
		return nil, err
	}

	tx.SetBlockReference(block)

	if err = tx.SetProposer(proposer, account.Key().Index()); err != nil {
		return nil, err
	}

	tx, err = tx.Sign()
	if err != nil {
		return nil, err
	}

	return tx, nil
}
