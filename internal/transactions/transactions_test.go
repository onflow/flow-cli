package transactions

import (
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
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

func Test_Send(t *testing.T) {
	srv, state, _ := util.TestMocks(t)

	t.Run("Success", func(t *testing.T) {
		const gas = uint64(1000)
		sendFlags.GasLimit = gas
		inArgs := []string{tests.TransactionArgString.Filename, "foo"}

		srv.SendTransaction.Run(func(args mock.Arguments) {
			roles := args.Get(1).(*flowkit.TransactionAccountRoles)
			acc := config.DefaultEmulatorServiceAccountName
			assert.Equal(t, acc, roles.Payer.Name())
			assert.Equal(t, acc, roles.Proposer.Name())
			assert.Equal(t, acc, roles.Authorizers[0].Name())
			script := args.Get(2).(*flowkit.Script)
			assert.Equal(t, tests.TransactionArgString.Filename, script.Location())
			assert.Equal(t, args.Get(3).(uint64), gas)
		}).Return(nil, nil, nil)

		result, err := send(inArgs, util.NoFlags, util.NoLogger, srv.Mock, state)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Fail non-existing account", func(t *testing.T) {
		sendFlags.Proposer = "invalid"
		_, err := send([]string{""}, util.NoFlags, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "proposer account: [invalid] doesn't exists in configuration")
		sendFlags.Proposer = "" // reset

		sendFlags.Payer = "invalid"
		_, err = send([]string{""}, util.NoFlags, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "payer account: [invalid] doesn't exists in configuration")
		sendFlags.Payer = "" // reset

		sendFlags.Authorizers = []string{"invalid"}
		_, err = send([]string{""}, util.NoFlags, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "authorizer account: [invalid] doesn't exists in configuration")
		sendFlags.Authorizers = nil // reset

		sendFlags.Signer = "invalid"
		_, err = send([]string{""}, util.NoFlags, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "signer account: [invalid] doesn't exists in configuration")
		sendFlags.Signer = "" // reset
	})

	t.Run("Fail signer and payer flag", func(t *testing.T) {
		sendFlags.Proposer = config.DefaultEmulatorServiceAccountName
		sendFlags.Signer = config.DefaultEmulatorServiceAccountName
		_, err := send([]string{""}, util.NoFlags, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "signer flag cannot be combined with payer/proposer/authorizer flags")
		sendFlags.Signer = "" // reset
	})

	t.Run("Fail loading transaction file", func(t *testing.T) {
		_, err := send([]string{"invalid"}, util.NoFlags, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "error loading transaction file: open invalid: file does not exist")
	})
}

func Test_SendSigned(t *testing.T) {
	srv, _, rw := util.TestMocks(t)

	t.Run("Success", func(t *testing.T) {
		inArgs := []string{"test"}
		payload := []byte("f8aaf8a6b8617472616e73616374696f6e2829207b0a097072657061726528617574686f72697a65723a20417574684163636f756e7429207b7d0a0965786563757465207b0a09096c65742078203d20310a090970616e696328227465737422290a097d0a7d0ac0a003d40910037d575d52831647b39814f445bc8cc7ba8653286c0eb1473778c34f8203e888f8d6e0586b0a20c7808088f8d6e0586b0a20c7c988f8d6e0586b0a20c7c0c0")
		_ = rw.WriteFile(inArgs[0], payload, 0677)

		srv.SendSignedTransaction.Run(func(args mock.Arguments) {
			tx := args.Get(1).(*flowkit.Transaction)
			assert.Equal(t, "f8d6e0586b0a20c7", tx.FlowTransaction().Payer.String())
			assert.Equal(t, "f8d6e0586b0a20c7", tx.FlowTransaction().Authorizers[0].String())
			assert.Equal(t, "f8d6e0586b0a20c7", tx.FlowTransaction().ProposalKey.Address.String())
		}).Return(nil, nil, nil)

		result, err := sendSigned(inArgs, command.GlobalFlags{Yes: true}, util.NoLogger, rw, srv.Mock)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Fail loading transaction", func(t *testing.T) {
		inArgs := []string{"invalid"}
		_, err := sendSigned(inArgs, command.GlobalFlags{Yes: true}, util.NoLogger, rw, srv.Mock)
		assert.EqualError(t, err, "error loading transaction payload: open invalid: file does not exist")
	})

	t.Run("Fail not approved", func(t *testing.T) {
		inArgs := []string{"test"}
		payload := []byte("f8aaf8a6b8617472616e73616374696f6e2829207b0a097072657061726528617574686f72697a65723a20417574684163636f756e7429207b7d0a0965786563757465207b0a09096c65742078203d20310a090970616e696328227465737422290a097d0a7d0ac0a003d40910037d575d52831647b39814f445bc8cc7ba8653286c0eb1473778c34f8203e888f8d6e0586b0a20c7808088f8d6e0586b0a20c7c988f8d6e0586b0a20c7c0c0")
		_ = rw.WriteFile(inArgs[0], payload, 0677)
		_, err := sendSigned(inArgs, util.NoFlags, util.NoLogger, rw, srv.Mock)
		assert.EqualError(t, err, "transaction was not approved for sending")
	})
}

func Test_Sign(t *testing.T) {
	srv, state, rw := util.TestMocks(t)

	t.Run("Success", func(t *testing.T) {
		inArgs := []string{"t1.rlp"}
		built := []byte("f884f880b83b7472616e73616374696f6e2829207b0a0909097072657061726528617574686f72697a65723a20417574684163636f756e7429207b7d0a09097d0ac0a003d40910037d575d52831647b39814f445bc8cc7ba8653286c0eb1473778c34f8203e888f8d6e0586b0a20c7808088f8d6e0586b0a20c7c988f8d6e0586b0a20c7c0c0")
		_ = rw.WriteFile(inArgs[0], built, 0677)

		srv.SignTransactionPayload.Run(func(args mock.Arguments) {
			assert.Equal(t, "emulator-account", args.Get(1).(*flowkit.Account).Name())
			assert.Equal(t, built, args.Get(2).([]byte))
		}).Return(flowkit.NewTransaction(), nil)

		result, err := sign(inArgs, command.GlobalFlags{Yes: true}, util.NoLogger, srv.Mock, state)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Fail filename arg required", func(t *testing.T) {
		_, err := sign([]string{}, command.GlobalFlags{Yes: true}, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "filename argument is required")
	})

	t.Run("Fail only use filename", func(t *testing.T) {
		signFlags.FromRemoteUrl = "foo"
		_, err := sign([]string{"test"}, command.GlobalFlags{Yes: true}, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "only use one, filename argument or --from-remote-url <url>")
		signFlags.FromRemoteUrl = ""
	})

	t.Run("Fail invalid signer", func(t *testing.T) {
		inArgs := []string{"t1.rlp"}
		built := []byte("f884f880b83b7472616e73616374696f6e2829207b0a0909097072657061726528617574686f72697a65723a20417574684163636f756e7429207b7d0a09097d0ac0a003d40910037d575d52831647b39814f445bc8cc7ba8653286c0eb1473778c34f8203e888f8d6e0586b0a20c7808088f8d6e0586b0a20c7c988f8d6e0586b0a20c7c0c0")
		_ = rw.WriteFile(inArgs[0], built, 0677)
		signFlags.Signer = []string{"invalid"}
		_, err := sign(inArgs, command.GlobalFlags{Yes: true}, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "signer account: [invalid] doesn't exists in configuration")
		signFlags.Signer = []string{}
	})
}
