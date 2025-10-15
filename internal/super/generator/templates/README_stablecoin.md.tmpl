# USDF Stablecoin PiggyBank Tutorial

This project demonstrates how to build a simple PiggyBank contract that works with the USDF stablecoin on Flow blockchain. USDF is a production stablecoin deployed on Flow Mainnet.

## Overview

This tutorial uses the USDF stablecoin contract deployed on mainnet (`1e4aa0b87d10b141`). For development and testing purposes, a mock version of USDF is available:

- **Mainnet**: Uses the real USDF contract at `1e4aa0b87d10b141`
- **Emulator**: Uses USDF_MOCK contract at `f8d6e0586b0a20c7`
- **Testnet**: Uses USDF_MOCK contract (alias configured in `flow.json`)

The contract aliases in `flow.json` ensure **the same scripts and transactions work seamlessly across all environments** without any code changes.

## Prerequisites

Before running this tutorial, you need to install the Flow CLI:

- **Installation Guide**: https://developers.flow.com/build/tools/flow-cli

The Flow CLI is required to run the emulator, deploy contracts, and execute transactions and scripts.

## Tutorial

Follow these steps to run the complete PiggyBank tutorial:

### 1. Start the Flow Emulator

```bash
flow emulator
```

This starts a local Flow blockchain for development.

### 2. Deploy the Contracts

```bash
flow project deploy
```

This deploys the USDF_MOCK and PiggyBank contracts to the emulator.

### 3. Setup the USDF Mock Vault

```bash
flow transactions send cadence/transactions/SetupUSDFMockVault.cdc --signer emulator-account
```

This creates a USDF vault in the emulator account's storage to hold tokens.

### 4. Check Initial PiggyBank Balance

```bash
flow scripts execute cadence/scripts/GetPiggyBankBalance.cdc
```

This should return `0.00` as the PiggyBank starts empty.

### 5. Mint USDF Tokens

```bash
flow transactions send cadence/transactions/MintUSDFMock.cdc 100.00 f8d6e0586b0a20c7 --signer emulator-account
```

This mints 100.00 USDF tokens to the emulator account (`f8d6e0586b0a20c7`).

### 6. Deposit to PiggyBank

```bash
flow transactions send cadence/transactions/DepositToPiggyBank.cdc 50.00 --signer emulator-account
```

This deposits 50.00 USDF tokens into the PiggyBank. Check the balance again to see it now shows `50.00`.

### 7. Withdraw from PiggyBank

```bash
flow transactions send cadence/transactions/WithdrawFromPiggyBank.cdc 25.00 --signer emulator-account
```

This withdraws 25.00 USDF tokens from the PiggyBank. The PiggyBank balance should now be `25.00`.

### 8. Check User USDF Balance

```bash
flow scripts execute cadence/scripts/GetUserUSDFBalance.cdc f8d6e0586b0a20c7
```

This checks the USDF balance in the user's vault. After the transactions above, it should show `75.00` (100 minted - 50 deposited + 25 withdrawn).

## Project Structure

- `cadence/contracts/` - Smart contracts (PiggyBank and USDF_MOCK)
- `cadence/transactions/` - Transaction files for interacting with contracts
- `cadence/scripts/` - Read-only scripts for querying blockchain state
- `flow.json` - Flow project configuration with contract aliases

## Next Steps

Try modifying the amounts in the transactions or create your own transactions to interact with the PiggyBank contract!
