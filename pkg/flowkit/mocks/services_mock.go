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
	AddContractFunc                  = "AddContract"
	BuildTransactionFunc             = "BuildTransaction"
	CreateAccountFunc                = "CreateAccount"
	DeployProjectFunc                = "DeployProject"
	DerivePrivateKeyFromMnemonicFunc = "DerivePrivateKeyFromMnemonic"
	GatewayFunc                      = "Gateway"
	GenerateKeyFunc                  = "GenerateKey"
	GenerateMnemonicKeyFunc          = "GenerateMnemonicKey"
	GetBlockFunc                     = "GetBlock"
	GetTransactionByIDFunc           = "GetTransactionByID"
	GetTransactionsByBlockIDFunc     = "GetTransactionsByBlockID"
	NetworkFunc                      = "Network"
	PingFunc                         = "Ping"
	RemoveContractFunc               = "RemoveContract"
	SendTransactionFunc              = "SendTransaction"
	SetLoggerFunc                    = "SetLogger"
	SignTransactionPayloadFunc       = "SignTransactionPayload"
	TestFunc                         = "Test"
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
			AddContractFunc,
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
			BuildTransactionFunc,
			mock.Anything,
			mock.AnythingOfType("*flowkit.TransactionAddressesRoles"),
			mock.AnythingOfType("int"),
			mock.AnythingOfType("*flowkit.Script"),
			mock.AnythingOfType("uint64"),
		),
		CreateAccount: m.On(
			CreateAccountFunc,
			mock.Anything,
			mock.AnythingOfType("*flowkit.Account"),
			mock.AnythingOfType("[]flowkit.AccountPublicKey"),
		),
		DeployProject: m.On(
			DeployProjectFunc,
			mock.Anything,
			mock.AnythingOfType("flowkit.UpdateContract"),
		),
		DerivePrivateKeyFromMnemonic: m.On(
			DerivePrivateKeyFromMnemonicFunc,
			mock.Anything,
			mock.AnythingOfType("string"),
			mock.AnythingOfType("crypto.SignatureAlgorithm"),
			mock.AnythingOfType("string"),
		),
		Gateway: m.On(GatewayFunc),
		GenerateKey: m.On(
			GenerateKeyFunc,
			mock.Anything,
			mock.AnythingOfType("crypto.SignatureAlgorithm"),
			mock.AnythingOfType("string"),
		),
		GenerateMnemonicKey: m.On(
			GenerateMnemonicKeyFunc,
			mock.Anything,
			mock.AnythingOfType("crypto.SignatureAlgorithm"),
			mock.AnythingOfType("string"),
		),
		GetBlock: m.On(
			GetBlockFunc,
			mock.Anything,
			mock.AnythingOfType("flowkit.BlockQuery"),
		),
		GetTransactionByID: m.On(
			GetTransactionByIDFunc,
			mock.Anything,
			mock.AnythingOfType("flow.Identifier"),
			mock.AnythingOfType("bool"),
		),
		GetTransactionsByBlockID: m.On(
			GetTransactionsByBlockIDFunc,
			mock.Anything,
			mock.AnythingOfType("flow.Identifier"),
		),
		RemoveContract: m.On(
			RemoveContractFunc,
			mock.Anything,
			mock.AnythingOfType("*flowkit.Account"),
			mock.AnythingOfType("string"),
		),
		SendTransaction: m.On(
			SendTransactionFunc,
			mock.Anything,
			mock.AnythingOfType("*flowkit.TransactionAccountRoles"),
			mock.AnythingOfType("*flowkit.Script"),
			mock.AnythingOfType("uint64"),
		),
		SignTransactionPayload: m.On(
			SignTransactionPayloadFunc,
			mock.Anything,
			mock.AnythingOfType("*flowkit.Account"),
			mock.AnythingOfType("[]uint8"),
		),
		Test: m.On(
			TestFunc,
			mock.Anything,
			mock.AnythingOfType("[]byte"),
			mock.AnythingOfType("string"),
		),
		Network:   m.On(NetworkFunc),
		Ping:      m.On(PingFunc),
		SetLogger: m.On(SetLoggerFunc, mock.AnythingOfType("output.Logger")),
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
