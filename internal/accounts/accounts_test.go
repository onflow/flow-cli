/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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
	"encoding/json"
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/test"
)

// One way to handle testing a result serializes as expected - convert it to a map[string]interface{} and make
// many type assertions
func TestAccountsResultJSON(t *testing.T) {

	testCases := []struct {
		Name    string
		Account *flow.Account
	}{
		{Name: "Generated Account", Account: test.AccountGenerator().New()},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			account := tc.Account
			result := AccountResult{
				Account: tc.Account,
				include: []string{"contracts"},
			}
			out := result.JSON()
			EvaluateAccountJson(t, out, account)
		})
	}
}

func EvaluateAccountJson(t *testing.T, out interface{}, account *flow.Account) {

	// we expect output to be marshallable to json
	jsonBytes, err := json.Marshal(out)
	assert.NoError(t, err)

	// should be able to round trip back to map[string]interface{}
	var jsonObject map[string]interface{}
	err = json.Unmarshal(jsonBytes, &jsonObject)
	assert.NoError(t, err)

	// expect the account address should exist and match the initial address
	address, has := jsonObject["address"]
	assert.True(t, has, "expected the account address to be present on the output as 'address'")
	assert.Equal(t, address, account.Address.String())

	balanceRaw, has := jsonObject["balance"]
	assert.True(t, has, "expected the account balance to be present on the output as 'balance'")
	balanceStr, valid := balanceRaw.(string)
	assert.True(t, valid, "expected balance to be expressed as a string in json output")

	// we should be able to parse the balance
	fixedPoint, err := cadence.NewUFix64(balanceStr)
	assert.NoError(t, err)
	assert.Equal(t, uint64(fixedPoint), account.Balance)

	// expect public keys to be present
	keysRaw, has := jsonObject["keys"]
	assert.True(t, has, "expected account keys to be present on the output as 'keys'")

	valueSlice, valid := keysRaw.([]interface{})
	assert.True(t, valid, "expected keys to be a slice in json output")

	for i, accountKey := range account.Keys {
		pubKeyAsStr, valid := valueSlice[i].(string)
		assert.True(t, valid, "expected public keys to be encoded as strings")

		// decode using same algo as source address key
		decoded, err := crypto.DecodePublicKeyHex(accountKey.PublicKey.Algorithm(), pubKeyAsStr)
		assert.NoError(t, err)
		assert.Equal(t, decoded, accountKey.PublicKey)
	}

	// expect contract names to be present
	contractsRaw, has := jsonObject["contracts"]
	assert.True(t, has, "expected deployed contracts keys to be present on the output as 'contracts'")

	valueSlice, valid = contractsRaw.([]interface{})
	assert.True(t, valid, "expected contracts to be a slice in json output")

	// assert contract names are strings
	contractNames := make([]string, len(valueSlice))
	for i, val := range valueSlice {
		contractNameAsStr, valid := val.(string)
		assert.True(t, valid, "expected contract names to be encoded as strings")
		contractNames[i] = contractNameAsStr
	}
	// all contract names on the account should be present
	for name := range account.Contracts {
		assert.Contains(t, contractNames, name)
	}

	// expect contract code to be present ('contracts' flag was passed)
	codeRaw, has := jsonObject["code"]
	assert.True(t, has, "expected contract code to be present if 'contracts' flag is passed")

	codeMap, valid := codeRaw.(map[string]interface{})
	assert.True(t, valid, "expected contract code to be marshalled as a json object")

	for key, rawCode := range codeMap {
		codeStr, valid := rawCode.(string)
		assert.True(t, valid, "expected contract code to be encoded as strings")
		assert.Equal(t, account.Contracts[key], codeStr)
	}
}
