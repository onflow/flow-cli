package templates

import (
	"fmt"
	"strings"

	"github.com/onflow/flow-cli/pkg/flowkit/config"
)

var emulator = config.DefaultEmulatorNetwork().Name
var testnet = config.DefaultTestnetNetwork().Name
var mainnet = config.DefaultMainnetNetwork().Name

type template struct {
	name   string
	source string
	// imports matching order of imports in script
	imports map[string][]string
}

func (t *template) Name() string {
	return t.name
}

func (t *template) Source(network string) ([]byte, error) {
	imports := t.imports[network]

	if len(imports) == 0 {
		return nil, fmt.Errorf("invalid network")
	}

	// converting each value since array conversion doesn't work
	replace := make([]interface{}, len(imports))
	for i, im := range imports {
		replace[i] = im
	}

	return []byte(fmt.Sprintf(t.source, replace...)), nil
}

func ByName(name string) (*template, error) {
	for _, c := range collection {
		if strings.ToLower(c.name) == strings.ToLower(name) {
			return c, nil
		}
	}

	return nil, fmt.Errorf("template not found by name")
}

var collection = []*template{{
	name: "transferFUSD",
	imports: map[string][]string{
		testnet: {"0x9a0766d93b6608b7", "0xe223d8a629e49c68"},
		mainnet: {"0xf233dcee88fe0abe", "0x3c5959b568896393"},
		// todo(sideninja): add emulator network, but first add fusd to emulator bootstrap procedure
	},
	// source: https://github.com/onflow/fusd
	source: `
		import FungibleToken from %s
		import FUSD from %s
		
		transaction(amount: UFix64, to: Address) {
		
			// The Vault resource that holds the tokens that are being transferred
			let sentVault: @FungibleToken.Vault
		
			prepare(signer: AuthAccount) {
				// Get a reference to the signer's stored vault
				let vaultRef = signer.borrow<&FUSD.Vault>(from: /storage/fusdVault)
					?? panic("Could not borrow reference to the owner's Vault!")
		
				// Withdraw tokens from the signer's stored vault
				self.sentVault <- vaultRef.withdraw(amount: amount)
			}
		
			execute {
				// Get the recipient's public account object
				let recipient = getAccount(to)
		
				// Get a reference to the recipient's Receiver
				let receiverRef = recipient.getCapability(/public/fusdReceiver)!.borrow<&{FungibleToken.Receiver}>()
					?? panic("Could not borrow receiver reference to the recipient's Vault")
		
				// Deposit the withdrawn tokens in the recipient's receiver
				receiverRef.deposit(from: <-self.sentVault)
			}
		}`,
}}
