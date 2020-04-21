package convert

import (
	model "github.com/dapperlabs/flow-go/model/flow"
	"github.com/onflow/flow/protobuf/go/flow/entities"

	"github.com/dapperlabs/flow-emulator/types"
)

func BlockToMessage(b types.Block) *entities.Block {
	return &entities.Block{
		Id:                   b.ID().Bytes(),
		ParentId:             b.ParentID.Bytes(),
		Height:               b.Height,
		CollectionGuarantees: CollectionGuaranteesToMessages(b.Guarantees),
	}
}

func CollectionGuaranteeToMessage(g model.CollectionGuarantee) *entities.CollectionGuarantee {
	return &entities.CollectionGuarantee{
		CollectionId: g.CollectionID[:],
	}
}

func CollectionGuaranteesToMessages(l []*model.CollectionGuarantee) []*entities.CollectionGuarantee {
	results := make([]*entities.CollectionGuarantee, len(l))

	for i, item := range l {
		results[i] = CollectionGuaranteeToMessage(*item)
	}

	return results
}

func MessageToCollection(m *entities.Collection) model.LightCollection {
	transactionIDs := make([]model.Identifier, len(m.GetTransactionIds()))
	for i, transactionID := range m.GetTransactionIds() {
		transactionIDs[i] = model.HashToID(transactionID)
	}

	return model.LightCollection{
		Transactions: transactionIDs,
	}
}

func CollectionToMessage(c model.LightCollection) *entities.Collection {
	transactionIDs := make([][]byte, len(c.Transactions))

	for i, transactionID := range c.Transactions {
		transactionIDs[i] = transactionID[:]
	}

	return &entities.Collection{
		TransactionIds: transactionIDs,
	}
}
