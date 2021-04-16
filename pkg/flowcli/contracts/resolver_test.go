package contracts

import (
	"regexp"
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/pkg/flowcli/project"
)

func cleanCode(code []byte) string {
	space := regexp.MustCompile(`\s+`)
	return space.ReplaceAllString(string(code), " ")
}

func TestResolver(t *testing.T) {

	contracts := []project.Contract{{
		Name:   "Kibble",
		Source: "./tests/Kibble.cdc",
		Target: flow.HexToAddress("0x1"),
	}, {
		Name:   "FT",
		Source: "./tests/FT.cdc",
		Target: flow.HexToAddress("0x2"),
	}}

	aliases := map[string]string{
		"./tests/NFT.cdc": flow.HexToAddress("0x4").String(),
	}

	paths := []string{
		"./tests/foo.cdc",
		"./scripts/bar/foo.cdc",
		"./scripts/bar/foo.cdc",
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
    `),
	}

	t.Run("Import exists", func(t *testing.T) {
		resolver, err := NewResolver([]byte(`
      import Kibble from "./Kibble.cdc"
      pub fun main() {}
    `))
		assert.NoError(t, err)
		assert.True(t, resolver.HasFileImports())
	})

	t.Run("Import doesn't exists", func(t *testing.T) {
		resolver, err := NewResolver([]byte(`
      pub fun main() {}
    `))
		assert.NoError(t, err)
		assert.False(t, resolver.HasFileImports())

		resolver, err = NewResolver([]byte(`
			import Foo from 0xf8d6e0586b0a20c7
      pub fun main() {}
    `))
		assert.NoError(t, err)
		assert.False(t, resolver.HasFileImports())
	})

	t.Run("Parse imports", func(t *testing.T) {
		resolver, err := NewResolver(scripts[0])
		assert.NoError(t, err)
		assert.Equal(t, resolver.getFileImports(), []string{
			"./Kibble.cdc", "./FT.cdc",
		})
	})

	t.Run("Resolve imports", func(t *testing.T) {
		for i, script := range scripts {
			resolver, err := NewResolver(script)
			assert.NoError(t, err)

			code, err := resolver.ResolveImports(paths[i], contracts, aliases)

			assert.NoError(t, err)
			assert.Equal(t, cleanCode(code), cleanCode(resolved[i]))
		}
	})

}
