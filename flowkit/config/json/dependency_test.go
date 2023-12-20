package json

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ConfigDependencies(t *testing.T) {
	b := []byte(`{
		"HelloWorld": "testnet://877931736ee77cff.HelloWorld"
		}`)

	var jsonDependencies jsonDependencies
	err := json.Unmarshal(b, &jsonDependencies)
	assert.NoError(t, err)

	dependencies, err := jsonDependencies.transformToConfig()
	assert.NoError(t, err)

	assert.Len(t, dependencies, 1)

	dependencyOne := dependencies.ByName("HelloWorld")
	assert.NotNil(t, dependencyOne)

	assert.NotNil(t, dependencyOne)
}

func Test_TransformDependenciesToJSON(t *testing.T) {
	b := []byte(`{
		"HelloWorld": "testnet://877931736ee77cff.HelloWorld"
	}`)

	bOut := []byte(`{
		"HelloWorld": {
			"remoteSource": "testnet://877931736ee77cff.HelloWorld",
			"aliases": {}
		}
	}`)

	var jsonContracts jsonContracts
	errContracts := json.Unmarshal(b, &jsonContracts)
	assert.NoError(t, errContracts)

	var jsonDependencies jsonDependencies
	err := json.Unmarshal(b, &jsonDependencies)
	assert.NoError(t, err)

	contracts, err := jsonContracts.transformToConfig()
	dependencies, err := jsonDependencies.transformToConfig()
	assert.NoError(t, err)

	j := transformDependenciesToJSON(dependencies, contracts)
	x, _ := json.Marshal(j)

	assert.Equal(t, cleanSpecialChars(bOut), cleanSpecialChars(x))
}
