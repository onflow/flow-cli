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

package flowkit_test

import (
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/tests"
)

func Test_AccountCreatedEvent(t *testing.T) {
	address := flow.HexToAddress("00c4fef62310c807")
	flowEvent := tests.NewEvent(0,
		flow.EventAccountCreated,
		[]cadence.Field{{
			Identifier: "address",
			Type:       cadence.AddressType{},
		}},
		[]cadence.Value{cadence.NewAddress(address)},
	)
	tx := tests.NewTransactionResult([]flow.Event{*flowEvent})
	events := flowkit.EventsFromTransaction(tx)

	assert.Len(t, events.GetCreatedAddresses(), 1)
	assert.Equal(t, address, *events.GetCreatedAddresses()[0])

	assert.Len(t, events, 1)
	assert.Equal(t, address, *events[0].GetAddress())
	assert.Equal(t, `flow.AccountCreated(address: 0x00c4fef62310c807)`, flowEvent.Value.String())
}

func TestAddress(t *testing.T) {
	address := flow.HexToAddress("cdfef0f4f0786e9")
	assert.Equal(t, "0cdfef0f4f0786e9", address.String())
}
