package emulator

import (
	"github.com/dapperlabs/flow-go-sdk"
	model "github.com/dapperlabs/flow-go/model/flow"

	"github.com/dapperlabs/flow-emulator/types"
)

// A pendingBlock contains the pending state required to form a new block.
type pendingBlock struct {
	height   uint64
	parentID flow.Identifier
	// mapping from transaction ID to transaction
	transactions map[flow.Identifier]*flow.Transaction
	// list of transaction IDs in the block
	transactionIDs []flow.Identifier
	// mapping from transaction ID to transaction result
	transactionResults map[flow.Identifier]*TransactionResult
	// current working ledger, updated after each transaction execution
	ledgerView *types.LedgerView
	// events emitted during execution
	events []flow.Event
	// index of transaction execution
	index int
}

// newPendingBlock creates a new pending block using the specified block as its parent.
func newPendingBlock(prevBlock *types.Block, ledgerView *types.LedgerView) *pendingBlock {
	return &pendingBlock{
		height:             prevBlock.Height + 1,
		parentID:           prevBlock.ID(),
		transactions:       make(map[flow.Identifier]*flow.Transaction),
		transactionIDs:     make([]flow.Identifier, 0),
		transactionResults: make(map[flow.Identifier]*TransactionResult),
		ledgerView:         ledgerView,
		events:             make([]flow.Event, 0),
		index:              0,
	}
}

// ID returns the ID of the pending block.
func (b *pendingBlock) ID() flow.Identifier {
	return b.Block().ID()
}

// Height returns the number of the pending block.
func (b *pendingBlock) Height() uint64 {
	return b.height
}

// Block returns the block information for the pending block.
func (b *pendingBlock) Block() *types.Block {
	collections := b.Collections()

	guarantees := make([]*model.CollectionGuarantee, len(collections))
	for i, collection := range collections {
		guarantees[i] = &model.CollectionGuarantee{
			CollectionID: collection.ID(),
		}
	}

	return &types.Block{
		Height:     b.height,
		ParentID:   b.parentID,
		Guarantees: guarantees,
	}
}

func (b *pendingBlock) Collections() []*model.LightCollection {
	if len(b.transactionIDs) == 0 {
		return []*model.LightCollection{}
	}

	transactionIDs := make([]model.Identifier, len(b.transactionIDs))

	// TODO: remove once SDK models are removed
	for i, transactionID := range b.transactionIDs {
		transactionIDs[i] = model.Identifier(transactionID)
	}

	collection := model.LightCollection{Transactions: transactionIDs}

	return []*model.LightCollection{&collection}
}

func (b *pendingBlock) Transactions() map[flow.Identifier]*flow.Transaction {
	return b.transactions
}

func (b *pendingBlock) TransactionResults() map[flow.Identifier]*TransactionResult {
	return b.transactionResults
}

// LedgerDelta returns the ledger delta for the pending block.
func (b *pendingBlock) LedgerDelta() types.LedgerDelta {
	return b.ledgerView.Delta()
}

// AddTransaction adds a transaction to the pending block.
func (b *pendingBlock) AddTransaction(tx flow.Transaction) {
	b.transactionIDs = append(b.transactionIDs, tx.ID())
	b.transactions[tx.ID()] = &tx
}

// ContainsTransaction checks if a transaction is included in the pending block.
func (b *pendingBlock) ContainsTransaction(txID flow.Identifier) bool {
	_, exists := b.transactions[txID]
	return exists
}

// GetTransaction retrieves a transaction in the pending block by ID.
func (b *pendingBlock) GetTransaction(txID flow.Identifier) *flow.Transaction {
	return b.transactions[txID]
}

// nextTransaction returns the next indexed transaction.
func (b *pendingBlock) nextTransaction() *flow.Transaction {
	txID := b.transactionIDs[b.index]
	return b.GetTransaction(txID)
}

// ExecuteNextTransaction executes the next transaction in the pending block.
//
// This function uses the provided execute function to perform the actual
// execution, then updates the pending block with the output.
func (b *pendingBlock) ExecuteNextTransaction(
	execute func(ledgerView *types.LedgerView, tx flow.Transaction) (*TransactionResult, error),
) (*TransactionResult, error) {
	tx := b.nextTransaction()

	result, err := execute(b.ledgerView, *tx)
	if err != nil {
		// fail fast if fatal error occurs
		return nil, err
	}

	// increment transaction index even if transaction reverts
	b.index++

	if result.Error == nil {
		b.events = append(b.events, result.Events...)
	}

	b.transactionResults[tx.ID()] = result

	return result, nil
}

// Events returns all events captured during the execution of the pending block.
func (b *pendingBlock) Events() []flow.Event {
	return b.events
}

// ExecutionStarted returns true if the pending block has started executing.
func (b *pendingBlock) ExecutionStarted() bool {
	return b.index > 0
}

// ExecutionComplete returns true if the pending block is fully executed.
func (b *pendingBlock) ExecutionComplete() bool {
	return b.index >= b.Size()
}

// Size returns the number of transactions in the pending block.
func (b *pendingBlock) Size() int {
	return len(b.transactionIDs)
}

// Empty returns true if the pending block is empty.
func (b *pendingBlock) Empty() bool {
	return b.Size() == 0
}
