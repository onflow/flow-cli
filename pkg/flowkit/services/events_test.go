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
	"testing"

	"github.com/onflow/flow-cli/tests"
	"github.com/stretchr/testify/assert"
)

func TestEvents(t *testing.T) {
	t.Run("Get Events", func(t *testing.T) {
		_, s, gw := setup()
		_, err := s.Events.Get("flow.CreateAccount", "0", "1")

		assert.NoError(t, err)
		gw.Mock.AssertCalled(t, tests.GetEventsFunc, "flow.CreateAccount", uint64(0), uint64(1))
	})

	t.Run("Get Events Latest", func(t *testing.T) {
		_, s, gw := setup()
		_, err := s.Events.Get("flow.CreateAccount", "0", "latest")

		assert.NoError(t, err)
		gw.Mock.AssertCalled(t, tests.GetLatestBlockFunc)
		gw.Mock.AssertCalled(t, tests.GetEventsFunc, "flow.CreateAccount", uint64(0), uint64(1))
	})

	t.Run("Fails to get events without name", func(t *testing.T) {
		_, s, _ := setup()
		_, err := s.Events.Get("", "0", "1")
		assert.Equal(t, err.Error(), "cannot use empty string as event name")
	})

	t.Run("Fails to get events with wrong height", func(t *testing.T) {
		_, s, _ := setup()
		_, err := s.Events.Get("test", "-1", "1")
		assert.Equal(t, err.Error(), "failed to parse start height of block range: -1")
	})

	t.Run("Fails to get events with wrong end height", func(t *testing.T) {
		_, s, _ := setup()
		_, err := s.Events.Get("test", "1", "-1")
		assert.Equal(t, err.Error(), "failed to parse end height of block range: -1")
	})

	t.Run("Fails to get events with wrong start height", func(t *testing.T) {
		_, s, _ := setup()
		_, err := s.Events.Get("test", "10", "5")
		assert.Equal(t, err.Error(), "cannot have end height (5) of block range less that start height (10)")
	})
}
