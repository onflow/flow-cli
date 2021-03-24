package services

import (
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-cli/tests"
)

func TestScripts(t *testing.T) {
	mock := &tests.MockGateway{}

	proj, err := project.InitProject(crypto.ECDSA_P256, crypto.SHA3_256)
	assert.NoError(t, err)

	scripts := NewScripts(mock, proj, output.NewStdoutLogger(output.InfoLog))

	t.Run("Execute Script", func(t *testing.T) {
		mock.ExecuteScriptMock = func(script []byte, arguments []cadence.Value) (cadence.Value, error) {
			assert.Equal(t, len(string(script)), 69)
			assert.Equal(t, arguments[0].String(), "\"Foo\"")
			return arguments[0], nil
		}

		_, err := scripts.Execute("../../../tests/script.cdc", []string{"String:Foo"}, "")

		assert.NoError(t, err)
	})
}
