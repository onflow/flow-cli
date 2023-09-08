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

package project

import (
	"strings"
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testContract struct {
	location       string
	code           []byte
	accountAddress flow.Address
	accountName    string
}

var addresses = test.AddressGenerator()

var testContractA = testContract{
	location:       "foobar/ContractA.cdc",
	code:           []byte(`pub contract ContractA {}`),
	accountAddress: addresses.New(),
}

var testContractB = testContract{
	location:       "ContractB.cdc",
	code:           []byte(`pub contract ContractB {}`),
	accountAddress: addresses.New(),
}

var testContractC = testContract{
	location: "foobar/ContractC.cdc",
	code: []byte(`
        import ContractA from "ContractA.cdc"
    
        pub contract ContractC {}
    `),
	accountAddress: addresses.New(),
}

var testContractD = testContract{
	location: "ContractD.cdc",
	code: []byte(`
        import ContractC from "foobar/ContractC.cdc"

        pub contract ContractD {}
    `),
	accountAddress: addresses.New(),
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

var testContractG = testContract{
	location: "ContractG.cdc",
	code: []byte(`
        import ContractA from "foobar/ContractA.cdc"
        import ContractB from "ContractB.cdc"

        pub contract ContractG {}
    `),
	accountAddress: addresses.New(),
}

var testContractH = testContract{
	location: "ContractH.cdc",
	code: []byte(`
        import ContractFoo from "Foo.cdc"

        pub contract ContractH {}
    `),
	accountAddress: addresses.New(),
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
			expectedDeploymentError: &CyclicImportError{},
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

func TestContractDeploymentOrder(t *testing.T) {
	testCases := getTestCases()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			contracts := make([]*Contract, len(testCase.contracts))
			for i, contract := range testCase.contracts {
				contracts[i] = NewContract(
					strings.Split(contract.location, ".")[0],
					contract.location,
					contract.code,
					contract.accountAddress,
					contract.accountName,
					nil,
				)
			}

			deployment, err := NewDeployment(contracts, nil)

			contracts, err = deployment.Sort()
			if !strings.Contains(testCase.name, "unresolved") && !strings.Contains(testCase.name, "cycle") {
				assert.NoError(t, err)
			}

			if strings.Contains(testCase.name, "unresolved") {
				assert.EqualError(t, err, "import from ContractH could not be found: Foo.cdc, make sure import path is correct, and the contract is added to deployments or has an alias")
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
				len(contracts),
				"deployed contract count does not match expected count",
			)

			for i, deployedContract := range contracts {
				assert.Equal(t, testCase.expectedDeploymentOrder[i].location, deployedContract.Location())
			}
		})
	}
}
