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

package resolvers_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/onflow/cadence"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-cli/pkg/flowkit/resolvers"
)

type testContract struct {
	location             string
	code                 []byte
	accountAddress       flow.Address
	accountName          string
	expectedDependencies []testContract
}

var addresses = test.AddressGenerator()

var testContractA = testContract{
	location:             "ContractA.cdc",
	code:                 []byte(`pub contract ContractA {}`),
	accountAddress:       addresses.New(),
	expectedDependencies: nil,
}

var testContractB = testContract{
	location:             "ContractB.cdc",
	code:                 []byte(`pub contract ContractB {}`),
	accountAddress:       addresses.New(),
	expectedDependencies: nil,
}

var testContractC = testContract{
	location: "ContractC.cdc",
	code: []byte(`
        import ContractA from "ContractA.cdc"
    
        pub contract ContractC {}
    `),
	accountAddress:       addresses.New(),
	expectedDependencies: []testContract{testContractA},
}

var testContractD = testContract{
	location: "ContractD.cdc",
	code: []byte(`
        import ContractC from "ContractC.cdc"

        pub contract ContractD {}
    `),
	accountAddress:       addresses.New(),
	expectedDependencies: []testContract{testContractC},
}

var testContractE = testContract{
	location: "ContractE.cdc",
	code: []byte(`
        import ContractF from "ContractF.cdc"

        pub contract ContractE {}
    `),
}

var testContractF = testContract{
	location: "ContractF.cdc",
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
	location: "ContractG.cdc",
	code: []byte(`
        import ContractA from "ContractA.cdc"
        import ContractB from "ContractB.cdc"

        pub contract ContractG {}
    `),
	accountAddress:       addresses.New(),
	expectedDependencies: []testContract{testContractA, testContractB},
}

var testContractH = testContract{
	location: "ContractH.cdc",
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
	case testContractA.location:
		return testContractA.code, nil
	case testContractB.location:
		return testContractB.code, nil
	case testContractC.location:
		return testContractC.code, nil
	case testContractD.location:
		return testContractD.code, nil
	case testContractE.location:
		return testContractE.code, nil
	case testContractF.location:
		return testContractF.code, nil
	case testContractG.location:
		return testContractG.code, nil
	case testContractH.location:
		return testContractH.code, nil
	}

	return nil, fmt.Errorf("failed to load %s", source)
}

func (t testLoader) Normalize(base, relative string) string {
	return relative
}

func contractBySource(all *resolvers.Deployments, source string) *resolvers.Contract {
	for _, c := range all.Contracts() {
		if c.Location() == source {
			return c
		}
	}

	return nil
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
			expectedDeploymentError: &resolvers.CyclicImportError{},
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
			c := resolvers.NewDeployments(testLoader{}, noAliases)

			for _, contract := range testCase.contracts {
				_, err := c.Add(
					contract.location,
					contract.accountAddress,
					contract.accountName,
					[]cadence.Value{nil},
				)
				assert.NoError(t, err)
			}

			err := c.Sort()
			if !strings.Contains(testCase.name, "unresolved") && !strings.Contains(testCase.name, "cycle") {
				assert.NoError(t, err)
			}

			for _, sourceContract := range testCase.contracts {

				contract := contractBySource(c, sourceContract.location)
				require.NotNil(t, sourceContract)

				require.Equal(
					t,
					len(sourceContract.expectedDependencies),
					len(contract.Dependencies()),
					"resolved dependency count does not match expected count",
				)

				for _, dependency := range sourceContract.expectedDependencies {
					require.Contains(t, contract.Dependencies(), dependency.location)

					contractDependency := contractBySource(c, dependency.location)
					require.NotNil(t, contractDependency)

					assert.Equal(t, contract.Dependencies()[dependency.location], contractDependency)

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
			c := resolvers.NewDeployments(testLoader{}, noAliases)

			for _, contract := range testCase.contracts {
				_, err := c.Add(
					contract.location,
					contract.accountAddress,
					contract.accountName,
					[]cadence.Value{nil},
				)
				assert.NoError(t, err)
			}

			err := c.Sort()
			if !strings.Contains(testCase.name, "unresolved") && !strings.Contains(testCase.name, "cycle") {
				assert.NoError(t, err)
			}

			if strings.Contains(testCase.name, "unresolved") {
				assert.EqualError(t, err, "import from ContractH could not be found: Foo.cdc, make sure import path is correct.")
				return
			}

			if testCase.expectedDeploymentError != nil {
				assert.IsType(t, testCase.expectedDeploymentError, err)
				return
			} else {
				assert.NoError(t, err, testCase.name)
			}

			require.Equal(
				t,
				len(testCase.expectedDeploymentOrder),
				len(c.Contracts()),
				"deployed contract count does not match expected count",
			)

			for i, deployedContract := range c.Contracts() {
				assert.Equal(t, testCase.expectedDeploymentOrder[i].location, deployedContract.Location())
			}
		})
	}
}
