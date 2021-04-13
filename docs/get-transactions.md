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
> flow transactions get ff35821007322405608c0d3da79312617f8d16e118afe63e764b5e68edc96dd5 --host access.mainnet.nodes.onflow.org:9000

ID	 ff35821007322405608c0d3da79312617f8d16e118afe63e764b5e68edc96dd5
Status	 SEALED
Payer	 12e354a23e4f791d
Events	 
	 Index	 0
	 Type	 flow.AccountCreated
	 Tx ID	 ff35821007322405608c0d3da79312617f8d16e118afe63e764b5e68edc96dd5
	 Values
		 address (Address)	18c4931b5f3c7151

	 Index	 1
	 Type	 flow.AccountKeyAdded
	 Tx ID	 ff35821007322405608c0d3da79312617f8d16e118afe63e764b5e68edc96dd5
	 Values
		 address (Address)	18c4931b5f3c7151
		 publicKey (Unknown)	f847b8404c296679364d2...7b168678cc762bc08f342d8d92e0a36e6ecfdcf15850721821823e8
```

## Arguments

### Transaction ID

- Name: `<tx_id>`
- Valid Input: transaction ID.

The first argument is the ID (hash) of the transaction.

## Flags
    
### Display Transaction Code

- Flag: `--code`
- Default: `false`

Indicate whether to print the transaction Cadence code.

### Wait for Seal

- Flag: `--sealed`
- Default: `false`

Indicate whether to wait for the transaction to be sealed
before displaying the result.

### Host

- Flag: `--host`
- Valid inputs: an IP address or hostname.
- Default: `127.0.0.1:3569` (Flow Emulator)

Specify the hostname of the Access API that will be
used to execute the command. This flag overrides
any host defined by the `--network` flag.

### Network

- Flag: `--network`
- Short Flag: `-n`
- Valid inputs: the name of a network defined in the configuration (`flow.json`)

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
