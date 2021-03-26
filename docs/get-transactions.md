---
title: Get a Transaction with the Flow CLI
sidebar_title: Get a Transaction
description: How to get a Flow transaction from the command line
---

The Flow CLI provides a command to get a transaction
that was previously submitted to an Access API.

`flow transactions get <tx_id>`

## Example Usage

```shell
> flow transactions get ff35821007322405608c0d3da79312617f8d16e118afe63e764b5e68edc96dd5 --host access.mainnet.nodes.onflow.org:9000

Hash	 ff35821007322405608c0d3da79312617f8d16e118afe63e764b5e68edc96dd5
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
- Valid Input: transaction hash

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
used to execute the commands.

### Network

- Flag: `--network`
- Short Flag: `-n`
- Valid inputs: the name of a network defined in the configuration (`flow.json`)

Specify which network you want the command to use for execution.

### Filter

- Flag: `--filter`
- Short Flag: `-x`
- Valid inputs: case-sensitive name of the result property.

Specify any property name from the result you want to return as the only value.

### Output

- Flag: `--output`
- Short Flag: `-o`
- Valid inputs: `json`, `inline`

Specify in which format you want to display the result.

### Save

- Flag: `--save`
- Short Flag: `-s`
- Valid inputs: valid filename

Specify the filename where you want the result to be saved.

### Log

- Flag: `--log`
- Short Flag: `-l`
- Valid inputs: `none`, `error`, `debug`
- Default: `info`

Specify the log level. Control how much output you want to see while command execution.

### Configuration

- Flag: `--conf`
- Short Flag: `-f`
- Valid inputs: valid filename

Specify a filename for the configuration files, you can provide multiple configuration
files by using `-f` flag multiple times.





