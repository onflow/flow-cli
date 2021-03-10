package gateway

import (
	"fmt"

	"github.com/onflow/flow-go-sdk/client"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/flow/cli"
	emulator "github.com/onflow/flow-emulator"
	"github.com/onflow/flow-go-sdk"
)

type EmulatorGateway struct {
	emulator *emulator.Blockchain
}

func NewEmulatorGateway() *EmulatorGateway {
	return &EmulatorGateway{
		emulator: newEmulator(),
	}
}

func newEmulator() *emulator.Blockchain {
	b, err := emulator.NewBlockchain()
	if err != nil {
		panic(err)
	}
	return b
}

func (g *EmulatorGateway) GetAccount(address flow.Address) (*flow.Account, error) {
	return g.emulator.GetAccount(address)
}

func (g *EmulatorGateway) SendTransaction(tx *flow.Transaction, signer *cli.Account) (*flow.Transaction, error) {
	return nil, fmt.Errorf("Not Supported Yet")
}

func (g *EmulatorGateway) GetTransactionResult(tx *flow.Transaction) (*flow.TransactionResult, error) {
	return g.emulator.GetTransactionResult(tx.ID())
}

func (g *EmulatorGateway) GetTransaction(id flow.Identifier) (*flow.Transaction, error) {
	return g.emulator.GetTransaction(id)
}

func (g *EmulatorGateway) ExecuteScript(script []byte, arguments []cadence.Value) (cadence.Value, error) {
	return nil, fmt.Errorf("Not Supported Yet")
}

func (g *EmulatorGateway) GetLatestBlock() (*flow.Block, error) {
	return nil, fmt.Errorf("Not Supported Yet")
}

func (g *EmulatorGateway) GetEvents(string, uint64, uint64) ([]client.BlockEvents, error) {
	return nil, fmt.Errorf("Not Supported Yet")
}

func (g *EmulatorGateway) GetCollection(id flow.Identifier) (*flow.Collection, error) {
	return nil, fmt.Errorf("Not Supported Yet")
}
