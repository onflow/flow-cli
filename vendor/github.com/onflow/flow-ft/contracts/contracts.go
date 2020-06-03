package contracts

//go:generate go run github.com/kevinburke/go-bindata/go-bindata -prefix ../src/contracts -o internal/assets/assets.go -pkg assets -nometadata -nomemcopy ../src/contracts

import (
	"strings"

	"github.com/onflow/flow-ft/contracts/internal/assets"
)

const (
	fungibleTokenFilename       = "FungibleToken.cdc"
	flowTokenFilename           = "FlowToken.cdc"
	defaultFungibleTokenAddress = "02"
)

// FungibleToken returns the FungibleToken contract interface.
func FungibleToken() []byte {
	return assets.MustAsset(fungibleTokenFilename)
}

// FlowToken returns the FlowToken contract.
//
// The returned contract will import the FungibleToken contract from the specified address.
func FlowToken(fungibleTokenAddr string) []byte {
	code := assets.MustAssetString(flowTokenFilename)

	code = strings.ReplaceAll(
		code,
		"0x"+defaultFungibleTokenAddress,
		"0x"+fungibleTokenAddr,
	)

	return []byte(code)
}
