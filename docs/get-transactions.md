---
title: Get a Transaction with the Flow CLI
sidebar_title: Get a Transaction
description: How to get a Flow transaction from the command line
---

The Flow CLI provides a command to fetch a transaction
that was previously submitted to an Access API.

```shell
flow transactions get <tx_id>
```

## Example Usage

```shell
> flow transactions get 40bc4b100c1930c61381c22e0f4c10a7f5827975ee25715527c1061b8d71e5aa --network mainnet 

Status		✅ SEALED
ID		40bc4b100c1930c61381c22e0f4c10a7f5827975ee25715527c1061b8d71e5aa
Payer		18eb4ee6b3c026d2
Authorizers	[18eb4ee6b3c026d2]

Proposal Key:	
    Address	18eb4ee6b3c026d2
    Index	11
    Sequence	17930

Payload Signature 0: 18eb4ee6b3c026d2
Payload Signature 1: 18eb4ee6b3c026d2
Envelope Signature 0: 18eb4ee6b3c026d2
Signatures (minimized, use --include signatures)

Events:		 
    Index	0
    Type	A.1654653399040a61.FlowToken.TokensWithdrawn
    Tx ID	40bc4b100c1930c61381c22e0f4c10a7f5827975ee25715527c1061b8d71e5aa
    Values
		- amount (UFix64):	0.00100000
		- from ({}?):			18eb4ee6b3c026d2

    Index	1
    Type	A.1654653399040a61.FlowToken.TokensDeposited
    Tx ID	40bc4b100c1930c61381c22e0f4c10a7f5827975ee25715527c1061b8d71e5aa
    Values
		- amount (UFix64):	0.00100000
		- to ({}?):			5068e27f275c546c

    Index	2
    Type	A.18eb4ee6b3c026d2.PrivateReceiverForwarder.PrivateDeposit
    Tx ID	40bc4b100c1930c61381c22e0f4c10a7f5827975ee25715527c1061b8d71e5aa
    Values
		- amount (UFix64):	0.00100000
		- to ({}?):			5068e27f275c546c



Code (hidden, use --include code)

Payload (hidden, use --include payload)
```

## Arguments

### Transaction ID

- Name: `<tx_id>`
- Valid Input: transaction ID.

The first argument is the ID (hash) of the transaction.

## Flags
    
### Include Fields

- Flag: `--include`
- Valid inputs: `code`, `payload`, `signatures`

Specify fields to include in the result output. Applies only to the text output.

### Code

- Flag: `--code`

⚠️  Deprecated: use include flag.

### Wait for Seal

- Flag: `--sealed`
- Default: `false`

Indicate whether to wait for the transaction to be sealed
before displaying the result.

### Exclude Fields

- Flag: `--exclude`
- Valid inputs: `events`

Specify fields to exclude from the result output. Applies only to the text output.

### Host

- Flag: `--host`
- Valid inputs: an IP address or hostname.
- Default: `127.0.0.1:3569` (Flow Emulator)

Specify the hostname of the Access API that will be
used to execute the command. This flag overrides
any host defined by the `--network` flag.

### Network Key

- Flag: `--network-key`
- Valid inputs: A valid network public key of the host in hex string format

Specify the network public key of the Access API that will be
used to create a secure GRPC client when executing the command.

### Network

- Flag: `--network`
- Short Flag: `-n`
- Valid inputs: the name of a network defined in the configuration (`flow.json`)
- Default: `emulator`

Specify which network you want the command to use for execution.

### Filter

- Flag: `--filter`
- Short Flag: `-x`
- Valid inputs: a case-sensitive name of the result property.

Specify any property name from the result you want to return as the only value.

### Output

- Flag: `--output`
- Short Flag: `-o`
- Valid inputs: `json`, `inline`

Specify the format of the command results.

### Save

- Flag: `--save`
- Short Flag: `-s`
- Valid inputs: a path in the current filesystem.

Specify the filename where you want the result to be saved

### Log

- Flag: `--log`
- Short Flag: `-l`
- Valid inputs: `none`, `error`, `debug`
- Default: `info`

Specify the log level. Control how much output you want to see during command execution.

### Configuration

- Flag: `--config-path`
- Short Flag: `-f`
- Valid inputs: a path in the current filesystem.
- Default: `flow.json`

Specify the path to the `flow.json` configuration file.
You can use the `-f` flag multiple times to merge
several configuration files.
