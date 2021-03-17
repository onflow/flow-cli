package services

import (
	"testing"

	flow2 "github.com/onflow/flow-cli/pkg/flow"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-cli/pkg/flow/util"
	"github.com/onflow/flow-cli/tests"
)

func TestBlocks(t *testing.T) {

	mock := &tests.MockGateway{}

	project, err := flow2.InitProject(crypto.ECDSA_P256, crypto.SHA3_256)
	require.NoError(t, err)

	blocks := NewBlocks(mock, project, util.NewStdoutLogger(util.InfoLog))

	t.Run("Get Latest Block", func(t *testing.T) {
		called := false
		mock.GetLatestBlockMock = func() (*flow.Block, error) {
			called = true
			return tests.NewBlock(), nil
		}

		mock.GetEventsMock = func(name string, start uint64, end uint64) ([]client.BlockEvents, error) {
			assert.Equal(t, name, "flow.AccountCreated")
			return nil, nil
		}

		_, _, err := blocks.GetBlock("latest", "flow.AccountCreated")

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

		mock.GetEventsMock = func(name string, start uint64, end uint64) ([]client.BlockEvents, error) {
			assert.Equal(t, name, "flow.AccountCreated")
			return nil, nil
		}

		_, _, err := blocks.GetBlock("10", "flow.AccountCreated")

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

		mock.GetEventsMock = func(name string, start uint64, end uint64) ([]client.BlockEvents, error) {
			assert.Equal(t, name, "flow.AccountCreated")
			return nil, nil
		}

		_, _, err := blocks.GetBlock("a310685082f0b09f2a148b2e8905f08ea458ed873596b53b200699e8e1f6536f", "flow.AccountCreated")

		assert.NoError(t, err)
		assert.True(t, called)
	})
}
