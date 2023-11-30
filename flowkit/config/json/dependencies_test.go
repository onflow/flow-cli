package json

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ConfigDependencies(t *testing.T) {
	b := []byte(`{
	"HelloWorld": {
		"remoteSource": "testnet/0x0123123123.HelloWorld",
		"aliases": {
				"testnet": "0x0123123123",
				"mainnet": "0x0123123124"
			}
		}
	}`)

	var jsonDependencies jsonDependencies
	err := json.Unmarshal(b, &jsonDependencies)
	assert.NoError(t, err)

	dependencies, err := jsonDependencies.transformToConfig()
	assert.NoError(t, err)

	assert.Len(t, dependencies, 1)

	dependencyOne := dependencies.ByName("HelloWorld")

	assert.NotNil(t, dependencyOne)
}
