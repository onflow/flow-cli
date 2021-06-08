/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gateway

import (
	"fmt"

	"github.com/onflow/cadence"
	emulator "github.com/onflow/flow-emulator"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"

	"github.com/onflow/flow-cli/pkg/flowkit/project"
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

func (g *EmulatorGateway) SendSignedTransaction(tx *flow.Transaction, signer *project.Account) (*flow.Transaction, error) {
	return nil, fmt.Errorf("Not Supported Yet")
}

func (g *EmulatorGateway) GetTransactionResult(tx *flow.Transaction, waitSeal bool) (*flow.TransactionResult, error) {
	return g.emulator.GetTransactionResult(tx.ID())
}

func (g *EmulatorGateway) GetTransaction(id flow.Identifier) (*flow.Transaction, error) {
	return g.emulator.GetTransaction(id)
}

func (g *EmulatorGateway) Ping() error {
	return nil
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

func (g *EmulatorGateway) GetBlockByID(id flow.Identifier) (*flow.Block, error) {
	return nil, fmt.Errorf("Not Supported Yet")
}

func (g *EmulatorGateway) GetBlockByHeight(height uint64) (*flow.Block, error) {
	return nil, fmt.Errorf("Not Supported Yet")
}
