/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
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
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/stretchr/testify/mock"

	"github.com/onflow/flow-cli/tests/mocks"
)

const (
	GetAccountFunc            = "GetAccount"
	SendSignedTransactionFunc = "SendSignedTransaction"
	GetCollectionFunc         = "GetCollection"
	GetTransactionResultFunc  = "GetTransactionResult"
	GetEventsFunc             = "GetEvents"
	GetLatestBlockFunc        = "GetLatestBlock"
	GetBlockByHeightFunc      = "GetBlockByHeight"
	GetBlockByIDFunc          = "GetBlockByID"
	ExecuteScriptFunc         = "ExecuteScript"
	GetTransactionFunc        = "GetTransaction"
)

type TestGateway struct {
	Mock                  *mocks.Gateway
	SendSignedTransaction *mock.Call
	GetAccount            *mock.Call
	GetCollection         *mock.Call
	GetTransactionResult  *mock.Call
	GetEvents             *mock.Call
	GetLatestBlock        *mock.Call
	GetBlockByHeight      *mock.Call
	GetBlockByID          *mock.Call
	ExecuteScript         *mock.Call
	GetTransaction        *mock.Call
}

func DefaultMockGateway() *TestGateway {
	m := &mocks.Gateway{}
	t := &TestGateway{
		Mock: m,
		SendSignedTransaction: m.On(
			SendSignedTransactionFunc,
			mock.AnythingOfType("*flowkit.Transaction"),
		),
		GetAccount: m.On(
			GetAccountFunc,
			mock.AnythingOfType("flow.Address"),
		),
		GetCollection: m.On(
			GetCollectionFunc,
			mock.AnythingOfType("flow.Identifier"),
		),
		GetTransactionResult: m.On(
			GetTransactionResultFunc,
			mock.AnythingOfType("*flow.Transaction"),
			mock.AnythingOfType("bool"),
		),
		GetTransaction: m.On(
			GetTransactionFunc,
			mock.AnythingOfType("flow.Identifier"),
		),
		GetEvents: m.On(
			GetEventsFunc,
			mock.AnythingOfType("string"),
			mock.AnythingOfType("uint64"),
			mock.AnythingOfType("uint64"),
		),
		ExecuteScript: m.On(
			ExecuteScriptFunc,
			mock.Anything,
			mock.Anything,
		),
		GetBlockByHeight: m.On(GetBlockByHeightFunc, mock.Anything),
		GetBlockByID:     m.On(GetBlockByIDFunc, mock.Anything),
		GetLatestBlock:   m.On(GetLatestBlockFunc),
	}

	// default return values
	t.SendSignedTransaction.Run(func(args mock.Arguments) {
		t.SendSignedTransaction.Return(NewTransaction(), nil)
	})

	t.GetAccount.Run(func(args mock.Arguments) {
		addr := args.Get(0).(flow.Address)
		t.GetAccount.Return(NewAccountWithAddress(addr.String()), nil)
	})

	t.ExecuteScript.Run(func(args mock.Arguments) {
		t.ExecuteScript.Return(cadence.MustConvertValue(""), nil)
	})

	t.GetTransaction.Return(NewTransaction(), nil)
	t.GetCollection.Return(NewCollection(), nil)
	t.GetTransactionResult.Return(NewTransactionResult(nil), nil)
	t.GetEvents.Return([]client.BlockEvents{}, nil)
	t.GetLatestBlock.Return(NewBlock(), nil)
	t.GetBlockByHeight.Return(NewBlock(), nil)
	t.GetBlockByID.Return(NewBlock(), nil)

	return t
}
