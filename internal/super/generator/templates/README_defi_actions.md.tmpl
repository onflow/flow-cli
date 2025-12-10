# DeFi Actions Starter

This project demonstrates how to build with Flow Actions - a standard for composable DeFi connectors on the Flow blockchain.

## Overview

This starter project includes a minimal example connector (`TokenSink`) that implements the `DeFiActions.Sink` interface. It shows how to:

- Create a connector that accepts fungible tokens via a receiver capability
- Compose transactions using Flow Actions patterns
- Test connectors using the Flow Testing Framework

## Quick Start

Get started in seconds:

```bash
flow test  # Run tests to verify everything works
```

That's it! The test deploys all contracts and executes an example transaction.

## Prerequisites

Before running this project, you need to install the Flow CLI:

- **Installation Guide**: https://developers.flow.com/tools/flow-cli/install

**Note**: All dependencies (FungibleToken, FlowToken, DeFiActions, etc.) are already installed in your project during initialization.

## Getting Started

### 1. Run Tests (Recommended First Step)

```bash
flow test
```

This runs the test suite to verify everything works. The tests automatically deploy all contracts and execute the example transaction in an isolated test environment.

### 2. Start the Flow Emulator

```bash
flow emulator
```

This starts a local Flow blockchain for development.

### 3. Deploy the Contracts

In a new terminal:

```bash
flow project deploy --network emulator
```

This deploys all required contracts to the emulator:
- `DeFiActionsUtils` - Helper utilities for DeFi Actions
- `DeFiActions` - Core DeFi Actions framework
- `ExampleConnectors` - Your TokenSink connector

### 4. Run the Example Transaction

Send tokens to yourself using the TokenSink:

```bash
flow transactions send cadence/transactions/DepositViaSink.cdc \
  --signer emulator-account \
  --network emulator \
  --args-json '[{"type":"Address","value":"0xf8d6e0586b0a20c7"},{"type":"UFix64","value":"1.0"}]'
```

This sends `1.0` FLOW from the emulator account to itself (`0xf8d6e0586b0a20c7`) using the `TokenSink` connector.

## Testing

This project includes two types of tests to demonstrate different testing approaches:

### Unit Tests

Run the test suite:

```bash
flow test cadence/tests/ExampleConnectors_test.cdc
```

This test runs in the Cadence Testing Framework and manually:
1. Deploys all DeFi Actions dependencies (`DeFiActionsUtils`, `DeFiActions`)
2. Deploys the `ExampleConnectors` contract
3. Executes the `DepositViaSink` transaction
4. Verifies tokens are deposited successfully

**Note**: The testing framework manages its own test environment - no emulator needs to be running.

### Fork Testing

This project includes a fork test (`test_incrementfi_swap_on_fork.cdc`) that demonstrates testing your contracts against real mainnet state. Fork testing allows you to:

- Test interactions with production DeFi protocols (like IncrementFi)
- Validate your connectors against real deployed contracts
- Use actual mainnet account data without deploying anything
- Debug issues with historical blockchain state

Run the fork test against mainnet:

```bash
flow test cadence/tests/test_incrementfi_swap_on_fork.cdc
```

The fork test executes a real swap from FLOW → stFlow using IncrementFi's deployed contracts on mainnet. It uses account impersonation to test transactions as any mainnet account, with all changes happening locally in your test environment.

**Learn more**: See the [Fork Testing Tutorial](https://developers.flow.com/blockchain-development-tutorials/cadence/fork-testing) and [Testing Strategy Guide](https://developers.flow.com/build/cadence/smart-contracts/testing-strategy) for detailed information on when and how to use fork testing.

## Project Structure

- `cadence/contracts/` - Smart contracts
  - `ExampleConnectors.cdc` - TokenSink connector implementation
- `cadence/transactions/` - Transaction files
  - `DepositViaSink.cdc` - Example transaction using TokenSink
  - `incrementfi_swap.cdc` - IncrementFi swap transaction
- `cadence/tests/` - Test files
  - `ExampleConnectors_test.cdc` - Integration test for TokenSink
  - `test_incrementfi_swap_on_fork.cdc` - Fork test against mainnet IncrementFi
- `flow.json` - Flow project configuration with DeFiActions dependencies

## Dependencies

This project includes the following dependencies (already installed):

**Core Dependencies:**
- `FungibleToken` - Standard fungible token interface
- `FlowToken` - Native FLOW token implementation

**DeFi Actions Framework:**
- `DeFiActions` - Core framework for composable DeFi connectors
- `DeFiActionsUtils` - Helper utilities

**DeFi Protocol Dependencies:**
- `stFlowToken` - Liquid staking token (used in fork test example)
- `SwapConfig` - IncrementFi swap configuration utilities
- `IncrementFiSwapConnectors` - IncrementFi swap connectors (used in fork test example)

**Network Configuration:**
- **Testnet**: DeFi Actions contracts at `0x0b11b1848a8aa2c0`, IncrementFi at `0x494536c102537e1e`
- **Mainnet**: DeFi Actions contracts at `0x6d888f175c158410`, IncrementFi at `0xe844c7cf7430a77c`
- **Emulator**: Contracts are deployed from source to your emulator account

## Understanding the TokenSink Connector

The `TokenSink` connector demonstrates a minimal implementation of the `DeFiActions.Sink` interface:

- Accepts a `FungibleToken.Receiver` capability (publicly available)
- Deposits tokens into the recipient's vault via `depositCapacity()`
- Includes type checks to ensure safe deposits

## Next Steps

- Explore the [Flow Actions FLIP](https://github.com/onflow/flips/pull/339) for more details on the standard
- Build your own connectors implementing `Source`, `Sink`, or `Swapper` interfaces
- Compose complex DeFi operations by chaining multiple connectors

## Resources

- **Flow Actions Repository**: https://github.com/onflow/FlowActions
- **Flow Documentation**: https://developers.flow.com/
- **Cadence Language**: https://cadence-lang.org/docs/language

## Community

- [Flow Community Forum](https://forum.flow.com/)
- [Flow Discord](https://discord.gg/flow)
- [Flow Twitter](https://x.com/flow_blockchain)

