package backend

import (
	flowgo "github.com/onflow/flow-go/model/flow"
	sdk "github.com/onflow/flow-go-sdk"

	"github.com/dapperlabs/flow-emulator/types"
)

// Emulator defines the method set of an emulated blockchain.
type Emulator interface {
	AddTransaction(tx sdk.Transaction) error
	ExecuteNextTransaction() (*types.TransactionResult, error)
	ExecuteBlock() ([]*types.TransactionResult, error)
	CommitBlock() (*flowgo.Block, error)
	ExecuteAndCommitBlock() (*flowgo.Block, []*types.TransactionResult, error)
	GetLatestBlock() (*flowgo.Block, error)
	GetBlockByID(id sdk.Identifier) (*flowgo.Block, error)
	GetBlockByHeight(height uint64) (*flowgo.Block, error)
	GetCollection(colID sdk.Identifier) (*sdk.Collection, error)
	GetTransaction(txID sdk.Identifier) (*sdk.Transaction, error)
	GetTransactionResult(txID sdk.Identifier) (*sdk.TransactionResult, error)
	GetAccount(address sdk.Address) (*sdk.Account, error)
	GetAccountAtBlock(address sdk.Address, blockHeight uint64) (*sdk.Account, error)
	GetEventsByHeight(blockHeight uint64, eventType string) ([]sdk.Event, error)
	ExecuteScript(script []byte, arguments [][]byte) (*types.ScriptResult, error)
	ExecuteScriptAtBlock(script []byte, arguments [][]byte, blockHeight uint64) (*types.ScriptResult, error)
}
