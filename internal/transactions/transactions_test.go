package transactions

import (
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func Test_Build(t *testing.T) {
	const serviceAccountAddress = "f8d6e0586b0a20c7"
	srv, state, _ := util.TestMocks(t)

	t.Run("Success", func(t *testing.T) {
		inArgs := []string{tests.TransactionSimple.Filename}

		srv.BuildTransaction.Run(func(args mock.Arguments) {
			roles := args.Get(1).(*flowkit.TransactionAddressesRoles)
			assert.Equal(t, serviceAccountAddress, roles.Payer.String())
			assert.Equal(t, serviceAccountAddress, roles.Proposer.String())
			assert.Equal(t, serviceAccountAddress, roles.Authorizers[0].String())
			assert.Equal(t, 0, args.Get(2).(int))
			script := args.Get(3).(*flowkit.Script)
			assert.Equal(t, tests.TransactionSimple.Filename, script.Location())
		}).Return(flowkit.NewTransaction(), nil)

		result, err := build(inArgs, command.GlobalFlags{Yes: true}, util.NoLogger, srv.Mock, state)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Fail not approved", func(t *testing.T) {
		inArgs := []string{tests.TransactionSimple.Filename}
		srv.BuildTransaction.Return(flowkit.NewTransaction(), nil)

		result, err := build(inArgs, util.NoFlags, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "transaction was not approved")
		assert.Nil(t, result)
	})

	t.Run("Fail parsing JSON", func(t *testing.T) {
		inArgs := []string{tests.TransactionArgString.Filename}
		srv.BuildTransaction.Return(flowkit.NewTransaction(), nil)
		buildFlags.ArgsJSON = `invalid`

		result, err := build(inArgs, util.NoFlags, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "error parsing transaction arguments: invalid character 'i' looking for beginning of value")
		assert.Nil(t, result)
		buildFlags.ArgsJSON = ""
	})

	t.Run("Fail invalid file", func(t *testing.T) {
		inArgs := []string{"invalid"}
		srv.BuildTransaction.Return(flowkit.NewTransaction(), nil)
		result, err := build(inArgs, util.NoFlags, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "error loading transaction file: open invalid: file does not exist")
		assert.Nil(t, result)
	})
}

func Test_Decode(t *testing.T) {

}
