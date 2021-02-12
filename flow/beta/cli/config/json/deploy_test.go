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

	//TODO: fix test to be sorted since its not necessary correct order
	assert.Equal(t, "account-2", deploys.GetByNetwork("testnet")[0].Account)
	assert.Equal(t, []string{"FungibleToken", "NonFungibleToken", "Kibble", "KittyItems"}, deploys.GetByNetwork("testnet")[0].Contracts)

	assert.Equal(t, 2, len(deploys.GetByNetwork("emulator")))
	assert.Equal(t, "account-3", deploys.GetByNetwork("emulator")[0].Account)
	assert.Equal(t, "account-4", deploys.GetByNetwork("emulator")[1].Account)
	assert.Equal(t, []string{"KittyItems", "KittyItemsMarket"}, deploys.GetByNetwork("emulator")[0].Contracts)
	assert.Equal(t, []string{"FungibleToken", "NonFungibleToken", "Kibble", "KittyItems", "KittyItemsMarket"}, deploys.GetByNetwork("emulator")[1].Contracts)
}

func Test_TransformDeployToJSON(t *testing.T) {
	b := []byte(`{"emulator":{"account-3":["KittyItems","KittyItemsMarket"],"account-4":["FungibleToken","NonFungibleToken","Kibble","KittyItems","KittyItemsMarket"]},"testnet":{"account-2":["FungibleToken","NonFungibleToken","Kibble","KittyItems"]}}`)

	var jsonDeploys jsonDeploys
	json.Unmarshal(b, &jsonDeploys)
	deploys := jsonDeploys.transformToConfig()

	j := jsonDeploys.transformToJSON(deploys)
	x, _ := json.Marshal(j)

	assert.Equal(t, string(b), string(x))
}
