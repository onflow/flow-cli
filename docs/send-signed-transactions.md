---
title: Send Signed Transaction with the Flow CLI
sidebar_title: Send Signed Transaction
description: How to send a signed Flow transaction from the command line
---

The Flow CLI provides a command to send signed transactions to
any Flow Access API.

Use this functionality in the following order:
1. Use the `build` command to build the transaction.
2. Use the `sign` command to sign with each account specified in the build process.
3. Use this command (`send-signed`) to submit the signed transaction to the Flow network.

```shell
flow transactions send-signed <signed transaction filename>
```

## Example Usage

```shell
> flow transactions send-signed ./signed.rlp
    
Status		✅ SEALED
ID		528332aceb288cdfe4d11d6522aa27bed94fb3266b812cb350eb3526ed489d99
Payer		f8d6e0586b0a20c7
Authorizers	[f8d6e0586b0a20c7]

Proposal Key:	
    Address	f8d6e0586b0a20c7
    Index	0
    Sequence	0

No Payload Signatures

Envelope Signature 0: f8d6e0586b0a20c7
Signatures (minimized, use --include signatures)

Events:	 None

Code (hidden, use --include code)

Payload (hidden, use --include payload)

```


## Arguments

### Signed Code Filename
- Name: `signed transaction filename`
- Valid inputs: Any filename and path valid on the system.

The first argument is a path to a Cadence file containing the
transaction to be executed.

## Flags

### Include Fields

- Flag: `--include`
- Valid inputs: `code`, `payload`

Specify fields to include in the result output. Applies only to the text output.

### Exclude Fields

- Flag: `--exclude`
- Valid inputs: `events`

Specify fields to exclude from the result output. Applies only to the text output.

### Filter

- Flag: `--filter`
- Short Flag: `-x`
- Valid inputs: a case-sensitive name of the result property.

Specify any property name from the result you want to return as the only value.

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
