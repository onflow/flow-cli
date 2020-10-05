package emulator

import (
	"errors"

	"github.com/onflow/flow-go/access"
	"github.com/onflow/flow-go/fvm"
	flowgo "github.com/onflow/flow-go/model/flow"

	"github.com/dapperlabs/flow-emulator/storage"
)

var _ fvm.Blocks = &blocks{}
var _ access.Blocks = &blocks{}

type blocks struct {
	blockchain *Blockchain
}

func newBlocks(b *Blockchain) *blocks {
	return &blocks{b}
}

func (b *blocks) ByHeight(height uint64) (*flowgo.Block, error) {
	if height == b.blockchain.pendingBlock.Height() {
		return b.blockchain.pendingBlock.Block(), nil
	}

	return b.blockchain.storage.BlockByHeight(height)
}

func (b *blocks) HeaderByID(id flowgo.Identifier) (*flowgo.Header, error) {
	block, err := b.blockchain.storage.BlockByID(id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return block.Header, nil
}

func (b *blocks) FinalizedHeader() (*flowgo.Header, error) {
	block, err := b.blockchain.storage.LatestBlock()
	if err != nil {
		return nil, err
	}

	return block.Header, nil
}
