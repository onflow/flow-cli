/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
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

package accounts

import (
	"fmt"
	"strings"
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/tests"
)

func Test_AddContract(t *testing.T) {
	srv, state, _ := util.TestMocks(t)

	t.Run("Success", func(t *testing.T) {
		inArgs := []string{tests.ContractSimpleWithArgs.Filename, "1"}

		srv.AddContract.Run(func(args mock.Arguments) {
			script := args.Get(2).(flowkit.Script)
			assert.Equal(t, tests.ContractSimpleWithArgs.Filename, script.Location)
			assert.Len(t, script.Args, 1)
			assert.Equal(t, inArgs[1], script.Args[0].String())
		})

		result, err := addContract(inArgs, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)

		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Success JSON arg", func(t *testing.T) {
		srv.AddContract.Run(func(args mock.Arguments) {
			script := args.Get(2).(flowkit.Script)
			assert.Equal(t, tests.ContractSimpleWithArgs.Filename, script.Location)
			assert.Len(t, script.Args, 1)
			assert.Equal(t, "1", script.Args[0].String())
		})

		addContractFlags.ArgsJSON = `[{"type": "UInt64", "value": "1"}]`
		args := []string{tests.ContractSimpleWithArgs.Filename}
		result, err := addContract(args, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)

		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Fail non-existing file", func(t *testing.T) {
		args := []string{"non-existing"}
		result, err := addContract(args, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)

		assert.Nil(t, result)
		assert.EqualError(t, err, "error loading contract file: open non-existing: file does not exist")
	})

	t.Run("Fail invalid-json", func(t *testing.T) {
		args := []string{tests.ContractA.Filename}
		addContractFlags.ArgsJSON = "invalid"
		result, err := addContract(args, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)

		assert.Nil(t, result)
		assert.EqualError(t, err, "error parsing transaction arguments: invalid character 'i' looking for beginning of value")
	})

	t.Run("Fail invalid signer", func(t *testing.T) {
		args := []string{tests.ContractA.Filename}
		addContractFlags.Signer = "invalid"
		result, err := addContract(args, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)

		assert.Nil(t, result)
		assert.EqualError(t, err, "could not find account with name invalid in the configuration")
	})

}

func Test_RemoveContract(t *testing.T) {
	srv, state, _ := util.TestMocks(t)

	t.Run("Success", func(t *testing.T) {
		inArgs := []string{"test"}

		srv.RemoveContract.Run(func(args mock.Arguments) {
			acc := args.Get(1).(*flowkit.Account)
			assert.Equal(t, "emulator-account", acc.Name)
			assert.Equal(t, inArgs[0], args.Get(2).(string))
		})

		result, err := removeContract(inArgs, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Fail non-existing account", func(t *testing.T) {
		inArgs := []string{"test"}
		flagsRemove.Signer = "invalid"

		_, err := removeContract(inArgs, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "could not find account with name invalid in the configuration")
	})
}

func Test_UpdateContract(t *testing.T) {
	srv, state, _ := util.TestMocks(t)

	t.Run("Success", func(t *testing.T) {
		inArgs := []string{tests.ContractSimpleWithArgs.Filename, "1"}

		srv.AddContract.Run(func(args mock.Arguments) {
			script := args.Get(2).(flowkit.Script)
			assert.Equal(t, tests.ContractSimpleWithArgs.Filename, script.Location)
			assert.Len(t, script.Args, 1)
			assert.Equal(t, inArgs[1], script.Args[0].String())
		})

		result, err := updateContract(inArgs, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)

		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Success JSON arg", func(t *testing.T) {
		updateContractFlags.ArgsJSON = `[{"type": "UInt64", "value": "1"}]`
		inArgs := []string{tests.ContractSimpleWithArgs.Filename}

		srv.AddContract.Run(func(args mock.Arguments) {
			script := args.Get(2).(flowkit.Script)
			assert.Equal(t, tests.ContractSimpleWithArgs.Filename, script.Location)
			assert.Len(t, script.Args, 1)
			assert.Equal(t, "1", script.Args[0].String())
		})

		result, err := updateContract(inArgs, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)

		require.NoError(t, err)
		assert.NotNil(t, result)
		updateContractFlags.ArgsJSON = "" // reset
	})

	t.Run("Fail invalid-json", func(t *testing.T) {
		args := []string{tests.ContractA.Filename}
		updateContractFlags.ArgsJSON = "invalid"
		result, err := updateContract(args, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)

		assert.Nil(t, result)
		assert.EqualError(t, err, "error parsing transaction arguments: invalid character 'i' looking for beginning of value")
		updateContractFlags.ArgsJSON = "" // reset
	})

	t.Run("Fail non-existing file", func(t *testing.T) {
		args := []string{"non-existing"}
		result, err := updateContract(args, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)

		assert.Nil(t, result)
		assert.EqualError(t, err, "error loading contract file: open non-existing: file does not exist")
	})
}

func Test_Create(t *testing.T) {
	srv, state, _ := util.TestMocks(t)

	t.Run("Success", func(t *testing.T) {
		pkey := "014d91eb68b5fddeca118821e74f70b48d9582c8546d8a2ae9d6835cdb7d1d008624945f55c4b409c628b63a89a54570ed028e8e68a1fe0c98ef08d7f488037b"
		createFlags.Keys = []string{pkey}

		srv.CreateAccount.Run(func(args mock.Arguments) {
			acc := args.Get(1).(*flowkit.Account)
			keys := args.Get(2).([]flowkit.AccountPublicKey)
			assert.Equal(t, "emulator-account", acc.Name)
			assert.Len(t, keys, 1)
			assert.Equal(t, fmt.Sprintf("0x%s", pkey), keys[0].Public.String())
			assert.Equal(t, crypto.ECDSA_P256, keys[0].SigAlgo)
			assert.Equal(t, crypto.SHA3_256, keys[0].HashAlgo)
		})

		result, err := create([]string{}, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)
		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("Success multiple keys", func(t *testing.T) {
		pkey1 := "014d91eb68b5fddeca118821e74f70b48d9582c8546d8a2ae9d6835cdb7d1d008624945f55c4b409c628b63a89a54570ed028e8e68a1fe0c98ef08d7f488037b"
		pkey2 := "c4bcde70e3c29cdc472ce7be46e219ca42f0ed2174369b3ba693c5655ed03f7027c571ba3881ed4b480fba41760572bcc167a8dbcf4e6ed952dcce831f82fc92"
		createFlags.Keys = []string{pkey1, pkey2}
		createFlags.SigAlgo = []string{"ECDSA_P256", "ECDSA_secp256k1"}
		createFlags.HashAlgo = []string{"SHA3_256", "SHA2_256"}
		createFlags.Weights = []int{500, 500}

		srv.CreateAccount.Run(func(args mock.Arguments) {
			acc := args.Get(1).(*flowkit.Account)
			keys := args.Get(2).([]flowkit.AccountPublicKey)
			assert.Equal(t, "emulator-account", acc.Name)
			assert.Len(t, keys, 2)

			assert.Equal(t, fmt.Sprintf("0x%s", pkey1), keys[0].Public.String())
			assert.Equal(t, crypto.ECDSA_P256, keys[0].SigAlgo)
			assert.Equal(t, crypto.SHA3_256, keys[0].HashAlgo)
			assert.Equal(t, 500, keys[0].Weight)

			assert.Equal(t, fmt.Sprintf("0x%s", pkey2), keys[1].Public.String())
			assert.Equal(t, crypto.ECDSA_secp256k1, keys[1].SigAlgo)
			assert.Equal(t, 500, keys[1].Weight)
		})

		result, err := create([]string{}, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)
		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("Fail not enough weights", func(t *testing.T) {
		pkey1 := "014d91eb68b5fddeca118821e74f70b48d9582c8546d8a2ae9d6835cdb7d1d008624945f55c4b409c628b63a89a54570ed028e8e68a1fe0c98ef08d7f488037b"
		pkey2 := "c4bcde70e3c29cdc472ce7be46e219ca42f0ed2174369b3ba693c5655ed03f7027c571ba3881ed4b480fba41760572bcc167a8dbcf4e6ed952dcce831f82fc92"
		createFlags.Keys = []string{pkey1, pkey2}
		createFlags.SigAlgo = []string{"ECDSA_P256", "ECDSA_secp256k1"}
		createFlags.HashAlgo = []string{"SHA3_256", "SHA2_256"}
		createFlags.Weights = []int{1000}

		result, err := create([]string{}, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)
		require.EqualError(t, err, "must provide a key weight for each key provided, keys provided: 2, weights provided: 1")
		require.Nil(t, result)
	})

	t.Run("Fail miss match algos", func(t *testing.T) {
		pkey1 := "014d91eb68b5fddeca118821e74f70b48d9582c8546d8a2ae9d6835cdb7d1d008624945f55c4b409c628b63a89a54570ed028e8e68a1fe0c98ef08d7f488037b"
		createFlags.Keys = []string{pkey1}
		createFlags.SigAlgo = []string{"ECDSA_P256", "ECDSA_secp256k1"}
		createFlags.HashAlgo = []string{"SHA3_256"}
		createFlags.Weights = []int{1000}

		result, err := create([]string{}, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)
		require.EqualError(t, err, "must provide a signature and hash algorithm for every key provided to --key: 1 keys, 2 signature algo, 1 hash algo")
		require.Nil(t, result)
	})

	t.Run("Fail parse keys", func(t *testing.T) {
		_, err := parsePublicKeys([]string{"invalid"}, []crypto.SignatureAlgorithm{crypto.ECDSA_P256})
		assert.EqualError(t, err, "failed decoding public key: invalid with error: encoding/hex: invalid byte: U+0069 'i'")
	})

	t.Run("Fail parse hash", func(t *testing.T) {
		_, err := parseHashingAlgorithms([]string{"invalid"})
		assert.EqualError(t, err, "invalid hash algorithm: invalid")
	})

	t.Run("Fail parse signature algorithm", func(t *testing.T) {
		_, err := parseSignatureAlgorithms([]string{"invalid"})
		assert.EqualError(t, err, "invalid signature algorithm: invalid")
	})
}

func Test_Get(t *testing.T) {
	srv, _, _ := util.TestMocks(t)

	t.Run("Success", func(t *testing.T) {
		inArgs := []string{"0x01"}

		srv.GetAccount.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, "0000000000000001", addr.String())
			srv.GetAccount.Return(tests.NewAccountWithAddress(inArgs[0]), nil)
		})

		result, err := get(inArgs, command.GlobalFlags{}, util.NoLogger, nil, srv.Mock)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func Test_Result(t *testing.T) {
	pkey, _ := crypto.DecodePublicKeyHex(crypto.ECDSA_P256, "a60b9c10a39070806d37d8f0e6be081e7af2d18cd92ee1bd850d10c994d61d538d2693eebe8faa94fea59ee579ea65a70ed897b05126e508e74f55b8669eec6b")
	account := &flow.Account{
		Address: flow.HexToAddress("0x01"),
		Balance: uint64(1),
		Keys: []*flow.AccountKey{{
			Index:          0,
			PublicKey:      pkey,
			SigAlgo:        crypto.ECDSA_P256,
			HashAlgo:       crypto.SHA3_256,
			Weight:         1000,
			SequenceNumber: 0,
			Revoked:        false,
		}},
		Contracts: nil,
	}
	result := AccountResult{Account: account}

	assert.Equal(t, strings.TrimPrefix(`
Address	 0x0000000000000001
Balance	 0.00000001
Keys	 1

Key 0	Public Key		 a60b9c10a39070806d37d8f0e6be081e7af2d18cd92ee1bd850d10c994d61d538d2693eebe8faa94fea59ee579ea65a70ed897b05126e508e74f55b8669eec6b
	Weight			 1000
	Signature Algorithm	 ECDSA_P256
	Hash Algorithm		 SHA3_256
	Revoked 		 false
	Sequence Number 	 0
	Index 			 0

Contracts Deployed: 0


Contracts (hidden, use --include contracts)`, "\n"), result.String())

	assert.Equal(t, map[string]interface{}{
		"address":   flow.Address{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1},
		"balance":   "0.00000001",
		"contracts": []string{},
		"keys": []string{
			"a60b9c10a39070806d37d8f0e6be081e7af2d18cd92ee1bd850d10c994d61d538d2693eebe8faa94fea59ee579ea65a70ed897b05126e508e74f55b8669eec6b",
		},
	}, result.JSON())

}
