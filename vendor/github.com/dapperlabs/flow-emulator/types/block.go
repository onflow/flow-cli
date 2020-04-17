package types

import (
	"github.com/dapperlabs/flow-go-sdk"
	"github.com/dapperlabs/flow-go/crypto"
	model "github.com/dapperlabs/flow-go/model/flow"
)

// Block is a naive data structure used to represent blocks in the emulator.
type Block struct {
	Height     uint64
	ParentID   flow.Identifier
	Guarantees []*model.CollectionGuarantee
}

// ID returns the hash of this block.
func (b Block) ID() flow.Identifier {
	hasher := crypto.NewSHA3_256()
	return flow.HashToID(hasher.ComputeHash(b.Encode()))
}

func (b Block) Encode() []byte {
	temp := struct {
		Height        uint64
		ParentID      flow.Identifier
		CollectionIDs []model.Identifier
	}{
		b.Height,
		b.ParentID,
		model.GetIDs(b.Guarantees),
	}

	return flow.DefaultEncoder.MustEncode(&temp)
}

func (b Block) Header() flow.BlockHeader {
	return flow.BlockHeader{
		ID:       b.ID(),
		ParentID: b.ParentID,
		Height:   b.Height,
	}
}

// GenesisBlock returns the genesis block for an emulated blockchain.
func GenesisBlock() Block {
	return Block{
		Height:     0,
		ParentID:   flow.ZeroID,
		Guarantees: []*model.CollectionGuarantee{},
	}
}
