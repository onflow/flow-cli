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
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	goeth "github.com/ethereum/go-ethereum/accounts"
	"github.com/lmars/go-slip10"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/access/grpc"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/pkg/errors"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"github.com/onflow/flow-cli/flowkit/accounts"
	"github.com/onflow/flow-cli/flowkit/config"
	"github.com/onflow/flow-cli/flowkit/gateway"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/flowkit/project"
	"github.com/onflow/flow-cli/flowkit/transactions"
)

// BlockQuery defines possible queries for block.
type BlockQuery struct {
	ID     *flow.Identifier
	Height uint64
	Latest bool
}

// LatestBlockQuery specifies the latest block.
var LatestBlockQuery = BlockQuery{Latest: true}

// NewBlockQuery creates block query based on the passed query value.
//
// Query string options:
// - "latest"                : return the latest block
// - height (e.g. 123456789) : return block at this height
// - ID                      : return block with this ID
// if none of the valid values are passed an error is returned.
func NewBlockQuery(query string) (BlockQuery, error) {
	if query == "latest" {
		return LatestBlockQuery, nil
	}
	if height, ce := strconv.ParseUint(query, 10, 64); ce == nil {
		return BlockQuery{Height: height}, nil
	}
	if id := flow.HexToID(query); id != flow.EmptyID {
		return BlockQuery{ID: &id}, nil
	}

	return BlockQuery{}, fmt.Errorf("invalid query: %s, valid are: \"latest\", block height or block ID", query)
}

// ScriptQuery defines block ID or height at which we should execute the script.
type ScriptQuery struct {
	Latest bool
	ID     flow.Identifier
	Height uint64
}

// LatestScriptQuery specifies the latest block at which query is executed.
var LatestScriptQuery = ScriptQuery{Latest: true}

// EventWorker defines how many workers do we want to start, each in its own subroutine, and how many blocks
// each worker fetches from the network. This is used to process the event requests concurrently.
type EventWorker struct {
	Count           int
	BlocksPerWorker uint64
}

var _ Services = &Flowkit{}

func NewFlowkit(
	state *State,
	network config.Network,
	gateway gateway.Gateway,
	logger output.Logger,
) *Flowkit {
	return &Flowkit{state, network, gateway, logger}
}

type Flowkit struct {
	state   *State
	network config.Network
	gateway gateway.Gateway
	logger  output.Logger
}

func (f *Flowkit) Network() config.Network {
	return f.network
}

func (f *Flowkit) Gateway() gateway.Gateway {
	return f.gateway
}

func (f *Flowkit) SetLogger(logger output.Logger) {
	f.logger = logger
}

func (f *Flowkit) State() (*State, error) {
	if f.state == nil {
		return nil, config.ErrDoesNotExist
	}
	return f.state, nil
}

func (f *Flowkit) Ping() error {
	return f.gateway.Ping()
}

// GetAccount fetches account on the Flow network.
func (f *Flowkit) GetAccount(_ context.Context, address flow.Address) (*flow.Account, error) {
	return f.gateway.GetAccount(address)
}

// CreateAccount on the Flow network with the provided keys and using the signer for creation transaction.
// Returns the newly created account as well as the ID of the transaction that created the account.
//
// Keys is a slice but only one can be passed as well. If the transaction fails or there are other issues an error is returned.
func (f *Flowkit) CreateAccount(
	_ context.Context,
	signer *accounts.Account,
	keys []accounts.PublicKey,
) (*flow.Account, flow.Identifier, error) {
	var accKeys []*flow.AccountKey
	for _, k := range keys {
		if k.Weight == 0 { // if key weight is not specified
			k.Weight = flow.AccountKeyWeightThreshold
		}

		accKey := &flow.AccountKey{
			PublicKey: k.Public,
			SigAlgo:   k.SigAlgo,
			HashAlgo:  k.HashAlgo,
			Weight:    k.Weight,
		}

		err := accKey.Validate()
		if err != nil {
			return nil, flow.EmptyID, fmt.Errorf("invalid account key: %w", err)
		}

		accKeys = append(accKeys, accKey)
	}

	tx, err := transactions.NewCreateAccount(signer, accKeys, nil)
	if err != nil {
		return nil, flow.EmptyID, err
	}

	tx, err = f.prepareTransaction(tx, signer)
	if err != nil {
		return nil, flow.EmptyID, err
	}

	f.logger.Info(fmt.Sprintf("Transaction ID: %s", tx.FlowTransaction().ID()))
	f.logger.StartProgress("Creating account...")
	defer f.logger.StopProgress()

	sentTx, err := f.gateway.SendSignedTransaction(tx.FlowTransaction())
	if err != nil {
		return nil, flow.EmptyID, errors.Wrap(err, "account creation transaction failed")
	}

	f.logger.StartProgress("Waiting for transaction to be sealed...")
	defer f.logger.StopProgress()

	result, err := f.gateway.GetTransactionResult(sentTx.ID(), true)
	if err != nil {
		return nil, flow.EmptyID, err
	}

	if result.Error != nil {
		return nil, flow.EmptyID, result.Error
	}

	events := EventsFromTransaction(result)
	newAccountAddress := events.GetCreatedAddresses()
	if len(newAccountAddress) == 0 {
		return nil, flow.EmptyID, fmt.Errorf("new account address couldn't be fetched")
	}

	account, err := f.gateway.GetAccount(*newAccountAddress[0]) // we know it's the only and first event
	if err != nil {
		return nil, flow.EmptyID, err
	}

	return account, sentTx.ID(), nil
}

// prepareTransaction prepares transaction for sending with data from network
func (f *Flowkit) prepareTransaction(
	tx *transactions.Transaction,
	account *accounts.Account,
) (*transactions.Transaction, error) {
	block, err := f.gateway.GetLatestBlock()
	if err != nil {
		return nil, err
	}

	proposer, err := f.gateway.GetAccount(account.Address)
	if err != nil {
		return nil, err
	}

	tx.SetBlockReference(block)
	if err = tx.SetProposer(proposer, account.Key.Index()); err != nil {
		return nil, err
	}

	tx, err = tx.Sign()
	if err != nil {
		return nil, err
	}

	return tx, nil
}

var errUpdateNoDiff = errors.New("contract already exists and is the same as the contract provided for update")

type UpdateContract func(existing []byte, new []byte) bool

func UpdateExistingContract(updateExisting bool) UpdateContract {
	return func(existing []byte, new []byte) bool {
		return updateExisting
	}
}

// AddContract to the Flow account provided and return the transaction ID.
//
// If the contract already exists on the account the operation will fail and error will be returned.
// Use UpdateExistingContract(bool) to define whether a contract should be updated or not, or you can also
// define a custom UpdateContract function which returns bool indicating whether a contract should be updated or not.
func (f *Flowkit) AddContract(
	ctx context.Context,
	account *accounts.Account,
	contract Script,
	update UpdateContract,
) (flow.Identifier, bool, error) {
	state, err := f.State()
	if err != nil {
		return flow.EmptyID, false, err
	}

	program, err := project.NewProgram(contract.Code, contract.Args, contract.Location)
	if err != nil {
		return flow.EmptyID, false, err
	}

	if program.HasImports() {
		contracts, err := state.DeploymentContractsByNetwork(f.network)
		if err != nil {
			return flow.EmptyID, false, err
		}

		importReplacer := project.NewImportReplacer(
			contracts,
			state.AliasesForNetwork(f.network),
		)

		program, err = importReplacer.Replace(program)
		if err != nil {
			return flow.EmptyID, false, err
		}
	}

	name, err := program.Name()
	if err != nil {
		return flow.EmptyID, false, err
	}

	tx, err := transactions.NewAddAccountContract(
		account,
		name,
		program.Code(),
		contract.Args,
	)
	if err != nil {
		return flow.EmptyID, false, err
	}

	f.logger.StartProgress(fmt.Sprintf("Checking contract '%s' on account '%s'...", name, account.Address))
	defer f.logger.StopProgress()

	// check if contract exists on account
	flowAccount, err := f.gateway.GetAccount(account.Address)
	if err != nil {
		return flow.EmptyID, false, err
	}
	existingContract, exists := flowAccount.Contracts[name]
	noDiffInContract := bytes.Equal(program.Code(), existingContract)

	if exists && noDiffInContract {
		return flow.EmptyID, false, errUpdateNoDiff
	}

	updateExisting := update(existingContract, program.Code())
	if exists && !updateExisting {
		return flow.EmptyID, false, fmt.Errorf(fmt.Sprintf("contract %s exists in account %s", name, account.Name))
	}

	if exists && updateExisting {
		tx, err = transactions.NewUpdateAccountContract(account, name, program.Code())
		if err != nil {
			return flow.EmptyID, false, err
		}
	}

	tx, err = f.prepareTransaction(tx, account)
	if err != nil {
		return flow.EmptyID, false, err
	}

	// send transaction with contract
	sentTx, err := f.gateway.SendSignedTransaction(tx.FlowTransaction())
	if err != nil {
		return tx.FlowTransaction().ID(), false, fmt.Errorf("failed to send transaction to deploy a contract: %w", err)
	}

	if exists {
		f.logger.StartProgress(fmt.Sprintf("Contract '%s' updating on the account '%s'.", name, account.Address))
	} else {
		f.logger.StartProgress(fmt.Sprintf("Contract '%s' deploying on the account '%s'.", name, account.Address))
	}

	// we wait for transaction to be sealed
	trx, err := f.gateway.GetTransactionResult(sentTx.ID(), true)
	if err != nil {
		return tx.FlowTransaction().ID(), false, err
	}
	if trx.Error != nil {
		return tx.FlowTransaction().ID(), false, trx.Error
	}

	d := state.Deployments().ByAccountAndNetwork(account.Name, f.network.Name)
	cd := config.ContractDeployment{
		Name: name,
		Args: contract.Args,
	}
	if d != nil {
		d.AddContract(cd)
	} else {
		deployment := config.Deployment{
			Network:   f.network.Name,
			Account:   account.Name,
			Contracts: []config.ContractDeployment{cd},
		}
		state.Deployments().AddOrUpdate(deployment)
	}

	// don't add contract if it already exists because it might overwrite existing data
	if c, _ := state.Contracts().ByName(name); c == nil {
		state.Contracts().AddOrUpdate(config.Contract{
			Name:     name,
			Location: contract.Location,
		})
	}

	return sentTx.ID(), updateExisting, err
}

// RemoveContract from the provided account by its name.
//
// If removal is successful transaction ID is returned.
func (f *Flowkit) RemoveContract(
	_ context.Context,
	account *accounts.Account,
	contractName string,
) (flow.Identifier, error) {
	// check if contracts exists on the account
	flowAcc, err := f.gateway.GetAccount(account.Address)
	if err != nil {
		return flow.EmptyID, err
	}

	existingContracts := maps.Keys(flowAcc.Contracts)
	if !slices.Contains(existingContracts, contractName) {
		return flow.EmptyID, fmt.Errorf(
			"can not remove a non-existing contract named '%s'. Account only contains the contracts: %v",
			contractName,
			strings.Join(existingContracts, ", "),
		)
	}

	tx, err := transactions.NewRemoveAccountContract(account, contractName)
	if err != nil {
		return flow.EmptyID, err
	}

	tx, err = f.prepareTransaction(tx, account)
	if err != nil {
		return flow.EmptyID, err
	}

	f.logger.StartProgress(
		fmt.Sprintf("Removing Contract %s from %s...", contractName, account.Address),
	)
	defer f.logger.StopProgress()

	sentTx, err := f.gateway.SendSignedTransaction(tx.FlowTransaction())
	if err != nil {
		return flow.EmptyID, err
	}

	txr, err := f.gateway.GetTransactionResult(sentTx.ID(), true)
	if err != nil {
		return flow.EmptyID, err
	}
	if txr != nil && txr.Error != nil {
		return flow.EmptyID, txr.Error
	}

	f.logger.StopProgress()

	return sentTx.ID(), nil
}

// GetBlock by the query from Flow blockchain. Query can define a block by ID, block by height or require the latest block.
func (f *Flowkit) GetBlock(_ context.Context, query BlockQuery) (*flow.Block, error) {
	var err error
	var block *flow.Block
	if query.Latest {
		block, err = f.gateway.GetLatestBlock()
	} else if query.ID != nil {
		block, err = f.gateway.GetBlockByID(*query.ID)
	} else {
		block, err = f.gateway.GetBlockByHeight(query.Height)
	}

	if err != nil {
		return nil, fmt.Errorf("error fetching block: %s", err.Error())
	}

	if block == nil {
		return nil, fmt.Errorf("block not found")
	}

	return block, err
}

// GetCollection by the ID from Flow network.
func (f *Flowkit) GetCollection(_ context.Context, ID flow.Identifier) (*flow.Collection, error) {
	return f.gateway.GetCollection(ID)
}

// GetEvents from Flow network by their event name in the specified height interval defined by start and end inclusive.
// Optional worker defines parameters for how many concurrent workers do we want to fetch our events,
// and how many blocks between the provided interval each worker fetches.
//
// Providing worker value will produce faster response as the interval will be scanned concurrently. This parameter is optional,
// if not provided only a single worker will be used.
func (f *Flowkit) GetEvents(
	_ context.Context,
	names []string,
	startHeight uint64,
	endHeight uint64,
	worker *EventWorker,
) ([]flow.BlockEvents, error) {
	if endHeight < startHeight {
		return nil, fmt.Errorf("cannot have end height (%d) of block range less that start height (%d)", endHeight, startHeight)
	}

	if worker == nil { // if no worker is passed, create a default one
		worker = &EventWorker{
			Count:           1,
			BlocksPerWorker: 250,
		}
	}

	queries := makeEventQueries(names, startHeight, endHeight, worker.BlocksPerWorker)

	jobChan := make(chan grpc.EventRangeQuery, worker.Count)
	results := make(chan eventWorkerResult)

	var wg sync.WaitGroup

	for i := 0; i < worker.Count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f.eventWorker(jobChan, results)
		}()
	}

	// wait on the workers to finish and close the result channel
	// to signal downstream that all work is done
	go func() {
		defer close(results)
		wg.Wait()
	}()

	go func() {
		defer close(jobChan)
		for _, query := range queries {
			jobChan <- query
		}
	}()

	var resultEvents []flow.BlockEvents
	for eventResult := range results {
		if eventResult.err != nil {
			return nil, eventResult.err
		}

		resultEvents = append(resultEvents, eventResult.events...)
	}

	return resultEvents, nil
}

func (f *Flowkit) eventWorker(jobChan <-chan grpc.EventRangeQuery, results chan<- eventWorkerResult) {
	for q := range jobChan {
		blockEvents, err := f.gateway.GetEvents(q.Type, q.StartHeight, q.EndHeight)
		if err != nil {
			results <- eventWorkerResult{nil, err}
		}
		results <- eventWorkerResult{blockEvents, nil}
	}
}

type eventWorkerResult struct {
	events []flow.BlockEvents
	err    error
}

func makeEventQueries(
	events []string,
	startHeight uint64,
	endHeight uint64,
	blockCount uint64,
) []grpc.EventRangeQuery {
	var queries []grpc.EventRangeQuery
	for startHeight <= endHeight {
		suggestedEndHeight := startHeight + blockCount - 1 // since we are inclusive
		end := endHeight
		if suggestedEndHeight < endHeight {
			end = suggestedEndHeight
		}
		for _, event := range events {
			queries = append(queries, grpc.EventRangeQuery{
				Type:        event,
				StartHeight: startHeight,
				EndHeight:   end,
			})
		}
		startHeight = suggestedEndHeight + 1
	}
	return queries
}

// GenerateKey using the signature algorithm and optional seed. If seed is not provided a random safe seed will be generated.
func (f *Flowkit) GenerateKey(
	_ context.Context,
	sigAlgo crypto.SignatureAlgorithm,
	inputSeed string,
) (crypto.PrivateKey, error) {
	seed := []byte(inputSeed)

	if inputSeed == "" {
		seed = make([]byte, crypto.MinSeedLength)
		_, err := rand.Read(seed)
		if err != nil {
			return nil, fmt.Errorf("failed to generate random seed: %v", err)
		}
	}

	privateKey, err := crypto.GeneratePrivateKey(sigAlgo, seed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	return privateKey, nil
}

// GenerateMnemonicKey will generate a new key with the signature algorithm and optional derivation path.
//
// If the derivation path is not provided a default "m/44'/539'/0'/0/0" will be used.
func (f *Flowkit) GenerateMnemonicKey(
	_ context.Context,
	sigAlgo crypto.SignatureAlgorithm,
	derivationPath string,
) (crypto.PrivateKey, string, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return nil, "", err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, "", err
	}

	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, "", fmt.Errorf("invalid mnemonic")
	}

	seed := bip39.NewSeed(mnemonic, "")

	key, err := f.derivePrivateKeyFromSeed(seed, sigAlgo, derivationPath)
	if err != nil {
		return nil, "", err
	}

	return key, mnemonic, nil
}

func (f *Flowkit) DerivePrivateKeyFromMnemonic(
	_ context.Context,
	mnemonic string,
	sigAlgo crypto.SignatureAlgorithm,
	derivationPath string,
) (crypto.PrivateKey, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic")
	}

	seed := bip39.NewSeed(mnemonic, "")

	return f.derivePrivateKeyFromSeed(seed, sigAlgo, derivationPath)
}

func (f *Flowkit) derivePrivateKeyFromSeed(
	seed []byte,
	sigAlgo crypto.SignatureAlgorithm,
	derivationPath string,
) (crypto.PrivateKey, error) {
	// sanity check of seed length
	if len(seed) < 16 {
		return nil, fmt.Errorf("seed length should be at least 16 bytes, got %d", len(seed))
	}

	if derivationPath == "" {
		derivationPath = "m/44'/539'/0'/0/0"
	}

	path, err := goeth.ParseDerivationPath(derivationPath)
	if err != nil {
		return nil, fmt.Errorf("invalid derivation path")
	}

	curve := slip10.CurveBitcoin // case ECDSA_secp256k1
	if sigAlgo == crypto.ECDSA_P256 {
		curve = slip10.CurveP256
	} else if sigAlgo != crypto.ECDSA_secp256k1 {
		return nil, fmt.Errorf("invalid signature algorithm")
	}

	accountKey, err := slip10.NewMasterKeyWithCurve(seed, curve)
	if err != nil {
		return nil, err
	}

	for _, n := range path {
		accountKey, err = accountKey.NewChildKey(n)

		if err != nil {
			return nil, err
		}
	}
	privateKey, err := crypto.DecodePrivateKey(sigAlgo, accountKey.Key)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

// DeployProject contracts to the Flow network or update if already exists and UpdateContracts returns true.
//
// Retrieve all the contracts for specified network, sort them for deployment deploy one by one and replace
// the imports in the contract source, so it corresponds to the account name the contract was deployed to.
// If contracts already exist use UpdateExistingContract(bool) to define whether a contract should be updated or not.
func (f *Flowkit) DeployProject(ctx context.Context, update UpdateContract) ([]*project.Contract, error) {
	state, err := f.State()
	if err != nil {
		return nil, err
	}

	contracts, err := state.DeploymentContractsByNetwork(f.network)
	if err != nil {
		return nil, err
	}

	deployment, err := project.NewDeployment(contracts, state.AliasesForNetwork(f.network))
	if err != nil {
		return nil, err
	}

	sorted, err := deployment.Sort()
	if err != nil {
		return nil, err
	}

	f.logger.Info(fmt.Sprintf(
		"\nDeploying %d contracts for accounts: %s\n",
		len(sorted),
		state.AccountsForNetwork(f.network).String(),
	))
	defer f.logger.StopProgress()

	deployErr := &ProjectDeploymentError{}
	for _, contract := range sorted {
		targetAccount, err := state.Accounts().ByName(contract.AccountName)
		if err != nil {
			return nil, fmt.Errorf("target account for deploying contract not found in configuration")
		}

		txID, updated, err := f.AddContract(
			ctx,
			targetAccount,
			Script{Code: contract.Code(), Args: contract.Args, Location: contract.Location()},
			update,
		)
		if err != nil && errors.Is(err, errUpdateNoDiff) {
			f.logger.Info(fmt.Sprintf(
				"%s -> 0x%s [skipping, no changes found]",
				output.Italic(contract.Name),
				contract.AccountAddress.String(),
			))
			continue
		} else if err != nil {
			deployErr.add(contract, err, fmt.Sprintf("failed to deploy contract %s", contract.Name))
			continue
		}

		f.logger.Info(fmt.Sprintf(
			"%s -> 0x%s (%s) %s",
			output.Green(contract.Name),
			contract.AccountAddress,
			txID.String(),
			map[bool]string{true: "[updated]", false: ""}[updated],
		))
	}

	if len(deployErr.contracts) > 0 {
		return nil, deployErr
	}

	f.logger.Info(fmt.Sprintf("\n%s All contracts deployed successfully", output.SuccessEmoji()))
	return sorted, nil
}

type ProjectDeploymentError struct {
	contracts map[string]error
}

func (d *ProjectDeploymentError) add(contract *project.Contract, err error, msg string) {
	if d.contracts == nil {
		d.contracts = make(map[string]error)
	}
	d.contracts[contract.Name] = fmt.Errorf("%s: %w", msg, err)
}

func (d *ProjectDeploymentError) Contracts() map[string]error {
	return d.contracts
}

func (d *ProjectDeploymentError) Error() string {
	err := ""
	for c, e := range d.contracts {
		err = fmt.Sprintf("%s %s: %s,", err, c, e.Error())
	}
	return err
}

// Script includes Cadence code and optional arguments and filename.
//
// Filename is only required to be passed if the code has imports  you want to resolve.
type Script struct {
	Code     []byte
	Args     []cadence.Value
	Location string
}

// ExecuteScript on the Flow network and return the Cadence value as a result. The script is executed at the
// block provided as part of the ScriptQuery value.
func (f *Flowkit) ExecuteScript(_ context.Context, script Script, query ScriptQuery) (cadence.Value, error) {
	state, err := f.State()
	if err != nil {
		return nil, err
	}

	program, err := project.NewProgram(script.Code, script.Args, script.Location)
	if err != nil {
		return nil, err
	}

	if program.HasImports() {
		contracts, err := state.DeploymentContractsByNetwork(f.network)
		if err != nil {
			return nil, err
		}

		importReplacer := project.NewImportReplacer(
			contracts,
			state.AliasesForNetwork(f.network),
		)

		if state == nil {
			return nil, config.ErrDoesNotExist
		}
		if f.network == config.EmptyNetwork {
			return nil, fmt.Errorf("missing network, specify which network to use to resolve imports in script code")
		}
		if script.Location == "" {
			return nil, fmt.Errorf("resolving imports in scripts not supported")
		}

		program, err = importReplacer.Replace(program)
		if err != nil {
			return nil, err
		}
	}

	if query.Latest {
		return f.gateway.ExecuteScript(program.Code(), script.Args)
	} else if query.ID != flow.EmptyID {
		return f.gateway.ExecuteScriptAtID(program.Code(), script.Args, query.ID)
	} else {
		return f.gateway.ExecuteScriptAtHeight(program.Code(), script.Args, query.Height)
	}
}

// GetTransactionByID from the Flow network including the transaction result. Using the waitSeal we can wait for the transaction to be sealed.
func (f *Flowkit) GetTransactionByID(
	_ context.Context,
	ID flow.Identifier,
	waitSeal bool,
) (*flow.Transaction, *flow.TransactionResult, error) {
	f.logger.StartProgress("Fetching Transaction...")
	defer f.logger.StopProgress()

	tx, err := f.gateway.GetTransaction(ID)
	if err != nil {
		return nil, nil, err
	}

	if waitSeal {
		f.logger.StartProgress("Waiting for transaction to be sealed...")
		defer f.logger.StopProgress()
	}

	result, err := f.gateway.GetTransactionResult(ID, waitSeal)
	return tx, result, err
}

func (f *Flowkit) GetTransactionsByBlockID(
	_ context.Context,
	blockID flow.Identifier,
) ([]*flow.Transaction, []*flow.TransactionResult, error) {
	tx, err := f.gateway.GetTransactionsByBlockID(blockID)
	if err != nil {
		return nil, nil, err
	}

	txRes, err := f.gateway.GetTransactionResultsByBlockID(blockID)
	if err != nil {
		return nil, nil, err
	}
	return tx, txRes, nil
}

// BuildTransaction builds a new transaction type for later signing and submitting to the network.
//
// AddressesRoles type defines the address for each role (payer, proposer, authorizers) and the script defines the transaction content.
func (f *Flowkit) BuildTransaction(
	_ context.Context,
	addresses transactions.AddressesRoles,
	proposerKeyIndex int,
	script Script,
	gasLimit uint64,
) (*transactions.Transaction, error) {
	state, err := f.State()
	if err != nil {
		return nil, err
	}

	latestBlock, err := f.gateway.GetLatestBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest sealed block: %w", err)
	}

	proposerAccount, err := f.gateway.GetAccount(addresses.Proposer)
	if err != nil {
		return nil, err
	}

	tx := transactions.New().
		SetPayer(addresses.Payer).
		SetComputeLimit(gasLimit).
		SetBlockReference(latestBlock)

	program, err := project.NewProgram(script.Code, script.Args, script.Location)
	if err != nil {
		return nil, err
	}

	if program.HasImports() {
		if f.network == config.EmptyNetwork {
			return nil, fmt.Errorf("missing network, specify which network to use to resolve imports in transaction code")
		}
		if script.Location == "" { // when used as lib with code we don't support imports
			return nil, fmt.Errorf("resolving imports in transactions not supported")
		}

		contracts, err := state.DeploymentContractsByNetwork(f.network)
		if err != nil {
			return nil, err
		}

		importReplacer := project.NewImportReplacer(
			contracts,
			state.AliasesForNetwork(f.network),
		)

		program, err = importReplacer.Replace(program)
		if err != nil {
			return nil, fmt.Errorf("error resolving imports: %w", err)
		}
	}

	if err := tx.SetProposer(proposerAccount, proposerKeyIndex); err != nil {
		return nil, err
	}

	if err := tx.SetScriptWithArgs(program.Code(), script.Args); err != nil {
		return nil, err
	}

	tx, err = tx.AddAuthorizers(addresses.Authorizers)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// SignTransactionPayload will use the signer account provided and the payload raw byte content to sign it.
//
// The payload should be RLP encoded transaction payload and is suggested to be used in pair with BuildTransaction function.
func (f *Flowkit) SignTransactionPayload(
	_ context.Context,
	signer *accounts.Account,
	payload []byte,
) (*transactions.Transaction, error) {
	tx, err := transactions.NewFromPayload(payload)
	if err != nil {
		return nil, err
	}

	err = tx.SetSigner(signer)
	if err != nil {
		return nil, err
	}

	return tx.Sign()
}

// SendSignedTransaction will send a prebuilt and signed transaction to the Flow network.
//
// You can build the transaction using the BuildTransaction method and then sign it using the SignTranscation method.
func (f *Flowkit) SendSignedTransaction(
	_ context.Context,
	tx *transactions.Transaction,
) (*flow.Transaction, *flow.TransactionResult, error) {
	sentTx, err := f.gateway.SendSignedTransaction(tx.FlowTransaction())
	if err != nil {
		return nil, nil, err
	}

	res, err := f.gateway.GetTransactionResult(sentTx.ID(), true)
	if err != nil {
		return nil, nil, err
	}

	return sentTx, res, nil
}

// SendTransaction will build and send a transaction to the Flow network, using the accounts provided for each role and
// contain the script. Transaction as well as transaction result will be returned in case the transaction is successfully submitted.
func (f *Flowkit) SendTransaction(
	ctx context.Context,
	accounts transactions.AccountRoles,
	script Script,
	gasLimit uint64,
) (*flow.Transaction, *flow.TransactionResult, error) {
	tx, err := f.BuildTransaction(
		ctx,
		accounts.AddressRoles(),
		accounts.Proposer.Key.Index(),
		script,
		gasLimit,
	)
	if err != nil {
		return nil, nil, err
	}

	for _, signer := range accounts.Signers() {
		err = tx.SetSigner(signer)
		if err != nil {
			return nil, nil, err
		}

		tx, err = tx.Sign()
		if err != nil {
			return nil, nil, err
		}
	}

	f.logger.Info(fmt.Sprintf("Transaction ID: %s", tx.FlowTransaction().ID()))
	f.logger.StartProgress("Sending transaction...")
	defer f.logger.StopProgress()

	sentTx, err := f.gateway.SendSignedTransaction(tx.FlowTransaction())
	if err != nil {
		return nil, nil, err
	}

	f.logger.StopProgress()
	f.logger.StartProgress("Waiting for transaction to be sealed...")
	defer f.logger.StopProgress()

	res, err := f.gateway.GetTransactionResult(sentTx.ID(), true)

	return sentTx, res, err
}

// this is added to resolve the issue with chainhash ambiguous import,
// the code is not used, but it's needed to force go.mod specify and retain chainhash version
// workaround for issue: https://github.com/golang/go/issues/27899
var _ = chainhash.Hash{}
