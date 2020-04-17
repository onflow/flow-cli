// Package emulator provides an emulated version of the Flow blockchain that can be used
// for development purposes.
//
// This package can be used as a library or as a standalone application.
//
// When used as a library, this package provides tools to write programmatic tests for
// Flow applications.
//
// When used as a standalone application, this package implements the Flow Observation API
// and is fully-compatible with Flow gRPC client libraries.
package emulator

import (
	"errors"
	"fmt"
	"sync"

	"github.com/dapperlabs/cadence/runtime"
	"github.com/dapperlabs/flow-go-sdk"
	sdkcrypto "github.com/dapperlabs/flow-go-sdk/crypto"
	"github.com/dapperlabs/flow-go-sdk/keys"
	"github.com/dapperlabs/flow-go-sdk/templates"
	"github.com/dapperlabs/flow-go/crypto"
	model "github.com/dapperlabs/flow-go/model/flow"

	"github.com/dapperlabs/flow-emulator/execution"
	"github.com/dapperlabs/flow-emulator/storage"
	"github.com/dapperlabs/flow-emulator/storage/memstore"
	"github.com/dapperlabs/flow-emulator/types"
)

// Blockchain emulates the functionality of the Flow blockchain.
type Blockchain struct {
	// committed chain state: blocks, transactions, registers, events
	storage storage.Store

	// mutex protecting pending block
	mu sync.RWMutex

	// pending block containing block info, register state, pending transactions
	pendingBlock *pendingBlock

	// runtime context used to execute transactions and scripts
	computer *computer

	rootAccountAddress    flow.Address
	rootAccountPrivateKey flow.AccountPrivateKey
	lastCreatedAddress    flow.Address
}

// BlockchainAPI defines the method set of an emulated blockchain.
type BlockchainAPI interface {
	AddTransaction(tx flow.Transaction) error
	ExecuteNextTransaction() (*TransactionResult, error)
	ExecuteBlock() ([]*TransactionResult, error)
	CommitBlock() (*types.Block, error)
	ExecuteAndCommitBlock() (*types.Block, []*TransactionResult, error)
	GetLatestBlock() (*types.Block, error)
	GetBlockByID(id flow.Identifier) (*types.Block, error)
	GetBlockByHeight(height uint64) (*types.Block, error)
	GetCollection(colID flow.Identifier) (*model.LightCollection, error)
	GetTransaction(txID flow.Identifier) (*flow.Transaction, error)
	GetTransactionResult(txID flow.Identifier) (*flow.TransactionResult, error)
	GetAccount(address flow.Address) (*flow.Account, error)
	GetAccountAtBlock(address flow.Address, blockHeight uint64) (*flow.Account, error)
	GetEventsByHeight(blockHeight uint64, eventType string) ([]flow.Event, error)
	ExecuteScript(script []byte) (*ScriptResult, error)
	ExecuteScriptAtBlock(script []byte, blockHeight uint64) (*ScriptResult, error)
	RootAccountAddress() flow.Address
	RootKey() RootKey
}

var _ BlockchainAPI = &Blockchain{}

type RootKey struct {
	ID             int
	Address        flow.Address
	SequenceNumber uint64
	PrivateKey     flow.AccountPrivateKey
}

func (r RootKey) Signer() sdkcrypto.Signer {
	return r.PrivateKey.Signer()
}

func (r RootKey) AccountKey() flow.AccountKey {
	accountKey := r.PrivateKey.ToAccountKey()
	accountKey.Weight = keys.PublicKeyWeightThreshold
	return accountKey
}

// config is a set of configuration options for an emulated blockchain.
type config struct {
	RootAccountKey flow.AccountPrivateKey
	Store          storage.Store
}

// defaultConfig is the default configuration for an emulated blockchain.
// NOTE: Instantiated in init function
var defaultConfig config

// Option is a function applying a change to the emulator config.
type Option func(*config)

// WithRootKey sets the root key.
func WithRootAccountKey(rootKey flow.AccountPrivateKey) Option {
	return func(c *config) {
		c.RootAccountKey = rootKey
	}
}

// WithStore sets the persistent storage provider.
func WithStore(store storage.Store) Option {
	return func(c *config) {
		c.Store = store
	}
}

// NewBlockchain instantiates a new emulated blockchain with the provided options.
func NewBlockchain(opts ...Option) (*Blockchain, error) {
	var pendingBlock *pendingBlock
	var rootAccount *flow.Account

	// apply options to the default config
	config := defaultConfig
	for _, opt := range opts {
		opt(&config)
	}

	// if no store is specified, use a memstore
	// NOTE: we don't initialize this in defaultConfig because otherwise the same
	// memstore is shared between Blockchain instances
	if config.Store == nil {
		config.Store = memstore.New()
	}
	store := config.Store

	latestBlock, err := store.LatestBlock()
	if err == nil && latestBlock.Height > 0 {
		// storage contains data, load state from storage
		latestLedgerView := store.LedgerViewByHeight(latestBlock.Height)

		// restore pending block header from store information
		pendingBlock = newPendingBlock(&latestBlock, latestLedgerView)
		rootAccount = getAccount(latestLedgerView, flow.RootAddress)
	} else if err != nil && !errors.Is(err, storage.ErrNotFound) {
		// internal storage error, fail fast
		return nil, err
	} else {
		genesisLedgerView := store.LedgerViewByHeight(0)

		// storage is empty, create the root account
		createAccount(genesisLedgerView, config.RootAccountKey)

		rootAccount = getAccount(genesisLedgerView, flow.RootAddress)

		// commit the genesis block to storage
		genesis := types.GenesisBlock()

		err := store.CommitBlock(
			&genesis,
			nil,
			nil,
			nil,
			genesisLedgerView.Delta(),
			nil,
		)
		if err != nil {
			return nil, err
		}

		// get empty ledger view
		ledgerView := store.LedgerViewByHeight(0)

		// create pending block from genesis
		pendingBlock = newPendingBlock(&genesis, ledgerView)
	}

	b := &Blockchain{
		storage:               config.Store,
		pendingBlock:          pendingBlock,
		rootAccountAddress:    rootAccount.Address,
		rootAccountPrivateKey: config.RootAccountKey,
		lastCreatedAddress:    rootAccount.Address,
	}

	interpreterRuntime := runtime.NewInterpreterRuntime()
	b.computer = newComputer(interpreterRuntime)

	return b, nil
}

// RootAccountAddress returns the root account address for this blockchain.
func (b *Blockchain) RootAccountAddress() flow.Address {
	return b.rootAccountAddress
}

// RootKey returns the root private key for this blockchain.
func (b *Blockchain) RootKey() RootKey {
	rootAccountKey := RootKey{
		Address:    b.rootAccountAddress,
		PrivateKey: b.rootAccountPrivateKey,
	}

	rootAccount, err := b.getAccount(b.rootAccountAddress)
	if err != nil {
		return rootAccountKey
	}

	if len(rootAccount.Keys) > 0 {
		rootAccountKey.ID = rootAccount.Keys[0].ID
		rootAccountKey.SequenceNumber = rootAccount.Keys[0].SequenceNumber
	}

	return rootAccountKey
}

// PendingBlockID returns the ID of the pending block.
func (b *Blockchain) PendingBlockID() flow.Identifier {
	return b.pendingBlock.ID()
}

// GetLatestBlock gets the latest sealed block.
func (b *Blockchain) GetLatestBlock() (*types.Block, error) {
	block, err := b.storage.LatestBlock()
	if err != nil {
		return nil, &ErrStorage{err}
	}
	return &block, nil
}

// GetBlockByID gets a block by ID.
func (b *Blockchain) GetBlockByID(id flow.Identifier) (*types.Block, error) {
	block, err := b.storage.BlockByID(id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, &ErrBlockNotFoundByID{ID: id}
		}
		return nil, &ErrStorage{err}
	}

	return &block, nil
}

// GetBlockByHeight gets a block by height.
func (b *Blockchain) GetBlockByHeight(height uint64) (*types.Block, error) {
	block, err := b.storage.BlockByHeight(height)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, &ErrBlockNotFoundByHeight{Height: height}
		}
		return nil, err
	}

	return &block, nil
}

func (b *Blockchain) GetCollection(colID flow.Identifier) (*model.LightCollection, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	col, err := b.storage.CollectionByID(colID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, &ErrCollectionNotFound{ID: colID}
		}
		return nil, &ErrStorage{err}
	}

	return &col, nil
}

// GetTransaction gets an existing transaction by ID.
//
// The function first looks in the pending block, then the current blockchain state.
func (b *Blockchain) GetTransaction(txID flow.Identifier) (*flow.Transaction, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	pendingTx := b.pendingBlock.GetTransaction(txID)
	if pendingTx != nil {
		return pendingTx, nil
	}

	tx, err := b.storage.TransactionByID(txID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, &ErrTransactionNotFound{ID: txID}
		}
		return nil, &ErrStorage{err}
	}

	return &tx, nil
}

func (b *Blockchain) GetTransactionResult(txID flow.Identifier) (*flow.TransactionResult, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.pendingBlock.ContainsTransaction(txID) {
		return &flow.TransactionResult{
			Status: flow.TransactionStatusPending,
		}, nil
	}

	result, err := b.storage.TransactionResultByID(txID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return &flow.TransactionResult{
				Status: flow.TransactionStatusUnknown,
			}, nil
		}
		return nil, &ErrStorage{err}
	}

	return &result, nil
}

// GetAccount returns the account for the given address.
func (b *Blockchain) GetAccount(address flow.Address) (*flow.Account, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.getAccount(address)
}

// getAccount returns the account for the given address.
func (b *Blockchain) getAccount(address flow.Address) (*flow.Account, error) {
	latestBlock, err := b.GetLatestBlock()
	if err != nil {
		return nil, err
	}

	latestLedgerView := b.storage.LedgerViewByHeight(latestBlock.Height)

	acct := getAccount(latestLedgerView, address)
	if acct == nil {
		return nil, &ErrAccountNotFound{Address: address}
	}

	return acct, nil
}

// TODO: Implement
func (b *Blockchain) GetAccountAtBlock(address flow.Address, blockHeight uint64) (*flow.Account, error) {
	panic("not implemented")
}

func getAccount(ledgerView *types.LedgerView, address flow.Address) *flow.Account {
	runtimeCtx := execution.NewRuntimeContext(ledgerView)
	return runtimeCtx.GetAccount(address)
}

// GetEventsByHeight returns the events in the block at the given height, optionally filtered by type.
func (b *Blockchain) GetEventsByHeight(blockHeight uint64, eventType string) ([]flow.Event, error) {
	return b.storage.EventsByHeight(blockHeight, eventType)
}

// AddTransaction validates a transaction and adds it to the current pending block.
func (b *Blockchain) AddTransaction(tx flow.Transaction) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// If Index > 0, pending block has begun execution (cannot add anymore txs)
	if b.pendingBlock.ExecutionStarted() {
		return &ErrPendingBlockMidExecution{BlockID: b.pendingBlock.ID()}
	}

	if b.pendingBlock.ContainsTransaction(tx.ID()) {
		return &ErrDuplicateTransaction{TxID: tx.ID()}
	}

	_, err := b.storage.TransactionByID(tx.ID())
	if err == nil {
		// Found the transaction, this is a dupe
		return &ErrDuplicateTransaction{TxID: tx.ID()}
	} else if !errors.Is(err, storage.ErrNotFound) {
		// Error in the storage provider
		return fmt.Errorf("failed to check storage for transaction %w", err)
	}

	if tx.ProposalKey == (flow.ProposalKey{}) {
		return &ErrInvalidTransaction{TxID: tx.ID(), MissingFields: []string{"proposal_key"}}
	}

	if err := b.verifySignatures(tx); err != nil {
		return err
	}

	// add transaction to pending block
	b.pendingBlock.AddTransaction(tx)

	return nil
}

// ExecuteBlock executes the remaining transactions in pending block.
func (b *Blockchain) ExecuteBlock() ([]*TransactionResult, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.executeBlock()
}

func (b *Blockchain) executeBlock() ([]*TransactionResult, error) {
	results := make([]*TransactionResult, 0)

	// empty blocks do not require execution, treat as a no-op
	if b.pendingBlock.Empty() {
		return results, nil
	}

	// cannot execute a block that has already executed
	if b.pendingBlock.ExecutionComplete() {
		return results, &ErrPendingBlockTransactionsExhausted{
			BlockID: b.pendingBlock.ID(),
		}
	}

	// continue executing transactions until execution is complete
	for !b.pendingBlock.ExecutionComplete() {
		result, err := b.executeNextTransaction()
		if err != nil {
			return results, err
		}

		results = append(results, result)
	}

	return results, nil
}

// ExecuteNextTransaction executes the next indexed transaction in pending block.
func (b *Blockchain) ExecuteNextTransaction() (*TransactionResult, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.executeNextTransaction()
}

// executeNextTransaction is a helper function for ExecuteBlock and ExecuteNextTransaction that
// executes the next transaction in the pending block.
func (b *Blockchain) executeNextTransaction() (*TransactionResult, error) {
	// check if there are remaining txs to be executed
	if b.pendingBlock.ExecutionComplete() {
		return nil, &ErrPendingBlockTransactionsExhausted{
			BlockID: b.pendingBlock.ID(),
		}
	}

	// use the computer to execute the next transaction
	result, err := b.pendingBlock.ExecuteNextTransaction(
		func(
			ledgerView *types.LedgerView,
			tx flow.Transaction,
		) (*TransactionResult, error) {
			return b.computer.ExecuteTransaction(ledgerView, tx)
		},
	)
	if err != nil {
		// fail fast if fatal error occurs
		return nil, err
	}

	return result, nil
}

// CommitBlock seals the current pending block and saves it to storage.
//
// This function clears the pending transaction pool and resets the pending block.
func (b *Blockchain) CommitBlock() (*types.Block, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.commitBlock()
}

func (b *Blockchain) commitBlock() (*types.Block, error) {
	// pending block cannot be committed before execution starts (unless empty)
	if !b.pendingBlock.ExecutionStarted() && !b.pendingBlock.Empty() {
		return nil, &ErrPendingBlockCommitBeforeExecution{BlockID: b.pendingBlock.ID()}
	}

	// pending block cannot be committed before execution completes
	if b.pendingBlock.ExecutionStarted() && !b.pendingBlock.ExecutionComplete() {
		return nil, &ErrPendingBlockMidExecution{BlockID: b.pendingBlock.ID()}
	}

	block := b.pendingBlock.Block()
	collections := b.pendingBlock.Collections()
	transactions := b.pendingBlock.Transactions()
	transactionResults := convertToSealedResults(b.pendingBlock.TransactionResults())
	delta := b.pendingBlock.LedgerDelta()
	events := b.pendingBlock.Events()

	// commit the pending block to storage
	err := b.storage.CommitBlock(block, collections, transactions, transactionResults, delta, events)
	if err != nil {
		return nil, err
	}

	// update system state based on emitted events
	b.handleEvents(events, block.Height)

	ledgerView := b.storage.LedgerViewByHeight(block.Height)

	// reset pending block using current block and ledger state
	b.pendingBlock = newPendingBlock(block, ledgerView)

	return block, nil
}

// ExecuteAndCommitBlock is a utility that combines ExecuteBlock with CommitBlock.
func (b *Blockchain) ExecuteAndCommitBlock() (*types.Block, []*TransactionResult, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	results, err := b.executeBlock()
	if err != nil {
		return nil, nil, err
	}

	block, err := b.commitBlock()
	if err != nil {
		return nil, results, err
	}

	return block, results, nil
}

// ResetPendingBlock clears the transactions in pending block.
func (b *Blockchain) ResetPendingBlock() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	latestBlock, err := b.GetLatestBlock()
	if err != nil {
		return err
	}

	latestLedgerView := b.storage.LedgerViewByHeight(latestBlock.Height)

	// reset pending block using latest committed block and ledger state
	b.pendingBlock = newPendingBlock(latestBlock, latestLedgerView)

	return nil
}

// ExecuteScript executes a read-only script against the world state and returns the result.
func (b *Blockchain) ExecuteScript(script []byte) (*ScriptResult, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	latestBlock, err := b.GetLatestBlock()
	if err != nil {
		return nil, err
	}

	latestLedgerView := b.storage.LedgerViewByHeight(latestBlock.Height)

	result, err := b.computer.ExecuteScript(latestLedgerView, script)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (b *Blockchain) ExecuteScriptAtBlock(script []byte, blockHeight uint64) (*ScriptResult, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	requestedBlock, err := b.GetBlockByHeight(blockHeight)
	if err != nil {
		return nil, err
	}

	requestedLedgerView := b.storage.LedgerViewByHeight(requestedBlock.Height)

	result, err := b.computer.ExecuteScript(requestedLedgerView, script)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// LastCreatedAccount returns the last account that was created in the blockchain.
func (b *Blockchain) LastCreatedAccount() flow.Account {
	account, _ := b.getAccount(b.lastCreatedAddress)
	return *account
}

// verifySignatures verifies that a transaction contains the necessary signatures.
//
// An error is returned if any of the expected signatures are invalid or missing.
func (b *Blockchain) verifySignatures(tx flow.Transaction) error {
	if tx.Payer == flow.ZeroAddress {
		// TODO: add error type for missing payer
		return fmt.Errorf("missing payer signature")
	}

	payloadWeights, proposalKeyVerifiedInPayload, err := b.aggregateAccountSignatures(
		tx.PayloadSignatures,
		tx.PayloadMessage(),
		tx.ProposalKey,
	)
	if err != nil {
		return err
	}

	envelopeWeights, proposalKeyVerifiedInEnvelope, err := b.aggregateAccountSignatures(
		tx.EnvelopeSignatures,
		tx.EnvelopeMessage(),
		tx.ProposalKey,
	)
	if err != nil {
		return err
	}

	proposalKeyVerified := proposalKeyVerifiedInPayload || proposalKeyVerifiedInEnvelope

	if !proposalKeyVerified {
		return fmt.Errorf(
			"missing signature for proposal key (address: %s, key: %d)",
			tx.ProposalKey.Address,
			tx.ProposalKey.KeyID,
		)
	}

	for _, addr := range tx.Authorizers {
		// Skip this authorizer if it is also the payer. In the case where an account is
		// both a PAYER as well as an AUTHORIZER or PROPOSER, that account is required
		// to sign only the envelope.
		if addr == tx.Payer {
			continue
		}

		if !hasSufficientKeyWeight(payloadWeights, addr) {
			return &ErrMissingSignature{addr}
		}
	}

	if !hasSufficientKeyWeight(envelopeWeights, tx.Payer) {
		return &ErrMissingSignature{tx.Payer}
	}

	return nil
}

func (b *Blockchain) aggregateAccountSignatures(
	signatures []flow.TransactionSignature,
	message []byte,
	proposalKey flow.ProposalKey,
) (
	weights map[flow.Address]int,
	proposalKeyVerified bool,
	err error,
) {
	weights = make(map[flow.Address]int)

	for _, txSig := range signatures {
		accountKey, err := b.verifyAccountSignature(txSig, message)
		if err != nil {
			return nil, false, err
		}

		if sigIsForProposalKey(txSig, proposalKey) {
			proposalKeyVerified = true
		}

		weights[txSig.Address] += accountKey.Weight
	}

	return
}

// CreateAccount submits a transaction to create a new account with the given
// account keys and code. The transaction is paid by the root account.
func (b *Blockchain) CreateAccount(publicKeys []flow.AccountKey, code []byte) (flow.Address, error) {
	createAccountScript, err := templates.CreateAccount(publicKeys, code)

	if err != nil {
		return flow.Address{}, err
	}

	rootKey := b.RootKey()

	tx := flow.NewTransaction().
		SetScript(createAccountScript).
		SetGasLimit(10).
		SetProposalKey(rootKey.Address, rootKey.ID, rootKey.SequenceNumber).
		SetPayer(rootKey.Address)

	err = tx.SignEnvelope(rootKey.Address, rootKey.ID, rootKey.Signer())
	if err != nil {
		return flow.Address{}, err
	}

	err = b.AddTransaction(*tx)
	if err != nil {
		return flow.Address{}, err
	}

	result, err := b.ExecuteNextTransaction()
	if err != nil {
		return flow.Address{}, err
	}

	_, err = b.CommitBlock()
	if err != nil {
		return flow.Address{}, err
	}

	if result.Reverted() {
		return flow.Address{}, result.Error
	}

	return b.LastCreatedAccount().Address, nil
}

// verifyAccountSignature verifies that an account signature is valid for the
// account and given message.
//
// If the signature is valid, this function returns the associated account key.
//
// An error is returned if the account does not contain a public key that
// correctly verifies the signature against the given message.
func (b *Blockchain) verifyAccountSignature(
	txSig flow.TransactionSignature,
	message []byte,
) (accountKey flow.AccountKey, err error) {
	account, err := b.getAccount(txSig.Address)
	if err != nil {
		return accountKey, &ErrInvalidSignatureAccount{Address: txSig.Address}
	}

	signature := crypto.Signature(txSig.Signature)

	if txSig.KeyID < 0 || txSig.KeyID >= len(account.Keys) {
		return accountKey, &ErrInvalidSignatureAccount{Address: txSig.Address}
	}

	accountKey = account.Keys[txSig.KeyID]

	hasher, err := crypto.NewHasher(accountKey.HashAlgo)
	if err != nil {
		return accountKey, fmt.Errorf("public key specifies invalid hash algorithm")
	}

	valid, err := accountKey.PublicKey.Verify(signature, message, hasher)
	if err != nil {
		return accountKey, fmt.Errorf("cannot verify public key")
	}

	if !valid {
		return accountKey, &ErrInvalidSignaturePublicKey{Account: txSig.Address, KeyID: txSig.KeyID}
	}

	return accountKey, nil
}

// handleEvents updates emulator state based on emitted system events.
func (b *Blockchain) handleEvents(events []flow.Event, blockHeight uint64) {
	for _, event := range events {
		// update lastCreatedAccount if this is an AccountCreated event
		if event.Type == flow.EventAccountCreated {
			acctCreatedEvent := flow.AccountCreatedEvent(event)

			b.lastCreatedAddress = acctCreatedEvent.Address()
		}

	}
}

// createAccount creates an account with the given private key and injects it
// into the given state, bypassing the need for a transaction.
func createAccount(ledgerView *types.LedgerView, privateKey flow.AccountPrivateKey) flow.Account {
	accountKey := privateKey.ToAccountKey()
	accountKey.Weight = keys.PublicKeyWeightThreshold

	publicKeyBytes, err := keys.EncodePublicKey(accountKey)
	if err != nil {
		panic(err)
	}

	runtimeContext := execution.NewRuntimeContext(ledgerView)
	accountAddress, err := runtimeContext.CreateAccount(
		[][]byte{publicKeyBytes},
	)
	if err != nil {
		panic(err)
	}

	account := runtimeContext.GetAccount(flow.Address(accountAddress))
	return *account
}

func sigIsForProposalKey(txSig flow.TransactionSignature, proposalKey flow.ProposalKey) bool {
	return txSig.Address == proposalKey.Address && txSig.KeyID == proposalKey.KeyID
}

func hasSufficientKeyWeight(weights map[flow.Address]int, address flow.Address) bool {
	return weights[address] >= keys.PublicKeyWeightThreshold
}

func convertToSealedResults(
	results map[flow.Identifier]*TransactionResult,
) map[flow.Identifier]*flow.TransactionResult {

	output := make(map[flow.Identifier]*flow.TransactionResult)

	for id, result := range results {
		output[id] = &flow.TransactionResult{
			Status: flow.TransactionStatusSealed,
			Error:  result.Error,
			Events: result.Events,
		}
	}

	return output
}

func init() {
	// Initialize default emulator options
	defaultRootKey, err := keys.GeneratePrivateKey(
		keys.ECDSA_P256_SHA3_256,
		[]byte("elephant ears space cowboy octopus rodeo potato cannon pineapple"))
	if err != nil {
		panic("Failed to generate default root key: " + err.Error())
	}

	defaultConfig.RootAccountKey = defaultRootKey
}
