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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProgram(t *testing.T) {

	t.Run("AddressImports", func(t *testing.T) {
		tests := []struct {
			code          []byte
			expectedCount int
		}{
			{
				code: []byte(`
                import Foo from 0x123
                import "Bar"
                import FooSpace from 0x124
                import "BarSpace"

                access(all) contract Foo {}
            `),
				expectedCount: 2,
			},
		}

		for i, test := range tests {
			program, err := NewProgram(test.code, nil, "")
			require.NoError(t, err, fmt.Sprintf("AddressImports test %d failed", i))
			addressImports := program.AddressImportDeclarations()
			assert.Len(t, addressImports, test.expectedCount, fmt.Sprintf("AddressImports test %d failed", i))
		}
	})

	t.Run("Imports", func(t *testing.T) {
		tests := []struct {
			code    []byte
			imports []string
		}{{
			code:    []byte(`access(all) contract Foo {}`),
			imports: []string{},
		}, {
			code:    []byte(`access(all) fun main() {}`),
			imports: []string{},
		}, {
			code: []byte(`
				import Bar from "./Bar.cdc"
				access(all) contract Foo {}
			`),
			imports: []string{"./Bar.cdc"},
		}, {
			code: []byte(`
				import Bar from "./Bar.cdc"
				import Zoo from "./zoo/Zoo.cdc"
				access(all) contract Foo {}
			`),
			imports: []string{"./Bar.cdc", "./zoo/Zoo.cdc"},
		}, { // new schema import
			code: []byte(`
				import "Bar"
				access(all) contract Foo {}
			`),
			imports: []string{"Bar"},
		}, {
			code: []byte(`
				import "Bar"
				import Zoo from "./Zoo.cdc"
				import Crypto
				import Foo from 0x01

				access(all) contract Foo {}
			`),
			imports: []string{"Bar", "./Zoo.cdc"},
		}}

		for i, test := range tests {
			program, err := NewProgram(test.code, nil, "")
			require.NoError(t, err, fmt.Sprintf("import test %d failed", i))
			assert.Equal(t, len(test.imports) > 0, program.HasImports(), fmt.Sprintf("import test %d failed", i))
			assert.Equal(t, test.imports, program.imports(), fmt.Sprintf("import test %d failed", i))
		}
	})

	t.Run("Name", func(t *testing.T) {
		tests := []struct {
			code []byte
			name string
		}{{
			code: []byte(`access(all) contract Foo {}`),
			name: "Foo",
		}, {
			code: []byte(`
				import Bar from "./Bar.cdc"
				access(all) contract Foo {}
			`),
			name: "Foo",
		}}

		for i, test := range tests {
			program, err := NewProgram(test.code, nil, "")
			require.NoError(t, err, fmt.Sprintf("import test %d failed", i))
			name, err := program.Name()
			require.NoError(t, err)
			assert.Equal(t, test.name, name)
		}

		failed := [][]byte{
			[]byte(`
				access(all) contract Foo {}
				access(all) contract Bar {}
			`),
			[]byte(`
				access(all) contract Foo {}
				access(all) resource interface Test {}
			`),
			[]byte(`
				access(all) contract Foo {}
				struct Bar {}
			`),
		}

		for _, code := range failed {
			program, err := NewProgram(code, nil, "")
			require.NoError(t, err)

			_, err = program.Name()
			assert.EqualError(t, err, "the code must declare exactly one contract or contract interface")
		}

		program, err := NewProgram([]byte(`access(all) fun main() {}`), nil, "")
		require.NoError(t, err)
		_, err = program.Name()
		assert.EqualError(t, err, "unable to determine contract name")
	})

	t.Run("Replace", func(t *testing.T) {
		code := []byte(`
			import Foo from "./Foo.cdc"
			import "Bar"
			import  FooSpace  from  "./FooSpace.cdc"
			import   "BarSpace"

			access(all) contract Foo {}
		`)

		replaced := []byte(`
			import Foo from 0x1
			import Bar from 0x2
			import FooSpace from 0x3
			import BarSpace from 0x4

			access(all) contract Foo {}
		`)

		program, err := NewProgram(code, nil, "")
		require.NoError(t, err)

		program.
			replaceImport("./Foo.cdc", "1").
			replaceImport("Bar", "2").
			replaceImport("./FooSpace.cdc", "3").
			replaceImport("BarSpace", "4")

		assert.Equal(t, string(replaced), string(program.Code()))
	})

	t.Run("Convert to Import Syntax", func(t *testing.T) {
		code := []byte(`
		import Foo from 0x123
		import "Bar"
		import FooSpace from 0x124
		import "BarSpace"

		access(all) contract Foo {}
	`)

		expected := []byte(`
		import "Foo"
		import "Bar"
		import "FooSpace"
		import "BarSpace"

		access(all) contract Foo {}
	`)

		program, err := NewProgram(code, nil, "")
		require.NoError(t, err)

		program.ConvertAddressImports()

		assert.Equal(t, string(expected), string(program.CodeWithUnprocessedImports()))
	})

}
