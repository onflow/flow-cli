package contracts

//go:generate go run github.com/kevinburke/go-bindata/go-bindata -prefix ../../../contracts -o internal/assets/assets.go -pkg assets -nometadata -nomemcopy ../../../contracts

import (
	"strings"

	"github.com/onflow/flow-ft/lib/go/contracts/internal/assets"
)

const (
	fungibleTokenFilename       = "FungibleToken.cdc"
	flowTokenFilename           = "FlowToken.cdc"
	exampleTokenFilename        = "ExampleToken.cdc"
	defaultFungibleTokenAddress = "FUNGIBLETOKENADDRESS"
	tokenForwardingFilename     = "TokenForwarding.cdc"
)

// FungibleToken returns the FungibleToken contract interface.
func FungibleToken() []byte {
	return assets.MustAsset(fungibleTokenFilename)
}

// ExampleToken returns the ExampleToken contract.
//
// The returned contract will import the FungibleToken interface from the specified address.
func ExampleToken(fungibleTokenAddr string) []byte {
	code := assets.MustAssetString(exampleTokenFilename)

	code = strings.ReplaceAll(
		code,
		"0x"+defaultFungibleTokenAddress,
		"0x"+fungibleTokenAddr,
	)

	return []byte(code)
}

// CustomToken returns the ExampleToken contract with a custom name.
//
// The returned contract will import the FungibleToken interface from the specified address.
func CustomToken(fungibleTokenAddr, tokenName, storageName, initialBalance string) []byte {
	code := assets.MustAssetString(exampleTokenFilename)

	code = strings.ReplaceAll(
		code,
		"0x"+defaultFungibleTokenAddress,
		"0x"+fungibleTokenAddr,
	)

	code = strings.ReplaceAll(
		code,
		"ExampleToken",
		tokenName,
	)

	code = strings.ReplaceAll(
		code,
		"exampleToken",
		storageName,
	)

	code = strings.ReplaceAll(
		code,
		"1000.0",
		initialBalance,
	)

	return []byte(code)
}

// TokenForwarding returns the TokenForwarding contract.
//
// The returned contract will import the FungibleToken contract from the specified address.
func TokenForwarding(fungibleTokenAddr string) []byte {
	code := assets.MustAssetString(tokenForwardingFilename)

	code = strings.ReplaceAll(
		code,
		"0x"+defaultFungibleTokenAddress,
		"0x"+fungibleTokenAddr,
	)

	return []byte(code)
}

// CustomTokenForwarding returns the TokenForwarding contract for a custom token
//
// The returned contract will import the FungibleToken interface from the specified address.
func CustomTokenForwarding(fungibleTokenAddr, tokenName, storageName string) []byte {
	code := assets.MustAssetString(tokenForwardingFilename)

	code = strings.ReplaceAll(
		code,
		"0x"+defaultFungibleTokenAddress,
		"0x"+fungibleTokenAddr,
	)

	code = strings.ReplaceAll(
		code,
		"ExampleToken",
		tokenName,
	)

	code = strings.ReplaceAll(
		code,
		"exampleToken",
		storageName,
	)

	return []byte(code)
}
