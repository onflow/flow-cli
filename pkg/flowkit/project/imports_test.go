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

package project_test

import (
	"fmt"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/project"
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
	contracts := []*project.Contract{
		project.NewContract("Kibble", "./tests/Kibble.cdc", nil, flow.HexToAddress("0x1"), "", nil),
		project.NewContract("FT", "./tests/FT.cdc", nil, flow.HexToAddress("0x2"), "", nil),
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

	t.Run("Import exists", func(t *testing.T) {
		tests := []struct {
			code       []byte
			hasImports bool
		}{{
			code: []byte(`
			  import Kibble from "./Kibble.cdc"
			  pub fun main() {}
			`),
			hasImports: true,
		}, {
			code: []byte(`
				pub fun main() {}
			`),
			hasImports: false,
		}, {
			code:       scripts[3],
			hasImports: true,
		}}

		for i, test := range tests {
			program, err := project.NewProgram(flowkit.NewScript(test.code, nil, ""))
			assert.NoError(t, err)
			assert.Equal(t, test.hasImports, program.HasImports(), fmt.Sprintf("failed with test vector %d", i))
		}
	})

	t.Run("Parse imports", func(t *testing.T) {
		program, err := project.NewProgram(flowkit.NewScript(scripts[3], nil, ""))
		assert.NoError(t, err)
		assert.Equal(t, program.Imports(), []string{
			"./Kibble.cdc",
		})
	})

	t.Run("Resolve imports", func(t *testing.T) {
		replacer := project.NewImportReplacer(contracts, aliases)
		for i, script := range scripts {
			program, err := project.NewProgram(flowkit.NewScript(script, nil, paths[i]))
			require.NoError(t, err)

			program, err = replacer.Replace(program)
			assert.NoError(t, err)
			assert.Equal(t, cleanCode(program.Code()), cleanCode(resolved[i]))
		}
	})
	t.Run("Get Contract Name", func(t *testing.T) {
		program, err := project.NewProgram(flowkit.NewScript(
			[]byte(`
				pub contract HelloWorld {}
			`),
			nil, "",
		))
		require.NoError(t, err)
		contractName, err := program.Name()
		assert.NoError(t, err)
		assert.Equal(t, "HelloWorld", contractName)
	})
	t.Run("Get Contract Name", func(t *testing.T) {
		program, err := project.NewProgram(flowkit.NewScript(
			[]byte(`
				pub struct SomeStruct {}
				pub contract HelloWorld {}
			`),
			nil, "",
		))
		require.NoError(t, err)
		name, err := program.Name()
		assert.ErrorContains(t, err, "the code must declare exactly one contract or contract interface")
		assert.Equal(t, "", name)
	})
}
