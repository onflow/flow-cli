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
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/tests"
)

func TestBlocks(t *testing.T) {
	t.Parallel()

	t.Run("Get Latest Block", func(t *testing.T) {
		t.Parallel()

		_, s, gw := setup()

		_, _, _, err := s.Blocks.GetBlock("latest", "flow.AccountCreated", false)

		gw.Mock.AssertCalled(t, tests.GetLatestBlockFunc)
		gw.Mock.AssertCalled(t, tests.GetEventsFunc, "flow.AccountCreated", uint64(1), uint64(1))
		gw.Mock.AssertNotCalled(t, tests.GetBlockByHeightFunc)
		gw.Mock.AssertNotCalled(t, tests.GetBlockByIDFunc)
		assert.NoError(t, err)
	})

	t.Run("Get latest block height", func(t *testing.T) {
		t.Parallel()
		_, s, gw := setup()
		height, err := s.Blocks.GetLatestBlockHeight()
		gw.Mock.AssertCalled(t, tests.GetLatestBlockFunc)
		assert.NoError(t, err)
		assert.Equal(t, height, uint64(1))

	})

	t.Run("Get Block by Height", func(t *testing.T) {
		t.Parallel()

		_, s, gw := setup()

		block := tests.NewBlock()
		block.Height = 10
		gw.GetBlockByHeight.Return(block, nil)

		_, _, _, err := s.Blocks.GetBlock("10", "flow.AccountCreated", false)

		gw.Mock.AssertCalled(t, tests.GetBlockByHeightFunc, uint64(10))
		gw.Mock.AssertCalled(t, tests.GetEventsFunc, "flow.AccountCreated", uint64(10), uint64(10))
		gw.Mock.AssertNotCalled(t, tests.GetLatestBlockFunc)
		gw.Mock.AssertNotCalled(t, tests.GetBlockByIDFunc)
		assert.NoError(t, err)
	})

	t.Run("Get Block by ID", func(t *testing.T) {
		t.Parallel()

		_, s, gw := setup()
		ID := "a310685082f0b09f2a148b2e8905f08ea458ed873596b53b200699e8e1f6536f"

		_, _, _, err := s.Blocks.GetBlock(ID, "flow.AccountCreated", false)

		assert.NoError(t, err)
		gw.Mock.AssertCalled(t, tests.GetBlockByIDFunc, flow.HexToID(ID))
		gw.Mock.AssertCalled(t, tests.GetEventsFunc, "flow.AccountCreated", uint64(1), uint64(1))
		gw.Mock.AssertNotCalled(t, tests.GetBlockByHeightFunc)
		gw.Mock.AssertNotCalled(t, tests.GetLatestBlockFunc)
	})

}

func TestBlocksGet_Integration(t *testing.T) {
	t.Parallel()

	t.Run("Get Block", func(t *testing.T) {
		t.Parallel()

		state, s := setupIntegration()
		srvAcc, _ := state.EmulatorServiceAccount()

		block, blockEvents, collection, err := s.Blocks.GetBlock("latest", "", true)

		assert.NoError(t, err)
		assert.Nil(t, blockEvents)
		assert.Equal(t, collection, []*flow.Collection{})
		assert.Equal(t, block.Height, uint64(0))
		assert.Equal(t, block.ID.String(), "7bc42fe85d32ca513769a74f97f7e1a7bad6c9407f0d934c2aa645ef9cf613c7")

		// create an event
		_, _ = s.Accounts.Create(srvAcc, tests.PubKeys(), nil, tests.SigAlgos(), tests.HashAlgos(), nil)

		block, blockEvents, _, err = s.Blocks.GetBlock("latest", "flow.AccountCreated", true)

		assert.NoError(t, err)
		assert.NotNil(t, block)

		assert.Len(t, blockEvents, 1)
		assert.Len(t, blockEvents[0].Events, 1)
	})

	t.Run("Get Block Invalid", func(t *testing.T) {
		t.Parallel()

		_, s := setupIntegration()

		_, _, _, err := s.Blocks.GetBlock("foo", "flow.AccountCreated", true)
		assert.Equal(t, err.Error(), "invalid query: foo, valid are: \"latest\", block height or block ID")
	})
}
