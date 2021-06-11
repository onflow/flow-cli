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
	"fmt"
	"reflect"
	"runtime"
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
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
	functionsCalled               []interface{}
}

func DefaultMockGateway() *TestGateway {
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
		GetTransactionResultMock: func(tx *flow.Transaction) (*flow.TransactionResult, error) {
			return NewTransactionResult(nil), nil
		},
	}
}

func (g *TestGateway) AssertFuncsCalled(t *testing.T, reset bool, funcs ...interface{}) {
	if len(funcs) > len(g.functionsCalled) {
		t.Error("not all required functions were called")
	}

	called, err := g.funcsCalled(funcs)
	if reset {
		g.functionsCalled = nil
	}

	if !called {
		t.Error(err)
	}
}

func (g *TestGateway) AssertFuncsNotCalled(t *testing.T, reset bool, funcs ...interface{}) {
	called, err := g.funcsCalled(funcs)
	if reset {
		g.functionsCalled = nil
	}

	if called {
		t.Error(err)
	}
}

func (g *TestGateway) funcCall(f interface{}) {
	g.functionsCalled = append(g.functionsCalled, f)
}

func (g *TestGateway) funcsCalled(funcs ...interface{}) (bool, error) {
	for _, f := range funcs {
		for x, c := range g.functionsCalled {
			fp := reflect.ValueOf(f).Pointer()
			if fp == reflect.ValueOf(c).Pointer() {
				break
			} else if x == len(g.functionsCalled)-1 {
				return false, fmt.Errorf(
					"required function %s not called",
					runtime.FuncForPC(fp).Name(),
				)
			}
		}
	}

	return true, nil
}

func (g *TestGateway) GetAccount(address flow.Address) (*flow.Account, error) {
	g.funcCall(g.GetAccount)
	return g.GetAccountMock(address)
}

func (g *TestGateway) SendSignedTransaction(tx *flowkit.Transaction) (*flow.Transaction, error) {
	g.funcCall(g.SendSignedTransaction)
	return g.SendSignedTransactionMock(tx)
}

func (g *TestGateway) GetTransactionResult(tx *flow.Transaction, waitSeal bool) (*flow.TransactionResult, error) {
	g.funcCall(g.GetTransactionResult)
	return g.GetTransactionResultMock(tx)
}

func (g *TestGateway) GetTransaction(id flow.Identifier) (*flow.Transaction, error) {
	g.funcCall(g.GetTransaction)
	return g.GetTransactionMock(id)
}

func (g *TestGateway) ExecuteScript(script []byte, arguments []cadence.Value) (cadence.Value, error) {
	g.funcCall(g.ExecuteScript)
	return g.ExecuteScriptMock(script, arguments)
}

func (g *TestGateway) GetLatestBlock() (*flow.Block, error) {
	g.funcCall(g.GetLatestBlock)
	return g.GetLatestBlockMock()
}

func (g *TestGateway) GetBlockByID(id flow.Identifier) (*flow.Block, error) {
	g.funcCall(g.GetBlockByID)
	return g.GetBlockByIDMock(id)
}

func (g *TestGateway) GetBlockByHeight(height uint64) (*flow.Block, error) {
	g.funcCall(g.GetBlockByHeight)
	return g.GetBlockByHeightMock(height)
}

func (g *TestGateway) GetEvents(name string, start uint64, end uint64) ([]client.BlockEvents, error) {
	g.funcCall(g.GetEvents)
	return g.GetEventsMock(name, start, end)
}

func (g *TestGateway) GetCollection(id flow.Identifier) (*flow.Collection, error) {
	g.funcCall(g.GetCollection)
	return g.GetCollectionMock(id)
}

func (g *TestGateway) Ping() error {
	g.funcCall(g.Ping)
	return nil
}
