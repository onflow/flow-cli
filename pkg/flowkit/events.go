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
	"bytes"
	"fmt"
	"strings"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
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

// TODO(sideninja):
// - Refactor this to flow.Address and err as return value instead of returning nil.
// - This section should be improved to support all the core events parsing to better Go struct representation and should be extracted to Go SDK

func (e *Events) GetAddress() *flow.Address {
	for _, event := range *e {
		if a, ok := event.Values["address"].(cadence.Address); ok {
			address := flow.HexToAddress(a.String())
			return &address
		}
	}

	return nil
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

func handleCadenceArrayValues(keyArray cadence.Array) []byte {
	parsedKey := make([]byte, len(keyArray.Values))
	for i, val := range keyArray.Values {
		parsedKey[i] = val.ToGoValue().(byte)
	}
	return parsedKey
}

func (e *Events) GetAddressForKeyAdded(publicKey crypto.PublicKey) *flow.Address {
	for _, event := range *e {
		if event.Type == flow.EventAccountKeyAdded {
			// new format
			if keyStruct, ok := event.Values["publicKey"].(cadence.Struct); ok {
				if keyArray, ok := keyStruct.Fields[0].(cadence.Array); ok {
					parsedKey := handleCadenceArrayValues(keyArray)
					if bytes.Equal(parsedKey, publicKey.Encode()) {
						return event.GetAddress()
					}
				}
			}

			// older format support, in previous versions of cadence the Public key was encoded into an array containing other data beside the key
			if p, ok := event.Values["publicKey"].(cadence.Array); ok {
				parsedKey := handleCadenceArrayValues(p)
				parsedKeyhex := fmt.Sprintf("%x", parsedKey)
				publicKeyHex := publicKey.String()
				//we have to remove 0x from beginning of publicKeyHex
				if strings.Contains(parsedKeyhex, publicKeyHex[2:]) {
					return event.GetAddress()
				}
			}
		}
	}

	return nil
}
