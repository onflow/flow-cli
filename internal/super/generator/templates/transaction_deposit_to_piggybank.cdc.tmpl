import "FungibleToken"
import "EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed"
import "PiggyBank"

/// This transaction withdraws USDF tokens from the signer's vault and deposits them into the piggy bank contract
transaction(amount: UFix64) {

    let usdVault: auth(FungibleToken.Withdraw) &EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.Vault

    prepare(signer: auth(BorrowValue) &Account) {

        // Borrow reference to the USDF vault
        self.usdVault = signer.storage.borrow<auth(FungibleToken.Withdraw) &EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.Vault>(
            from: /storage/EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabedVault
        ) ?? panic("Could not borrow reference to USDF vault. Make sure you have set up your USDF vault first.")
    }

    execute {
        // Withdraw USDF tokens from the signer's vault
        let tokens <- self.usdVault.withdraw(amount: amount)

        // Deposit into the piggy bank contract vault
        PiggyBank.deposit(from: <-tokens)

        log("Successfully deposited ".concat(amount.toString()).concat(" USDF tokens into piggy bank"))
    }
}
