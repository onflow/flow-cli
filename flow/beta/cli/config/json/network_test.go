package json

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ConfigNetworkSimple(t *testing.T) {
	b := []byte(`{
    "testnet": "access.testnet.nodes.onflow.org:9000"
	}`)

	var jsonNetworks jsonNetworks
	json.Unmarshal(b, &jsonNetworks)

	networks := jsonNetworks.transformToConfig()

	assert.Equal(t, networks.GetByName("testnet").Host, "access.testnet.nodes.onflow.org:9000")
	assert.Equal(t, networks.GetByName("testnet").Name, "testnet")
}

func Test_ConfigNetworkMultiple(t *testing.T) {
	b := []byte(`{
    "emulator": {
      "host": "127.0.0.1:3569",
			"chain": "flow-emulator",
      "serviceAccount": "emulator-service"
    },
    "testnet": "access.testnet.nodes.onflow.org:9000"
	}`)

	var jsonNetworks jsonNetworks
	json.Unmarshal(b, &jsonNetworks)

	networks := jsonNetworks.transformToConfig()

	assert.Equal(t, networks.GetByName("testnet").Host, "access.testnet.nodes.onflow.org:9000")
	assert.Equal(t, networks.GetByName("testnet").Name, "testnet")

	assert.Equal(t, networks.GetByName("emulator").Name, "emulator")
	assert.Equal(t, networks.GetByName("emulator").Host, "127.0.0.1:3569")
}

func Test_TransformNetworkToJSON(t *testing.T) {
	b := []byte(`{"emulator":{"host":"127.0.0.1:3569","chain":"flow-emulator"},"testnet":"access.testnet.nodes.onflow.org:9000"}`)

	var jsonNetworks jsonNetworks
	json.Unmarshal(b, &jsonNetworks)
	networks := jsonNetworks.transformToConfig()

	j := jsonNetworks.transformToJSON(networks)
	x, _ := json.Marshal(j)

	assert.Equal(t, string(b), string(x))
}
