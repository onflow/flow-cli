import "FungibleToken"
import "EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed"

/// This transaction mints USDF tokens and deposits them into the recipient's vault.
/// Works with USDF_MOCK on emulator (public mint) and EVMVMBridgedToken on mainnet (requires admin).
/// If the recipient doesn't have a vault set up, the transaction will create one for them.
transaction(amount: UFix64, recipient: Address) {

    let recipientVault: &{FungibleToken.Receiver}

    prepare(signer: auth(BorrowValue, IssueStorageCapabilityController, PublishCapability, SaveValue) &Account) {

        // Get the recipient's account
        let recipientAccount = getAccount(recipient)

        // Check if recipient has a vault capability
        self.recipientVault = recipientAccount.capabilities.borrow<&{FungibleToken.Receiver}>(
            /public/EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabedReceiver
        ) ?? panic("Could not borrow receiver reference to recipient's vault")
    }

    execute {
        // Mint the requested amount of tokens
        let mintedVault <- EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.mintTokens(amount: amount)

        // Deposit the newly minted tokens into the recipient's vault
        self.recipientVault.deposit(from: <-mintedVault)

        log("Successfully minted ".concat(amount.toString()).concat(" USDF tokens to ").concat(recipient.toString()))
    }
}
