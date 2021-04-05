package services

import (
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-cli/tests"
)

func TestEvents(t *testing.T) {
	mock := &tests.MockGateway{}

	proj, err := project.Init(crypto.ECDSA_P256, crypto.SHA3_256)
	assert.NoError(t, err)

	events := NewEvents(mock, proj, output.NewStdoutLogger(output.InfoLog))

	t.Run("Get Events", func(t *testing.T) {
		called := false
		mock.GetEventsMock = func(s string, u uint64, u2 uint64) ([]client.BlockEvents, error) {
			called = true
			return nil, nil
		}

		_, err := events.Get("flow.CreateAccount", "0", "1")

		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("Get Events Latest", func(t *testing.T) {
		called := 0
		mock.GetEventsMock = func(s string, u uint64, u2 uint64) ([]client.BlockEvents, error) {
			called++
			return nil, nil
		}

		mock.GetLatestBlockMock = func() (*flow.Block, error) {
			called++
			return tests.NewBlock(), nil
		}

		_, err := events.Get("flow.CreateAccount", "0", "latest")

		assert.NoError(t, err)
		assert.Equal(t, called, 2)
	})

	t.Run("Fails to get events without name", func(t *testing.T) {
		_, err := events.Get("", "0", "1")
		assert.Equal(t, err.Error(), "cannot use empty string as event name")
	})

	t.Run("Fails to get events with wrong height", func(t *testing.T) {
		_, err := events.Get("test", "-1", "1")
		assert.Equal(t, err.Error(), "failed to parse start height of block range: -1")
	})

	t.Run("Fails to get events with wrong end height", func(t *testing.T) {
		_, err := events.Get("test", "1", "-1")
		assert.Equal(t, err.Error(), "failed to parse end height of block range: -1")
	})

	t.Run("Fails to get events with wrong start height", func(t *testing.T) {
		_, err := events.Get("test", "10", "5")
		assert.Equal(t, err.Error(), "cannot have end height (5) of block range less that start height (10)")
	})
}
