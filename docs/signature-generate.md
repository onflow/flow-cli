---
title: Generate a Signature with the Flow CLI
sidebar_title: Generate a Signature
description: How to generate a new signature from the command line
---

Generate a signature using the private key of the signer account.

```shell
flow signatures generate <message>  
```

⚠️ _Make sure the account you want to use for signing is saved in the `flow.json` configuration. 
The address of the account is not important, just the private key._

## Example Usage

```shell
> flow signatures generate 'The quick brown fox jumps over the lazy dog' --signer alice

Signature 		 b33eabfb05d374b...f09929da96f5beec167fd1f123ec
Message 		 The quick brown fox jumps over the lazy dog
Public Key 		 0xc92a7c...042c4025d241fd430242368ce662d39636987
Hash Algorithm 		 SHA3_256
Signature Algorithm 	 ECDSA_P256
```

## Arguments

### Message
- Name: `message`

Message used for signing.

## Flags

### Signer

- Flag: `--signer`
- Valid inputs: the name of an account defined in the configuration (`flow.json`)

Specify the name of the account that will be used to sign the transaction.

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

- Flag: `--config-path`
- Short Flag: `-f`
- Valid inputs: valid filename

Specify a filename for the configuration files, you can provide multiple configuration
files by using `-f` flag multiple times.





