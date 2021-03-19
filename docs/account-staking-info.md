---
title: Get Account Staking Info with the Fow CLI
sidebar_title: Staking Info
description: How to get staking info
---

Retrieve staking information for the account on the Flow network using Flow CLI.

`flow accounts staking-info <address>`

## Example Usage

```shell
accounts staking-info 535b975637fb6bee --host access.testnet.nodes.onflow.org:9000
```

### Example response

```shell
Account Staking Info:
ID: 			 "ca00101101010100001011010101010101010101010101011010101010101010"
Initial Weight: 	 100
Networking Address: 	 "ca00101101010100001011010101010101010101010101011010101010101010"
Networking Key: 	 "ca00101101010100001011010101010101010101010101011010101010101010ca00101101010100001011010101010101010101010101011010101010101010"
Role: 			 1
Staking Key: 		 "ca00101101010100001011010101010101010101010101011010101010101010ca00101101010100001011010101010101010101010101011010101010101010ca00101101010100001011010101010101010101010101011010101010101010"
Tokens Committed: 	 0.00000000
Tokens To Unstake: 	 0.00000000
Tokens Rewarded: 	 82627.77000000
Tokens Staked: 		 250000.00000000
Tokens Unstaked: 	 0.00000000
Tokens Unstaking: 	 0.00000000
Total Tokens Staked: 	 250000.00000000


Account Delegation Info:
ID: 			 7
Tokens Committed: 	 0.00000000
Tokens To Unstake: 	 0.00000000
Tokens Rewarded: 	 30397.81936000
Tokens Staked: 		 100000.00000000
Tokens Unstaked: 	 0.00000000
Tokens Unstaking: 	 0.00000000

```

## Arguments

### Address
- Name: `address`
- Valid Input: Flow account address

Flow [account address](https://docs.onflow.org/concepts/accounts-and-keys/) (prefixed with `0x` or not).

## Flags

### Host
- Flag: `--host`
- Valid inputs: an IP address or hostname.
- Default: `127.0.0.1:3569` (Flow Emulator)

Specify the hostname of the Access API that will be
used to execute the commands.