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

package json

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ConfigDependencies(t *testing.T) {
	b := []byte(`{
		"HelloWorld": "testnet://877931736ee77cff.HelloWorld"
		}`)

	var jsonDependencies jsonDependencies
	err := json.Unmarshal(b, &jsonDependencies)
	assert.NoError(t, err)

	dependencies, err := jsonDependencies.transformToConfig()
	assert.NoError(t, err)

	assert.Len(t, dependencies, 1)

	dependencyOne := dependencies.ByName("HelloWorld")
	assert.NotNil(t, dependencyOne)

	assert.NotNil(t, dependencyOne)
}

func Test_TransformDependenciesToJSON(t *testing.T) {
	b := []byte(`{
		"HelloWorld": "testnet://877931736ee77cff.HelloWorld"
	}`)

	bOut := []byte(`{
		"HelloWorld": {
			"remoteSource": "testnet://877931736ee77cff.HelloWorld",
			"aliases": {}
		}
	}`)

	var jsonContracts jsonContracts
	errContracts := json.Unmarshal(b, &jsonContracts)
	assert.NoError(t, errContracts)

	var jsonDependencies jsonDependencies
	err := json.Unmarshal(b, &jsonDependencies)
	assert.NoError(t, err)

	contracts, err := jsonContracts.transformToConfig()
	dependencies, err := jsonDependencies.transformToConfig()
	assert.NoError(t, err)

	j := transformDependenciesToJSON(dependencies, contracts)
	x, _ := json.Marshal(j)

	assert.Equal(t, cleanSpecialChars(bOut), cleanSpecialChars(x))
}
