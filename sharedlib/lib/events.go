package lib

import (
	"fmt"
	"github.com/onflow/flow-go-sdk"
	"strings"
)

type Event struct {
	Type   string
	Values map[string]string
}

type Events []Event

func NewEventsFromResult(tx *flow.TransactionResult) Events {
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
		values[name] = fmt.Sprintf("%v", field)
	}

	return Event{
		Type:   event.Type,
		Values: values,
	}
}

func (e *Events) GetAddress() *flow.Address {
	addr := ""
	for _, event := range *e {
		if event.Values["address"] != "" {
			addr = event.Values["address"]
		}
	}

	if addr == "" {
		return nil
	}

	address := flow.HexToAddress(
		strings.ReplaceAll(addr, "0x", ""),
	)
	return &address // todo: maybe not
}
