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
- `DeFiActionsMathUtils` - Math utilities for DeFi Actions
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

Run the test suite to verify the connector works correctly:

```bash
flow test
```

The tests run in an isolated environment and automatically:
1. Deploy all DeFi Actions dependencies (`DeFiActionsMathUtils`, `DeFiActionsUtils`, `DeFiActions`)
2. Deploy the `ExampleConnectors` contract
3. Execute the `DepositViaSink` transaction
4. Verify tokens are deposited successfully

**Note**: Tests don't require the emulator to be running - they use their own test environment.

## Project Structure

- `cadence/contracts/` - Smart contracts
  - `ExampleConnectors.cdc` - TokenSink connector implementation
- `cadence/transactions/` - Transaction files
  - `DepositViaSink.cdc` - Example transaction using TokenSink
- `cadence/tests/` - Test files
  - `ExampleConnectors_test.cdc` - Integration test for TokenSink
- `flow.json` - Flow project configuration with DeFiActions dependencies

## Dependencies

This project includes the following dependencies (already installed):

**Core Dependencies:**
- `FungibleToken` - Standard fungible token interface
- `FlowToken` - Native FLOW token implementation

**DeFi Actions Framework:**
- `DeFiActions` - Core framework for composable DeFi connectors
- `DeFiActionsUtils` - Helper utilities
- `DeFiActionsMathUtils` - Math utilities for DeFi operations

**Network Configuration:**
- **Testnet**: All DeFi Actions contracts available at `0x4c2ff9dd03ab442f`
- **Mainnet**: All DeFi Actions contracts available at `0x92195d814edf9cb0`
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

