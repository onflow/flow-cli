---
title: Remove a Contract with the Flow CLI
sidebar_title: Remove a Contract
---

Remove an existing contract deployed to a Flow account using the Flow CLI.

```shell
flow accounts remove-contract <name>
```

## Example Usage

```shell
> flow accounts remove-contract FungibleToken

Contract 'FungibleToken' removed from account '0xf8d6e0586b0a20c7'

Address	 0xf8d6e0586b0a20c7
Balance	 99999999999.70000000
Keys	 1

Key 0	Public Key		 640a5a359bf3536d15192f18d872d57c98a96cb871b92b70cecb0739c2d5c37b4be12548d3526933c2cda9b0b9c69412f45ffb6b85b6840d8569d969fe84e5b7
	Weight			 1000
	Signature Algorithm	 ECDSA_P256
	Hash Algorithm		 SHA3_256
	Revoked 		 false
	Sequence Number 	 6
	Index 			 0

Contracts Deployed: 0
```
**Testnet Example**
```
> flow accounts remove-contract FungibleToken --signer alice --network testnet

Contract 'FungibleToken' removed from account '0xf8d6e0586b0a20c7'

Address	 0xf8d6e0586b0a20c7
Balance	 99999999999.70000000
Keys	 1

Key 0	Public Key		 640a5a359bf3536d15192f18d872d57c98a96cb871b92b70cecb0739c2d5c37b4be12548d3526933c2cda9b0b9c69412f45ffb6b85b6840d8569d969fe84e5b7
	Weight			 1000
	Signature Algorithm	 ECDSA_P256
	Hash Algorithm		 SHA3_256
	Revoked 		 false
	Sequence Number 	 6
	Index 			 0

Contracts Deployed: 0

```
*Make sure alice account is defined in flow.json*

## Arguments

### Name

- Name: `name`
- Valid inputs: any string value.

Name of the contract as it is defined in the contract source code.

## Flags

### Signer

- Flag: `--signer`
- Valid inputs: the name of an account defined in the configuration (`flow.json`).

Specify the name of the account that will be used to sign the transaction.

### Include Fields

- Flag: `--include`
- Valid inputs: `contracts`

Specify fields to include in the result output. Applies only to the text output.


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
- Valid inputs: a case-sensitive name of the result property

Specify any property name from the result you want to return as the only value.

### Output

- Flag: `--output`
- Short Flag: `-o`
- Valid inputs: `json`, `inline`

Specify the format of the command results.

### Save

- Flag: `--save`
- Short Flag: `-s`
- Valid inputs: a path in the current filesystem

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
- Valid inputs: a path in the current filesystem
- Default: `flow.json`

Specify the path to the `flow.json` configuration file. 
You can use the `-f` flag multiple times to merge
several configuration files.
