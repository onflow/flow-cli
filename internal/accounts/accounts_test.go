package accounts

import (
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/mocks"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/tests"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

var NoFlags = command.GlobalFlags{}
var NoLogger = output.NewStdoutLogger(output.NoneLog)

func CommandWithState(t *testing.T) (*mocks.MockServices, *flowkit.State, flowkit.ReaderWriter) {
	services := mocks.DefaultMockServices()
	rw, _ := tests.ReaderWriter()
	state, err := flowkit.Init(rw, crypto.ECDSA_P256, crypto.SHA3_256)
	require.NoError(t, err)

	return services, state, rw
}

func Test_AddContract(t *testing.T) {
	srv, state, _ := CommandWithState(t)

	t.Run("Success", func(t *testing.T) {
		srv.AddContract.Run(func(args mock.Arguments) {
			script := args.Get(2).(*flowkit.Script)
			assert.Equal(t, tests.ContractSimpleWithArgs.Filename, script.Location())
			assert.Len(t, script.Args, 1)
			assert.Equal(t, "1", script.Args[0].String())
		})

		args := []string{tests.ContractSimpleWithArgs.Filename, "1"}
		result, err := addContract(args, NoFlags, NoLogger, srv.Mock, state)

		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Success JSON arg", func(t *testing.T) {
		srv.AddContract.Run(func(args mock.Arguments) {
			script := args.Get(2).(*flowkit.Script)
			assert.Equal(t, tests.ContractSimpleWithArgs.Filename, script.Location())
			assert.Len(t, script.Args, 1)
			assert.Equal(t, "1", script.Args[0].String())
		})

		addContractFlags.ArgsJSON = `[{"type": "UInt64", "value": "1"}]`
		args := []string{tests.ContractSimpleWithArgs.Filename}
		result, err := addContract(args, NoFlags, NoLogger, srv.Mock, state)

		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Fail non-existing file", func(t *testing.T) {
		args := []string{"non-existing"}
		result, err := addContract(args, NoFlags, NoLogger, srv.Mock, state)

		assert.Nil(t, result)
		assert.EqualError(t, err, "error loading contract file: open non-existing: file does not exist")
	})

	t.Run("Fail invalid-json", func(t *testing.T) {
		args := []string{tests.ContractA.Filename}
		addContractFlags.ArgsJSON = "invalid"
		result, err := addContract(args, NoFlags, NoLogger, srv.Mock, state)

		assert.Nil(t, result)
		assert.EqualError(t, err, "error parsing transaction arguments: invalid character 'i' looking for beginning of value")
	})

	t.Run("Fail invalid signer", func(t *testing.T) {
		args := []string{tests.ContractA.Filename}
		addContractFlags.Signer = "invalid"
		result, err := addContract(args, NoFlags, NoLogger, srv.Mock, state)

		assert.Nil(t, result)
		assert.EqualError(t, err, "could not find account with name invalid in the configuration")
	})

}

func Test_RemoveContract(t *testing.T) {
	t.Run("Success", func(t *testing.T) {

	})
}
