import "FungibleToken"
import "FungibleTokenMetadataViews"
import "MetadataViews"
import "ViewResolver"

/// USDF_MOCK is a simplified mock version of the EVMVMBridgedToken_USDF contract
/// designed for testing purposes on Flow testnet. It implements the FungibleToken
/// interface but removes all EVM bridge complexity.
///
/// This contract includes a public mint function so users can mint tokens for
/// testing without needing a faucet.
///
access(all) contract EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed: FungibleToken {
    /// Total supply of tokens in existence
    access(all) var totalSupply: UFix64

    /// Token metadata
    access(all) let name: String
    access(all) let symbol: String
    access(all) let decimals: UInt8

    /// Events
    access(all) event TokensInitialized(initialSupply: UFix64)
    access(all) event TokensWithdrawn(amount: UFix64, from: Address?)
    access(all) event TokensDeposited(amount: UFix64, to: Address?)
    access(all) event TokensMinted(amount: UFix64, to: Address?)
    access(all) event TokensBurned(amount: UFix64, from: Address?)

    /// Storage and Public Paths
    access(all) let VaultStoragePath: StoragePath
    access(all) let VaultPublicPath: PublicPath
    access(all) let ReceiverPublicPath: PublicPath
    access(all) let MinterStoragePath: StoragePath

    /// The Vault resource that holds the tokens
    access(all) resource Vault: FungibleToken.Vault {
        /// The total balance of this vault
        access(all) var balance: UFix64

        init(balance: UFix64) {
            self.balance = balance
        }

        /// getSupportedVaultTypes returns a list of vault types that this receiver accepts
        access(all) view fun getSupportedVaultTypes(): {Type: Bool} {
            return {Type<@EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.Vault>(): true}
        }

        access(all) view fun isSupportedVaultType(type: Type): Bool {
            return self.getSupportedVaultTypes()[type] ?? false
        }

        /// Asks if the amount can be withdrawn from this vault
        access(all) view fun isAvailableToWithdraw(amount: UFix64): Bool {
            return amount <= self.balance
        }

        /// Gets the token name - for interface compatibility with original contract
        access(all) view fun getName(): String {
            return EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.name
        }

        /// Gets the token symbol - for interface compatibility with original contract
        access(all) view fun getSymbol(): String {
            return EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.symbol
        }

        /// Gets the token decimals - for interface compatibility with original contract
        access(all) view fun getDecimals(): UInt8 {
            return EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.decimals
        }

        /// Returns the mock EVM contract address - for interface compatibility with original contract
        access(all) view fun getEVMContractAddress(): String? {
            return nil
        }

        /// withdraw subtracts `amount` from the vault's balance
        /// and returns a new vault with the subtracted balance
        access(FungibleToken.Withdraw) fun withdraw(amount: UFix64): @{FungibleToken.Vault} {
            self.balance = self.balance - amount
            emit TokensWithdrawn(amount: amount, from: self.owner?.address)
            return <-create Vault(balance: amount)
        }

        /// deposit takes a vault and adds its balance to the balance of this vault
        access(all) fun deposit(from: @{FungibleToken.Vault}) {
            let vault <- from as! @EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.Vault
            self.balance = self.balance + vault.balance
            emit TokensDeposited(amount: vault.balance, to: self.owner?.address)
            vault.balance = 0.0
            destroy vault
        }

        /// createEmptyVault allows any user to create a new Vault that has a zero balance
        access(all) fun createEmptyVault(): @{FungibleToken.Vault} {
            return <-create Vault(balance: 0.0)
        }

        access(all) view fun getViews(): [Type] {
            return EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.getContractViews(resourceType: nil)
        }

        access(all) fun resolveView(_ view: Type): AnyStruct? {
            return EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.resolveContractView(resourceType: nil, viewType: view)
        }

        /// Called when a fungible token is burned via the `Burner.burn()` method
        access(contract) fun burnCallback() {
            if self.balance > 0.0 {
                EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.totalSupply = EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.totalSupply - self.balance
                emit TokensBurned(amount: self.balance, from: self.owner?.address)
            }
            self.balance = 0.0
        }
    }

    /// createEmptyVault allows any user to create a new Vault that has a zero balance
    access(all) fun createEmptyVault(vaultType: Type): @{FungibleToken.Vault} {
        return <-create Vault(balance: 0.0)
    }

    /// Minter resource allows minting tokens
    /// In this mock version, we'll make minting public for testing
    access(all) resource Minter {
        /// mintTokens mints new tokens and returns them
        access(all) fun mintTokens(amount: UFix64): @{FungibleToken.Vault} {
            pre {
                amount > 0.0: "Amount minted must be greater than zero"
            }
            EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.totalSupply = EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.totalSupply + amount
            return <-create Vault(balance: amount)
        }
    }

    /// Public mint function for testing - allows anyone to mint tokens
    access(all) fun mintTokens(amount: UFix64): @{FungibleToken.Vault} {
        pre {
            amount > 0.0: "Amount minted must be greater than zero"
            amount <= 1000.0: "Cannot mint more than 1000 tokens at once (for testing)"
        }
        EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.totalSupply = EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.totalSupply + amount
        emit TokensMinted(amount: amount, to: nil)
        return <-create Vault(balance: amount)
    }

    /// Gets the name of the token
    access(all) view fun getName(): String {
        return self.name
    }

    /// Gets the symbol of the token
    access(all) view fun getSymbol(): String {
        return self.symbol
    }

    /// Gets the number of decimals
    access(all) view fun getDecimals(): UInt8 {
        return self.decimals
    }

    /// Returns nil for EVM contract address (not needed in mock)
    access(all) view fun getEVMContractAddress(): String? {
        return nil
    }

    /// Function that returns all the Metadata Views implemented by this contract
    access(all) view fun getContractViews(resourceType: Type?): [Type] {
        return [
            Type<FungibleTokenMetadataViews.FTView>(),
            Type<FungibleTokenMetadataViews.FTDisplay>(),
            Type<FungibleTokenMetadataViews.FTVaultData>(),
            Type<FungibleTokenMetadataViews.TotalSupply>(),
            Type<MetadataViews.EVMBridgedMetadata>()
        ]
    }

    /// Function that resolves a metadata view for this contract
    access(all) fun resolveContractView(resourceType: Type?, viewType: Type): AnyStruct? {
        switch viewType {
            case Type<FungibleTokenMetadataViews.FTView>():
                return FungibleTokenMetadataViews.FTView(
                    ftDisplay: self.resolveContractView(resourceType: nil, viewType: Type<FungibleTokenMetadataViews.FTDisplay>()) as! FungibleTokenMetadataViews.FTDisplay?,
                    ftVaultData: self.resolveContractView(resourceType: nil, viewType: Type<FungibleTokenMetadataViews.FTVaultData>()) as! FungibleTokenMetadataViews.FTVaultData?
                )
            case Type<FungibleTokenMetadataViews.FTDisplay>():
                let media = MetadataViews.Media(
                    file: MetadataViews.HTTPFile(url: "https://assets.website-files.com/5f6294c0c7a8cdd643b1c820/5f6294c0c7a8cda55cb1c936_Flow_Wordmark.svg"),
                    mediaType: "image/svg+xml"
                )
                let medias = MetadataViews.Medias([media])
                return FungibleTokenMetadataViews.FTDisplay(
                    name: self.name,
                    symbol: self.symbol,
                    description: "A mock version of USDF token for testing purposes on Flow blockchain",
                    externalURL: MetadataViews.ExternalURL("https://flow.com"),
                    logos: medias,
                    socials: {
                        "twitter": MetadataViews.ExternalURL("https://twitter.com/flow_blockchain")
                    }
                )
            case Type<FungibleTokenMetadataViews.FTVaultData>():
                return FungibleTokenMetadataViews.FTVaultData(
                    storagePath: self.VaultStoragePath,
                    receiverPath: self.ReceiverPublicPath,
                    metadataPath: self.VaultPublicPath,
                    receiverLinkedType: Type<&EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.Vault>(),
                    metadataLinkedType: Type<&EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.Vault>(),
                    createEmptyVaultFunction: (fun(): @{FungibleToken.Vault} {
                        return <-EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.createEmptyVault(vaultType: Type<@EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.Vault>())
                    })
                )
            case Type<FungibleTokenMetadataViews.TotalSupply>():
                return FungibleTokenMetadataViews.TotalSupply(totalSupply: self.totalSupply)
            case Type<MetadataViews.EVMBridgedMetadata>():
                return MetadataViews.EVMBridgedMetadata(
                    name: self.name,
                    symbol: self.symbol,
                    uri: MetadataViews.URI(baseURI: nil, value: "")
                )
        }
        return nil
    }

    init() {
        // Same data as the original contract on EVM side
        // https://evm.flowscan.io/token/0x2aaBea2058b5aC2D339b163C6Ab6f2b6d53aabED?tab=contract
        self.totalSupply = 0.0
        self.name = "USDF MOCK"
        self.symbol = "USDF"
        self.decimals = 6

        // Same paths as mainnet USDF contract for compatibility
        self.VaultStoragePath = /storage/EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabedVault
        self.VaultPublicPath = /public/EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabedVault
        self.ReceiverPublicPath = /public/EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabedReceiver
        self.MinterStoragePath = /storage/USDFMockMinter

        // Create the initial empty vault for the contract account
        let vault <- create Vault(balance: 0.0)
        self.account.storage.save(<-vault, to: self.VaultStoragePath)

        // Create a public capability for the vault
        let vaultCap = self.account.capabilities.storage.issue<&EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.Vault>(self.VaultStoragePath)
        self.account.capabilities.publish(vaultCap, at: self.VaultPublicPath)

        let receiverCap = self.account.capabilities.storage.issue<&{FungibleToken.Receiver}>(self.VaultStoragePath)
        self.account.capabilities.publish(receiverCap, at: self.ReceiverPublicPath)

        // Create and save minter resource
        let minter <- create Minter()
        self.account.storage.save(<-minter, to: self.MinterStoragePath)

        emit TokensInitialized(initialSupply: self.totalSupply)
    }
}
