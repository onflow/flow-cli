import "FungibleToken"
import "FlowToken"
import "DeFiActions"
import "ExampleConnectors"

/// Deposit FlowToken into a recipient's vault using ExampleConnectors.TokenSink
///
/// This transaction demonstrates:
/// - Creating a TokenSink pointing to a recipient's vault capability
/// - Withdrawing tokens from the signer's vault
/// - Depositing via the sink's depositCapacity method
///
/// @param recipient: Address of the account to receive tokens
/// @param amount: Amount of FlowToken to send
transaction(recipient: Address, amount: UFix64) {
    let senderVault: auth(FungibleToken.Withdraw) &FlowToken.Vault
    let sink: ExampleConnectors.TokenSink

    prepare(signer: auth(BorrowValue) &Account) {
        // Borrow the signer's FlowToken vault
        self.senderVault = signer.storage.borrow<auth(FungibleToken.Withdraw) &FlowToken.Vault>(
            from: /storage/flowTokenVault
        ) ?? panic("Could not borrow FlowToken vault from signer")

        // Get a capability to the recipient's FlowToken receiver
        let recipientCap = getAccount(recipient)
            .capabilities.get<&{FungibleToken.Receiver}>(/public/flowTokenReceiver)

        // Create a TokenSink that will deposit into the recipient's vault
        self.sink = ExampleConnectors.TokenSink(
            vault: recipientCap,
            uniqueID: nil
        )
    }

    execute {
        // Withdraw tokens from signer
        let tokens <- self.senderVault.withdraw(amount: amount)

        // Deposit via the sink
        self.sink.depositCapacity(from: &tokens as auth(FungibleToken.Withdraw) &{FungibleToken.Vault})

        // Ensure everything was deposited
        assert(tokens.balance == 0.0, message: "Tokens remaining after deposit")
        destroy tokens
    }
}

