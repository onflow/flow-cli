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
	"strings"

	"github.com/onflow/flow-go-sdk"
)

const addressLength = 16

type Event struct {
	Type   string
	Values map[string]string
}

type Events []Event

func EventsFromTransaction(tx *flow.TransactionResult) Events {
	var events Events
	for _, event := range tx.Events {
		events = append(events, newEvent(event))
	}

	return events
}

func newEvent(event flow.Event) Event {
	var names []string

	for _, eventType := range event.Value.EventType.Fields {
		names = append(names, eventType.Identifier)
	}
	values := map[string]string{}
	for id, field := range event.Value.Fields {
		name := names[id]
		values[name] = field.String()
	}

	return Event{
		Type:   event.Type,
		Values: values,
	}
}

// TODO(sideninja): Refactor this to flow.Address and err as return value instead of returning nil.

func (e *Events) GetAddress() *flow.Address {
	addr := ""
	for _, event := range *e {
		if strings.Contains(event.Type, flow.EventAccountCreated) {
			addr = event.Values["address"]
		}
	}

	if addr == "" {
		return nil
	}

	// add 0 to beginning of address due to them being stripped
	if len(addr) < addressLength {
		addr = strings.Repeat("0", addressLength-len(addr)) + addr
	}

	address := flow.HexToAddress(strings.ReplaceAll(addr, `"`, ""))

	return &address
}
