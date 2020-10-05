// Package storage defines the interface and implementations for interacting with
// persistent chain state.
package storage

import (
	"github.com/onflow/flow-go/engine/execution/state/delta"
	flowgo "github.com/onflow/flow-go/model/flow"

	"github.com/dapperlabs/flow-emulator/types"
)

// Store defines the storage layer for persistent chain state.
//
// This includes finalized blocks and transactions, and the resultant register
// states and emitted events. It does not include pending state, such as pending
// transactions and register states.
//
// Implementations must distinguish between not found errors and errors with
// the underlying storage by returning an instance of store.ErrNotFound if a
// resource cannot be found.
//
// Implementations must be safe for use by multiple goroutines.
type Store interface {

	// LatestBlock returns the block with the highest block height.
	LatestBlock() (flowgo.Block, error)

	// Store stores the block. If the exactly same block is already in a storage, return successfully
	StoreBlock(block *flowgo.Block) error

	// BlockByID returns the block with the given hash. It is available for
	// finalized and ambiguous blocks.
	BlockByID(blockID flowgo.Identifier) (*flowgo.Block, error)

	// BlockByHeight returns the block at the given height. It is only available
	// for finalized blocks.
	BlockByHeight(height uint64) (*flowgo.Block, error)

	// CommitBlock atomically saves the execution results for a block.
	CommitBlock(
		block flowgo.Block,
		collections []*flowgo.LightCollection,
		transactions map[flowgo.Identifier]*flowgo.TransactionBody,
		transactionResults map[flowgo.Identifier]*types.StorableTransactionResult,
		delta delta.Delta,
		events []flowgo.Event,
	) error

	// CollectionByID gets the collection (transaction IDs only) with the given ID.
	CollectionByID(flowgo.Identifier) (flowgo.LightCollection, error)

	// TransactionByID gets the transaction with the given ID.
	TransactionByID(flowgo.Identifier) (flowgo.TransactionBody, error)

	// TransactionResultByID gets the transaction result with the given ID.
	TransactionResultByID(flowgo.Identifier) (types.StorableTransactionResult, error)

	// LedgerViewByHeight returns a view into the ledger state at a given block.
	LedgerViewByHeight(blockHeight uint64) *delta.View

	// EventsByHeight returns the events in the block at the given height, optionally filtered by type.
	EventsByHeight(blockHeight uint64, eventType string) ([]flowgo.Event, error)
}
