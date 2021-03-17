package tests

import (
	"github.com/onflow/flow-cli/flow/gateway"
	"github.com/onflow/flow-cli/flow/lib"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
)

type MockGateway struct {
	GetAccountMock           func(address flow.Address) (*flow.Account, error)
	SendTransactionMock      func(tx *flow.Transaction, signer *lib.Account) (*flow.Transaction, error)
	GetTransactionResultMock func(tx *flow.Transaction) (*flow.TransactionResult, error)
	GetTransactionMock       func(id flow.Identifier) (*flow.Transaction, error)
	ExecuteScriptMock        func(script []byte, arguments []cadence.Value) (cadence.Value, error)
	GetLatestBlockMock       func() (*flow.Block, error)
	GetEventsMock            func(string, uint64, uint64) ([]client.BlockEvents, error)
	GetCollectionMock        func(id flow.Identifier) (*flow.Collection, error)
	GetBlockByHeightMock     func(uint64) (*flow.Block, error)
	GetBlockByIDMock         func(flow.Identifier) (*flow.Block, error)
}

func NewMockGateway() gateway.Gateway {
	return &MockGateway{}
}

func (g *MockGateway) GetAccount(address flow.Address) (*flow.Account, error) {
	return g.GetAccountMock(address)
}

func (g *MockGateway) SendTransaction(tx *flow.Transaction, signer *lib.Account) (*flow.Transaction, error) {
	return g.SendTransactionMock(tx, signer)
}

func (g *MockGateway) GetTransactionResult(tx *flow.Transaction, waitSeal bool) (*flow.TransactionResult, error) {
	return g.GetTransactionResultMock(tx)
}

func (g *MockGateway) GetTransaction(id flow.Identifier) (*flow.Transaction, error) {
	return g.GetTransactionMock(id)
}

func (g *MockGateway) ExecuteScript(script []byte, arguments []cadence.Value) (cadence.Value, error) {
	return g.ExecuteScriptMock(script, arguments)
}

func (g *MockGateway) GetLatestBlock() (*flow.Block, error) {
	return g.GetLatestBlockMock()
}

func (g *MockGateway) GetBlockByID(id flow.Identifier) (*flow.Block, error) {
	return g.GetBlockByIDMock(id)
}

func (g *MockGateway) GetBlockByHeight(height uint64) (*flow.Block, error) {
	return g.GetBlockByHeightMock(height)
}

func (g *MockGateway) GetEvents(name string, start uint64, end uint64) ([]client.BlockEvents, error) {
	return g.GetEventsMock(name, start, end)
}

func (g *MockGateway) GetCollection(id flow.Identifier) (*flow.Collection, error) {
	return g.GetCollectionMock(id)
}
