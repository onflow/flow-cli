package memstore

import (
	"fmt"
	"sync"

	"github.com/dapperlabs/flow-go/engine/execution/state/delta"
	"github.com/dapperlabs/flow-go/fvm/state"
	flowgo "github.com/dapperlabs/flow-go/model/flow"

	"github.com/dapperlabs/flow-emulator/storage"
	"github.com/dapperlabs/flow-emulator/types"
)

// Store implements the Store interface with an in-memory store.
type Store struct {
	mu sync.RWMutex
	// block ID to block height
	blockIDToHeight map[flowgo.Identifier]uint64
	// blocks by height
	blocks map[uint64]flowgo.Block
	// collections by ID
	collections map[flowgo.Identifier]flowgo.LightCollection
	// transactions by ID
	transactions map[flowgo.Identifier]flowgo.TransactionBody
	// Transaction results by ID
	transactionResults map[flowgo.Identifier]types.StorableTransactionResult
	// Ledger states by block height
	ledger map[uint64]*state.MapLedger
	// events by block height
	eventsByBlockHeight map[uint64][]flowgo.Event
	// highest block height
	blockHeight uint64
}

// New returns a new in-memory Store implementation.
func New() *Store {
	return &Store{
		mu:                  sync.RWMutex{},
		blockIDToHeight:     make(map[flowgo.Identifier]uint64),
		blocks:              make(map[uint64]flowgo.Block),
		collections:         make(map[flowgo.Identifier]flowgo.LightCollection),
		transactions:        make(map[flowgo.Identifier]flowgo.TransactionBody),
		transactionResults:  make(map[flowgo.Identifier]types.StorableTransactionResult),
		ledger:              make(map[uint64]*state.MapLedger),
		eventsByBlockHeight: make(map[uint64][]flowgo.Event),
	}
}

var _ storage.Store = &Store{}

func (s *Store) BlockByID(id flowgo.Identifier) (*flowgo.Block, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	blockHeight := s.blockIDToHeight[id]
	block, ok := s.blocks[blockHeight]
	if !ok {
		return nil, storage.ErrNotFound
	}

	return &block, nil
}

func (s *Store) BlockByHeight(blockHeight uint64) (*flowgo.Block, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	block, ok := s.blocks[blockHeight]
	if !ok {
		return nil, storage.ErrNotFound
	}

	return &block, nil
}

func (s *Store) LatestBlock() (flowgo.Block, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	latestBlock, ok := s.blocks[s.blockHeight]
	if !ok {
		return flowgo.Block{}, storage.ErrNotFound
	}
	return latestBlock, nil
}

func (s *Store) StoreBlock(block *flowgo.Block) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.store(block)
}

func (s *Store) store(block *flowgo.Block) error {
	s.blocks[block.Header.Height] = *block
	if block.Header.Height > s.blockHeight {
		s.blockHeight = block.Header.Height
	}

	return nil
}

func (s *Store) CommitBlock(
	block flowgo.Block,
	collections []*flowgo.LightCollection,
	transactions map[flowgo.Identifier]*flowgo.TransactionBody,
	transactionResults map[flowgo.Identifier]*types.StorableTransactionResult,
	delta delta.Delta,
	events []flowgo.Event,
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

	err := s.store(&block)
	if err != nil {
		return err
	}

	for _, col := range collections {
		err := s.insertCollection(*col)
		if err != nil {
			return err
		}
	}

	for _, tx := range transactions {
		err := s.insertTransaction(tx.ID(), *tx)
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

	err = s.insertLedgerDelta(block.Header.Height, delta)
	if err != nil {
		return err
	}

	err = s.insertEvents(block.Header.Height, events)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) CollectionByID(colID flowgo.Identifier) (flowgo.LightCollection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tx, ok := s.collections[colID]
	if !ok {
		return flowgo.LightCollection{}, storage.ErrNotFound
	}
	return tx, nil
}

func (s *Store) insertCollection(col flowgo.LightCollection) error {
	s.collections[col.ID()] = col
	return nil
}

func (s *Store) TransactionByID(txID flowgo.Identifier) (flowgo.TransactionBody, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tx, ok := s.transactions[txID]
	if !ok {
		return flowgo.TransactionBody{}, storage.ErrNotFound
	}
	return tx, nil
}

func (s *Store) insertTransaction(txID flowgo.Identifier, tx flowgo.TransactionBody) error {
	s.transactions[txID] = tx
	return nil
}

func (s *Store) TransactionResultByID(txID flowgo.Identifier) (types.StorableTransactionResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result, ok := s.transactionResults[txID]
	if !ok {
		return types.StorableTransactionResult{}, storage.ErrNotFound
	}
	return result, nil
}

func (s *Store) insertTransactionResult(txID flowgo.Identifier, result types.StorableTransactionResult) error {
	s.transactionResults[txID] = result
	return nil
}

func (s *Store) LedgerViewByHeight(blockHeight uint64) *delta.View {
	return delta.NewView(func(key flowgo.RegisterID) (value flowgo.RegisterValue, err error) {

		s.mu.RLock()
		defer s.mu.RUnlock()

		ledger, ok := s.ledger[blockHeight]
		if !ok {
			return nil, nil
		}

		return ledger.Get(key)
	})
}

func (s *Store) insertLedgerDelta(blockHeight uint64, delta delta.Delta) error {
	var oldLedger *state.MapLedger

	// use empty ledger if this is the genesis block
	if blockHeight == 0 {
		oldLedger = state.NewMapLedger()
	} else {
		oldLedger = s.ledger[blockHeight-1]
	}

	newLedger := state.NewMapLedger()

	// copy values from the previous ledger
	for keyString, value := range oldLedger.Registers {
		key := flowgo.RegisterID(keyString)
		// do not copy deleted values
		if !delta.HasBeenDeleted(key) {
			newLedger.Set(key, value)
		}
	}

	// write all updated values
	ids, values := delta.RegisterUpdates()
	for i, value := range values {
		key := ids[i]
		if value != nil {
			newLedger.Set(key, value)
		}
	}

	s.ledger[blockHeight] = newLedger

	return nil
}

func (s *Store) EventsByHeight(blockHeight uint64, eventType string) ([]flowgo.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	allEvents := s.eventsByBlockHeight[blockHeight]

	events := make([]flowgo.Event, 0)

	for _, event := range allEvents {
		if eventType == "" {
			events = append(events, event)
		} else {
			if string(event.Type) == eventType {
				events = append(events, event)
			}
		}
	}

	return events, nil
}

func (s *Store) insertEvents(blockHeight uint64, events []flowgo.Event) error {
	if s.eventsByBlockHeight[blockHeight] == nil {
		s.eventsByBlockHeight[blockHeight] = events
	} else {
		s.eventsByBlockHeight[blockHeight] = append(s.eventsByBlockHeight[blockHeight], events...)
	}

	return nil
}
