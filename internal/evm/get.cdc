import EVM from 0xf8d6e0586b0a20c7 // todo dynamically set

access(all) fun main(address: [UInt8; 20]): UFix64 {
    let fundAddress = EVM.EVMAddress(bytes: address)

    return fundAddress.balance().flow
}