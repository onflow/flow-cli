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

package arguments

import (
	"fmt"
	"testing"

	"github.com/onflow/cadence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ParseWithoutType(t *testing.T) {
	t.Parallel()

	t.Run("types", func(t *testing.T) {
		t.Parallel()

		type testCase struct {
			Name           string
			Literal        string
			InvalidLiteral string
			Value          cadence.Value
			Type           cadence.Type
		}

		testCases := []testCase{
			{
				Name:  "Address",
				Value: cadence.NewAddress([8]byte{0, 0, 0, 0, 0, 0, 0, 1}),
				Type:  cadence.AddressType{},
			},
			{
				Name:  "Bool",
				Value: cadence.NewBool(true),
				Type:  cadence.BoolType{},
			},
			{
				Name:  "Int",
				Value: cadence.NewInt(-42),
				Type:  cadence.IntType{},
			},
			{
				Name:  "Int8",
				Value: cadence.NewInt8(-42),
				Type:  cadence.Int8Type{},
			},
			{
				Name:  "Int16",
				Value: cadence.NewInt16(-42),
				Type:  cadence.Int16Type{},
			},
			{
				Name:  "Int32",
				Value: cadence.NewInt32(-42),
				Type:  cadence.Int32Type{},
			},
			{
				Name:  "Int64",
				Value: cadence.NewInt64(-42),
				Type:  cadence.Int64Type{},
			},
			{
				Name:  "Int128",
				Value: cadence.NewInt128(-42),
				Type:  cadence.Int128Type{},
			},
			{
				Name:  "UInt8",
				Value: cadence.NewUInt8(42),
				Type:  cadence.UInt8Type{},
			},
			{
				Name:  "UInt16",
				Value: cadence.NewUInt16(42),
				Type:  cadence.UInt16Type{},
			},
			{
				Name:  "UInt32",
				Value: cadence.NewUInt32(42),
				Type:  cadence.UInt32Type{},
			},
			{
				Name:  "UInt64",
				Value: cadence.NewUInt64(42),
				Type:  cadence.UInt64Type{},
			},
			{
				Name:  "UInt128",
				Value: cadence.NewUInt128(42),
				Type:  cadence.UInt128Type{},
			},
			{
				Name:  "String",
				Value: cadence.String("42"),
				Type:  cadence.StringType{},
			},
			{
				Name:    "String, no quoting",
				Literal: `foo`,
				Value:   cadence.String("foo"),
				Type:    cadence.StringType{},
			},
			{
				Name:  "optional String, nil",
				Value: cadence.NewOptional(nil),
				Type: &cadence.OptionalType{
					Type: cadence.StringType{},
				},
			},
			{
				Name: "optional String, value",
				Value: cadence.NewOptional(
					cadence.String("test"),
				),
				Type: &cadence.OptionalType{
					Type: cadence.StringType{},
				},
			},
			// TODO: depends on https://github.com/onflow/cadence/pull/2469
			//{
			//	Name: "doubly optional String, nil",
			//	Value: cadence.NewOptional(
			//		cadence.NewOptional(nil),
			//	),
			//	Type: &cadence.OptionalType{
			//		Type: &cadence.OptionalType{
			//			Type: cadence.StringType{},
			//		},
			//	},
			//},
			{
				Name: "variable-sized array",
				Value: cadence.NewArray([]cadence.Value{
					cadence.String("42"),
				}),
				Type: &cadence.VariableSizedArrayType{
					ElementType: cadence.StringType{},
				},
			},
			{
				Name: "constant-sized array",
				Value: cadence.NewArray([]cadence.Value{
					cadence.String("42"),
				}),
				Type: &cadence.ConstantSizedArrayType{
					ElementType: cadence.StringType{},
					Size:        1,
				},
			},
			{
				Name:           "identifier (invalid)",
				InvalidLiteral: "foo",
				Type:           cadence.IntType{},
			},
			{
				Name:           "expression (invalid)",
				InvalidLiteral: "1 + 1",
				Type:           cadence.IntType{},
			},
		}

		test := func(testCase testCase) {
			t.Run(testCase.Name, func(t *testing.T) {
				t.Parallel()

				literal := testCase.Literal
				if testCase.Value != nil {
					literal = testCase.Value.String()
				} else if len(testCase.InvalidLiteral) > 0 {
					literal = testCase.InvalidLiteral
				}

				args, err := ParseWithoutType(
					[]string{literal},
					[]byte(fmt.Sprintf(
						`pub fun main(test: %s) {}`,
						testCase.Type.ID(),
					)),
					"",
				)
				if len(testCase.InvalidLiteral) > 0 {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					assert.Equal(t,
						[]cadence.Value{testCase.Value},
						args,
					)
				}
			})
		}

		for _, testCase := range testCases {
			test(testCase)
		}
	})

	t.Run("entrypoints", func(t *testing.T) {
		t.Parallel()

		template := map[string]string{
			"script":      `pub fun main(foo: String): Void {}`,
			"contract":    `pub contract Foo { init(foo: String) {} }`,
			"transaction": `transaction(foo: String) {}`,
		}

		test := func(name string, code string) {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				args, err := ParseWithoutType([]string{"hello"}, []byte(code), "")
				assert.NoError(t, err)
				assert.Len(t, args, 1)
				v, _ := cadence.NewString("hello")
				assert.Equal(t, []cadence.Value{v}, args)
			})
		}

		for name, code := range template {
			test(name, code)
		}
	})
}

func Test_ParseJSON(t *testing.T) {
	t.Parallel()

	jsonInput := `[{"type": "String", "value": "Hello World"}]`

	values, err := ParseJSON(jsonInput)
	require.NoError(t, err)

	require.Len(t, values, 1)
	assert.Equal(t, `"Hello World"`, values[0].String())
	assert.Equal(t, "String", values[0].Type().ID())
}
