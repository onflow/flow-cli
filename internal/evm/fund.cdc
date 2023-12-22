import EVM from 0xf8d6e0586b0a20c7 // todo dynamically set
import FungibleToken from 0xee82856bf20e2aa6
import FlowToken from 0x0ae53cb6e3f42a79

transaction(address: [UInt8; 20], amount: UFix64) {
    let fundVault: @FlowToken.Vault

    prepare(signer: auth(Storage) &Account) {
        let vaultRef = signer.storage.borrow<auth(FungibleToken.Withdrawable) &FlowToken.Vault>(from: /storage/flowTokenVault)
            ?? panic("Could not borrow reference to the owner's Vault!")

        self.fundVault <- vaultRef.withdraw(amount: amount) as! @FlowToken.Vault
    }

    execute {
        let fundAddress = EVM.EVMAddress(bytes: address)
        fundAddress.deposit(from: <-self.fundVault)
    }
}