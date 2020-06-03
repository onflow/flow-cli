package entity

import (
	"time"

	"github.com/dapperlabs/flow-go/model/flow"
)

type CompleteCollection struct {
	Guarantee    *flow.CollectionGuarantee
	Transactions []*flow.TransactionBody
}

type ExecutableBlock struct {
	Block               *flow.Block
	CompleteCollections map[flow.Identifier]*CompleteCollection
	StartState          flow.StateCommitment
}

type BlocksByCollection struct {
	CollectionID     flow.Identifier
	ExecutableBlocks map[flow.Identifier]*ExecutableBlock
	TimeoutTimer     *time.Timer
}

func (c *CompleteCollection) Collection() flow.Collection {
	return flow.Collection{Transactions: c.Transactions}
}

func (b *BlocksByCollection) ID() flow.Identifier {
	return b.CollectionID
}

func (b *BlocksByCollection) Checksum() flow.Identifier {
	return b.CollectionID
}

func (b *ExecutableBlock) ID() flow.Identifier {
	return b.Block.ID()
}

func (b *ExecutableBlock) Checksum() flow.Identifier {
	return b.Block.Checksum()
}

func (b *ExecutableBlock) Height() uint64 {
	return b.Block.Header.Height
}

func (b *ExecutableBlock) ParentID() flow.Identifier {
	return b.Block.Header.ParentID
}

func (b *ExecutableBlock) Collections() []*CompleteCollection {
	collections := make([]*CompleteCollection, len(b.Block.Payload.Guarantees))

	for i, cg := range b.Block.Payload.Guarantees {
		collections[i] = b.CompleteCollections[cg.ID()]
	}

	return collections
}

func (b *ExecutableBlock) HasAllTransactions() bool {
	for _, collection := range b.Block.Payload.Guarantees {

		completeCollection, ok := b.CompleteCollections[collection.ID()]
		if ok && completeCollection.Transactions != nil {
			continue
		}
		return false
	}
	return true
}

func (b *ExecutableBlock) IsComplete() bool {
	return b.HasAllTransactions() && len(b.StartState) > 0
}
