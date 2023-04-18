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

package flowkit

import (
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

type Event struct {
	Type   string
	Values map[string]cadence.Value
}

func (e *Event) GetAddress() *flow.Address {
	if a, ok := e.Values["address"].(cadence.Address); ok {
		address := flow.HexToAddress(a.String())
		return &address
	}

	return nil
}

type Events []Event

func EventsFromTransaction(tx *flow.TransactionResult) Events {
	var events Events
	for _, event := range tx.Events {
		events = append(events, NewEvent(event))
	}

	return events
}

func NewEvent(event flow.Event) Event {
	var names []string

	for _, eventType := range event.Value.EventType.Fields {
		names = append(names, eventType.Identifier)
	}
	values := make(map[string]cadence.Value)
	for id, field := range event.Value.Fields {
		values[names[id]] = field
	}

	return Event{
		Type:   event.Type,
		Values: values,
	}
}

func (e *Events) GetCreatedAddresses() []*flow.Address {
	addresses := make([]*flow.Address, 0)
	for _, event := range *e {
		if event.Type == flow.EventAccountCreated {
			addresses = append(addresses, event.GetAddress())
		}
	}

	return addresses
}
