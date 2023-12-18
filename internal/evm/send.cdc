import EVM from 0xf8d6e0586b0a20c7 // todo dynamically set

transaction(encodedTx: [UInt8]) {
    let bridgedAccount: @EVM.BridgedAccount
    let auth: AuthAccount

    prepare(signer: AuthAccount) {
        self.auth = signer
        self.bridgedAccount <- signer.load<@EVM.BridgedAccount>(from: StoragePath(identifier: "evm")!)!
    }

    pre {
        self.bridgedAccount != nil : "the transaction signer should have already created a bridged account and stored it in its storage under 'evm' path"
    }

    execute {
        // we use the signer of the transaction for coinbase
        EVM.run(tx: encodedTx, coinbase: self.bridgedAccount.address())

        log("transaction executed")
        self.auth.save<@EVM.BridgedAccount>(<-self.bridgedAccount, to: StoragePath(identifier: "evm")!)
    }
}
