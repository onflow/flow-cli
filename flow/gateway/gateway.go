package gateway

import (
	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/flow/lib"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
)

type Gateway interface {
	GetAccount(flow.Address) (*flow.Account, error)
	SendTransaction(*flow.Transaction, *lib.Account) (*flow.Transaction, error)
	GetTransactionResult(*flow.Transaction) (*flow.TransactionResult, error)
	GetTransaction(flow.Identifier) (*flow.Transaction, error)
	ExecuteScript([]byte, []cadence.Value) (cadence.Value, error)
	GetLatestBlock() (*flow.Block, error)
	GetBlockByHeight(uint64) (*flow.Block, error)
	GetBlockByID(flow.Identifier) (*flow.Block, error)
	GetEvents(string, uint64, uint64) ([]client.BlockEvents, error)
	GetCollection(flow.Identifier) (*flow.Collection, error)
}
