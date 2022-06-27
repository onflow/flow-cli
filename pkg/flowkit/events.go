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
	"encoding/json"
	"fmt"

	"github.com/onflow/cadence"

	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flow-go-sdk"
)

const addressLength = 16

type Event struct {
	Type   string
	Values map[string]cadence.Value
}

type Events []Event

func EventsFromTransaction(tx *flow.TransactionResult) Events {
	var events Events
	for _, event := range tx.Events {
		events = append(events, NewEvent(event))
	}

	return events
}

func NewEvents(events []flow.Event) Events {
	flowkitEvents := make(Events, len(events))
	for i, e := range events {
		flowkitEvents[i] = NewEvent(e)
	}
	return flowkitEvents
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

// TODO(sideninja): Refactor this to flow.Address and err as return value instead of returning nil.

func (e *Events) GetAddress() *flow.Address {
	for _, event := range *e {
		if event.Type == flow.EventAccountCreated {
			if a, ok := event.Values["address"].(cadence.Address); ok {
				address := flow.HexToAddress(a.String())
				return &address
			}
		}
	}

	return nil
}

func (e *Events) GetAddressForKeyAdded(publicKey crypto.PublicKey) *flow.Address {
	for _, event := range *e {
		if event.Type == flow.EventAccountAdded {
			if p, ok := event.Values["publicKey"].(cadence.Array); ok { // todo this is older format, support also new format and potentialy move to go sdk
				var parsedKey []byte
				_ = json.Unmarshal([]byte(p.String()), &parsedKey)

				if publicKey.String() == fmt.Sprintf("0x%x", parsedKey[4:len(parsedKey)-5]) {
					return e.GetAddress()
				}
			}
		}
	}

	return nil
}
