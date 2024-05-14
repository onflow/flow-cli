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

type mockNetworkAccount struct {
	address         flow.Address
	contracts       map[string][]byte
	stagedContracts map[string][]byte
}

func setupValidatorMocks(
	t *testing.T,
	accounts []mockNetworkAccount,
) *flowkitMocks.Services {
	t.Helper()
	srv := flowkitMocks.NewServices(t)

	// Mock all accounts & staged contracts
	for _, acc := range accounts {
		mockAcct := &flow.Account{
			Address:   acc.address,
			Balance:   1000,
			Keys:      nil,
			Contracts: acc.contracts,
		}

		srv.On("GetAccount", mock.Anything, acc.address).Return(mockAcct, nil).Maybe()

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

	// Mock trying to get staged contract code for a contract that doesn't exist
	// This is the fallback mock for all other staged contract code requests
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

			if callContractAddress.Type() != cadence.AddressType {
				return false
			}

			if callContractName.Type() != cadence.StringType {
				return false
			}

			return true
		}),
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
		location := simpleAddressLocation("0x01.Test")
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
		srv := setupValidatorMocks(t, []mockNetworkAccount{
			{
				address:         flow.HexToAddress("01"),
				contracts:       map[string][]byte{"Test": []byte(oldContract)},
				stagedContracts: nil,
			},
		})

		validator := newStagingValidator(srv)
		err := validator.Validate([]stagedContractUpdate{{location, sourceCodeLocation, []byte(newContract)}})
		require.NoError(t, err)
	})

	t.Run("contract update with update error", func(t *testing.T) {
		location := simpleAddressLocation("0x01.Test")
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
		srv := setupValidatorMocks(t, []mockNetworkAccount{
			{
				address:         flow.HexToAddress("01"),
				contracts:       map[string][]byte{"Test": []byte(oldContract)},
				stagedContracts: nil,
			},
		})

		validator := newStagingValidator(srv)
		err := validator.Validate([]stagedContractUpdate{{location, sourceCodeLocation, []byte(newContract)}})

		var validatorErr *stagingValidatorError
		require.ErrorAs(t, err, &validatorErr)

		var updateErr *stdlib.ContractUpdateError
		require.ErrorAs(t, validatorErr.errors[simpleAddressLocation("0x01.Test")], &updateErr)
	})

	t.Run("contract update with checker error", func(t *testing.T) {
		location := simpleAddressLocation("0x01.Test")
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
		srv := setupValidatorMocks(t, []mockNetworkAccount{
			{
				address:         flow.HexToAddress("01"),
				contracts:       map[string][]byte{"Test": []byte(oldContract)},
				stagedContracts: nil,
			},
		})

		validator := newStagingValidator(srv)
		err := validator.Validate([]stagedContractUpdate{{location, sourceCodeLocation, []byte(newContract)}})

		var validatorErr *stagingValidatorError
		require.ErrorAs(t, err, &validatorErr)

		var checkerErr *sema.CheckerError
		require.ErrorAs(t, validatorErr.errors[simpleAddressLocation("0x01.Test")], &checkerErr)
	})

	t.Run("valid contract update with dependencies", func(t *testing.T) {
		location := simpleAddressLocation("0x01.Test")
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
		srv := setupValidatorMocks(t, []mockNetworkAccount{
			{
				address:   flow.HexToAddress("01"),
				contracts: map[string][]byte{"Test": []byte(oldContract)},
			},
			{
				address:         flow.HexToAddress("02"),
				stagedContracts: map[string][]byte{"ImpContract": []byte(impContract)},
			},
		})

		// validate
		validator := newStagingValidator(srv)
		err := validator.Validate([]stagedContractUpdate{{location, sourceCodeLocation, []byte(newContract)}})
		require.NoError(t, err)
	})

	t.Run("contract update missing dependency", func(t *testing.T) {
		location := simpleAddressLocation("0x01.Test")
		impLocation := simpleAddressLocation("0x02.ImpContract")
		sourceCodeLocation := common.StringLocation("./Test.cdc")
		oldContract := `
		pub contract Test {
			pub fun test() {}
		}`
		newContract := `
		// staged contract does not exist
		import ImpContract from 0x02
		access(all) contract Test {
			access(all) fun test() {}
		}`

		// setup mocks
		srv := setupValidatorMocks(t, []mockNetworkAccount{
			{
				address:   flow.HexToAddress("01"),
				contracts: map[string][]byte{"Test": []byte(oldContract)},
			},
			{
				address: flow.HexToAddress("02"),
			},
		})

		validator := newStagingValidator(srv)
		err := validator.Validate([]stagedContractUpdate{{location, sourceCodeLocation, []byte(newContract)}})

		var validatorErr *stagingValidatorError
		require.ErrorAs(t, err, &validatorErr)
		require.Equal(t, 1, len(validatorErr.errors))

		var missingDependenciesErr *missingDependenciesError
		require.ErrorAs(t, validatorErr.errors[simpleAddressLocation("0x01.Test")], &missingDependenciesErr)
		require.Equal(t, 1, len(missingDependenciesErr.MissingContracts))
		require.Equal(t, impLocation, missingDependenciesErr.MissingContracts[0])
	})

	t.Run("valid contract update with system contract imports", func(t *testing.T) {
		location := simpleAddressLocation("0x01.Test")
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
		srv := setupValidatorMocks(t, []mockNetworkAccount{
			{
				address:   flow.HexToAddress("01"),
				contracts: map[string][]byte{"Test": []byte(oldContract)},
			},
		})

		validator := newStagingValidator(srv)
		err := validator.Validate([]stagedContractUpdate{{location, sourceCodeLocation, []byte(newContract)}})
		require.NoError(t, err)
	})

	t.Run("resolves account access correctly", func(t *testing.T) {
		// setup mocks
		srv := setupValidatorMocks(t, []mockNetworkAccount{
			{
				address: flow.HexToAddress("01"),
				contracts: map[string][]byte{
					"Test": []byte(`
					import ImpContract from 0x01
					pub contract Test {
						pub fun test() {}
					}`),
					"Imp2": []byte(`
					pub contract Imp2 {
						access(account) fun test() {}
					}`),
				},
				stagedContracts: map[string][]byte{"ImpContract": []byte(`
				access(all) contract ImpContract {
					access(account) fun test() {}
					init() {}
				}`)},
			},
		})

		// validate
		validator := newStagingValidator(srv)
		err := validator.Validate([]stagedContractUpdate{
			{
				simpleAddressLocation("0x01.Test"),
				common.StringLocation("./Test.cdc"),
				[]byte(`
			import ImpContract from 0x01
			import Imp2 from 0x01
			access(all) contract Test {
				access(all) fun test() {}
				init() {
					ImpContract.test()
					Imp2.test()
				}
			}`),
			},
			{
				simpleAddressLocation("0x01.Imp2"),
				common.StringLocation("./Imp2.cdc"),
				[]byte(`
			access(all) contract Imp2 {
				access(account) fun test() {}
			}`),
			},
		})
		require.NoError(t, err)
	})

	t.Run("validates multiple contracts, no error", func(t *testing.T) {
		// setup mocks
		srv := setupValidatorMocks(t, []mockNetworkAccount{
			{
				address: flow.HexToAddress("01"),
				contracts: map[string][]byte{"Foo": []byte(`
				pub contract Foo {
					pub fun test() {}
				}`)},
			},
			{
				address: flow.HexToAddress("02"),
				contracts: map[string][]byte{"Bar": []byte(`
				import Foo from 0x01
				pub contract Bar {
					pub fun test() {}
					init() {
						Foo.test()
					}
				}`)},
			},
		})

		validator := newStagingValidator(srv)
		err := validator.Validate([]stagedContractUpdate{
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
		srv := setupValidatorMocks(t, []mockNetworkAccount{
			{
				address: flow.HexToAddress("01"),
				contracts: map[string][]byte{"Foo": []byte(`
				pub contract Foo {
					pub fun test() {}
					init() {}
				}`)},
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
			},
		})

		validator := newStagingValidator(srv)
		err := validator.Validate([]stagedContractUpdate{
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

		require.Equal(t, 2, len(validatorErr.errors))

		// check that error exists & ensure that the local contract names are used (not the deploy locations)
		fooErr := validatorErr.errors[simpleAddressLocation("0x01.Foo")]
		require.ErrorContains(t, fooErr, "mismatched types")
		require.ErrorContains(t, fooErr, "Foo.cdc")

		// Bar should have an error related to
		var upstreamErr *upstreamValidationError
		require.ErrorAs(t, validatorErr.errors[simpleAddressLocation("0x02.Bar")], &upstreamErr)
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
		srv := setupValidatorMocks(t, []mockNetworkAccount{
			{
				address:         flow.HexToAddress("01"),
				contracts:       map[string][]byte{"Test": []byte(oldContract)},
				stagedContracts: map[string][]byte{"ImpContract": []byte(impContract)},
			},
		})

		// validate
		validator := newStagingValidator(srv)
		err := validator.Validate([]stagedContractUpdate{{location, sourceCodeLocation, []byte(newContract)}})
		require.NoError(t, err)
	})

	t.Run("validates multiple contracts, no error", func(t *testing.T) {
		// setup mocks
		srv := setupValidatorMocks(t, []mockNetworkAccount{
			{
				address: flow.HexToAddress("01"),
				contracts: map[string][]byte{"Foo": []byte(`
				pub contract Foo {
					pub fun test() {}
				}`)},
			},
			{
				address: flow.HexToAddress("02"),
				contracts: map[string][]byte{"Bar": []byte(`
				import Foo from 0x01
				pub contract Bar {
					pub fun test() {}
					init() {
						Foo.test()
					}
				}`)},
			},
		})

		validator := newStagingValidator(srv)
		err := validator.Validate([]stagedContractUpdate{
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

	t.Run("validates cyclic imports", func(t *testing.T) {
		// setup mocks
		srv := setupValidatorMocks(t, []mockNetworkAccount{
			{
				address: flow.HexToAddress("01"),
				contracts: map[string][]byte{"Foo": []byte(`
				pub contract Foo {
					pub fun test() {}
					init() {}
				}`)},
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
			},
		})

		validator := newStagingValidator(srv)
		err := validator.Validate([]stagedContractUpdate{
			{
				DeployLocation: simpleAddressLocation("0x01.Foo"),
				SourceLocation: common.StringLocation("./Foo.cdc"),
				Code: []byte(`
				import Bar from 0x02
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

		require.Equal(t, 2, len(validatorErr.errors))

		// check that error exists & ensure that the local contract names are used (not the deploy locations)
		var cyclicImportError *sema.CyclicImportsError
		require.ErrorAs(t, validatorErr.errors[simpleAddressLocation("0x01.Foo")], &cyclicImportError)
		require.ErrorAs(t, validatorErr.errors[simpleAddressLocation("0x02.Bar")], &cyclicImportError)
	})

	t.Run("upstream missing dependency errors", func(t *testing.T) {
		// setup mocks
		srv := setupValidatorMocks(t, []mockNetworkAccount{
			{
				address: flow.HexToAddress("01"),
				contracts: map[string][]byte{"Foo": []byte(`
				import ImpContract from 0x03
				pub contract Foo {
					pub fun test() {}
					init() {}
				}`)},
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
			},
			{
				address: flow.HexToAddress("03"),
				contracts: map[string][]byte{"ImpContract": []byte(`
				pub contract ImpContract {}
				`)},
			},
			{
				address: flow.HexToAddress("04"),
				contracts: map[string][]byte{"AnotherImp": []byte(`
				pub contract AnotherImp {}
				`)},
			},
		})

		validator := newStagingValidator(srv)

		// ordering is important here, e.g. even though Foo is checked
		// first, Bar will still recognize the missing dependency
		err := validator.Validate([]stagedContractUpdate{
			{
				DeployLocation: simpleAddressLocation("0x01.Foo"),
				SourceLocation: common.StringLocation("./Foo.cdc"),
				Code: []byte(`
				// staged contract does not exist
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
				import AnotherImp from 0x04
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
		require.Equal(t, 2, len(validatorErr.errors))

		var missingDependenciesErr *missingDependenciesError
		require.ErrorAs(t, validatorErr.errors[simpleAddressLocation("0x01.Foo")], &missingDependenciesErr)
		require.Equal(t, 1, len(missingDependenciesErr.MissingContracts))
		require.Equal(t, simpleAddressLocation("0x03.ImpContract"), missingDependenciesErr.MissingContracts[0])

		require.ErrorAs(t, validatorErr.errors[simpleAddressLocation("0x02.Bar")], &missingDependenciesErr)
		require.Equal(t, 2, len(missingDependenciesErr.MissingContracts))
		require.ElementsMatch(t, []common.AddressLocation{
			simpleAddressLocation("0x03.ImpContract"),
			simpleAddressLocation("0x04.AnotherImp"),
		}, missingDependenciesErr.MissingContracts)
	})

	t.Run("import Crypto checker", func(t *testing.T) {
		// setup mocks
		srv := setupValidatorMocks(t, []mockNetworkAccount{
			{
				address: flow.HexToAddress("01"),
				contracts: map[string][]byte{"Foo": []byte(`
				import Crypto
				pub contract Foo {
					init() {}
				}`)},
			},
		})

		validator := newStagingValidator(srv)
		err := validator.Validate([]stagedContractUpdate{
			{
				DeployLocation: simpleAddressLocation("0x01.Foo"),
				SourceLocation: common.StringLocation("./Foo.cdc"),
				Code: []byte(`
				import Crypto
				access(all) contract Foo {
					init() {}
				}`),
			},
		})

		require.Nil(t, err)
	})
}
