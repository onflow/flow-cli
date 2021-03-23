package tests

import (
	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/pkg/flow"
	"github.com/onflow/flow-cli/pkg/flow/gateway"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
)

type MockGateway struct {
	GetAccountMock           func(address flowsdk.Address) (*flowsdk.Account, error)
	SendTransactionMock      func(tx *flowsdk.Transaction, signer *flow.Account) (*flowsdk.Transaction, error)
	GetTransactionResultMock func(tx *flowsdk.Transaction) (*flowsdk.TransactionResult, error)
	GetTransactionMock       func(id flowsdk.Identifier) (*flowsdk.Transaction, error)
	ExecuteScriptMock        func(script []byte, arguments []cadence.Value) (cadence.Value, error)
	GetLatestBlockMock       func() (*flowsdk.Block, error)
	GetEventsMock            func(string, uint64, uint64) ([]client.BlockEvents, error)
	GetCollectionMock        func(id flowsdk.Identifier) (*flowsdk.Collection, error)
	GetBlockByHeightMock     func(uint64) (*flowsdk.Block, error)
	GetBlockByIDMock         func(flowsdk.Identifier) (*flowsdk.Block, error)
}

func NewMockGateway() gateway.Gateway {
	return &MockGateway{}
}

func (g *MockGateway) GetAccount(address flowsdk.Address) (*flowsdk.Account, error) {
	return g.GetAccountMock(address)
}

func (g *MockGateway) SendTransaction(tx *flowsdk.Transaction, signer *flow.Account) (*flowsdk.Transaction, error) {
	return g.SendTransactionMock(tx, signer)
}

func (g *MockGateway) GetTransactionResult(tx *flowsdk.Transaction, waitSeal bool) (*flowsdk.TransactionResult, error) {
	return g.GetTransactionResultMock(tx)
}

func (g *MockGateway) GetTransaction(id flowsdk.Identifier) (*flowsdk.Transaction, error) {
	return g.GetTransactionMock(id)
}

func (g *MockGateway) ExecuteScript(script []byte, arguments []cadence.Value) (cadence.Value, error) {
	return g.ExecuteScriptMock(script, arguments)
}

func (g *MockGateway) GetLatestBlock() (*flowsdk.Block, error) {
	return g.GetLatestBlockMock()
}

func (g *MockGateway) GetBlockByID(id flowsdk.Identifier) (*flowsdk.Block, error) {
	return g.GetBlockByIDMock(id)
}

func (g *MockGateway) GetBlockByHeight(height uint64) (*flowsdk.Block, error) {
	return g.GetBlockByHeightMock(height)
}

func (g *MockGateway) GetEvents(name string, start uint64, end uint64) ([]client.BlockEvents, error) {
	return g.GetEventsMock(name, start, end)
}

func (g *MockGateway) GetCollection(id flowsdk.Identifier) (*flowsdk.Collection, error) {
	return g.GetCollectionMock(id)
}
