import EVM from 0xf8d6e0586b0a20c7 // todo dynamically set
import FungibleToken from 0xee82856bf20e2aa6
import FlowToken from 0x0ae53cb6e3f42a79

transaction(amount: UFix64) {
    let sentVault: @FlowToken.Vault
    let auth: auth(Storage) &Account

    prepare(signer: auth(Storage) &Account) {
        let vaultRef = signer.storage.borrow<auth(FungibleToken.Withdrawable) &FlowToken.Vault>(from: /storage/flowTokenVault)
            ?? panic("Could not borrow reference to the owner's Vault!")

        self.sentVault <- vaultRef.withdraw(amount: amount) as! @FlowToken.Vault
        self.auth = signer
    }

    execute {
        let account <- EVM.createBridgedAccount()
        account.address().deposit(from: <-self.sentVault)

        log(account.balance())
        self.auth.storage.save<@EVM.BridgedAccount>(<-account, to: StoragePath(identifier: "evm")!)
    }
}
