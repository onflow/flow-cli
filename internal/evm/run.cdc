import EVM from 0xf8d6e0586b0a20c7 // todo dynamically set

pub fun main(caller: Address, contractAddress: [UInt8; 20], data: [UInt8]): [UInt8] {
    let flowAccount = getAuthAccount(caller)
    let bridgedAccount <- flowAccount.load<@EVM.BridgedAccount>(from: StoragePath(identifier: "evm")!)!
    let evmAddress = EVM.EVMAddress(contractAddress)

    let evmResult = bridgedAccount.call(
        to: evmAddress,
        data: data,
        gasLimit: 300000,
        value: EVM.Balance(flow: 0.0)
    )

    destroy bridgedAccount
    return evmResult
}