package gateway

import (
	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
)

type Gateway interface {
	GetAccount(flow.Address) (*flow.Account, error)
	SendTransaction(*flow.Transaction, *cli.Account) (*flow.Transaction, error)
	GetTransactionResult(*flow.Transaction) (*flow.TransactionResult, error)
	GetTransaction(flow.Identifier) (*flow.Transaction, error)
	ExecuteScript([]byte, []cadence.Value) (cadence.Value, error)
	GetLatestBlock() (*flow.Block, error)
	GetEvents(string, uint64, uint64) ([]client.BlockEvents, error)
}
