# Cadence Script Recipes

Reference repos for existing scripts (clone to /tmp if you need full source):
- [flow-core-contracts](https://github.com/onflow/flow-core-contracts) — staking, epoch, fees, token
- [flow-ft](https://github.com/onflow/flow-ft) — fungible token standard
- [flow-nft](https://github.com/onflow/flow-nft) — NFT standard, metadata views

All scripts below use **mainnet** addresses. For testnet, substitute addresses from the contract address tables in SKILL.md.

---

## Token Queries

### FLOW Token Balance

```cadence
import FungibleToken from 0xf233dcee88fe0abe
import FlowToken from 0x1654653399040a61

access(all) fun main(address: Address): UFix64 {
    let account = getAccount(address)
    let vaultRef = account.capabilities
        .borrow<&{FungibleToken.Balance}>(/public/flowTokenBalance)
        ?? panic("Could not borrow FungibleToken Balance capability for account ".concat(address.toString()).concat(" at path /public/flowTokenBalance. Make sure the account has a FlowToken Vault set up properly."))
    return vaultRef.balance
}
```

### FLOW Total Supply

```cadence
import FlowToken from 0x1654653399040a61

access(all) fun main(): UFix64 {
    return FlowToken.totalSupply
}
```

### Generic FT Balance (Any Token by Path)

```cadence
import FungibleToken from 0xf233dcee88fe0abe

access(all) fun main(address: Address, path: PublicPath): UFix64 {
    return getAccount(address).capabilities
        .borrow<&{FungibleToken.Balance}>(path)?.balance
        ?? panic("Could not borrow FungibleToken Balance capability for account ".concat(address.toString()).concat(" at path ").concat(path.toString()).concat(". Make sure the account has a Fungible Token Vault set up at this path."))
}
```

### FT Metadata (Display Info)

```cadence
import FungibleToken from 0xf233dcee88fe0abe
import FungibleTokenMetadataViews from 0xf233dcee88fe0abe

access(all) fun main(address: Address, vaultPath: PublicPath): FungibleTokenMetadataViews.FTDisplay? {
    let account = getAccount(address)
    let vaultRef = account.capabilities
        .borrow<&{FungibleToken.Vault}>(vaultPath)
        ?? panic("Could not borrow FungibleToken Vault capability for account ".concat(address.toString()).concat(" at path ").concat(vaultPath.toString()).concat(". Make sure the account has a Fungible Token Vault set up at this path."))
    return FungibleTokenMetadataViews.getFTDisplay(vaultRef)
}
```

---

## Account & Storage Queries

### Account Storage Capacity

```cadence
import FlowStorageFees from 0xe467b9dd11fa00df

access(all) fun main(address: Address): UFix64 {
    return FlowStorageFees.calculateAccountCapacity(address)
}
```

### Account Available Balance (After Storage Reservation)

```cadence
import FlowStorageFees from 0xe467b9dd11fa00df

access(all) fun main(address: Address): UFix64 {
    return FlowStorageFees.defaultTokenAvailableBalance(address)
}
```

### Account Creation Fee

```cadence
import FlowServiceAccount from 0xe467b9dd11fa00df

access(all) fun main(): UFix64 {
    return FlowServiceAccount.accountCreationFee
}
```

### Transaction Fee Parameters

```cadence
import FlowFees from 0xf919ee77447b7497

access(all) fun main(): FlowFees.FeeParameters {
    return FlowFees.getFeeParameters()
}
```

---

## Epoch Queries

### Current Epoch Counter

```cadence
import FlowEpoch from 0x8624b52f9ddcd04a

access(all) fun main(): UInt64 {
    return FlowEpoch.currentEpochCounter
}
```

### Current Epoch Phase

```cadence
import FlowEpoch from 0x8624b52f9ddcd04a

// Returns: 0=StakingAuction, 1=EpochSetup, 2=EpochCommit
access(all) fun main(): UInt8 {
    return FlowEpoch.currentEpochPhase.rawValue
}
```

### Epoch Metadata

```cadence
import FlowEpoch from 0x8624b52f9ddcd04a

access(all) fun main(counter: UInt64): FlowEpoch.EpochMetadata {
    return FlowEpoch.getEpochMetadata(counter)!
}
```

### Epoch Timing Config

```cadence
import FlowEpoch from 0x8624b52f9ddcd04a

access(all) fun main(): FlowEpoch.EpochTimingConfig {
    return FlowEpoch.getEpochTimingConfig()
}
```

---

## Staking Queries

### Node Info

```cadence
import FlowIDTableStaking from 0x8624b52f9ddcd04a

access(all) fun main(nodeID: String): FlowIDTableStaking.NodeInfo {
    return FlowIDTableStaking.NodeInfo(nodeID: nodeID)
}
```

### All Staked Node IDs

```cadence
import FlowIDTableStaking from 0x8624b52f9ddcd04a

access(all) fun main(): [String] {
    return FlowIDTableStaking.getStakedNodeIDs()
}
```

### Total FLOW Staked

```cadence
import FlowIDTableStaking from 0x8624b52f9ddcd04a

access(all) fun main(): UFix64 {
    return FlowIDTableStaking.getTotalStaked()
}
```

### Total Staked by Node Role

```cadence
import FlowIDTableStaking from 0x8624b52f9ddcd04a

// Roles: 1=Collector, 2=Consensus, 3=Execution, 4=Verification, 5=Access
access(all) fun main(role: UInt8): UFix64 {
    return FlowIDTableStaking.getTotalTokensStakedByNodeType()[role]!
}
```

### Stake Requirements by Node Type

```cadence
import FlowIDTableStaking from 0x8624b52f9ddcd04a

access(all) fun main(): {UInt8: UFix64} {
    return FlowIDTableStaking.getMinimumStakeRequirements()
}
```

### Weekly Epoch Reward Payout

```cadence
import FlowIDTableStaking from 0x8624b52f9ddcd04a

access(all) fun main(): UFix64 {
    return FlowIDTableStaking.getEpochTokenPayout()
}
```

### Delegator Reward Cut Percentage

```cadence
import FlowIDTableStaking from 0x8624b52f9ddcd04a

access(all) fun main(): UFix64 {
    return FlowIDTableStaking.getRewardCutPercentage()
}
```

### Delegator Info

```cadence
import FlowIDTableStaking from 0x8624b52f9ddcd04a

access(all) fun main(nodeID: String, delegatorID: UInt32): FlowIDTableStaking.DelegatorInfo {
    return FlowIDTableStaking.DelegatorInfo(nodeID: nodeID, delegatorID: delegatorID)
}
```

### Staking Collection — All Node Info for an Account

```cadence
import FlowStakingCollection from 0x8d0e87b65159ae63

access(all) fun main(address: Address): [FlowStakingCollection.NodeInfo] {
    return FlowStakingCollection.getAllNodeInfo(address: address)
}
```

### Staking Collection — All Delegator Info for an Account

```cadence
import FlowStakingCollection from 0x8d0e87b65159ae63

access(all) fun main(address: Address): [FlowStakingCollection.DelegatorInfo] {
    return FlowStakingCollection.getAllDelegatorInfo(address: address)
}
```

---

## Protocol Queries

### Current Node Version

```cadence
import NodeVersionBeacon from 0xe467b9dd11fa00df

access(all) fun main(): NodeVersionBeacon.Semver {
    return NodeVersionBeacon.getCurrentVersionBoundary().version
}
```

### Latest Random Beacon Source

```cadence
import RandomBeaconHistory from 0xe467b9dd11fa00df

access(all) fun main(): RandomBeaconHistory.RandomSourceHistoryEntry {
    return RandomBeaconHistory.getLatestSourceOfRandomness()
}
```

---

## NFT Queries

### NFT Collection IDs

```cadence
import NonFungibleToken from 0x1d7e57aa55817448

access(all) fun main(address: Address, collectionPublicPath: PublicPath): [UInt64] {
    let account = getAccount(address)
    let collectionRef = account.capabilities
        .borrow<&{NonFungibleToken.Collection}>(collectionPublicPath)
        ?? panic("Could not borrow NonFungibleToken Collection capability for account ".concat(address.toString()).concat(" at path ").concat(collectionPublicPath.toString()).concat(". Make sure the account has an NFT Collection set up at this path."))
    return collectionRef.getIDs()
}
```

### NFT Metadata (Display)

```cadence
import NonFungibleToken from 0x1d7e57aa55817448
import MetadataViews from 0x1d7e57aa55817448

access(all) fun main(address: Address, collectionPublicPath: PublicPath, id: UInt64): MetadataViews.Display? {
    let account = getAccount(address)
    let collectionRef = account.capabilities
        .borrow<&{NonFungibleToken.Collection}>(collectionPublicPath)
        ?? panic("Could not borrow NonFungibleToken Collection capability for account ".concat(address.toString()).concat(" at path ").concat(collectionPublicPath.toString()).concat(". Make sure the account has an NFT Collection set up at this path."))
    let nft = collectionRef.borrowNFT(id)!
    return MetadataViews.getDisplay(nft)
}
```

---

## Key Data Structures

### FlowIDTableStaking.NodeInfo
- `id: String` (64 hex chars), `role: UInt8` (1-5)
- `networkingAddress`, `networkingKey`, `stakingKey`: String
- Token buckets: `tokensStaked`, `tokensCommitted`, `tokensUnstaking`, `tokensUnstaked`, `tokensRewarded`: UFix64
- `delegators`, `delegatorIDCounter`, `initialWeight`

### FlowIDTableStaking.DelegatorInfo
- `id: UInt32`, `nodeID: String`
- Token buckets: `tokensCommitted`, `tokensStaked`, `tokensUnstaking`, `tokensRewarded`, `tokensUnstaked`: UFix64

### FlowEpoch.EpochMetadata
- `counter: UInt64`, `seed: String`, `startView`, `endView`, `stakingEndView`: UInt64
- `totalRewards: UFix64`, `rewardsPaid: Bool`
- `collectorClusters`, `clusterQCs`, `dkgKeys`

### FlowEpoch.EpochPhase
- `STAKINGAUCTION (0)`, `EPOCHSETUP (1)`, `EPOCHCOMMIT (2)`

### Node Roles
- 1=Collector, 2=Consensus, 3=Execution, 4=Verification, 5=Access
