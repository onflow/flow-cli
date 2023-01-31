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

// limited test scripter implementation for running tests
type testScript struct {
	code     []byte
	location string
}

func (t *testScript) Code() []byte {
	return t.code
}

func (t *testScript) SetCode(code []byte) {
	t.code = code
}

func (t *testScript) Location() string {
	return t.location
}

func TestProgram(t *testing.T) {

	t.Run("Imports", func(t *testing.T) {
		tests := []struct {
			code    []byte
			imports []string
		}{{
			code:    []byte(`pub contract Foo {}`),
			imports: []string{},
		}, {
			code:    []byte(`pub fun main() {}`),
			imports: []string{},
		}, {
			code: []byte(`
				import Bar from "./Bar.cdc"
				pub contract Foo {}
			`),
			imports: []string{"./Bar.cdc"},
		}, {
			code: []byte(`
				import Bar from "./Bar.cdc"
				import Zoo from "./zoo/Zoo.cdc"
				pub contract Foo {}
			`),
			imports: []string{"./Bar.cdc", "./zoo/Zoo.cdc"},
		}, { // new schema import
			code: []byte(`
				import "Bar"
				pub contract Foo {}
			`),
			imports: []string{"Bar"},
		}, {
			code: []byte(`
				import "Bar"
				import Zoo from "./Zoo.cdc"
				import Crypto
				import Foo from 0x01

				pub contract Foo {}
			`),
			imports: []string{"Bar", "./Zoo.cdc"},
		}}

		for i, test := range tests {
			program, err := NewProgram(&testScript{code: test.code})
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
			code: []byte(`pub contract Foo {}`),
			name: "Foo",
		}, {
			code: []byte(`
				import Bar from "./Bar.cdc"
				pub contract Foo {}
			`),
			name: "Foo",
		}}

		for i, test := range tests {
			program, err := NewProgram(&testScript{code: test.code})
			require.NoError(t, err, fmt.Sprintf("import test %d failed", i))
			name, err := program.Name()
			require.NoError(t, err)
			assert.Equal(t, test.name, name)
		}

		failed := [][]byte{
			[]byte(`
				pub contract Foo {}
				pub contract Bar {}
			`),
			[]byte(`
				pub contract Foo {}
				pub resource interface Test {}
			`),
			[]byte(`
				pub contract Foo {}
				struct Bar {}
			`),
		}

		for _, code := range failed {
			program, err := NewProgram(&testScript{code: code})
			require.NoError(t, err)

			_, err = program.Name()
			assert.EqualError(t, err, "the code must declare exactly one contract or contract interface")
		}

		program, err := NewProgram(&testScript{code: []byte(`pub fun main() {}`)})
		require.NoError(t, err)
		_, err = program.Name()
		assert.EqualError(t, err, "unable to determine contract name")
	})

	t.Run("Replace", func(t *testing.T) {
		code := []byte(`
			import Foo from "./Foo.cdc"
			import "Bar"

			pub contract Foo {}
		`)

		replaced := []byte(`
			import Foo from 0x1
			import Bar from 0x2

			pub contract Foo {}
		`)

		program, err := NewProgram(&testScript{code: code})
		require.NoError(t, err)

		program.
			replaceImport("./Foo.cdc", "1").
			replaceImport("Bar", "2")

		assert.Equal(t, string(replaced), string(program.Code()))
	})

}
