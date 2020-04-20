package badger

import (
	"errors"
	"fmt"

	"github.com/dapperlabs/flow-go-sdk"
	model "github.com/dapperlabs/flow-go/model/flow"
	"github.com/dgraph-io/badger/v2"

	"github.com/dapperlabs/flow-emulator/storage"
	"github.com/dapperlabs/flow-emulator/types"
)

// Store is an embedded storage implementation using Badger as the underlying
// persistent key-value store.
type Store struct {
	db              *badger.DB
	ledgerChangeLog changelog
}

// New returns a new Badger Store.
func New(opts ...Opt) (*Store, error) {
	badgerOptions := getBadgerOptions(opts...)

	db, err := badger.Open(badgerOptions)
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}

	store := &Store{db, newChangelog()}
	if err = store.setup(); err != nil {
		return nil, err
	}

	return store, nil
}

// setup sets up in-memory indexes and prepares the store for use.
func (s *Store) setup() error {
	s.db.RLock()
	defer s.db.RUnlock()

	iterOpts := badger.DefaultIteratorOptions
	// only search for changelog entries
	iterOpts.Prefix = []byte(ledgerChangelogKeyPrefix)
	// create a buffer for copying changelists, this is reused for each register
	clistBuf := make([]byte, 256)

	// read the changelist from disk for each register
	return s.db.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(iterOpts)
		defer iter.Close()

		for iter.Rewind(); iter.Valid(); iter.Next() {
			item := iter.Item()
			registerID := registerIDFromLedgerChangelogKey(item.Key())
			// ensure the register ID is value
			if len(registerID) == 0 {
				return errors.New("found changelist for invalid register ID")
			}

			// decode the changelist
			encClist, err := item.ValueCopy(clistBuf)
			if err != nil {
				return err
			}
			var clist changelist
			if err := decodeChangelist(&clist, encClist); err != nil {
				return err
			}

			// add to the changelog
			s.ledgerChangeLog.setChangelist(registerID, clist)
		}
		return nil
	})
}

func (s *Store) LatestBlock() (block types.Block, err error) {
	err = s.db.View(func(txn *badger.Txn) error {
		// get latest block height
		latestBlockHeight, err := getLatestBlockHeightTx(txn)
		if err != nil {
			return err
		}

		// get corresponding block
		encBlock, err := getTx(txn)(blockKey(latestBlockHeight))
		if err != nil {
			return err
		}
		return decodeBlock(&block, encBlock)
	})
	return
}

func (s *Store) BlockByID(blockID flow.Identifier) (block types.Block, err error) {
	err = s.db.View(func(txn *badger.Txn) error {
		// get block height by block ID
		encBlockHeight, err := getTx(txn)(blockIDIndexKey(blockID))
		if err != nil {
			return err
		}

		// decode block height
		var blockHeight uint64
		if err := decodeUint64(&blockHeight, encBlockHeight); err != nil {
			return err
		}

		// get block by block height and decode
		encBlock, err := getTx(txn)(blockKey(blockHeight))
		if err != nil {
			return err
		}
		return decodeBlock(&block, encBlock)
	})
	return
}

func (s *Store) BlockByHeight(blockHeight uint64) (block types.Block, err error) {
	err = s.db.View(func(txn *badger.Txn) error {
		encBlock, err := getTx(txn)(blockKey(blockHeight))
		if err != nil {
			return err
		}
		return decodeBlock(&block, encBlock)
	})
	return
}

func (s *Store) InsertBlock(block types.Block) error {
	return s.db.Update(insertBlock(block))
}

func insertBlock(block types.Block) func(txn *badger.Txn) error {
	return func(txn *badger.Txn) error {
		encBlock, err := encodeBlock(block)
		if err != nil {
			return err
		}
		encBlockHeight, err := encodeUint64(block.Height)
		if err != nil {
			return err
		}

		// get latest block height
		latestBlockHeight, err := getLatestBlockHeightTx(txn)
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			return err
		}

		// insert the block by block height
		if err := txn.Set(blockKey(block.Height), encBlock); err != nil {
			return err
		}
		// add block ID to ID->height lookup
		if err := txn.Set(blockIDIndexKey(block.ID()), encBlockHeight); err != nil {
			return err
		}

		// if this is latest block, set latest block
		if block.Height >= latestBlockHeight {
			return txn.Set(latestBlockKey(), encBlockHeight)
		}

		return nil
	}
}

func (s *Store) CommitBlock(
	block *types.Block,
	collections []*model.LightCollection,
	transactions map[flow.Identifier]*flow.Transaction,
	transactionResults map[flow.Identifier]*types.StorableTransactionResult,
	delta types.LedgerDelta,
	events []flow.Event,
) (err error) {
	if len(transactions) != len(transactionResults) {
		return fmt.Errorf(
			"transactions count (%d) does not match result count (%d)",
			len(transactions),
			len(transactionResults),
		)
	}

	err = s.db.Update(func(txn *badger.Txn) error {
		err := insertBlock(*block)(txn)
		if err != nil {
			return err
		}

		for _, col := range collections {
			err := insertCollection(*col)(txn)
			if err != nil {
				return err
			}
		}

		for txID, tx := range transactions {
			err := insertTransaction(txID, *tx)(txn)
			if err != nil {
				return err
			}
		}

		for txID, result := range transactionResults {
			err := insertTransactionResult(txID, *result)(txn)
			if err != nil {
				return err
			}
		}

		err = s.insertLedgerDelta(block.Height, delta)(txn)
		if err != nil {
			return err
		}

		if events != nil {
			err = insertEvents(block.Height, events)(txn)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func (s *Store) CollectionByID(colID flow.Identifier) (col model.LightCollection, err error) {
	err = s.db.View(func(txn *badger.Txn) error {
		encCol, err := getTx(txn)(collectionKey(colID))
		if err != nil {
			return err
		}
		return decodeCollection(&col, encCol)
	})
	return
}

func (s *Store) InsertCollection(col model.LightCollection) error {
	return s.db.Update(insertCollection(col))
}

func insertCollection(col model.LightCollection) func(txn *badger.Txn) error {
	return func(txn *badger.Txn) error {
		encCol, err := encodeCollection(col)
		if err != nil {
			return err
		}

		return txn.Set(collectionKey(flow.Identifier(col.ID())), encCol)
	}
}

func (s *Store) TransactionByID(txID flow.Identifier) (tx flow.Transaction, err error) {
	err = s.db.View(func(txn *badger.Txn) error {
		encTx, err := getTx(txn)(transactionKey(txID))
		if err != nil {
			return err
		}
		return decodeTransaction(&tx, encTx)
	})
	return
}

func (s *Store) InsertTransaction(tx flow.Transaction) error {
	return s.db.Update(insertTransaction(tx.ID(), tx))
}

func insertTransaction(txID flow.Identifier, tx flow.Transaction) func(txn *badger.Txn) error {
	return func(txn *badger.Txn) error {
		encTx, err := encodeTransaction(tx)
		if err != nil {
			return err
		}

		return txn.Set(transactionKey(txID), encTx)
	}
}

func (s *Store) TransactionResultByID(txID flow.Identifier) (result types.StorableTransactionResult, err error) {
	err = s.db.View(func(txn *badger.Txn) error {
		encResult, err := getTx(txn)(transactionResultKey(txID))
		if err != nil {
			return err
		}
		return decodeTransactionResult(&result, encResult)
	})
	return
}

func (s *Store) InsertTransactionResult(txID flow.Identifier, result types.StorableTransactionResult) error {
	return s.db.Update(insertTransactionResult(txID, result))
}

func insertTransactionResult(txID flow.Identifier, result types.StorableTransactionResult) func(txn *badger.Txn) error {
	return func(txn *badger.Txn) error {
		encResult, err := encodeTransactionResult(result)
		if err != nil {
			return err
		}

		return txn.Set(transactionResultKey(txID), encResult)
	}
}

func (s *Store) LedgerViewByHeight(blockHeight uint64) *types.LedgerView {
	return types.NewLedgerView(func(key string) (value []byte, err error) {
		s.ledgerChangeLog.RLock()
		defer s.ledgerChangeLog.RUnlock()

		lastChangedBlock := s.ledgerChangeLog.getMostRecentChange(key, blockHeight)

		err = s.db.View(func(txn *badger.Txn) error {
			value, err = getTx(txn)(ledgerValueKey(key, lastChangedBlock))
			if err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			// silence not found errors
			if errors.Is(err, storage.ErrNotFound) {
				return nil, nil
			}

			return nil, err
		}

		return value, nil
	})
}

func (s *Store) InsertLedgerDelta(blockHeight uint64, delta types.LedgerDelta) error {
	return s.db.Update(s.insertLedgerDelta(blockHeight, delta))
}

func (s *Store) insertLedgerDelta(blockHeight uint64, delta types.LedgerDelta) func(txn *badger.Txn) error {
	return func(txn *badger.Txn) error {
		s.ledgerChangeLog.Lock()
		defer s.ledgerChangeLog.Unlock()

		for registerID, value := range delta.Updates() {
			if value != nil {
				// if register has an updated value, write it at this block
				err := txn.Set(ledgerValueKey(registerID, blockHeight), value)
				if err != nil {
					return err
				}
			}

			// otherwise register has been deleted, so record change
			// and keep value as nil

			// update the in-memory changelog
			s.ledgerChangeLog.addChange(registerID, blockHeight)

			// encode and write the changelist for the register to disk
			encChangelist, err := encodeChangelist(s.ledgerChangeLog.getChangelist(registerID))
			if err != nil {
				return err
			}

			if err := txn.Set(ledgerChangelogKey(registerID), encChangelist); err != nil {
				return err
			}
		}
		return nil
	}
}

func (s *Store) EventsByHeight(blockHeight uint64, eventType string) (events []flow.Event, err error) {
	// set up an iterator over all events in the block
	iterOpts := badger.DefaultIteratorOptions
	iterOpts.Prefix = eventKeyBlockPrefix(blockHeight)

	eventTypeBytes := []byte(eventType)

	err = s.db.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(iterOpts)
		defer iter.Close()

		// start from lowest possible event key for this block
		startKey := eventKey(blockHeight, 0, 0, "")

		// iteration happens in byte-wise lexicographical sorting order
		for iter.Seek(startKey); iter.Valid(); iter.Next() {
			item := iter.Item()

			// filter by event type if specified
			if eventType != "" {
				if !eventKeyHasType(item.Key(), eventTypeBytes) {
					continue
				}
			}

			err = item.Value(func(b []byte) error {
				var event flow.Event

				err := decodeEvent(&event, b)
				if err != nil {
					return err
				}

				events = append(events, event)

				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	return
}

func (s *Store) InsertEvents(blockHeight uint64, events []flow.Event) error {
	return s.db.Update(insertEvents(blockHeight, events))
}

func insertEvents(blockHeight uint64, events []flow.Event) func(txn *badger.Txn) error {
	return func(txn *badger.Txn) error {
		for _, event := range events {
			b, err := encodeEvent(event)
			if err != nil {
				return err
			}

			key := eventKey(blockHeight, event.TransactionIndex, event.EventIndex, event.Type)

			err = txn.Set(key, b)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

// Close closes the underlying Badger database. It is necessary to close
// a Store before exiting to ensure all writes are persisted to disk.
func (s *Store) Close() error {
	return s.db.Close()
}

// Sync syncs database content to disk.
func (s Store) Sync() error {
	return s.db.Sync()
}

// getTx returns a getter function bound to the input transaction that can be
// used to get values from Badger.
//
// The getter function checks for key-not-found errors and wraps them in
// storage.NotFound in order to comply with the storage.Store interface.
//
// This saves a few lines of converting a badger.Item to []byte.
func getTx(txn *badger.Txn) func([]byte) ([]byte, error) {
	return func(key []byte) ([]byte, error) {
		// Badger returns an "item" upon GETs, we need to copy the actual value
		// from the item and return it.
		item, err := txn.Get(key)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return nil, storage.ErrNotFound
			}
			return nil, err
		}

		val := make([]byte, item.ValueSize())
		return item.ValueCopy(val)
	}
}

// getLatestBlockHeightTx retrieves the latest block height and returns it.
// Must be called from within a Badger transaction.
func getLatestBlockHeightTx(txn *badger.Txn) (uint64, error) {
	encBlockHeight, err := getTx(txn)(latestBlockKey())
	if err != nil {
		return 0, err
	}

	var blockHeight uint64
	if err := decodeUint64(&blockHeight, encBlockHeight); err != nil {
		return 0, err
	}

	return blockHeight, nil
}
