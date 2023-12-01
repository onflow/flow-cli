import EVM from 0xf8d6e0586b0a20c7 // todo dynamically set

pub fun main(address: Address): [String] {
    let acc = getAuthAccount(address)

    let evmAcc <- acc.load<@EVM.BridgedAccount>(from: StoragePath(identifier: "evm")!)!
    let balance = evmAcc.balance()

    let address = evmAcc.address().bytes

    destroy evmAcc

    return ["1","2"]
}