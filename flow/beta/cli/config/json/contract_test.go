package json

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ConfigContractsSimple(t *testing.T) {
	b := []byte(`{
    "KittyItems": "./cadence/kittyItems/contracts/KittyItems.cdc",
    "KittyItemsMarket": "./cadence/kittyItems/contracts/KittyItemsMarket.cdc"
  }`)

	var jsonContracts jsonContracts
	json.Unmarshal(b, &jsonContracts)
	contracts := jsonContracts.transformToConfig()

	assert.Equal(t, "./cadence/kittyItems/contracts/KittyItems.cdc", contracts.GetByName("KittyItems").Source)
	assert.Equal(t, "./cadence/kittyItems/contracts/KittyItemsMarket.cdc", contracts.GetByName("KittyItemsMarket").Source)
}

func Test_ConfigContractsComplex(t *testing.T) {
	b := []byte(`{
    "KittyItems": "./cadence/kittyItems/contracts/KittyItems.cdc",
    "KittyItemsMarket": {
      "testnet": "0x123123123",
      "emulator": "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc"
    }
  }`)

	var jsonContracts jsonContracts
	json.Unmarshal(b, &jsonContracts)
	contracts := jsonContracts.transformToConfig()

	assert.Equal(t, "./cadence/kittyItems/contracts/KittyItems.cdc", contracts.GetByName("KittyItems").Source)
	assert.Equal(t, "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc", contracts.GetByNameAndNetwork("KittyItemsMarket", "emulator").Source)
	assert.Equal(t, "0x123123123", contracts.GetByNameAndNetwork("KittyItemsMarket", "testnet").Source)

	assert.Equal(t, 2, len(contracts.GetByNetwork("testnet")))
	assert.Equal(t, 2, len(contracts.GetByNetwork("emulator")))

	assert.Equal(t, "./cadence/kittyItems/contracts/KittyItems.cdc", contracts.GetByNetwork("testnet")[0].Source)
	assert.Equal(t, "0x123123123", contracts.GetByNetwork("testnet")[1].Source)

	assert.Equal(t, "./cadence/kittyItems/contracts/KittyItems.cdc", contracts.GetByNetwork("emulator")[0].Source)
	assert.Equal(t, "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc", contracts.GetByNetwork("emulator")[1].Source)
}
