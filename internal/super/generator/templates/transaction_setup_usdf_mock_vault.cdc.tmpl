import "FungibleToken"
import "EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed"

/// This transaction sets up a USDF vault for the signer's account.
/// It creates the vault, saves it to storage, and creates the necessary public capabilities.
transaction() {

    prepare(signer: auth(BorrowValue, IssueStorageCapabilityController, PublishCapability, SaveValue, UnpublishCapability) &Account) {

        // Check if account already has a vault
        if signer.storage.borrow<&EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.Vault>(from: /storage/EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabedVault) != nil {
            log("Account already has a vault set up")
            return
        }

        // Create a new empty vault
        let vault <- EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.createEmptyVault(vaultType: Type<@EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.Vault>())

        // Save the vault to storage
        signer.storage.save(<-vault, to: /storage/EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabedVault)

        // Create and publish public capabilities
        let vaultCap = signer.capabilities.storage.issue<&EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.Vault>(
            /storage/EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabedVault
        )
        signer.capabilities.publish(vaultCap, at: /public/EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabedVault)

        let receiverCap = signer.capabilities.storage.issue<&{FungibleToken.Receiver}>(
            /storage/EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabedVault
        )
        signer.capabilities.publish(receiverCap, at: /public/EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabedReceiver)
    }

    execute {
        log("Vault setup completed successfully")
    }
}
