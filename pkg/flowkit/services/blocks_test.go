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

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/tests"
)

func TestBlocks(t *testing.T) {

	mock := tests.DefaultMockGateway()
	readerWriter := tests.ReaderWriter()
	state, err := flowkit.Init(readerWriter, crypto.ECDSA_P256, crypto.SHA3_256)
	assert.NoError(t, err)

	blocks := NewBlocks(mock, state, output.NewStdoutLogger(output.InfoLog))

	t.Run("Get Latest Block", func(t *testing.T) {
		mock.GetEventsMock = func(name string, start uint64, end uint64) ([]client.BlockEvents, error) {
			assert.Equal(t, name, "flow.AccountCreated")
			return nil, nil
		}

		_, _, _, err := blocks.GetBlock("latest", "flow.AccountCreated", false)

		mock.AssertFuncsCalled(t, false, mock.GetLatestBlock)
		mock.AssertFuncsNotCalled(t, true, mock.GetBlockByID, mock.GetBlockByHeight)
		assert.NoError(t, err)
	})

	t.Run("Get Block by Height", func(t *testing.T) {
		mock.GetBlockByHeightMock = func(height uint64) (*flow.Block, error) {
			assert.Equal(t, height, uint64(10))
			return tests.NewBlock(), nil
		}

		mock.GetEventsMock = func(name string, start uint64, end uint64) ([]client.BlockEvents, error) {
			assert.Equal(t, name, "flow.AccountCreated")
			return nil, nil
		}

		_, _, _, err := blocks.GetBlock("10", "flow.AccountCreated", false)

		mock.AssertFuncsCalled(t, false, mock.GetBlockByHeight, mock.GetEvents)
		mock.AssertFuncsNotCalled(t, true, mock.GetBlockByID, mock.GetLatestBlock)
		assert.NoError(t, err)
	})

	t.Run("Get Block by ID", func(t *testing.T) {
		called := false
		mock.GetBlockByIDMock = func(id flow.Identifier) (*flow.Block, error) {
			called = true

			assert.Equal(t, id.String(), "a310685082f0b09f2a148b2e8905f08ea458ed873596b53b200699e8e1f6536f")
			return tests.NewBlock(), nil
		}

		mock.GetEventsMock = func(name string, start uint64, end uint64) ([]client.BlockEvents, error) {
			assert.Equal(t, name, "flow.AccountCreated")
			return nil, nil
		}

		_, _, _, err := blocks.GetBlock("a310685082f0b09f2a148b2e8905f08ea458ed873596b53b200699e8e1f6536f", "flow.AccountCreated", false)

		assert.NoError(t, err)
		assert.True(t, called)
		mock.AssertFuncsNotCalled(t, false, mock.GetBlockByHeight, mock.GetLatestBlock)
		mock.AssertFuncsCalled(t, true, mock.GetBlockByID, mock.GetEvents)
	})
}
