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

package tests

import (
	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"

	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
)

type TestGateway struct {
	GetAccountMock                func(address flow.Address) (*flow.Account, error)
	SendTransactionMock           func(tx *flowkit.Transaction) (*flow.Transaction, error)
	PrepareTransactionPayloadMock func(tx *flowkit.Transaction) (*flowkit.Transaction, error)
	SendSignedTransactionMock     func(tx *flowkit.Transaction) (*flow.Transaction, error)
	GetTransactionResultMock      func(tx *flow.Transaction) (*flow.TransactionResult, error)
	GetTransactionMock            func(id flow.Identifier) (*flow.Transaction, error)
	ExecuteScriptMock             func(script []byte, arguments []cadence.Value) (cadence.Value, error)
	GetLatestBlockMock            func() (*flow.Block, error)
	GetEventsMock                 func(string, uint64, uint64) ([]client.BlockEvents, error)
	GetCollectionMock             func(id flow.Identifier) (*flow.Collection, error)
	GetBlockByHeightMock          func(uint64) (*flow.Block, error)
	GetBlockByIDMock              func(flow.Identifier) (*flow.Block, error)
	PingMock                      func() error
}

func DefaultMockGateway() gateway.Gateway {
	return &TestGateway{
		SendSignedTransactionMock: func(tx *flowkit.Transaction) (*flow.Transaction, error) {
			return tx.FlowTransaction(), nil
		},
		GetLatestBlockMock: func() (*flow.Block, error) {
			return NewBlock(), nil
		},
		GetAccountMock: func(address flow.Address) (*flow.Account, error) {
			return NewAccountWithAddress(address.String()), nil
		},
	}
}

func (g *TestGateway) GetAccount(address flow.Address) (*flow.Account, error) {
	return g.GetAccountMock(address)
}

func (g *TestGateway) SendSignedTransaction(tx *flowkit.Transaction) (*flow.Transaction, error) {
	return g.SendSignedTransactionMock(tx)
}

func (g *TestGateway) GetTransactionResult(tx *flow.Transaction, waitSeal bool) (*flow.TransactionResult, error) {
	return g.GetTransactionResultMock(tx)
}

func (g *TestGateway) GetTransaction(id flow.Identifier) (*flow.Transaction, error) {
	return g.GetTransactionMock(id)
}

func (g *TestGateway) ExecuteScript(script []byte, arguments []cadence.Value) (cadence.Value, error) {
	return g.ExecuteScriptMock(script, arguments)
}

func (g *TestGateway) GetLatestBlock() (*flow.Block, error) {
	return g.GetLatestBlockMock()
}

func (g *TestGateway) GetBlockByID(id flow.Identifier) (*flow.Block, error) {
	return g.GetBlockByIDMock(id)
}

func (g *TestGateway) GetBlockByHeight(height uint64) (*flow.Block, error) {
	return g.GetBlockByHeightMock(height)
}

func (g *TestGateway) GetEvents(name string, start uint64, end uint64) ([]client.BlockEvents, error) {
	return g.GetEventsMock(name, start, end)
}

func (g *TestGateway) GetCollection(id flow.Identifier) (*flow.Collection, error) {
	return g.GetCollectionMock(id)
}

func (g *TestGateway) Ping() error {
	return nil
}
