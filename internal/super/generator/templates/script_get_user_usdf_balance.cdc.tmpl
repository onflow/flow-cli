import "EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed"

/// This script returns the balance of USDF tokens for a given account.
/// Returns 0.0 if the account doesn't have a vault set up.
access(all) fun main(account: Address): UFix64 {
    let accountRef = getAccount(account)

    // Try to borrow the vault reference
    if let vaultRef = accountRef.capabilities.borrow<&EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.Vault>(/public/EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabedVault) {
        return vaultRef.balance
    } else {
        // Account doesn't have a vault set up
        return 0.0
    }
}
