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

package migrate

import (
	"strings"
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/sema"
	"github.com/onflow/cadence/runtime/stdlib"
	"github.com/onflow/contract-updater/lib/go/templates"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	flowkitMocks "github.com/onflow/flowkit/v2/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockAccount struct {
	address         flow.Address
	contracts       map[string][]byte
	stagedContracts map[string][]byte
}

func setupValidatorMocks(
	t *testing.T,
	accounts []mockAccount,
) *flowkitMocks.Services {
	t.Helper()
	srv := flowkitMocks.NewServices(t)

	// Mock all accounts & staged contracts
	for _, acc := range accounts {
		mockAccount := &flow.Account{
			Address:   acc.address,
			Balance:   1000,
			Keys:      nil,
			Contracts: acc.contracts,
		}

		srv.On("GetAccount", mock.Anything, acc.address).Return(mockAccount, nil).Maybe()

		for contractName, code := range acc.stagedContracts {
			srv.On(
				"ExecuteScript",
				mock.Anything,
				mock.MatchedBy(func(script flowkit.Script) bool {
					if string(script.Code) != string(templates.GenerateGetStagedContractCodeScript(MigrationContractStagingAddress("testnet"))) {
						return false
					}

					if len(script.Args) != 2 {
						return false
					}

					callContractAddress, callContractName := script.Args[0], script.Args[1]
					if callContractName != cadence.String(contractName) {
						return false
					}
					if callContractAddress != cadence.Address(acc.address) {
						return false
					}

					return true
				}),
				mock.Anything,
			).Return(cadence.NewOptional(cadence.String(code)), nil).Maybe()
		}
	}

	// Mock missing contract, fallback if not found
	srv.On(
		"ExecuteScript",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(cadence.NewOptional(nil), nil).Maybe()

	srv.On("Network", mock.Anything).Return(config.Network{
		Name: "testnet",
	}, nil).Maybe()

	return srv
}

// Helper for creating address locations from strings in tests
func simpleAddressLocation(location string) common.AddressLocation {
	split := strings.Split(location, ".")
	addr, _ := common.HexToAddress(split[0])
	return common.NewAddressLocation(nil, addr, split[1])
}

func Test_StagingValidator(t *testing.T) {
	t.Run("valid contract update with no dependencies", func(t *testing.T) {
		location := common.NewAddressLocation(nil, common.Address{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, "Test")
		sourceCodeLocation := common.StringLocation("./Test.cdc")
		oldContract := `
		pub contract Test {
			pub fun test() {}
		}`
		newContract := `
		access(all) contract Test {
			access(all) fun test() {}
		}`

		// setup mocks
		srv := setupValidatorMocks(t, []mockAccount{
			{
				address:         flow.HexToAddress("01"),
				contracts:       map[string][]byte{"Test": []byte(oldContract)},
				stagedContracts: nil,
			},
		})

		validator := newStagingValidator(srv)
		err := validator.Validate([]StagedContract{{location, sourceCodeLocation, []byte(newContract)}})
		require.NoError(t, err)
	})

	t.Run("contract update with update error", func(t *testing.T) {
		location := common.NewAddressLocation(nil, common.Address{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, "Test")
		sourceCodeLocation := common.StringLocation("./Test.cdc")
		oldContract := `
		pub contract Test {
			pub fun test() {}
		}`
		newContract := `
		access(all) contract Test {
			access(all) let x: Int
			access(all) fun test() {}

			init() {
				self.x = 1
			}
		}`

		// setup mocks
		srv := setupValidatorMocks(t, []mockAccount{
			{
				address:         flow.HexToAddress("01"),
				contracts:       map[string][]byte{"Test": []byte(oldContract)},
				stagedContracts: nil,
			},
		})

		validator := newStagingValidator(srv)
		err := validator.Validate([]StagedContract{{location, sourceCodeLocation, []byte(newContract)}})
		var updateErr *stdlib.ContractUpdateError
		require.ErrorAs(t, err, &updateErr)
	})

	t.Run("contract update with checker error", func(t *testing.T) {
		location := common.NewAddressLocation(nil, common.Address{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, "Test")
		sourceCodeLocation := common.StringLocation("./Test.cdc")
		oldContract := `
		pub contract Test {
			let x: Int
			init() {
				self.x = 1
			}
		}`
		newContract := `
		access(all) contract Test {
			access(all) let x: Int
			init() {
				self.x = "bad type :("
			}
		}`

		// setup mocks
		srv := setupValidatorMocks(t, []mockAccount{
			{
				address:         flow.HexToAddress("01"),
				contracts:       map[string][]byte{"Test": []byte(oldContract)},
				stagedContracts: nil,
			},
		})

		validator := newStagingValidator(srv)
		err := validator.Validate([]StagedContract{{location, sourceCodeLocation, []byte(newContract)}})
		var checkerErr *sema.CheckerError
		require.ErrorAs(t, err, &checkerErr)
	})

	t.Run("valid contract update with dependencies", func(t *testing.T) {
		location := common.NewAddressLocation(nil, common.Address{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, "Test")
		sourceCodeLocation := common.StringLocation("./Test.cdc")
		oldContract := `
		pub contract Test {
			pub fun test() {}
		}`
		newContract := `
		import ImpContract from 0x02
		access(all) contract Test {
			access(all) fun test() {}
		}`
		impContract := `
		access(all) contract ImpContract {
			access(all) let x: Int
			init() {
				self.x = 1
			}
		}`

		// setup mocks
		srv := setupValidatorMocks(t, []mockAccount{
			{
				address:         flow.HexToAddress("01"),
				contracts:       map[string][]byte{"Test": []byte(oldContract)},
				stagedContracts: nil,
			},
			{
				address:         flow.HexToAddress("02"),
				contracts:       nil,
				stagedContracts: map[string][]byte{"ImpContract": []byte(impContract)},
			},
		})

		// validate
		validator := newStagingValidator(srv)
		err := validator.Validate([]StagedContract{{location, sourceCodeLocation, []byte(newContract)}})
		require.NoError(t, err)
	})

	t.Run("contract update missing dependency", func(t *testing.T) {
		location := common.NewAddressLocation(nil, common.Address{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, "Test")
		impLocation := common.NewAddressLocation(nil, common.Address{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02}, "ImpContract")
		sourceCodeLocation := common.StringLocation("./Test.cdc")
		oldContract := `
		pub contract Test {
			pub fun test() {}
		}`
		newContract := `
		import ImpContract from 0x02
		access(all) contract Test {
			access(all) fun test() {}
		}`

		// setup mocks
		srv := setupValidatorMocks(t, []mockAccount{
			{
				address:         flow.HexToAddress("01"),
				contracts:       map[string][]byte{"Test": []byte(oldContract)},
				stagedContracts: nil,
			},
		})

		validator := newStagingValidator(srv)
		err := validator.Validate([]StagedContract{{location, sourceCodeLocation, []byte(newContract)}})

		var validatorErr *stagingValidatorError
		require.ErrorAs(t, err, &validatorErr)
		require.Equal(t, 1, len(validatorErr.Unwrap()))

		var missingDependenciesErr *missingDependenciesError
		require.ErrorAs(t, validatorErr.Unwrap()[0], &missingDependenciesErr)
		require.Equal(t, 1, len(missingDependenciesErr.MissingContracts))
		require.Equal(t, impLocation, missingDependenciesErr.MissingContracts[0])
	})

	t.Run("valid contract update with system contract imports", func(t *testing.T) {
		location := common.NewAddressLocation(nil, common.Address{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, "Test")
		sourceCodeLocation := common.StringLocation("./Test.cdc")
		oldContract := `
		import FlowToken from 0x7e60df042a9c0868
		pub contract Test {
			pub fun test() {}
		}`
		newContract := `
		import FlowToken from 0x7e60df042a9c0868
		import Burner from 0x9a0766d93b6608b7
		access(all) contract Test {
			access(all) fun test() {}
		}`

		// setup mocks
		srv := setupValidatorMocks(t, []mockAccount{
			{
				address:         flow.HexToAddress("01"),
				contracts:       map[string][]byte{"Test": []byte(oldContract)},
				stagedContracts: nil,
			},
		})

		validator := newStagingValidator(srv)
		err := validator.Validate([]StagedContract{{location, sourceCodeLocation, []byte(newContract)}})
		require.NoError(t, err)
	})

	t.Run("resolves account access correctly", func(t *testing.T) {
		location := simpleAddressLocation("0x01.Test")
		sourceCodeLocation := common.StringLocation("./Test.cdc")
		oldContract := `
		import ImpContract from 0x01
		pub contract Test {
			pub fun test() {}
		}`
		newContract := `
		import ImpContract from 0x01
		access(all) contract Test {
			access(all) fun test() {}
			init() {
				ImpContract.test()
			}
		}`
		impContract := `
		access(all) contract ImpContract {
			access(account) fun test() {}
			init() {}
		}`

		// setup mocks
		srv := setupValidatorMocks(t, []mockAccount{
			{
				address:         flow.HexToAddress("01"),
				contracts:       map[string][]byte{"Test": []byte(oldContract)},
				stagedContracts: map[string][]byte{"ImpContract": []byte(impContract)},
			},
		})

		// validate
		validator := newStagingValidator(srv)
		err := validator.Validate([]StagedContract{{location, sourceCodeLocation, []byte(newContract)}})
		require.NoError(t, err)
	})

	t.Run("validates multiple contracts, no error", func(t *testing.T) {
		// setup mocks
		srv := setupValidatorMocks(t, []mockAccount{
			{
				address: flow.HexToAddress("01"),
				contracts: map[string][]byte{"Foo": []byte(`
				pub contract Foo {
					pub fun test() {}
				}`)},
				stagedContracts: nil,
			},
			{
				address: flow.HexToAddress("02"),
				contracts: map[string][]byte{"Bar": []byte(`
				pub contract Bar {
					pub fun test() {}
					init() {
						Foo.test()
					}
				}`)},
				stagedContracts: nil,
			},
		})

		validator := newStagingValidator(srv)
		err := validator.Validate([]StagedContract{
			{
				DeployLocation: simpleAddressLocation("0x01.Foo"),
				SourceLocation: common.StringLocation("./Foo.cdc"),
				Code: []byte(`
				access(all) contract Foo {
					access(all) fun test() {}
				}`),
			},
			{
				DeployLocation: simpleAddressLocation("0x02.Bar"),
				SourceLocation: common.StringLocation("./Bar.cdc"),
				Code: []byte(`
				import Foo from 0x01
				access(all) contract Bar {
					access(all) fun test() {}
					init() {
						Foo.test()
					}
				}`),
			},
		})

		require.NoError(t, err)
	})

	t.Run("validates multiple contracts with errors", func(t *testing.T) {
		// setup mocks
		srv := setupValidatorMocks(t, []mockAccount{
			{
				address: flow.HexToAddress("01"),
				contracts: map[string][]byte{"Foo": []byte(`
				pub contract Foo {
					pub fun test() {}
					init() {}
				}`)},
				stagedContracts: nil,
			},
			{
				address: flow.HexToAddress("02"),
				contracts: map[string][]byte{"Bar": []byte(`
				pub contract Bar {
					pub fun test() {}
					init() {
						Foo.test()
					}
				}`)},
				stagedContracts: nil,
			},
		})

		validator := newStagingValidator(srv)
		err := validator.Validate([]StagedContract{
			{
				DeployLocation: simpleAddressLocation("0x01.Foo"),
				SourceLocation: common.StringLocation("./Foo.cdc"),
				Code: []byte(`
				access(all) contract Foo {
					access(all) fun test() {}
					init() {
						let x: Int = "bad type"
					}
				}`),
			},
			{
				DeployLocation: simpleAddressLocation("0x02.Bar"),
				SourceLocation: common.StringLocation("./Bar.cdc"),
				Code: []byte(`
				import Foo from 0x01
				access(all) contract Bar {
					access(all) fun test() {}
					init() {
						Foo.test()
					}
				}`),
			},
		})

		var validatorErr *stagingValidatorError
		require.ErrorAs(t, err, &validatorErr)
		require.Equal(t, 2, len(validatorErr.Unwrap()))

		// check that error exists & ensure that the local contract names are used (not the deploy locations)
		require.ErrorContains(t, validatorErr.Unwrap()[0], "mismatched types")
		require.ErrorContains(t, validatorErr.Unwrap()[0], "Foo.cdc")
		require.ErrorContains(t, validatorErr.Unwrap()[1], "mismatched types")
		require.ErrorContains(t, validatorErr.Unwrap()[1], "Foo.cdc")
		require.ErrorContains(t, validatorErr.Unwrap()[1], "Bar.cdc")
	})

	t.Run("cached downstream missing dependency errors", func(t *testing.T) {
		// setup mocks
		// ordering is important here, e.g. even though Foo is checked
		// first, Bar will still recognize the missing dependency
		srv := setupValidatorMocks(t, []mockAccount{
			{
				address: flow.HexToAddress("01"),
				contracts: map[string][]byte{"Foo": []byte(`
				pub contract Foo {
					pub fun test() {}
					init() {}
				}`)},
				stagedContracts: nil,
			},
			{
				address: flow.HexToAddress("02"),
				contracts: map[string][]byte{"Bar": []byte(`
				pub contract Bar {
					pub fun test() {}
					init() {
						Foo.test()
					}
				}`)},
				stagedContracts: nil,
			},
		})

		validator := newStagingValidator(srv)
		err := validator.Validate([]StagedContract{
			{
				DeployLocation: simpleAddressLocation("0x01.Foo"),
				SourceLocation: common.StringLocation("./Foo.cdc"),
				Code: []byte(`
				import ImpContract from 0x03
				access(all) contract Foo {
					access(all) fun test() {}
					init() {}
				}`),
			},
			{
				DeployLocation: simpleAddressLocation("0x02.Bar"),
				SourceLocation: common.StringLocation("./Bar.cdc"),
				Code: []byte(`
				import Foo from 0x01
				access(all) contract Bar {
					access(all) fun test() {}
					init() {
						Foo.test()
					}
				}`),
			},
		})

		var validatorErr *stagingValidatorError
		require.ErrorAs(t, err, &validatorErr)
		require.Equal(t, 2, len(validatorErr.Unwrap()))

		var missingDependenciesErr *missingDependenciesError
		require.ErrorAs(t, validatorErr.Unwrap()[0], &missingDependenciesErr)
		require.Equal(t, 1, len(missingDependenciesErr.MissingContracts))
		require.Equal(t, simpleAddressLocation("0x03.ImpContract"), missingDependenciesErr.MissingContracts[0])

		require.ErrorAs(t, validatorErr.Unwrap()[1], &missingDependenciesErr)
		require.Equal(t, 1, len(missingDependenciesErr.MissingContracts))
		require.Equal(t, simpleAddressLocation("0x03.ImpContract"), missingDependenciesErr.MissingContracts[0])
	})
}
