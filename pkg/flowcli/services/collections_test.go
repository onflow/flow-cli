package services

import (
	"testing"

	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"

	"github.com/onflow/flow-cli/tests"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
)

func TestCollections(t *testing.T) {
	mock := &tests.MockGateway{}

	proj, err := project.InitProject(crypto.ECDSA_P256, crypto.SHA3_256)
	assert.NoError(t, err)

	collections := NewCollections(mock, proj, output.NewStdoutLogger(output.InfoLog))

	t.Run("Get Collection", func(t *testing.T) {
		called := false
		mock.GetCollectionMock = func(id flowsdk.Identifier) (*flowsdk.Collection, error) {
			called = true
			return tests.NewCollection(), nil
		}

		_, err := collections.Get("a310685082f0b09f2a148b2e8905f08ea458ed873596b53b200699e8e1f6536f")

		assert.NoError(t, err)
		assert.True(t, called)
	})
}
