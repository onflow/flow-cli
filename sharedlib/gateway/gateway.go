package gateway

import (
	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-go-sdk"
)

type Gateway interface {
	GetAccount(flow.Address) (*flow.Account, error)
	SendTransaction(*flow.Transaction, *cli.Account) (*flow.Transaction, error)
	GetTransactionResult(*flow.Transaction) (*flow.TransactionResult, error)
	ExecuteScript([]byte, []cadence.Value) (cadence.Value, error)
}
