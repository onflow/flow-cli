import EVM from 0xf8d6e0586b0a20c7 // todo dynamically set
import FungibleToken from 0xee82856bf20e2aa6
import FlowToken from 0x0ae53cb6e3f42a79

transaction(bytecode: String) {
    let sentVault: @FlowToken.Vault

    prepare(signer: AuthAccount) {
        let vaultRef = signer.borrow<&FlowToken.Vault>(from: /storage/flowTokenVault)
            ?? panic("Could not borrow reference to the owner's Vault!")

        self.sentVault <- vaultRef.withdraw(amount: 1.0) as! @FlowToken.Vault
    }

    execute {
        let decodedCode = bytecode.decodeHex()

        let bridgedAccount <- EVM.createBridgedAccount()
        bridgedAccount.address().deposit(from: <-self.sentVault)

        let address = bridgedAccount.deploy(
           code: decodedCode,
           gasLimit: 300000,
           value: EVM.Balance(flow: 0.0)
        )

        destroy bridgedAccount
    }
}