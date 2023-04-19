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

package args

import (
	"fmt"
	"testing"

	"github.com/onflow/cadence"
	"github.com/stretchr/testify/assert"
)

func Test_WithoutType(t *testing.T) {
	sampleValues := []cadence.Value{
		cadence.NewAddress([8]byte{0, 0, 0, 0, 0, 0, 0, 1}),
		cadence.NewBool(true),
		cadence.NewInt(-42),
		cadence.NewInt128(-424242),
		cadence.NewInt32(-42),
		cadence.NewInt64(-42),
		cadence.NewInt8(-42),
		cadence.NewUInt128(424242),
		cadence.NewUInt32(42),
		cadence.NewUInt64(42),
		cadence.NewUInt8(42),
		cadence.String("42"),
	}

	for _, sample := range sampleValues {

		sampleType := sample.Type().ID()

		args, err := ParseWithoutType([]string{sample.String()}, []byte(fmt.Sprintf(`pub fun main(test: %s): Void {}`, sampleType)), "")
		assert.NoError(t, err)
		assert.Len(t, args, 1)
		assert.Equal(t, []cadence.Value{sample}, args)
	}
}

func Test_WithoutTypeContracts(t *testing.T) {
	template := []string{
		`pub fun main(foo: String): Void {}`,
		`pub contract Foo { init(foo: String) {} }`,
		`transaction(foo: String) {}`,
	}

	for _, tmp := range template {
		args, err := ParseWithoutType([]string{"hello"}, []byte(tmp), "")
		assert.NoError(t, err)
		assert.Len(t, args, 1)
		v, _ := cadence.NewString("hello")
		assert.Equal(t, []cadence.Value{v}, args)
	}

}

func Test_ParseJSON(t *testing.T) {
	jsonInput := `[{"type": "String", "value": "Hello World"}]`

	values, err := ParseJSON(jsonInput)
	assert.NoError(t, err)
	assert.Len(t, values, 1)
	assert.Equal(t, `"Hello World"`, values[0].String())
	assert.Equal(t, "String", values[0].Type().ID())
}
