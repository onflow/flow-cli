/*
 * Flow CLI
 *
 * Copyright Flow Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package transactions

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/accounts"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"
	"github.com/onflow/flowkit/v2/tests"
	"github.com/onflow/flowkit/v2/transactions"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

func Test_Build(t *testing.T) {
	const serviceAccountAddress = "f8d6e0586b0a20c7"
	srv, state, _ := util.TestMocks(t)

	t.Run("Success", func(t *testing.T) {
		inArgs := []string{tests.TransactionSimple.Filename}

		srv.BuildTransaction.Run(func(args mock.Arguments) {
			roles := args.Get(1).(transactions.AddressesRoles)
			assert.Equal(t, serviceAccountAddress, roles.Payer.String())
			assert.Equal(t, serviceAccountAddress, roles.Proposer.String())
			assert.Equal(t, serviceAccountAddress, roles.Authorizers[0].String())
			assert.Equal(t, uint32(0), args.Get(2).(uint32))
			script := args.Get(3).(flowkit.Script)
			assert.Equal(t, tests.TransactionSimple.Filename, script.Location)
		}).Return(transactions.New(), nil)

		result, err := build(inArgs, command.GlobalFlags{Yes: true}, util.NoLogger, srv.Mock, state)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Fail not approved", func(t *testing.T) {
		inArgs := []string{tests.TransactionSimple.Filename}
		srv.BuildTransaction.Return(transactions.New(), nil)

		result, err := build(inArgs, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "transaction was not approved")
		assert.Nil(t, result)
	})

	t.Run("Fail parsing JSON", func(t *testing.T) {
		inArgs := []string{tests.TransactionArgString.Filename}
		srv.BuildTransaction.Return(transactions.New(), nil)
		buildFlags.ArgsJSON = `invalid`

		result, err := build(inArgs, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "error parsing transaction arguments: invalid character 'i' looking for beginning of value")
		assert.Nil(t, result)
		buildFlags.ArgsJSON = ""
	})

	t.Run("Fail invalid file", func(t *testing.T) {
		inArgs := []string{"invalid"}
		srv.BuildTransaction.Return(transactions.New(), nil)
		result, err := build(inArgs, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)
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

		result, err := decode(inArgs, command.GlobalFlags{}, util.NoLogger, rw, srv.Mock)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Fail decode", func(t *testing.T) {
		inArgs := []string{"test"}
		_ = rw.WriteFile(inArgs[0], []byte("invalid"), 0677)

		result, err := decode(inArgs, command.GlobalFlags{}, util.NoLogger, rw, srv.Mock)
		assert.EqualError(t, err, "failed to decode partial transaction from invalid: encoding/hex: invalid byte: U+0069 'i'")
		assert.Nil(t, result)
	})

	t.Run("Fail to read file", func(t *testing.T) {
		inArgs := []string{"invalid"}
		result, err := decode(inArgs, command.GlobalFlags{}, util.NoLogger, rw, srv.Mock)
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

		result, err := get(inArgs, command.GlobalFlags{}, util.NoLogger, rw, srv.Mock)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func Test_GetSystem(t *testing.T) {
	srv, _, rw := util.TestMocks(t)

	t.Run("Success", func(t *testing.T) {
		inArgs := []string{"100"}

		returnBlock := tests.NewBlock()
		returnBlock.Height = uint64(100)

		srv.GetBlock.Run(func(args mock.Arguments) {
			assert.Equal(t, uint64(100), args.Get(1).(flowkit.BlockQuery).Height)
		}).Return(returnBlock, nil)

		srv.GetSystemTransactionWithID.Return(tests.NewTransaction(), tests.NewTransactionResult(nil), nil)

		res, err := getSystemTransaction(inArgs, command.GlobalFlags{}, util.NoLogger, rw, srv.Mock)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("Fail invalid block ID", func(t *testing.T) {
		inArgs := []string{""}

		returnBlock := tests.NewBlock()
		returnBlock.Height = uint64(100)

		srv.GetBlock.Run(func(args mock.Arguments) {
			assert.Equal(t, uint64(100), args.Get(1).(flowkit.BlockQuery).Height)
		}).Return(returnBlock, nil)

		srv.GetSystemTransactionWithID.Return(nil, nil, fmt.Errorf("block not found"))

		res, err := getSystemTransaction(inArgs, command.GlobalFlags{}, util.NoLogger, rw, srv.Mock)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func Test_Send(t *testing.T) {
	srv, state, _ := util.TestMocks(t)

	t.Run("Success", func(t *testing.T) {
		const compute = uint64(1000)
		flags.ComputeLimit = compute
		inArgs := []string{tests.TransactionArgString.Filename, "test"}

		srv.SendTransaction.Run(func(args mock.Arguments) {
			roles := args.Get(1).(transactions.AccountRoles)
			acc := config.DefaultEmulator.ServiceAccount
			assert.Equal(t, acc, roles.Payer.Name)
			assert.Equal(t, acc, roles.Proposer.Name)
			assert.Equal(t, acc, roles.Authorizers[0].Name)
			script := args.Get(2).(flowkit.Script)
			assert.Equal(t, tests.TransactionArgString.Filename, script.Location)
			assert.Equal(t, args.Get(3).(uint64), compute)
		}).Return(nil, nil, nil)

		result, err := send(inArgs, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Fail non-existing account", func(t *testing.T) {
		flags.Proposer = "invalid"
		_, err := send([]string{""}, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "proposer account: [invalid] doesn't exists in configuration")
		flags.Proposer = "" // reset

		flags.Payer = "invalid"
		_, err = send([]string{""}, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "payer account: [invalid] doesn't exists in configuration")
		flags.Payer = "" // reset

		flags.Authorizers = []string{"invalid"}
		_, err = send([]string{""}, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "authorizer account: [invalid] doesn't exists in configuration")
		flags.Authorizers = nil // reset

		flags.Signer = "invalid"
		_, err = send([]string{""}, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "signer account: [invalid] doesn't exists in configuration")
		flags.Signer = "" // reset
	})

	t.Run("Fail signer and payer flag", func(t *testing.T) {
		flags.Proposer = config.DefaultEmulator.ServiceAccount
		flags.Signer = config.DefaultEmulator.ServiceAccount
		_, err := send([]string{""}, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "signer flag cannot be combined with payer/proposer/authorizer flags")
		flags.Signer = "" // reset
	})

	t.Run("Fail signer not used and payer flag not set", func(t *testing.T) {
		flags.Payer = ""
		flags.Proposer = config.DefaultEmulator.ServiceAccount
		_, err := send([]string{""}, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "proposer/payer flags are required when signer flag is not used")
		flags.Signer = "" // reset
	})

	t.Run("Fail signer not used and proposer flag not set", func(t *testing.T) {
		flags.Proposer = ""
		flags.Payer = config.DefaultEmulator.ServiceAccount
		_, err := send([]string{""}, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "proposer/payer flags are required when signer flag is not used")
		flags.Signer = "" // reset
	})

	t.Run("Fail loading transaction file", func(t *testing.T) {
		flags.Proposer = config.DefaultEmulator.ServiceAccount
		flags.Payer = config.DefaultEmulator.ServiceAccount
		_, err := send([]string{"invalid"}, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)
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
			tx := args.Get(1).(*transactions.Transaction)
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
		_, err := sendSigned(inArgs, command.GlobalFlags{}, util.NoLogger, rw, srv.Mock)
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
			assert.Equal(t, "emulator-account", args.Get(1).(*accounts.Account).Name)
			assert.Equal(t, built, args.Get(2).([]byte))
		}).Return(transactions.New(), nil)

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

func Test_Result(t *testing.T) {
	tx := &flow.Transaction{
		Script:           []byte(`transaction {}`),
		ReferenceBlockID: flow.HexToID("6cde7f812897d22ee7633b82b059070be24faccdc47997bc0f765420e6e28bb6"),
		GasLimit:         flow.DefaultTransactionGasLimit,
		ProposalKey: flow.ProposalKey{
			Address:        flow.HexToAddress("0x01"),
			KeyIndex:       0,
			SequenceNumber: 1,
		},
		Payer: flow.HexToAddress("0x02"),
		PayloadSignatures: []flow.TransactionSignature{{
			Address:     flow.HexToAddress("0x01"),
			SignerIndex: 0,
			KeyIndex:    0,
			Signature:   []byte("6cde7f812897d22ee7633b82b059070be24faccdc47997bc0f765420e6e28bb6"),
		}},
		EnvelopeSignatures: []flow.TransactionSignature{{
			Address:     flow.HexToAddress("0x01"),
			SignerIndex: 0,
			KeyIndex:    0,
			Signature:   []byte("6cde7f812897d22ee7633b82b059070be24faccdc47997bc0f765420e6e28bb6"),
		}},
	}

	event := tests.NewEvent(
		0,
		"A.foo",
		[]cadence.Field{{Type: cadence.StringType, Identifier: "bar"}},
		[]cadence.Value{cadence.NewInt(1)},
	)

	withdrawFlowEvent := tests.NewEvent(
		1,
		"A.1654653399040a61.FlowToken.TokensWithdrawn",
		[]cadence.Field{{Type: cadence.StringType, Identifier: "bar"}},
		[]cadence.Value{cadence.NewInt(1)},
	)
	depositFlowEvent := tests.NewEvent(
		2,
		"A.1654653399040a61.FlowToken.TokensDeposited",
		[]cadence.Field{{Type: cadence.StringType, Identifier: "bar"}},
		[]cadence.Value{cadence.NewInt(1)},
	)
	storageUsedEvent := tests.NewEvent(
		3,
		"A.1654653399040a61.FlowStorageFees.StorageCapacityUsed",
		[]cadence.Field{{Type: cadence.StringType, Identifier: "bar"}},
		[]cadence.Value{cadence.NewInt(1)},
	)
	feeEvent := tests.NewEvent(
		4,
		"A.f919ee77447b7497.FlowFees.FeesDeducted",
		[]cadence.Field{{Type: cadence.StringType, Identifier: "bar"}},
		[]cadence.Value{cadence.NewInt(1)},
	)
	txResult := &flow.TransactionResult{
		Status:      flow.TransactionStatusSealed,
		Error:       nil,
		Events:      []flow.Event{*event},
		BlockID:     flow.HexToID("7aa74143741c1c3b837d389fcffa7a5e251b67b4ffef6d6887b40cd9c803f537"),
		BlockHeight: 1,
	}

	txResultFeeEvents := &flow.TransactionResult{
		Status:      flow.TransactionStatusSealed,
		Error:       nil,
		Events:      []flow.Event{*event, *withdrawFlowEvent, *depositFlowEvent, *storageUsedEvent, *feeEvent},
		BlockID:     flow.HexToID("7aa74143741c1c3b837d389fcffa7a5e251b67b4ffef6d6887b40cd9c803f537"),
		BlockHeight: 1,
	}
	t.Run("Success with no result", func(t *testing.T) {
		result := transactionResult{tx: tx}

		assert.Equal(t, strings.TrimPrefix(`
ID		e913d1f3e431c7df49c99845bea9ebff9db11bbf25d507b9ad0fad45652d515f
Payer		0000000000000002
Authorizers	[]

Proposal Key:	
    Address	0000000000000001
    Index	0
    Sequence	1

Payload Signature 0: 0000000000000001
Envelope Signature 0: 0000000000000001
Signatures (minimized, use --include signatures)

Code (hidden, use --include code)

Payload (hidden, use --include payload)

Fee Events (hidden, use --include fee-events)`, "\n"), result.String())

		assert.Equal(t, map[string]any{
			"authorizers": "[]",
			"id":          "e913d1f3e431c7df49c99845bea9ebff9db11bbf25d507b9ad0fad45652d515f",
			"payer":       "0000000000000002",
			"payload":     "f8dbf8498e7472616e73616374696f6e207b7dc0a06cde7f812897d22ee7633b82b059070be24faccdc47997bc0f765420e6e28bb682270f8800000000000000018001880000000000000002c0f846f8448080b84036636465376638313238393764323265653736333362383262303539303730626532346661636364633437393937626330663736353432306536653238626236f846f8448080b84036636465376638313238393764323265653736333362383262303539303730626532346661636364633437393937626330663736353432306536653238626236",
		}, result.JSON())
	})

	t.Run("Success with result", func(t *testing.T) {
		result := transactionResult{tx: tx, result: txResult}

		expectedString := strings.TrimPrefix(fmt.Sprintf(`
Block ID	7aa74143741c1c3b837d389fcffa7a5e251b67b4ffef6d6887b40cd9c803f537
Block Height	1
Status		%s SEALED
ID		e913d1f3e431c7df49c99845bea9ebff9db11bbf25d507b9ad0fad45652d515f
Payer		0000000000000002
Authorizers	[]

Proposal Key:	
    Address	0000000000000001
    Index	0
    Sequence	1

Payload Signature 0: 0000000000000001
Envelope Signature 0: 0000000000000001
Signatures (minimized, use --include signatures)

Events:		 
    Index	0
    Type	A.foo
    Tx ID	0000000000000000000000000000000000000000000000000000000000000000
    Values
		- bar (String): 1 



Code (hidden, use --include code)

Payload (hidden, use --include payload)

Fee Events (hidden, use --include fee-events)`, output.OkEmoji()), "\n")

		assert.Equal(t, expectedString, result.String())

		assert.Equal(t, map[string]any{
			"authorizers":  "[]",
			"block_height": uint64(1),
			"block_id":     "7aa74143741c1c3b837d389fcffa7a5e251b67b4ffef6d6887b40cd9c803f537",
			"events": []any{
				map[string]any{
					"index":  0,
					"type":   "A.foo",
					"values": json.RawMessage{0x7b, 0x22, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x7b, 0x22, 0x69, 0x64, 0x22, 0x3a, 0x22, 0x41, 0x2e, 0x66, 0x6f, 0x6f, 0x22, 0x2c, 0x22, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x22, 0x3a, 0x5b, 0x7b, 0x22, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x7b, 0x22, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x22, 0x31, 0x22, 0x2c, 0x22, 0x74, 0x79, 0x70, 0x65, 0x22, 0x3a, 0x22, 0x49, 0x6e, 0x74, 0x22, 0x7d, 0x2c, 0x22, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x3a, 0x22, 0x62, 0x61, 0x72, 0x22, 0x7d, 0x5d, 0x7d, 0x2c, 0x22, 0x74, 0x79, 0x70, 0x65, 0x22, 0x3a, 0x22, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x22, 0x7d, 0xa},
				},
			},
			"id":      "e913d1f3e431c7df49c99845bea9ebff9db11bbf25d507b9ad0fad45652d515f",
			"payer":   "0000000000000002",
			"payload": "f8dbf8498e7472616e73616374696f6e207b7dc0a06cde7f812897d22ee7633b82b059070be24faccdc47997bc0f765420e6e28bb682270f8800000000000000018001880000000000000002c0f846f8448080b84036636465376638313238393764323265653736333362383262303539303730626532346661636364633437393937626330663736353432306536653238626236f846f8448080b84036636465376638313238393764323265653736333362383262303539303730626532346661636364633437393937626330663736353432306536653238626236",
			"status":  "SEALED",
		}, result.JSON())
	})

	t.Run("Result without fee events", func(t *testing.T) {
		result := transactionResult{tx: tx, result: txResultFeeEvents}

		assert.Equal(t, strings.TrimPrefix(fmt.Sprintf(`
Block ID	7aa74143741c1c3b837d389fcffa7a5e251b67b4ffef6d6887b40cd9c803f537
Block Height	1
Status		%s SEALED
ID		e913d1f3e431c7df49c99845bea9ebff9db11bbf25d507b9ad0fad45652d515f
Payer		0000000000000002
Authorizers	[]

Proposal Key:	
    Address	0000000000000001
    Index	0
    Sequence	1

Payload Signature 0: 0000000000000001
Envelope Signature 0: 0000000000000001
Signatures (minimized, use --include signatures)

Events:		 
    Index	0
    Type	A.foo
    Tx ID	0000000000000000000000000000000000000000000000000000000000000000
    Values
		- bar (String): 1 



Code (hidden, use --include code)

Payload (hidden, use --include payload)

Fee Events (hidden, use --include fee-events)`, output.OkEmoji()), "\n"), result.String())
	})
	t.Run("Result with fee events", func(t *testing.T) {
		result := transactionResult{tx: tx, result: txResultFeeEvents, include: []string{"fee-events"}}

		assert.Equal(t, strings.TrimPrefix(fmt.Sprintf(`
Block ID	7aa74143741c1c3b837d389fcffa7a5e251b67b4ffef6d6887b40cd9c803f537
Block Height	1
Status		%s SEALED
ID		e913d1f3e431c7df49c99845bea9ebff9db11bbf25d507b9ad0fad45652d515f
Payer		0000000000000002
Authorizers	[]

Proposal Key:	
    Address	0000000000000001
    Index	0
    Sequence	1

Payload Signature 0: 0000000000000001
Envelope Signature 0: 0000000000000001
Signatures (minimized, use --include signatures)

Events:		 
    Index	0
    Type	A.foo
    Tx ID	0000000000000000000000000000000000000000000000000000000000000000
    Values
		- bar (String): 1 

    Index	1
    Type	A.1654653399040a61.FlowToken.TokensWithdrawn
    Tx ID	0000000000000000000000000000000000000000000000000000000000000000
    Values
		- bar (String): 1 

    Index	2
    Type	A.1654653399040a61.FlowToken.TokensDeposited
    Tx ID	0000000000000000000000000000000000000000000000000000000000000000
    Values
		- bar (String): 1 

    Index	3
    Type	A.1654653399040a61.FlowStorageFees.StorageCapacityUsed
    Tx ID	0000000000000000000000000000000000000000000000000000000000000000
    Values
		- bar (String): 1 

    Index	4
    Type	A.f919ee77447b7497.FlowFees.FeesDeducted
    Tx ID	0000000000000000000000000000000000000000000000000000000000000000
    Values
		- bar (String): 1 



Code (hidden, use --include code)

Payload (hidden, use --include payload)`, output.OkEmoji()), "\n"), result.String())
	})

	t.Run("Block explorer link for mainnet", func(t *testing.T) {
		result := transactionResult{tx: tx, result: txResult, network: "mainnet"}

		output := result.String()
		assert.Contains(t, output, "🔗 View on Block Explorer:")
		assert.Contains(t, output, "https://www.flowscan.io/tx/e913d1f3e431c7df49c99845bea9ebff9db11bbf25d507b9ad0fad45652d515f")

		jsonResult := result.JSON()
		jsonMap, ok := jsonResult.(map[string]any)
		assert.True(t, ok)
		assert.Contains(t, jsonMap, "view_on_block_explorer")
		assert.Equal(t, "https://www.flowscan.io/tx/e913d1f3e431c7df49c99845bea9ebff9db11bbf25d507b9ad0fad45652d515f", jsonMap["view_on_block_explorer"])
	})

	t.Run("Block explorer link for testnet", func(t *testing.T) {
		result := transactionResult{tx: tx, result: txResult, network: "testnet"}

		output := result.String()
		assert.Contains(t, output, "🔗 View on Block Explorer:")
		assert.Contains(t, output, "https://testnet.flowscan.io/tx/e913d1f3e431c7df49c99845bea9ebff9db11bbf25d507b9ad0fad45652d515f")

		jsonResult := result.JSON()
		jsonMap, ok := jsonResult.(map[string]any)
		assert.True(t, ok)
		assert.Contains(t, jsonMap, "view_on_block_explorer")
		assert.Equal(t, "https://testnet.flowscan.io/tx/e913d1f3e431c7df49c99845bea9ebff9db11bbf25d507b9ad0fad45652d515f", jsonMap["view_on_block_explorer"])
	})

	t.Run("No block explorer link for emulator", func(t *testing.T) {
		result := transactionResult{tx: tx, result: txResult, network: "emulator"}

		output := result.String()
		assert.NotContains(t, output, "🔗 View on Block Explorer:")

		jsonResult := result.JSON()
		jsonMap, ok := jsonResult.(map[string]any)
		assert.True(t, ok)
		assert.NotContains(t, jsonMap, "view_on_block_explorer")
	})

	t.Run("No block explorer link for empty network", func(t *testing.T) {
		result := transactionResult{tx: tx, result: txResult, network: ""}

		output := result.String()
		assert.NotContains(t, output, "🔗 View on Block Explorer:")

		jsonResult := result.JSON()
		jsonMap, ok := jsonResult.(map[string]any)
		assert.True(t, ok)
		assert.NotContains(t, jsonMap, "view_on_block_explorer")
	})
}
