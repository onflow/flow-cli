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
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/test"
)

var accounts = test.AccountGenerator()
var transactions = test.TransactionGenerator()
var transactionResults = test.TransactionResultGenerator()

func NewAccountWithAddress(address string) *flow.Account {
	account := accounts.New()
	account.Address = flow.HexToAddress(address)
	return account
}

func NewTransaction() *flow.Transaction {
	return transactions.New()
}

func NewBlock() *flow.Block {
	return test.BlockGenerator().New()
}

func NewCollection() *flow.Collection {
	return test.CollectionGenerator().New()
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
	res.Error = nil

	return &res
}

func NewAccountCreateResult(address flow.Address) *flow.TransactionResult {
	events := []flow.Event{
		*NewEvent(0,
			"flow.AccountCreated",
			[]cadence.Field{{
				Identifier: "address",
				Type:       cadence.AddressType{},
			}},
			[]cadence.Value{
				cadence.String(address.String()),
			},
		),
	}

	return NewTransactionResult(events)
}
