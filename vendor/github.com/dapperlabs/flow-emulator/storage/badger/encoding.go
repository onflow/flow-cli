package badger

import (
	"fmt"

	flowgo "github.com/onflow/flow-go/model/flow"
	"github.com/fxamacker/cbor/v2"

	"github.com/dapperlabs/flow-emulator/types"
)

var em cbor.EncMode

func init() {
	opts := cbor.CanonicalEncOptions()
	opts.Time = cbor.TimeRFC3339Nano
	var err error
	em, err = opts.EncMode()
	if err != nil {
		panic(fmt.Sprintf("could not initialize cbor encoding mode: %s", err.Error()))
	}
}

func encodeBlock(block flowgo.Block) ([]byte, error) {
	return em.Marshal(block)
}

func decodeBlock(block *flowgo.Block, from []byte) error {
	return cbor.Unmarshal(from, block)
}

func encodeCollection(col flowgo.LightCollection) ([]byte, error) {
	return em.Marshal(col)
}

func decodeCollection(col *flowgo.LightCollection, from []byte) error {
	return cbor.Unmarshal(from, col)
}

func encodeTransaction(tx flowgo.TransactionBody) ([]byte, error) {
	return em.Marshal(tx)
}

func decodeTransaction(tx *flowgo.TransactionBody, from []byte) error {
	return cbor.Unmarshal(from, tx)
}

func encodeTransactionResult(result types.StorableTransactionResult) ([]byte, error) {
	return em.Marshal(result)
}

func decodeTransactionResult(result *types.StorableTransactionResult, from []byte) error {
	return cbor.Unmarshal(from, result)
}

func encodeUint64(v uint64) ([]byte, error) {
	return em.Marshal(v)
}

func decodeUint64(v *uint64, from []byte) error {
	return cbor.Unmarshal(from, v)
}

func encodeEvent(event flowgo.Event) ([]byte, error) {
	return em.Marshal(event)
}

func decodeEvent(event *flowgo.Event, from []byte) error {
	return cbor.Unmarshal(from, event)
}

func encodeChangelist(clist changelist) ([]byte, error) {
	return em.Marshal(clist.blocks)
}

func decodeChangelist(clist *changelist, from []byte) error {
	return cbor.Unmarshal(from, &clist.blocks)
}
