---
name: query-blockchain
description: Use when you need to read any data from the Flow blockchain — account state, blocks, events, transaction results, collections, or custom contract state via Cadence scripts.
---

# Querying the Flow Blockchain with flow-cli

## Overview

Use this skill any time you need on-chain data. Choose the right command from the decision table, then run it.

**When the entity commands don't expose what you need, use `flow scripts execute`** — Cadence scripts can query any on-chain state and are the primary tool for anything not covered by the commands below. See [Cadence Scripts](#cadence-scripts).

## Network

Default: **mainnet**. Infer from conversation context:

| Context | Flag |
|---|---|
| Default / production | `--network mainnet` |
| Testnet discussion | `--network testnet` |
| Local development / emulator | `--network emulator` |

Access node endpoints (built-in):
- Mainnet: `access.mainnet.nodes.onflow.org:9000`
- Testnet: `access.devnet.nodes.onflow.org:9000`
- Emulator: `127.0.0.1:3569`

Override with `--host <endpoint>` to point at a custom access node.

## Decision Table

| What you need | Command |
|---|---|
| Account balance, keys, deployed contracts | `flow accounts get` |
| Account staking info | `flow accounts staking-info` |
| Block info | `flow blocks get` |
| Events emitted in a block range | `flow events get` |
| Regular transaction status / result | `flow transactions get` |
| System transaction for a block | `flow transactions get-system` |
| Scheduled transaction details | `flow schedule get` / `flow schedule list` |
| Collection contents | `flow collections get` |
| Network status (online/offline) | `flow status` |
| Protocol state snapshot | `flow snapshot save` |
| Anything not covered above | `flow scripts execute` |

---

## Commands

### Accounts

```bash
flow accounts get <address|name> [--include contracts] [--network mainnet]
```

- `--include contracts` adds deployed contract source code to the output
- Flow addresses must include the `0x` prefix (e.g. `0xf8d6e0586b0a20c7`)
- `<name>` resolves via `flow.json` — only use names when a `flow.json` is present

```bash
flow accounts get 0xe467b9dd11fa00df --network mainnet
flow accounts get 0xe467b9dd11fa00df --include contracts --network mainnet
flow accounts staking-info 0xe467b9dd11fa00df --network mainnet
```

### Blocks

```bash
flow blocks get <block_id|latest|block_height> [--include transactions] [--events <event_name>] [--network mainnet]
```

```bash
flow blocks get latest --network mainnet
flow blocks get 12884163 --include transactions --network mainnet
flow blocks get latest --events A.1654653399040a61.FlowToken.TokensDeposited --network mainnet
```

### Events

```bash
flow events get <event_name> [<event_name2> ...] [--last 10] [--start N --end M] [--network mainnet]
```

- Default: last 10 blocks. Use `--last N` to widen.
- `--start`/`--end` for explicit block height range.
- Multiple event types are fetched in parallel.
- Event name format: `A.<address>.<ContractName>.<EventName>`

```bash
flow events get A.1654653399040a61.FlowToken.TokensDeposited --last 20 --network mainnet
flow events get A.1654653399040a61.FlowToken.TokensDeposited --start 11559500 --end 11559600 --network mainnet
flow events get A.1654653399040a61.FlowToken.TokensDeposited A.1654653399040a61.FlowToken.TokensWithdrawn --network mainnet
```

### Transactions

```bash
# Regular transaction
flow transactions get <tx_id> [--include signatures,code,payload,fee-events] [--exclude events] [--network mainnet]

# System transaction (by block, not tx hash)
flow transactions get-system <block_id|latest|block_height> [tx_id] [--network mainnet]

# Scheduled transactions
flow schedule get <numeric-id> [--network mainnet]
flow schedule list <address|account-name> [--network mainnet]
```

### Collections

```bash
flow collections get <collection_id> [--network mainnet]
```

### Network Status

```bash
flow status --network mainnet
```

---

## Output Format

All commands support `--output json` for machine-readable output.

```bash
flow accounts get 0xe467b9dd11fa00df --output json --network mainnet
flow events get A.1654653399040a61.FlowToken.TokensDeposited --output json --network mainnet
```

Use `--filter <property>` to extract specific fields from results.

---

## Cadence Scripts

`flow scripts execute` is the most powerful read tool. Use it when:

- You need data from a contract that has no dedicated CLI command
- You need to call a `view` function or read a field from a contract
- You need to combine data from multiple contracts in one query
- You need a historical snapshot at a specific block height

```bash
flow scripts execute <script.cdc> [args...] [--args-json '[{"type":"...","value":"..."}]'] [--block-height N] [--block-id <id>] [--network mainnet]
```

- Simple types (Address, UInt64, String, Bool) can be passed as positional args
- Use `--args-json` for complex types (UFix64, optionals, structs, arrays)
- `--block-height` / `--block-id` execute against historical state
- Write a temporary `.cdc` file, execute it, then clean up

### Writing and Running Scripts

Write script to a temp file, execute, clean up:

```bash
# Write
cat > /tmp/query.cdc << 'EOF'
import FungibleToken from 0xf233dcee88fe0abe
import FlowToken from 0x1654653399040a61

access(all) fun main(address: Address): UFix64 {
    let account = getAccount(address)
    let vaultRef = account.capabilities
        .borrow<&{FungibleToken.Balance}>(/public/flowTokenBalance)
        ?? panic("Could not borrow balance capability")
    return vaultRef.balance
}
EOF

# Execute
flow scripts execute /tmp/query.cdc 0xe467b9dd11fa00df --network mainnet

# Clean up
rm /tmp/query.cdc
```

### Passing Arguments

```bash
# Simple types as positional args
flow scripts execute /tmp/query.cdc 0xe467b9dd11fa00df --network mainnet

# Complex types with --args-json (JSON-Cadence encoding)
flow scripts execute /tmp/query.cdc --args-json '[{"type":"UFix64","value":"100.0"},{"type":"Address","value":"0xe467b9dd11fa00df"}]' --network mainnet

# Historical state
flow scripts execute /tmp/query.cdc 0xe467b9dd11fa00df --block-height 12884163 --network mainnet
```

---

## Contract Addresses

| Contract | Mainnet | Testnet | Emulator |
|---|---|---|---|
| FungibleToken | `0xf233dcee88fe0abe` | `0x9a0766d93b6608b7` | `0xee82856bf20e2aa6` |
| FungibleTokenMetadataViews | `0xf233dcee88fe0abe` | `0x9a0766d93b6608b7` | `0xf8d6e0586b0a20c7` |
| FungibleTokenSwitchboard | `0xf233dcee88fe0abe` | `0x9a0766d93b6608b7` | `0xf8d6e0586b0a20c7` |
| Burner | `0xf233dcee88fe0abe` | `0x9a0766d93b6608b7` | `0xf8d6e0586b0a20c7` |
| FlowToken | `0x1654653399040a61` | `0x7e60df042a9c0868` | `0x0ae53cb6e3f42a79` |
| NonFungibleToken | `0x1d7e57aa55817448` | `0x631e88ae7f1d7c20` | `0xf8d6e0586b0a20c7` |
| MetadataViews | `0x1d7e57aa55817448` | `0x631e88ae7f1d7c20` | `0xf8d6e0586b0a20c7` |
| ViewResolver | `0x1d7e57aa55817448` | `0x631e88ae7f1d7c20` | `0xf8d6e0586b0a20c7` |
| FlowFees | `0xf919ee77447b7497` | `0x912d5440f7e3769e` | `0xe5a8b7f23e8b548f` |
| FlowServiceAccount | `0xe467b9dd11fa00df` | `0x8c5303eaa26202d6` | `0xf8d6e0586b0a20c7` |
| FlowStorageFees | `0xe467b9dd11fa00df` | `0x8c5303eaa26202d6` | `0xf8d6e0586b0a20c7` |
| NodeVersionBeacon | `0xe467b9dd11fa00df` | `0x8c5303eaa26202d6` | `0xf8d6e0586b0a20c7` |
| RandomBeaconHistory | `0xe467b9dd11fa00df` | `0x8c5303eaa26202d6` | `0xf8d6e0586b0a20c7` |
| FlowIDTableStaking | `0x8624b52f9ddcd04a` | `0x9eca2b38b18b5dfe` | `0xf8d6e0586b0a20c7` |
| FlowEpoch | `0x8624b52f9ddcd04a` | `0x9eca2b38b18b5dfe` | `0xf8d6e0586b0a20c7` |
| FlowClusterQC | `0x8624b52f9ddcd04a` | `0x9eca2b38b18b5dfe` | `0xf8d6e0586b0a20c7` |
| FlowDKG | `0x8624b52f9ddcd04a` | `0x9eca2b38b18b5dfe` | `0xf8d6e0586b0a20c7` |
| FlowStakingCollection | `0x8d0e87b65159ae63` | `0x95e019a17d0e23d7` | `0xf8d6e0586b0a20c7` |
| LockedTokens | `0x8d0e87b65159ae63` | `0x95e019a17d0e23d7` | `0xf8d6e0586b0a20c7` |
| StakingProxy | `0x62430cf28c26d095` | `0x7aad92e5a0715d21` | `0xf8d6e0586b0a20c7` |
| EVM | `0xe467b9dd11fa00df` | `0x8c5303eaa26202d6` | `0xf8d6e0586b0a20c7` |

---

## Cadence Script Recipes & Data Structures

See [cadence-scripts.md](cadence-scripts.md) for 20+ ready-to-use Cadence scripts organized by category:
- **Token queries** — FLOW balance, total supply, generic FT balance, FT metadata
- **Account & storage** — storage capacity, available balance, account creation fee, fee parameters
- **Epoch** — counter, phase, metadata, timing config
- **Staking** — node info, staked node IDs, total staked, by role, requirements, rewards, delegator info, staking collections
- **Protocol** — node version beacon, random beacon source
- **NFT** — collection IDs, NFT metadata (Display)
- **Key data structures** — NodeInfo, DelegatorInfo, EpochMetadata, EpochPhase, Node Roles

---

## Common Event Types

| Event | Description |
|---|---|
| `A.f233dcee88fe0abe.FungibleToken.Deposited` | Any fungible token deposited |
| `A.f233dcee88fe0abe.FungibleToken.Withdrawn` | Any fungible token withdrawn |
| `A.f233dcee88fe0abe.FungibleToken.Burned` | Any fungible token burned |
| `A.1d7e57aa55817448.NonFungibleToken.Deposited` | Any NFT deposited to a collection |
| `A.1d7e57aa55817448.NonFungibleToken.Withdrawn` | Any NFT withdrawn from a collection |
| `A.8624b52f9ddcd04a.FlowEpoch.NewEpoch` | New epoch started |
| `A.8624b52f9ddcd04a.FlowEpoch.EpochSetup` | Epoch setup phase began |
| `A.8624b52f9ddcd04a.FlowEpoch.EpochCommit` | Epoch commit phase began |
| `A.8624b52f9ddcd04a.FlowIDTableStaking.NewNodeCreated` | New staking node registered |
| `A.8624b52f9ddcd04a.FlowIDTableStaking.TokensCommitted` | Tokens committed to stake |
| `A.8624b52f9ddcd04a.FlowIDTableStaking.RewardsPaid` | Staking rewards distributed |
| `A.8624b52f9ddcd04a.FlowIDTableStaking.NewDelegatorCreated` | New delegator registered |
| `A.f919ee77447b7497.FlowFees.FeesDeducted` | Transaction fees paid |
| `A.f919ee77447b7497.FlowFees.TokensDeposited` | Fees deposited to fee vault |

---

## Available Script Libraries

For more complex queries, clone these repos to `/tmp` and use their scripts directly:

| Repo | Scripts Path | Use For |
|---|---|---|
| [flow-core-contracts](https://github.com/onflow/flow-core-contracts) | `transactions/*/scripts/` | Staking, epoch, fees, locked tokens, version beacon, random beacon |
| [flow-ft](https://github.com/onflow/flow-ft) | `transactions/scripts/`, `transactions/metadata/scripts/` | FT balances, supply, metadata, switchboard |
| [flow-nft](https://github.com/onflow/flow-nft) | `transactions/scripts/` | NFT collections, metadata views, cross-VM views |

```bash
# Example: use an existing script from flow-core-contracts
git clone --depth 1 https://github.com/onflow/flow-core-contracts.git /tmp/flow-core-contracts
flow scripts execute /tmp/flow-core-contracts/transactions/idTableStaking/scripts/get_node_info.cdc "abc123...def456" --network mainnet
```

Note: Some repo scripts use `import "ContractName"` syntax (no address). These require a `flow.json` with address mappings. For ad-hoc queries, replace with explicit addresses:
```cadence
// Repo style (requires flow.json aliases):
import "FlowToken"
// Direct style (works without flow.json):
import FlowToken from 0x1654653399040a61
```
