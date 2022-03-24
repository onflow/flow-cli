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

package services

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-go-sdk/client"

	"github.com/onflow/flow-cli/tests"
)

func TestEvents(t *testing.T) {
	t.Parallel()

	t.Run("Get Events", func(t *testing.T) {
		t.Parallel()

		_, s, gw := setup()
		_, err := s.Events.Get([]string{"flow.CreateAccount"}, 0, 0, 250, 1)

		assert.NoError(t, err)
		gw.Mock.AssertCalled(t, tests.GetEventsFunc, "flow.CreateAccount", uint64(0), uint64(0))
	})

	t.Run("Should have larger endHeight then startHeight", func(t *testing.T) {
		t.Parallel()

		_, s, _ := setup()
		_, err := s.Events.Get([]string{"flow.CreateAccount"}, 10, 0, 250, 1)
		assert.EqualError(t, err, "cannot have end height (0) of block range less that start height (10)")
	})

	t.Run("Test create queries", func(t *testing.T) {

		names := []string{"first", "second"}
		queries := makeEventQueries(names, 0, 400, 250)
		expected := []client.EventRangeQuery{
			{Type: "first", StartHeight: 0, EndHeight: 249},
			{Type: "second", StartHeight: 0, EndHeight: 249},
			{Type: "first", StartHeight: 250, EndHeight: 400},
			{Type: "second", StartHeight: 250, EndHeight: 400},
		}
		assert.Equal(t, expected, queries)
	})

	t.Run("Should handle error from get events in goroutine", func(t *testing.T) {
		t.Parallel()

		_, s, gw := setup()

		gw.GetEvents.Return([]client.BlockEvents{}, errors.New("failed getting event"))

		_, err := s.Events.Get([]string{"flow.CreateAccount"}, 0, 1, 250, 1)

		assert.EqualError(t, err, "failed getting event")
	})

}

func TestEvents_Integration(t *testing.T) {
	t.Parallel()

	t.Run("Get Events for non existent event", func(t *testing.T) {
		t.Parallel()

		_, s := setupIntegration()

		events, err := s.Events.Get([]string{"nonexisting"}, 0, 0, 250, 1)
		assert.NoError(t, err)
		assert.Len(t, events, 1)
		assert.Len(t, events[0].Events, 0)
	})

	t.Run("Get Events while adding contracts", func(t *testing.T) {
		t.Parallel()

		state, s := setupIntegration()
		srvAcc, _ := state.EmulatorServiceAccount()

		// create events
		_, err := s.Accounts.AddContract(srvAcc, tests.ContractEvents.Name, tests.ContractEvents.Source, false)
		assert.NoError(t, err)
		assert.NoError(t, err)
		for x := 'A'; x <= 'J'; x++ { // test contract emits events named from A to J
			eName := fmt.Sprintf("A.%s.ContractEvents.Event%c", srvAcc.Address().String(), x)
			events, err := s.Events.Get([]string{eName}, 0, 1, 250, 1)
			assert.NoError(t, err)
			assert.Len(t, events, 2)
			assert.Len(t, events[1].Events, 1)

		}
	})

	t.Run("Get Events while adding contracts in parallel", func(t *testing.T) {
		t.Parallel()

		state, s := setupIntegration()
		srvAcc, _ := state.EmulatorServiceAccount()

		// create events
		_, err := s.Accounts.AddContract(srvAcc, tests.ContractEvents.Name, tests.ContractEvents.Source, false)
		assert.NoError(t, err)

		assert.NoError(t, err)
		var eventNames []string
		for x := 'A'; x <= 'J'; x++ { // test contract emits events named from A to J
			eName := fmt.Sprintf("A.%s.ContractEvents.Event%c", srvAcc.Address().String(), x)
			eventNames = append(eventNames, eName)
		}

		events, err := s.Events.Get(eventNames, 0, 1, 250, 5)
		assert.NoError(t, err)
		assert.Len(t, events, 20)
		assert.Len(t, events[1].Events, 1)
	})
}
