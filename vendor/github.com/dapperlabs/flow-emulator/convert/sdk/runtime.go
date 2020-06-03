package sdk

import (
	"github.com/onflow/cadence"
	sdk "github.com/onflow/flow-go-sdk"
)

func RuntimeEventToSDK(runtimeEvent cadence.Event, txID sdk.Identifier, txIndex int, eventIndex int) sdk.Event {
	return sdk.Event{
		Type:             runtimeEvent.EventType.ID(),
		TransactionID:    txID,
		TransactionIndex: txIndex,
		EventIndex:       eventIndex,
		Value:            runtimeEvent,
	}
}

func RuntimeEventsToSDK(runtimeEvents []cadence.Event, txID sdk.Identifier, txIndex int) []sdk.Event {
	ret := make([]sdk.Event, len(runtimeEvents))
	for i, runtimeEvent := range runtimeEvents {
		ret[i] = RuntimeEventToSDK(runtimeEvent, txID, txIndex, i)
	}
	return ret
}
