import EVM from 0xf8d6e0586b0a20c7 // todo dynamically set

transaction(encodedTx: [UInt8]) {

    prepare(signer: auth(Storage) &Account) {}

    execute {
        // we don't care for fees in demo just create random bridged acc
        let feeAcc <- EVM.createdBridgedAccount()
        // we use the signer of the transaction for coinbase
        EVM.run(tx: encodedTx, coinbase: feeAcc.address())

        destroy feeAcc
        log("transaction executed")
    }
}
