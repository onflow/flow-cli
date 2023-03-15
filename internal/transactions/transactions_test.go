package transactions

import (
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/tests"
	"github.com/onflow/flow-go-sdk"
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
	srv, _, rw := util.TestMocks(t)

	t.Run("Success", func(t *testing.T) {
		inArgs := []string{"test"}
		payload := []byte("f8aaf8a6b8617472616e73616374696f6e2829207b0a097072657061726528617574686f72697a65723a20417574684163636f756e7429207b7d0a0965786563757465207b0a09096c65742078203d20310a090970616e696328227465737422290a097d0a7d0ac0a003d40910037d575d52831647b39814f445bc8cc7ba8653286c0eb1473778c34f8203e888f8d6e0586b0a20c7808088f8d6e0586b0a20c7c988f8d6e0586b0a20c7c0c0")
		_ = rw.WriteFile(inArgs[0], payload, 0677)

		result, err := decode(inArgs, util.NoFlags, util.NoLogger, rw, srv.Mock)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Fail decode", func(t *testing.T) {
		inArgs := []string{"test"}
		_ = rw.WriteFile(inArgs[0], []byte("invalid"), 0677)

		result, err := decode(inArgs, util.NoFlags, util.NoLogger, rw, srv.Mock)
		assert.EqualError(t, err, "failed to decode partial transaction from invalid: encoding/hex: invalid byte: U+0069 'i'")
		assert.Nil(t, result)
	})

	t.Run("Fail to read file", func(t *testing.T) {
		inArgs := []string{"invalid"}
		result, err := decode(inArgs, util.NoFlags, util.NoLogger, rw, srv.Mock)
		assert.EqualError(t, err, "failed to read transaction from invalid: open invalid: file does not exist")
		assert.Nil(t, result)
	})
}

func Test_Get(t *testing.T) {
	srv, _, rw := util.TestMocks(t)

	t.Run("Success", func(t *testing.T) {
		inArgs := []string{"0x01"}

		srv.GetTransactionByID.Run(func(args mock.Arguments) {
			id := args.Get(1).(flow.Identifier)
			assert.Equal(t, "0100000000000000000000000000000000000000000000000000000000000000", id.String())
		}).Return(nil, nil, nil)

		result, err := get(inArgs, util.NoFlags, util.NoLogger, rw, srv.Mock)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}
