import "EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed"
import "FungibleTokenMetadataViews"

/// This script returns information about the USDF token.
access(all) struct TokenInfo {
    access(all) let name: String
    access(all) let symbol: String
    access(all) let decimals: UInt8
    access(all) let totalSupply: UFix64

    init(name: String, symbol: String, decimals: UInt8, totalSupply: UFix64) {
        self.name = name
        self.symbol = symbol
        self.decimals = decimals
        self.totalSupply = totalSupply
    }
}

access(all) fun main(): TokenInfo {
    // The mainnet contract may not have all the same methods as our mock
    // Try to get basic info that should be available on both contracts
    let ftView = EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.resolveContractView(resourceType: nil, viewType: Type<FungibleTokenMetadataViews.FTDisplay>()) as! FungibleTokenMetadataViews.FTDisplay?

    return TokenInfo(
        name: ftView?.name ?? "USDF",
        symbol: ftView?.symbol ?? "USDF",
        decimals: 6, // USDF has 6 decimals
        totalSupply: EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.totalSupply
    )
}
