---
title: Update a Contract with the Fow CLI
sidebar_title: Update a Contract
description: How to update an existing contract
---

Update an existing contract deployed to an account on the Flow network using Flow CLI.

`flow accounts update-contract <name> <filename>`

## Example Usage

```shell
flow accounts update-contract FungibleToken ./FungibleToken.cdc
```

### Example response

```shell
Contract 'FungibleToken' updated on account '0xf8d6e0586b0a20c7'

Address	 0xf8d6e0586b0a20c7
Balance	 9999999999970000000
Keys	 1

Key 0	Public Key		 640a5a359bf3536d15192f18d872d57c98a96cb871b92b70cecb0739c2d5c37b4be12548d3526933c2cda9b0b9c69412f45ffb6b85b6840d8569d969fe84e5b7
	Weight			 1000
	Signature Algorithm	 ECDSA_P256
	Hash Algorithm		 SHA3_256

Contracts Deployed: 1
Contract: 'FungibleToken'
```

## Arguments

### Name
- Name: `name`
- Valid inputs: Any string value

Name of the contract as it is defined in the contract source code.

### Filename
- Name: `filename`
- Valid inputs: Any filename and path valid on the system.

Filename of the file containing contract source code.


## Flags

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