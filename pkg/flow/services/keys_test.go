package services

import (
	"testing"

	"github.com/onflow/flow-cli/pkg/flow/output"

	"github.com/onflow/flow-cli/pkg/flow"

	"github.com/onflow/flow-cli/tests"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
)

func TestKeys(t *testing.T) {
	mock := &tests.MockGateway{}

	project, err := flow.InitProject(crypto.ECDSA_P256, crypto.SHA3_256)
	assert.NoError(t, err)

	keys := NewKeys(mock, project, output.NewStdoutLogger(output.InfoLog))

	t.Run("Generate Keys", func(t *testing.T) {
		key, err := keys.Generate("", "ECDSA_P256")

		assert.NoError(t, err)
		assert.Equal(t, len(key.PrivateKey.String()), 66)
	})

	t.Run("Generate Keys with seed", func(t *testing.T) {
		key, err := keys.Generate("aaaaaaaaaaaaaaaaaaaaaaannndddddd_its_gone", "ECDSA_P256")

		assert.NoError(t, err)
		assert.Equal(t, key.PrivateKey.String(), "0x134f702d0872dba9c7aea15498aab9b2ffedd5aeebfd8ac3cf47c591f0d7ce52")
	})

	t.Run("Fail generate keys, too short seed", func(t *testing.T) {
		_, err := keys.Generate("im_short", "ECDSA_P256")

		assert.Equal(t, err.Error(), "crypto: insufficient seed length 8, must be at least 32 bytes for ECDSA_P256")
	})

	t.Run("Fail generate keys, invalid sig algo", func(t *testing.T) {
		_, err := keys.Generate("", "JUSTNO")

		assert.Equal(t, err.Error(), "invalid signature algorithm: JUSTNO")
	})

}
