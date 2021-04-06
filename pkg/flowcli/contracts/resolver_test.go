package contracts

import (
	"testing"

	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-go-sdk"

	"github.com/stretchr/testify/assert"
)

func TestResolver(t *testing.T) {

	contracts := []project.Contract{{
		Name:   "Kibble",
		Source: "./../contracts/Kibble.cdc",
		Target: flow.HexToAddress("0x1"),
	}, {
		Name:   "FT",
		Source: "./../contracts/FT.cdc",
		Target: flow.HexToAddress("0x2"),
	}, {
		Name:   "NFT",
		Source: "./../contracts/NFT.cdc",
		Target: flow.HexToAddress("0x3"),
	}}

	t.Run("Import exists", func(t *testing.T) {
		resolver, err := NewResolver([]byte(`
      import Kibble from "../../contracts/Kibble.cdc"
      pub fun main() {}
    `))
		assert.NoError(t, err)
		assert.True(t, resolver.ImportExists())
	})

	t.Run("Import doesn't exists", func(t *testing.T) {
		resolver, err := NewResolver([]byte(`
      pub fun main() {}
    `))
		assert.NoError(t, err)
		assert.False(t, resolver.ImportExists())
	})

	t.Run("Parse imports", func(t *testing.T) {
		resolver, err := NewResolver([]byte(`
      import Kibble from "../../contracts/Kibble.cdc"
      import FT from "../../contracts/FT.cdc"
      pub fun main() {}
    `))
		assert.NoError(t, err)
		assert.Equal(t, resolver.parseImports(), []string{
			"../../contracts/Kibble.cdc", "../../contracts/FT.cdc",
		})
	})

	t.Run("Resolve imports", func(t *testing.T) {
		resolver, err := NewResolver([]byte(`
      import Kibble from "./Kibble.cdc"
      import FT from "./FT.cdc"
      pub fun main() {}
    `))

		code := resolver.ResolveImports(contracts, make(map[string]string))

		assert.NoError(t, err)
		assert.Equal(t, code, "")
	})

}
