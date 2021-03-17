package tests

import (
	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/test"
)

var accounts = test.AccountGenerator()
var transactions = test.TransactionGenerator()
var transactionResults = test.TransactionResultGenerator()
var events = test.EventGenerator()
var blocks = test.BlockGenerator()

func NewAccountWithAddress(address string) *flow.Account {
	account := accounts.New()
	account.Address = flow.HexToAddress(address)
	return account
}

func NewTransaction() *flow.Transaction {
	return transactions.New()
}

func NewBlock() *flow.Block {
	return blocks.New()
}

func NewEvent(index int, eventId string, fields []cadence.Field, values []cadence.Value) *flow.Event {
	location := common.StringLocation("test")

	testEventType := &cadence.EventType{
		Location:            location,
		QualifiedIdentifier: eventId,
		Fields:              fields,
	}

	testEvent := cadence.
		NewEvent(values).
		WithType(testEventType)

	typeID := location.TypeID(eventId)

	event := flow.Event{
		Type:             string(typeID),
		TransactionID:    flow.Identifier{},
		TransactionIndex: index,
		EventIndex:       index,
		Value:            testEvent,
	}

	return &event
}

func NewTransactionResult(events []flow.Event) *flow.TransactionResult {
	res := transactionResults.New()
	res.Events = events

	return &res
}

func NewAccountCreateResult(address string) *flow.TransactionResult {
	events := []flow.Event{
		*NewEvent(0,
			"flow.AccountCreated",
			[]cadence.Field{{
				Identifier: "address",
				Type:       cadence.AddressType{},
			}},
			[]cadence.Value{
				cadence.NewString(address),
			},
		),
	}

	return NewTransactionResult(events)
}
