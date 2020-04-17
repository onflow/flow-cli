package memstore

import (
	"fmt"
	"sync"

	"github.com/dapperlabs/flow-go-sdk"
	vm "github.com/dapperlabs/flow-go/engine/execution/computation/virtualmachine"
	model "github.com/dapperlabs/flow-go/model/flow"

	"github.com/dapperlabs/flow-emulator/storage"
	"github.com/dapperlabs/flow-emulator/types"
)

// Store implements the Store interface with an in-memory store.
type Store struct {
	mu sync.RWMutex
	// block ID to block height
	blockIDToHeight map[flow.Identifier]uint64
	// blocks by height
	blocks map[uint64]types.Block
	// collections by ID
	collections map[model.Identifier]model.LightCollection
	// transactions by ID
	transactions map[flow.Identifier]flow.Transaction
	// transaction results by ID
	transactionResults map[flow.Identifier]flow.TransactionResult
	// ledger states by block height
	ledger map[uint64]vm.MapLedger
	// events by block height
	eventsByBlockHeight map[uint64][]flow.Event
	// highest block height
	blockHeight uint64
}

// New returns a new in-memory Store implementation.
func New() *Store {
	return &Store{
		mu:                  sync.RWMutex{},
		blockIDToHeight:     make(map[flow.Identifier]uint64),
		blocks:              make(map[uint64]types.Block),
		collections:         make(map[model.Identifier]model.LightCollection),
		transactions:        make(map[flow.Identifier]flow.Transaction),
		transactionResults:  make(map[flow.Identifier]flow.TransactionResult),
		ledger:              make(map[uint64]vm.MapLedger),
		eventsByBlockHeight: make(map[uint64][]flow.Event),
	}
}

func (s *Store) LatestBlock() (types.Block, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	latestBlock, ok := s.blocks[s.blockHeight]
	if !ok {
		return types.Block{}, storage.ErrNotFound
	}
	return latestBlock, nil
}

func (s *Store) BlockByID(id flow.Identifier) (types.Block, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	blockHeight := s.blockIDToHeight[id]
	block, ok := s.blocks[blockHeight]
	if !ok {
		return types.Block{}, storage.ErrNotFound
	}

	return block, nil
}

func (s *Store) BlockByHeight(blockHeight uint64) (types.Block, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	block, ok := s.blocks[blockHeight]
	if !ok {
		return types.Block{}, storage.ErrNotFound
	}

	return block, nil
}

func (s *Store) insertBlock(block types.Block) error {
	s.blocks[block.Height] = block
	if block.Height > s.blockHeight {
		s.blockHeight = block.Height
	}

	return nil
}

func (s *Store) CommitBlock(
	block *types.Block,
	collections []*model.LightCollection,
	transactions map[flow.Identifier]*flow.Transaction,
	transactionResults map[flow.Identifier]*flow.TransactionResult,
	delta types.LedgerDelta,
	events []flow.Event,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(transactions) != len(transactionResults) {
		return fmt.Errorf(
			"transactions count (%d) does not match result count (%d)",
			len(transactions),
			len(transactionResults),
		)
	}

	err := s.insertBlock(*block)
	if err != nil {
		return err
	}

	for _, col := range collections {
		err := s.insertCollection(*col)
		if err != nil {
			return err
		}
	}

	for txID, tx := range transactions {
		err := s.insertTransaction(txID, *tx)
		if err != nil {
			return err
		}
	}

	for txID, result := range transactionResults {
		err := s.insertTransactionResult(txID, *result)
		if err != nil {
			return err
		}
	}

	err = s.insertLedgerDelta(block.Height, delta)
	if err != nil {
		return err
	}

	err = s.insertEvents(block.Height, events)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) CollectionByID(colID flow.Identifier) (model.LightCollection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tx, ok := s.collections[model.Identifier(colID)]
	if !ok {
		return model.LightCollection{}, storage.ErrNotFound
	}
	return tx, nil
}

func (s *Store) insertCollection(col model.LightCollection) error {
	s.collections[col.ID()] = col
	return nil
}

func (s *Store) TransactionByID(txID flow.Identifier) (flow.Transaction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tx, ok := s.transactions[txID]
	if !ok {
		return flow.Transaction{}, storage.ErrNotFound
	}
	return tx, nil
}

func (s *Store) insertTransaction(txID flow.Identifier, tx flow.Transaction) error {
	s.transactions[txID] = tx
	return nil
}

func (s *Store) TransactionResultByID(txID flow.Identifier) (flow.TransactionResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result, ok := s.transactionResults[txID]
	if !ok {
		return flow.TransactionResult{}, storage.ErrNotFound
	}
	return result, nil
}

func (s *Store) insertTransactionResult(txID flow.Identifier, result flow.TransactionResult) error {
	s.transactionResults[txID] = result
	return nil
}

func (s *Store) LedgerViewByHeight(blockHeight uint64) *types.LedgerView {
	return types.NewLedgerView(func(key string) ([]byte, error) {
		s.mu.RLock()
		defer s.mu.RUnlock()

		ledger, ok := s.ledger[blockHeight]
		if !ok {
			return nil, nil
		}

		return ledger[key], nil
	})
}

func (s *Store) insertLedgerDelta(blockHeight uint64, delta types.LedgerDelta) error {
	var oldLedger vm.MapLedger

	// use empty ledger if this is the genesis block
	if blockHeight == 0 {
		oldLedger = make(vm.MapLedger)
	} else {
		oldLedger = s.ledger[blockHeight-1]
	}

	newLedger := make(vm.MapLedger)

	// copy values from the previous ledger
	for key, value := range oldLedger {
		// do not copy deleted values
		if !delta.HasBeenDeleted(key) {
			newLedger[key] = value
		}
	}

	// write all updated values
	for key, value := range delta.Updates() {
		if value != nil {
			newLedger[key] = value
		}
	}

	s.ledger[blockHeight] = newLedger

	return nil
}

func (s *Store) EventsByHeight(blockHeight uint64, eventType string) ([]flow.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	allEvents := s.eventsByBlockHeight[blockHeight]

	events := make([]flow.Event, 0)

	for _, event := range allEvents {
		if eventType == "" {
			events = append(events, event)
		} else {
			if event.Type == eventType {
				events = append(events, event)
			}
		}
	}

	return events, nil
}

func (s *Store) insertEvents(blockHeight uint64, events []flow.Event) error {
	if s.eventsByBlockHeight[blockHeight] == nil {
		s.eventsByBlockHeight[blockHeight] = events
	} else {
		s.eventsByBlockHeight[blockHeight] = append(s.eventsByBlockHeight[blockHeight], events...)
	}

	return nil
}
