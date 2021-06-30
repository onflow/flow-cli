/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/tests"
)

func TestEvents(t *testing.T) {
	t.Parallel()

	t.Run("Get Events", func(t *testing.T) {
		t.Parallel()

		_, s, gw := setup()
		_, err := s.Events.Get([]string{"flow.CreateAccount"}, 0, 0, 250,1)

		assert.NoError(t, err)
		gw.Mock.AssertCalled(t, tests.GetEventsFunc, "flow.CreateAccount", uint64(0), uint64(0))
	})


}

func TestEvents_Integration(t *testing.T) {
	t.Parallel()

	t.Run("Get Events for non existant event", func(t *testing.T) {
		t.Parallel()

		_, s := setupIntegration()

		events, err := s.Events.Get([]string{"nonexisting"}, 0, 1, 250, 1)
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

		for x := 'A'; x <= 'J'; x++ { // test contract emits events named from A to J
			eName := fmt.Sprintf("A.%s.ContractEvents.Event%c", srvAcc.Address().String(), x)
			events, err := s.Events.Get([]string{ eName}, 0, 1, 250, 1)
			assert.NoError(t, err)
			assert.Len(t, events, 2)
			assert.Len(t, events[1].Events, 1)

		}
	})

	t.Run("Parse event start stop", func(t *testing.T) {
		t.Parallel()

		state, s := setupIntegration()
		srvAcc, _ := state.EmulatorServiceAccount()

		// create events
		_, err := s.Accounts.AddContract(srvAcc, tests.ContractEvents.Name, tests.ContractEvents.Source, false)
		assert.NoError(t, err)

		start, end, err := s.Events.CalculateStartEnd(1, 0, 1)
		assert.Equal(t, start, uint64(1))
		assert.Equal(t, end, uint64(1))
		assert.NoError(t, err)


		start, end, err = s.Events.CalculateStartEnd(1, 1, 1)
		assert.Equal(t, start, uint64(1))
		assert.Equal(t, end, uint64(1))
		assert.NoError(t, err)

		start, end, err = s.Events.CalculateStartEnd(0, 0, 1)
		assert.Equal(t, start, uint64(0))
		assert.Equal(t, end, uint64(1))
		assert.NoError(t, err)

		start, end, err = s.Events.CalculateStartEnd(2, 1, 1)
		assert.Error(t, err)
	})
}
