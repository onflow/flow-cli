import EVM from 0xf8d6e0586b0a20c7 // todo dynamically set

access(all) fun main(caller: Address, contractAddress: [UInt8; 20], data: [UInt8]): [UInt8] {
    let bridgedAccount <- EVM.createBridgedAccount()
    let evmAddress = EVM.EVMAddress(bytes: contractAddress)

    let evmResult = bridgedAccount.call(
        to: evmAddress,
        data: data,
        gasLimit: 300000,
        value: EVM.Balance(flow: 0.0)
    )

    log(evmResult)
    destroy bridgedAccount
    return evmResult
}