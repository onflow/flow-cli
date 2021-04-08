---
title: Send a Transaction with the Flow CLI
sidebar_title: Send a Transaction
description: How to send a Flow transaction from the command line
---

The Flow CLI provides a command to sign and send transactions to
any Flow Access API.

```shell
flow transactions send
```

## Example Usage

```shell
# Submit a transaction to Flow Testnet
> flow transactions send <filename>
    --signer my-testnet-account \
    --host access.testnet.nodes.onflow.org:9000
    
ID	 f23582ba17322405608c0d3da79312617f8d16e118afe63e764b5e68edc5f211
Status	 SEALED
Payer	 a2c4941b5f3c7151
Events
```

In the above example, the `flow.json` file would look something like this:

```json
{
  "accounts": {
    "my-testnet-account": {
      "address": "a2c4941b5f3c7151",
      "keys": "12c5dfde...bb2e542f1af710bd1d40b2"
    }
  }
}
```

## Arguments

### Filename
- Name: `filename`
- Valid inputs: Any filename and path valid on the system.

The first argument is a path to a Cadence file containing the
transaction to be executed.

## Flags

### Code

- Flag: `--code`

⚠️  No longer supported: use filename argument.

### Results

- Flag: `--results`

⚠️  No longer supported: all transactions will provide result.

### Signer

- Flag: `--signer`
- Valid inputs: the name of an account defined in the configuration (`flow.json`)

Specify the name of the account that will be used to sign the transaction.

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
