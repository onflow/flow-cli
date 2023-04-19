package config

import (
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
)

func TestAliases_Add(t *testing.T) {
	aliases := Aliases{}
	aliases.Add("testnet", flow.HexToAddress("0xabcdef"))

	alias := aliases.ByNetwork("testnet")
	assert.NotNil(t, alias)
}

func TestAliases_Add_Duplicate(t *testing.T) {
	aliases := Aliases{}
	aliases.Add("testnet", flow.HexToAddress("0xabcdef"))
	aliases.Add("testnet", flow.HexToAddress("0x123456"))

	assert.Len(t, aliases, 1)
}

func TestContracts_AddOrUpdate_Add(t *testing.T) {
	contracts := Contracts{}
	contracts.AddOrUpdate(Contract{Name: "mycontract", Location: "path/to/contract.cdc"})

	assert.Len(t, contracts, 1)

	contract, err := contracts.ByName("mycontract")
	assert.NoError(t, err)
	assert.Equal(t, "path/to/contract.cdc", contract.Location)
}

func TestContracts_AddOrUpdate_Update(t *testing.T) {
	contracts := Contracts{
		Contract{Name: "mycontract", Location: "path/to/contract.cdc"},
	}
	contracts.AddOrUpdate(Contract{Name: "mycontract", Location: "new/path/to/contract.cdc"})

	assert.Len(t, contracts, 1)

	contract, err := contracts.ByName("mycontract")
	assert.NoError(t, err)
	assert.Equal(t, "new/path/to/contract.cdc", contract.Location)
}

func TestContracts_Remove(t *testing.T) {
	contracts := Contracts{
		Contract{Name: "mycontract", Location: "path/to/contract.cdc"},
	}
	err := contracts.Remove("mycontract")
	assert.NoError(t, err)
	assert.Len(t, contracts, 0)
}

func TestContracts_Remove_NotFound(t *testing.T) {
	contracts := Contracts{
		Contract{Name: "mycontract", Location: "path/to/contract.cdc"},
	}
	err := contracts.Remove("nonexistent")
	assert.Error(t, err)
	assert.Len(t, contracts, 1)
}
