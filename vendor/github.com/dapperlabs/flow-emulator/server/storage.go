package server

import (
	"time"

	"github.com/pkg/errors"
	"github.com/psiemens/graceland"
	"github.com/sirupsen/logrus"

	"github.com/dapperlabs/flow-emulator/storage"
	"github.com/dapperlabs/flow-emulator/storage/badger"
	"github.com/dapperlabs/flow-emulator/storage/memstore"
)

type Storage interface {
	graceland.Routine
	Store() storage.Store
}

type MemoryStorage struct {
	store *memstore.Store
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{store: memstore.New()}
}

func (s *MemoryStorage) Start() error {
	return nil
}

func (s *MemoryStorage) Stop() {}

func (s *MemoryStorage) Store() storage.Store {
	return s.store
}

type BadgerStorage struct {
	logger         *logrus.Logger
	store          *badger.Store
	ticker         *time.Ticker
	done           chan bool
	gcInterval     time.Duration
	gcDiscardRatio float64
}

func NewBadgerStorage(
	logger *logrus.Logger,
	dbPath string,
	gcInterval time.Duration,
	gcDiscardRatio float64,
) (*BadgerStorage, error) {
	store, err := badger.New(
		badger.WithPath(dbPath),
		badger.WithLogger(logger),
		badger.WithTruncate(true),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize Badger store")
	}

	return &BadgerStorage{
		logger:         logger,
		store:          store,
		ticker:         time.NewTicker(gcInterval),
		done:           make(chan bool, 1),
		gcInterval:     gcInterval,
		gcDiscardRatio: gcDiscardRatio,
	}, nil
}

func (s *BadgerStorage) Start() error {
	for {
		select {
		case <-s.ticker.C:
			err := s.store.RunValueLogGC(s.gcDiscardRatio)
			if err != nil {
				return errors.Wrap(err, "failed to perform garbage collection on Badger DB")
			}

			s.logger.
				WithFields(logrus.Fields{
					"interval":     s.gcInterval,
					"discardRatio": s.gcDiscardRatio,
				}).
				Debug("Performed garbage collection on Badger value log")
		case <-s.done:
			return s.store.Close()
		}
	}
}

func (s *BadgerStorage) Stop() {
	s.done <- true
}

func (s *BadgerStorage) Store() storage.Store {
	return s.store
}
