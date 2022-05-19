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

package contracts_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/onflow/cadence"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-cli/pkg/flowkit/contracts"
)

type testContract struct {
	name                 string
	source               string
	code                 []byte
	accountAddress       flow.Address
	accountName          string
	expectedDependencies []testContract
}

var addresses = test.AddressGenerator()

var testContractA = testContract{
	name:                 "ContractA",
	source:               "ContractA.cdc",
	code:                 []byte(`pub contract ContractA {}`),
	accountAddress:       addresses.New(),
	expectedDependencies: nil,
}

var testContractB = testContract{
	name:                 "ContractB",
	source:               "ContractB.cdc",
	code:                 []byte(`pub contract ContractB {}`),
	accountAddress:       addresses.New(),
	expectedDependencies: nil,
}

var testContractC = testContract{
	name:   "ContractC",
	source: "ContractC.cdc",
	code: []byte(`
        import ContractA from "ContractA.cdc"
    
        pub contract ContractC {}
    `),
	accountAddress:       addresses.New(),
	expectedDependencies: []testContract{testContractA},
}

var testContractD = testContract{
	name:   "ContractD",
	source: "ContractD.cdc",
	code: []byte(`
        import ContractC from "ContractC.cdc"

        pub contract ContractD {}
    `),
	accountAddress:       addresses.New(),
	expectedDependencies: []testContract{testContractC},
}

var testContractE = testContract{
	name:   "ContractE",
	source: "ContractE.cdc",
	code: []byte(`
        import ContractF from "ContractF.cdc"

        pub contract ContractE {}
    `),
}

var testContractF = testContract{
	name:   "ContractF",
	source: "ContractF.cdc",
	code: []byte(`
        import ContractE from "ContractE.cdc"

        pub contract ContractF {}
    `),
	accountAddress: addresses.New(),
}

func init() {
	// create import cycle, cannot be done statically
	testContractE.expectedDependencies = []testContract{testContractF}
	testContractF.expectedDependencies = []testContract{testContractE}
}

var testContractG = testContract{
	name:   "ContractG",
	source: "ContractG.cdc",
	code: []byte(`
        import ContractA from "ContractA.cdc"
        import ContractB from "ContractB.cdc"

        pub contract ContractG {}
    `),
	accountAddress:       addresses.New(),
	expectedDependencies: []testContract{testContractA, testContractB},
}

var testContractH = testContract{
	name:   "ContractH",
	source: "ContractH.cdc",
	code: []byte(`
        import ContractFoo from "Foo.cdc"

        pub contract ContractH {}
    `),
	accountAddress:       addresses.New(),
	expectedDependencies: nil,
}

var noAliases = map[string]string{}

type testLoader struct{}

func (t testLoader) Load(source string) ([]byte, error) {
	switch source {
	case testContractA.source:
		return testContractA.code, nil
	case testContractB.source:
		return testContractB.code, nil
	case testContractC.source:
		return testContractC.code, nil
	case testContractD.source:
		return testContractD.code, nil
	case testContractE.source:
		return testContractE.code, nil
	case testContractF.source:
		return testContractF.code, nil
	case testContractG.source:
		return testContractG.code, nil
	case testContractH.source:
		return testContractH.code, nil
	}

	return nil, fmt.Errorf("failed to load %s", source)
}

func (t testLoader) Normalize(base, relative string) string {
	return relative
}

type contractTestCase struct {
	name                    string
	contracts               []testContract
	expectedDeploymentOrder []testContract
	expectedDeploymentError error
}

func getTestCases() []contractTestCase {
	return []contractTestCase{
		{
			name:                    "No contracts",
			contracts:               []testContract{},
			expectedDeploymentOrder: []testContract{},
		},
		{
			name:                    "Single contract no imports",
			contracts:               []testContract{testContractA},
			expectedDeploymentOrder: []testContract{testContractA},
		},
		{
			name:                    "Two contracts no imports",
			contracts:               []testContract{testContractA, testContractB},
			expectedDeploymentOrder: []testContract{testContractA, testContractB},
		},
		{
			name:                    "Two contracts with imports",
			contracts:               []testContract{testContractA, testContractC},
			expectedDeploymentOrder: []testContract{testContractA, testContractC},
		},
		{
			name:                    "Three contracts with imports",
			contracts:               []testContract{testContractA, testContractC, testContractD},
			expectedDeploymentOrder: []testContract{testContractA, testContractC, testContractD},
		},
		{
			name:                    "Two contracts with import cycle",
			contracts:               []testContract{testContractE, testContractF},
			expectedDeploymentError: &contracts.CyclicImportError{},
		},
		{
			name:                    "Single contract with two imports",
			contracts:               []testContract{testContractA, testContractB, testContractG},
			expectedDeploymentOrder: []testContract{testContractA, testContractB, testContractG},
		},
		{
			name:                    "Single contract with unresolved import",
			contracts:               []testContract{testContractH},
			expectedDeploymentOrder: []testContract{testContractH},
		},
	}
}

func TestResolveImports(t *testing.T) {
	testCases := getTestCases()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			p := contracts.NewPreprocessor(testLoader{}, noAliases)

			for _, contract := range testCase.contracts {
				err := p.AddContractSource(
					contract.name,
					contract.source,
					contract.accountAddress,
					contract.accountName,
					[]cadence.Value{nil},
				)
				assert.NoError(t, err)
			}

			err := p.ResolveImports()
			if !strings.Contains(testCase.name, "unresolved") {
				assert.NoError(t, err)
			}

			for _, sourceContract := range testCase.contracts {

				contract := p.ContractBySource(sourceContract.source)
				require.NotNil(t, contract)

				require.Equal(
					t,
					len(sourceContract.expectedDependencies),
					len(contract.Dependencies()),
					"resolved dependency count does not match expected count",
				)

				for _, dependency := range sourceContract.expectedDependencies {
					require.Contains(t, contract.Dependencies(), dependency.source)

					contractDependency := p.ContractBySource(dependency.source)
					require.NotNil(t, contractDependency)

					assert.Equal(t, contract.Dependencies()[dependency.source], contractDependency)

					assert.Contains(t, contract.TranspiledCode(), dependency.accountAddress.Hex())
				}
			}
		})
	}
}

func TestContractDeploymentOrder(t *testing.T) {
	testCases := getTestCases()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			p := contracts.NewPreprocessor(testLoader{}, noAliases)

			for _, contract := range testCase.contracts {
				err := p.AddContractSource(
					contract.name,
					contract.source,
					contract.accountAddress,
					contract.accountName,
					[]cadence.Value{nil},
				)
				assert.NoError(t, err)
			}

			err := p.ResolveImports()
			if !strings.Contains(testCase.name, "unresolved") {
				assert.NoError(t, err)
			}

			deployedContracts, err := p.ContractDeploymentOrder()

			if testCase.expectedDeploymentError != nil {
				assert.IsType(t, testCase.expectedDeploymentError, err)
				return
			} else {
				assert.NoError(t, err)
			}

			require.Equal(
				t,
				len(testCase.expectedDeploymentOrder),
				len(deployedContracts),
				"deployed contract count does not match expected count",
			)

			for i, deployedContract := range deployedContracts {
				assert.Equal(t, testCase.expectedDeploymentOrder[i].name, deployedContract.Name())
			}
		})
	}
}
