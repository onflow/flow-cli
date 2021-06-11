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

	"github.com/spf13/afero"

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/tests"
)

func TestBlocks(t *testing.T) {

	mock := &tests.TestGateway{}
	af := afero.Afero{afero.NewMemMapFs()}

	project, err := flowkit.Init(af, crypto.ECDSA_P256, crypto.SHA3_256)
	assert.NoError(t, err)

	blocks := NewBlocks(mock, project, output.NewStdoutLogger(output.InfoLog))

	t.Run("Get Latest Block", func(t *testing.T) {
		called := false
		mock.GetLatestBlockMock = func() (*flow.Block, error) {
			called = true
			return tests.NewBlock(), nil
		}

		mock.GetBlockByIDMock = func(identifier flow.Identifier) (*flow.Block, error) {
			assert.Fail(t, "shouldn't be called")
			return nil, nil
		}

		mock.GetBlockByHeightMock = func(height uint64) (*flow.Block, error) {
			assert.Fail(t, "shouldn't be called")
			return nil, nil
		}

		mock.GetEventsMock = func(name string, start uint64, end uint64) ([]client.BlockEvents, error) {
			assert.Equal(t, name, "flow.AccountCreated")
			return nil, nil
		}

		_, _, _, err := blocks.GetBlock("latest", "flow.AccountCreated", false)

		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("Get Block by Height", func(t *testing.T) {
		called := false
		mock.GetBlockByHeightMock = func(height uint64) (*flow.Block, error) {
			called = true
			assert.Equal(t, height, uint64(10))
			return tests.NewBlock(), nil
		}

		mock.GetBlockByIDMock = func(identifier flow.Identifier) (*flow.Block, error) {
			assert.Fail(t, "shouldn't be called")
			return nil, nil
		}

		mock.GetLatestBlockMock = func() (*flow.Block, error) {
			assert.Fail(t, "shouldn't be called")
			return nil, nil
		}

		mock.GetEventsMock = func(name string, start uint64, end uint64) ([]client.BlockEvents, error) {
			assert.Equal(t, name, "flow.AccountCreated")
			return nil, nil
		}

		_, _, _, err := blocks.GetBlock("10", "flow.AccountCreated", false)

		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("Get Block by ID", func(t *testing.T) {
		called := false
		mock.GetBlockByIDMock = func(id flow.Identifier) (*flow.Block, error) {
			called = true

			assert.Equal(t, id.String(), "a310685082f0b09f2a148b2e8905f08ea458ed873596b53b200699e8e1f6536f")
			return tests.NewBlock(), nil
		}

		mock.GetBlockByHeightMock = func(u uint64) (*flow.Block, error) {
			assert.Fail(t, "shouldn't be called")
			return nil, nil
		}

		mock.GetLatestBlockMock = func() (*flow.Block, error) {
			assert.Fail(t, "shouldn't be called")
			return nil, nil
		}

		mock.GetEventsMock = func(name string, start uint64, end uint64) ([]client.BlockEvents, error) {
			assert.Equal(t, name, "flow.AccountCreated")
			return nil, nil
		}

		_, _, _, err := blocks.GetBlock("a310685082f0b09f2a148b2e8905f08ea458ed873596b53b200699e8e1f6536f", "flow.AccountCreated", false)

		assert.NoError(t, err)
		assert.True(t, called)
	})
}
