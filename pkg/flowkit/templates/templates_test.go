package templates

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplateCollection(t *testing.T) {

	t.Run("Get by name transaction", func(t *testing.T) {
		tmp, err := TransactionByName(collection[0].name)
		assert.NoError(t, err)
		assert.Equal(t, tmp.Name(), collection[0].name)
	})

	t.Run("Get by name script", func(t *testing.T) {
		tmp, err := ScriptByName(collection[2].name)
		assert.NoError(t, err)
		assert.Equal(t, tmp.Name(), collection[2].name)
	})

	t.Run("Get by name invalid", func(t *testing.T) {
		tmp, err := TransactionByName(collection[0].name + "invalid")
		assert.EqualError(t, err, "template not found by name")
		assert.Nil(t, tmp)
	})
}

func TestTemplates(t *testing.T) {

	t.Run("Template by name for network", func(t *testing.T) {
		for _, c := range collection {
			var tmp *template
			var err error
			if c.kind == tx {
				tmp, err = TransactionByName(c.name)
			} else if c.kind == script {
				tmp, err = ScriptByName(c.name)
			}

			assert.NoError(t, err)

			for _, n := range []string{testnet, mainnet} {
				src, err := tmp.Source(n)
				assert.NoError(t, err)

				for _, i := range c.imports[n] {
					assert.True(t, strings.Index(string(src), fmt.Sprintf("from %s", i)) > 0, string(src))
				}
			}
		}
	})

	t.Run("Template by name for invalid network", func(t *testing.T) {
		otmp := collection[0]
		tmp, err := TransactionByName(otmp.name)
		assert.NoError(t, err)

		src, err := tmp.Source("foo")
		assert.Nil(t, src)
		assert.EqualError(t, err, "invalid network")
	})

}
