package json

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ConfigDeploySimple(t *testing.T) {
	b := []byte(`{
		"testnet": {
			"account-2": ["FungibleToken", "NonFungibleToken", "Kibble", "KittyItems"]
		}, 
		"emulator": {
			"account-3": ["KittyItems", "KittyItemsMarket"],
			"account-4": ["FungibleToken", "NonFungibleToken", "Kibble", "KittyItems", "KittyItemsMarket"]
		}
	}`)

	var jsonDeploys jsonDeploys
	json.Unmarshal(b, &jsonDeploys)
	deploys := jsonDeploys.transformToConfig()

	assert.Equal(t, "account-2", deploys.GetByNetwork("testnet")[0].Account)
	assert.Equal(t, []string{"FungibleToken", "NonFungibleToken", "Kibble", "KittyItems"}, deploys.GetByNetwork("testnet")[0].Contracts)

	assert.Equal(t, 2, len(deploys.GetByNetwork("emulator")))
	assert.Equal(t, "account-3", deploys.GetByNetwork("emulator")[0].Account)
	assert.Equal(t, "account-4", deploys.GetByNetwork("emulator")[1].Account)
	assert.Equal(t, []string{"KittyItems", "KittyItemsMarket"}, deploys.GetByNetwork("emulator")[0].Contracts)
	assert.Equal(t, []string{"FungibleToken", "NonFungibleToken", "Kibble", "KittyItems", "KittyItemsMarket"}, deploys.GetByNetwork("emulator")[1].Contracts)
}
