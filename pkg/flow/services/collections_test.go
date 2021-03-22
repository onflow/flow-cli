package services

import (
	"testing"

	flow2 "github.com/onflow/flow-cli/pkg/flow"

	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/pkg/flow/util"
	"github.com/onflow/flow-cli/tests"
	"github.com/onflow/flow-go-sdk/crypto"
)

func TestCollections(t *testing.T) {
	mock := &tests.MockGateway{}

	project, err := flow2.InitProject(crypto.ECDSA_P256, crypto.SHA3_256)
	assert.NoError(t, err)

	collections := NewCollections(mock, project, util.NewStdoutLogger(util.InfoLog))

	t.Run("Get Collection", func(t *testing.T) {
		called := false
		mock.GetCollectionMock = func(id flow.Identifier) (*flow.Collection, error) {
			called = true
			return tests.NewCollection(), nil
		}

		_, err := collections.Get("a310685082f0b09f2a148b2e8905f08ea458ed873596b53b200699e8e1f6536f")

		assert.NoError(t, err)
		assert.True(t, called)
	})
}
