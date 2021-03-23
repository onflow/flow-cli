package flow_test

import (
	"fmt"
	"testing"

	"github.com/onflow/flow-cli/pkg/flow"

	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/tests"
)

func TestEvent(t *testing.T) {
	flowEvent := tests.NewEvent(0,
		"flow.AccountCreated",
		[]cadence.Field{{
			Identifier: "address",
			Type:       cadence.AddressType{},
		}},
		[]cadence.Value{
			cadence.NewString("00c4fef62310c807"),
		},
	)
	tx := tests.NewTransactionResult([]flowsdk.Event{*flowEvent})
	e := flow.EventsFromTransaction(tx)

	fmt.Println(e.GetAddress())
	fmt.Println(flowEvent.Value.String())
}

func TestAddress(t *testing.T) {
	address := flowsdk.HexToAddress("cdfef0f4f0786e9")
	assert.Equal(t, "0cdfef0f4f0786e9", address.String())
}
