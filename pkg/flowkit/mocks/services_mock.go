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

package mocks

import (
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/mock"

	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway/mocks"
	"github.com/onflow/flow-cli/pkg/flowkit/tests"
)

const (
	addContractFunc                  = "AddContract"
	buildTransactionFunc             = "BuildTransaction"
	createAccountFunc                = "CreateAccount"
	deployProjectFunc                = "DeployProject"
	derivePrivateKeyFromMnemonicFunc = "DerivePrivateKeyFromMnemonic"
	gatewayFunc                      = "Gateway"
	generateKeyFunc                  = "GenerateKey"
	generateMnemonicKeyFunc          = "GenerateMnemonicKey"
	getBlockFunc                     = "GetBlock"
	getTransactionByIDFunc           = "GetTransactionByID"
	getTransactionsByBlockIDFunc     = "GetTransactionsByBlockID"
	networkFunc                      = "Network"
	pingFunc                         = "Ping"
	removeContractFunc               = "RemoveContract"
	sendTransactionFunc              = "SendTransaction"
	setLoggerFunc                    = "SetLogger"
	signTransactionPayloadFunc       = "SignTransactionPayload"
	testFunc                         = "Test"
)

type MockServices struct {
	Mock                         *Services
	AddContract                  *mock.Call
	BuildTransaction             *mock.Call
	CreateAccount                *mock.Call
	DeployProject                *mock.Call
	DerivePrivateKeyFromMnemonic *mock.Call
	Gateway                      *mock.Call
	GenerateKey                  *mock.Call
	GenerateMnemonicKey          *mock.Call
	GetBlock                     *mock.Call
	GetTransactionByID           *mock.Call
	GetTransactionsByBlockID     *mock.Call
	Network                      *mock.Call
	Ping                         *mock.Call
	RemoveContract               *mock.Call
	SendTransaction              *mock.Call
	SetLogger                    *mock.Call
	SignTransactionPayload       *mock.Call
	Test                         *mock.Call
	GetAccount                   *mock.Call
	ExecuteScript                *mock.Call
	SendSignedTransaction        *mock.Call
	GetEvents                    *mock.Call
	GetCollection                *mock.Call
}

func DefaultMockServices() *MockServices {
	m := &Services{}
	t := &MockServices{
		Mock: m,
		GetAccount: m.On(
			mocks.GetAccountFunc,
			mock.Anything,
			mock.AnythingOfType("flow.Address"),
		),
		ExecuteScript: m.On(
			mocks.ExecuteScriptFunc,
			mock.Anything,
			mock.AnythingOfType("flowkit.Script"),
			mock.AnythingOfType("flowkit.ScriptQuery"),
		),
		SendSignedTransaction: m.On(
			mocks.SendSignedTransactionFunc,
			mock.Anything,
			mock.AnythingOfType("*flowkit.Transaction"),
		),
		AddContract: m.On(
			addContractFunc,
			mock.Anything,
			mock.AnythingOfType("*flowkit.Account"),
			mock.AnythingOfType("flowkit.Script"),
			mock.AnythingOfType("flowkit.UpdateContract"),
		),
		GetCollection: m.On(
			mocks.GetCollectionFunc,
			mock.Anything,
			mock.AnythingOfType("flow.Identifier"),
		),
		GetEvents: m.On(
			mocks.GetEventsFunc,
			mock.Anything,
			mock.AnythingOfType("[]string"),
			mock.AnythingOfType("uint64"),
			mock.AnythingOfType("uint64"),
			mock.AnythingOfType("*flowkit.EventWorker"),
		),
		BuildTransaction: m.On(
			buildTransactionFunc,
			mock.Anything,
			mock.AnythingOfType("flowkit.TransactionAddressesRoles"),
			mock.AnythingOfType("int"),
			mock.AnythingOfType("flowkit.Script"),
			mock.AnythingOfType("uint64"),
		),
		CreateAccount: m.On(
			createAccountFunc,
			mock.Anything,
			mock.AnythingOfType("*flowkit.Account"),
			mock.AnythingOfType("[]flowkit.AccountPublicKey"),
		),
		DeployProject: m.On(
			deployProjectFunc,
			mock.Anything,
			mock.AnythingOfType("flowkit.UpdateContract"),
		),
		DerivePrivateKeyFromMnemonic: m.On(
			derivePrivateKeyFromMnemonicFunc,
			mock.Anything,
			mock.AnythingOfType("string"),
			mock.AnythingOfType("crypto.SignatureAlgorithm"),
			mock.AnythingOfType("string"),
		),
		Gateway: m.On(gatewayFunc),
		GenerateKey: m.On(
			generateKeyFunc,
			mock.Anything,
			mock.AnythingOfType("crypto.SignatureAlgorithm"),
			mock.AnythingOfType("string"),
		),
		GenerateMnemonicKey: m.On(
			generateMnemonicKeyFunc,
			mock.Anything,
			mock.AnythingOfType("crypto.SignatureAlgorithm"),
			mock.AnythingOfType("string"),
		),
		GetBlock: m.On(
			getBlockFunc,
			mock.Anything,
			mock.AnythingOfType("flowkit.BlockQuery"),
		),
		GetTransactionByID: m.On(
			getTransactionByIDFunc,
			mock.Anything,
			mock.AnythingOfType("flow.Identifier"),
			mock.AnythingOfType("bool"),
		),
		GetTransactionsByBlockID: m.On(
			getTransactionsByBlockIDFunc,
			mock.Anything,
			mock.AnythingOfType("flow.Identifier"),
		),
		RemoveContract: m.On(
			removeContractFunc,
			mock.Anything,
			mock.AnythingOfType("*flowkit.Account"),
			mock.AnythingOfType("string"),
		),
		SendTransaction: m.On(
			sendTransactionFunc,
			mock.Anything,
			mock.AnythingOfType("flowkit.TransactionAccountRoles"),
			mock.AnythingOfType("flowkit.Script"),
			mock.AnythingOfType("uint64"),
		),
		SignTransactionPayload: m.On(
			signTransactionPayloadFunc,
			mock.Anything,
			mock.AnythingOfType("*flowkit.Account"),
			mock.AnythingOfType("[]uint8"),
		),
		Test: m.On(
			testFunc,
			mock.Anything,
			mock.AnythingOfType("[]byte"),
			mock.AnythingOfType("string"),
		),
		Network:   m.On(networkFunc),
		Ping:      m.On(pingFunc),
		SetLogger: m.On(setLoggerFunc, mock.AnythingOfType("output.Logger")),
	}

	t.GetAccount.Run(func(args mock.Arguments) {
		addr := args.Get(1).(flow.Address)
		t.GetAccount.Return(tests.NewAccountWithAddress(addr.String()), nil)
	})

	t.ExecuteScript.Run(func(args mock.Arguments) {
		t.ExecuteScript.Return(cadence.MustConvertValue(""), nil)
	})

	t.GetTransactionByID.Return(tests.NewTransaction(), nil)
	t.GetCollection.Return(tests.NewCollection(), nil)
	t.GetEvents.Return([]flow.BlockEvents{}, nil)
	t.GetBlock.Return(tests.NewBlock(), nil)
	t.AddContract.Return(flow.EmptyID, false, nil)
	t.RemoveContract.Return(flow.EmptyID, nil)
	t.CreateAccount.Return(tests.NewAccountWithAddress("0x01"), flow.EmptyID, nil)
	t.Network.Return(config.EmulatorNetwork)

	return t
}
