---
title: Send a Transaction with the Flow CLI
sidebar_title: Send a Transaction
description: How to send a Flow transaction from the command line
---

The Flow CLI provides a command to sign and send transactions to
any Flow Access API.

```shell
flow transactions send <code filename>
```

## Example Usage

```shell
> flow transactions send ./transaction.cdc
    --args-json --args-json '[{"type": "String","value": "Hello World"}]'
    --signer my-testnet-account
    
Status		✅ SEALED
ID		b6430b35ba23849a8acb4fa1a4a1d5cce3ed4589111ecbb3984de1b6bd1ba39e
Payer		a2c4941b5f3c7151
Authorizers	[a2c4941b5f3c7151]

Proposal Key:	
    Address	a2c4941b5f3c7151
    Index	0
    Sequence	9

No Payload Signatures

Envelope Signature 0:
    Address	a2c4941b5f3c7151
    Signature	5391a6fed0fe...2742048166f9d5c925a8dcb78a6d8c710921d67
    Key Index	0


Events:	 None


Arguments (1):
    - Argument 0: {"type":"String","value":"Hello World"}


Code

transaction(greeting: String) {
  let guest: Address

  prepare(authorizer: AuthAccount) {
    self.guest = authorizer.address
  }

  execute {
    log(greeting.concat(",").concat(self.guest.toString()))
  }
}


Payload:
f90184f90138...8a9462751237da2742048166f9d5c925a8dcb78a6d8c710921d67

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

### Code Filename
- Name: `code filename`
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

### Arguments

- Flag: `--arg`
- Valid inputs: argument in `Type:Value` format.

Arguments passed to the Cadence transaction in `Type:Value` format.
The `Type` must be the same as type in the transaction source code for that argument.

### Arguments JSON

- Flag: `--argsJSON`
- Valid inputs: arguments in JSON-Cadence form.

Arguments passed to the Cadence transaction in Cadence JSON format.
Cadence JSON format contains `type` and `value` keys and is 
[documented here](https://docs.onflow.org/cadence/json-cadence-spec/).

### Gas Limit

- Flag: `--gas-limit`
- Valid inputs: an integer greater than zero.
- Default: `1000`

Specify the gas limit for this transaction.

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
