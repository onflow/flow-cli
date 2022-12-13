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
	"regexp"
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cleanCode(code []byte) string {
	space := regexp.MustCompile(`\s+`)
	return space.ReplaceAllString(string(code), " ")
}

func TestResolver(t *testing.T) {

	t.Run("Resolve imports", func(t *testing.T) {
		contracts := []*Contract{
			NewContract("Kibble", "./tests/Kibble.cdc", nil, flow.HexToAddress("0x1"), "", nil),
			NewContract("FT", "./tests/FT.cdc", nil, flow.HexToAddress("0x2"), "", nil),
		}

		aliases := map[string]string{
			"./tests/NFT.cdc": flow.HexToAddress("0x4").String(),
		}

		paths := []string{
			"./tests/foo.cdc",
			"./scripts/bar/foo.cdc",
			"./scripts/bar/foo.cdc",
			"./tests/foo.cdc",
		}

		scripts := [][]byte{
			[]byte(`
			import Kibble from "./Kibble.cdc"
			import FT from "./FT.cdc"
			pub fun main() {}
    `), []byte(`
			import Kibble from "../../tests/Kibble.cdc"
			import FT from "../../tests/FT.cdc"
			pub fun main() {}
    `), []byte(`
			import Kibble from "../../tests/Kibble.cdc"
			import NFT from "../../tests/NFT.cdc"
			pub fun main() {}
    `), []byte(`
			import Kibble from "./Kibble.cdc"
			import crypto
			import Foo from 0x0000000000000001
			pub fun main() {}
	`),
		}

		resolved := [][]byte{
			[]byte(`
			import Kibble from 0x0000000000000001 
			import FT from 0x0000000000000002 
			pub fun main() {}
    `), []byte(`
			import Kibble from 0x0000000000000001 
			import FT from 0x0000000000000002 
			pub fun main() {}
    `), []byte(`
			import Kibble from 0x0000000000000001 
			import NFT from 0x0000000000000004 
			pub fun main() {}
    `), []byte(`
			import Kibble from 0x0000000000000001
			import crypto
			import Foo from 0x0000000000000001
			pub fun main() {}
	`),
		}

		replacer := NewImportReplacer(contracts, aliases)
		for i, script := range scripts {
			program, err := NewProgram(&testScript{
				code:     script,
				location: paths[i],
			})
			require.NoError(t, err)

			program, err = replacer.Replace(program)
			assert.NoError(t, err)
			assert.Equal(t, cleanCode(program.Code()), cleanCode(resolved[i]))
		}
	})

	t.Run("Resolve new schema", func(t *testing.T) {
		contracts := []*Contract{
			NewContract("Bar", "./Bar.cdc", nil, flow.HexToAddress("0x2"), "", nil),
			NewContract("Foo", "./Foo.cdc", nil, flow.HexToAddress("0x1"), "", nil),
		}

		replacer := NewImportReplacer(contracts, nil)

		code := []byte(`
			import Foo from "./Foo.cdc"
			import Bar
			
			pub contract Zoo {}
		`)
		program, err := NewProgram(&testScript{code: code, location: "./Zoo.cdc"})
		require.NoError(t, err)

		replaced, err := replacer.Replace(program)
		require.NoError(t, err)

		expected := []byte(`
			import Foo from 0x0000000000000001
			import Bar from 0x0000000000000002
			
			pub contract Zoo {}
		`)

		assert.Equal(t, string(expected), string(replaced.Code()))
	})

}
