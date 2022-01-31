---
title: Sign a Transaction with the Flow CLI
sidebar_title: Sign a Transaction
description: How to sign a Flow transaction from the command line
---

The Flow CLI provides a command to sign transactions with options to specify
authorizer accounts, payer accounts and proposer accounts.

Use this functionality in the following order:
1. Use the `build` command to build the transaction.
2. Use this command (`sign`) to sign with each account specified in the build process.
3. Use the `send-signed` command to submit the signed transaction to the Flow network.

```shell
flow transactions sign <built transaction filename>
```

## Example Usage

```shell
> flow transactions sign ./built.rlp --signer alice \
  --filter payload --save signed.rlp

Hash		b03b18a8d9d30ff7c9f0fdaa80fcaab242c2f36eedb687dd9b368326311fe376
Payer		f8d6e0586b0a20c7
Authorizers	[f8d6e0586b0a20c7]

Proposal Key:	
    Address	f8d6e0586b0a20c7
    Index	0
    Sequence	6

No Envelope Signatures

Payload Signature 0:
    Address	f8d6e0586b0a20c7
    Signature	b5b1dfed2a899037...164e1b224a7ac924018e7033b68b0df86769dd54
    Key Index	0


Arguments (1):
    - Argument 0: {"type":"String","value":"Meow"}


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
f90184f...a199bfd9b837a11a0885f9104b54014750f5e3e5bfe4a5795968b0df86769dd54c0
```

## Arguments

### Built Transaction Filename
- Name: `built transaction filename`
- Valid inputs: Any filename and path valid on the system.

Specify the filename containing valid transaction payload that will be used for signing.
To be used with the `flow transaction build` command.

## Flags

### Include Fields

- Flag: `--include`
- Valid inputs: `code`, `payload`, `signatures`

Specify fields to include in the result output. Applies only to the text output.

### Signer

- Flag: `--signer`
- Valid inputs: the name of an account defined in the configuration (`flow.json`)

Specify the name of the account that will be used to sign the transaction.

### Host
- Flag: `--host`
- Valid inputs: an IP address or hostname.
- Default: `127.0.0.1:3569` (Flow Emulator)

Specify the hostname of the Access API that will be
used to execute the commands.

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
