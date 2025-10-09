import "FungibleToken"
import "EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed"

/// This script returns the balance of USDF tokens in a user's vault
access(all) fun main(address: Address): UFix64 {

    let account = getAccount(address)

    // Get the public capability for the USDF vault
    let vaultRef = account.capabilities.borrow<&EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.Vault>(
        /public/EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabedVault
    ) ?? panic("Could not borrow reference to USDF vault for address ".concat(address.toString()))

    return vaultRef.balance
}
