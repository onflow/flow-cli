import "FungibleToken"
import "EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed"

/// PiggyBank is a simple contract that stores USDF tokens in a shared vault
/// Anyone can deposit USDF tokens and withdraw them
access(all) contract PiggyBank {

    /// Events
    access(all) event Deposited(depositor: Address, amount: UFix64)
    access(all) event Withdrawn(withdrawer: Address, amount: UFix64)

    /// The vault that holds all deposited USDF tokens
    access(self) var vault: @{FungibleToken.Vault}

    /// Deposit USDF tokens into the piggy bank
    access(all) fun deposit(from: @{FungibleToken.Vault}) {
        let amount = from.balance
        self.vault.deposit(from: <-from)

        emit Deposited(depositor: self.account.address, amount: amount)
    }

    /// Withdraw USDF tokens from the piggy bank
    access(all) fun withdraw(amount: UFix64): @{FungibleToken.Vault} {
        pre {
            amount > 0.0: "Withdrawal amount must be greater than zero"
            amount <= self.vault.balance: "Insufficient balance in piggy bank"
        }

        let withdrawn <- self.vault.withdraw(amount: amount)

        emit Withdrawn(withdrawer: self.account.address, amount: amount)

        return <-withdrawn
    }

    /// Get the current balance in the piggy bank
    access(all) fun getBalance(): UFix64 {
        return self.vault.balance
    }

    init() {
        // Create an empty USDF vault to hold all deposits
        self.vault <- EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.createEmptyVault(
            vaultType: Type<@EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.Vault>()
        )
    }
}
