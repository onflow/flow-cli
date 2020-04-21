package badger

import (
	"bytes"
	"fmt"

	"github.com/onflow/flow-go-sdk"
)

const (
	blockKeyPrefix             = "block_by_height"
	blockIDIndexKeyPrefix      = "block_id_to_height"
	collectionKeyPrefix        = "collection_by_id"
	transactionKeyPrefix       = "transaction_by_id"
	transactionResultKeyPrefix = "transaction_result_by_id"
	ledgerKeyPrefix            = "ledger_by_block_height" // TODO remove
	eventKeyPrefix             = "event_by_block_height"
	ledgerChangelogKeyPrefix   = "ledger_changelog_by_register_id"
	ledgerValueKeyPrefix       = "ledger_value_by_block_height_register_id"
)

// The following *Key functions return keys to use when reading/writing values
// to Badger. The key name describes how the value is indexed (eg. by block
// height or by ID).
//
// Keys for which numeric ordering is defined, (eg. block height), have the
// numeric component of the key left-padded with zeros (%032d) so that
// lexicographic ordering matches numeric ordering.

func latestBlockKey() []byte {
	return []byte("latest_block_height")
}

func blockKey(blockHeight uint64) []byte {
	return []byte(fmt.Sprintf("%s-%032d", blockKeyPrefix, blockHeight))
}

func blockIDIndexKey(blockID flow.Identifier) []byte {
	return []byte(fmt.Sprintf("%s-%x", blockIDIndexKeyPrefix, blockID))
}

func collectionKey(colID flow.Identifier) []byte {
	return []byte(fmt.Sprintf("%s-%x", collectionKeyPrefix, colID))
}

func transactionKey(txID flow.Identifier) []byte {
	return []byte(fmt.Sprintf("%s-%x", transactionKeyPrefix, txID))
}

func transactionResultKey(txID flow.Identifier) []byte {
	return []byte(fmt.Sprintf("%s-%x", transactionResultKeyPrefix, txID))
}

func eventKey(blockHeight uint64, txIndex, eventIndex int, eventType string) []byte {
	return []byte(fmt.Sprintf(
		"%s-%032d-%032d-%032d-%s",
		eventKeyPrefix,
		blockHeight,
		txIndex,
		eventIndex,
		eventType,
	))
}

func eventKeyBlockPrefix(blockHeight uint64) []byte {
	return []byte(fmt.Sprintf(
		"%s-%032d",
		eventKeyPrefix,
		blockHeight,
	))
}

func eventKeyHasType(key []byte, eventType []byte) bool {
	// event type is at the end of the key, so we can simply compare suffixes
	return bytes.HasSuffix(key, eventType)
}

// TODO remove this
func ledgerKey(blockHeight uint64) []byte {
	return []byte(fmt.Sprintf("%s-%032d", ledgerKeyPrefix, blockHeight))
}

func ledgerChangelogKey(registerID string) []byte {
	return []byte(fmt.Sprintf("%s-%s", ledgerChangelogKeyPrefix, registerID))
}

func ledgerValueKey(registerID string, blockHeight uint64) []byte {
	return []byte(fmt.Sprintf("%s-%s-%032d", ledgerValueKeyPrefix, registerID, blockHeight))
}

// registerIDFromLedgerChangelogKey recovers the register ID from a ledger
// changelog key.
func registerIDFromLedgerChangelogKey(key []byte) string {
	var registerID string
	_, _ = fmt.Sscanf(string(key), ledgerChangelogKeyPrefix+"-%s", &registerID)
	return registerID
}
