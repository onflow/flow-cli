package json

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ConfigDependencies(t *testing.T) {
	b := []byte(`{
	"HelloWorld": {
		"remoteSource": "testnet/0x877931736ee77cff.HelloWorld",
		"aliases": {
				"testnet": "877931736ee77cff",
				"mainnet": "0b2a3299cc857e29"
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
	assert.Len(t, dependencyOne.Aliases, 2)
}

func Test_TransformDependenciesToJSON(t *testing.T) {
	b := []byte(`{
		"HelloWorld": {
			"remoteSource": "testnet/0x877931736ee77cff.HelloWorld",
			"aliases": {
 				"mainnet": "0b2a3299cc857e29",
				"testnet": "877931736ee77cff"
			}
		}
	}`)

	var jsonDependencies jsonDependencies
	err := json.Unmarshal(b, &jsonDependencies)
	assert.NoError(t, err)

	dependencies, err := jsonDependencies.transformToConfig()
	assert.NoError(t, err)

	j := jsonDependencies.transformDependenciesToJSON(dependencies)
	x, _ := json.Marshal(j)

	assert.Equal(t, cleanSpecialChars(b), cleanSpecialChars(x))
}
