import EVM from 0xf8d6e0586b0a20c7 // todo dynamically set

access(all) fun main(address: Address): UFix64 {
    let acc = getAuthAccount(address)

    let evmAcc <- acc.load<@EVM.BridgedAccount>(from: StoragePath(identifier: "evm")!)!
    let balance = evmAcc.balance()

    let address = evmAcc.address().bytes

    var x: [UInt8] = []

    for e in address {
        x.append(e)
    }

    log(String.encodeHex(x))
    destroy evmAcc

    return balance.flow
}