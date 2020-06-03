package contracts

//go:generate go run github.com/kevinburke/go-bindata/go-bindata -prefix ../src/contracts -o internal/assets/assets.go -pkg assets -nometadata -nomemcopy ../src/contracts

import (
	"strings"

	ftcontracts "github.com/onflow/flow-ft/contracts"

	"github.com/dapperlabs/flow-core-contracts/contracts/internal/assets"
)

const (
	flowFeesFilename           = "FlowFees.cdc"
	flowServiceAccountFilename = "FlowServiceAccount.cdc"
	hexPrefix                  = "0x"
	defaultFungibleTokenAddr   = "02"
	defaultFlowTokenAddr       = "03"
	defaultFlowFeesAddr        = "04"
)

// FungibleToken returns the FungibleToken contract interface.
func FungibleToken() []byte {
	return ftcontracts.FungibleToken()
}

// FlowToken returns the FlowToken contract. importing the
//
// The returned contract will import the FungibleToken contract from the specified address.
func FlowToken(fungibleTokenAddr string) []byte {
	return ftcontracts.FlowToken(fungibleTokenAddr)
}

// FlowFees returns the FlowFees contract.
//
// The returned contract will import the FungibleToken and FlowToken
// contracts from the specified addresses.
func FlowFees(fungibleTokenAddr, flowTokenAddr string) []byte {
	code := assets.MustAssetString(flowFeesFilename)

	code = strings.ReplaceAll(
		code,
		hexPrefix+defaultFungibleTokenAddr,
		hexPrefix+fungibleTokenAddr,
	)

	code = strings.ReplaceAll(
		code,
		hexPrefix+defaultFlowTokenAddr,
		hexPrefix+flowTokenAddr,
	)

	return []byte(code)
}

// FlowServiceAccount returns the FlowServiceAccount contract.
//
// The returned contract will import the FungibleToken, FlowToken and FlowFees
// contracts from the specified addresses.
func FlowServiceAccount(fungibleTokenAddr, flowTokenAddr, flowFeesAddr string) []byte {
	code := assets.MustAssetString(flowServiceAccountFilename)

	code = strings.ReplaceAll(
		code,
		hexPrefix+defaultFungibleTokenAddr,
		hexPrefix+fungibleTokenAddr,
	)

	code = strings.ReplaceAll(
		code,
		hexPrefix+defaultFlowTokenAddr,
		hexPrefix+flowTokenAddr,
	)

	code = strings.ReplaceAll(
		code,
		hexPrefix+defaultFlowFeesAddr,
		hexPrefix+flowFeesAddr,
	)

	return []byte(code)
}
