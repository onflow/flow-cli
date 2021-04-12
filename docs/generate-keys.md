---
title: Generate Key Pair with the Flow CLI
sidebar_title: Generate Keys
description: How to generate key pair from the command line
---

The Flow CLI provides a command to generate ECDSA key pairs
that can be [attached to new or existing Flow accounts](https://docs.onflow.org/concepts/accounts-and-keys).

```shell
flow keys generate
```

⚠️ Store private key safely and don't share with anyone!

## Example Usage

```shell
flow keys generate
```

### Example response

```shell
> flow keys generate

🔴️ Store Private Key safely and don't share with anyone! 
Private Key 	 c778170793026a9a7a3815dabed68ded445bde7f40a8c66889908197412be89f 
Public Key 	 584245c57e5316d6606c53b1ce46dae29f5c9bd26e9e8...aaa5091b2eebcb2ac71c75cf70842878878a2d650f7 
```

## Flags

### Seed

- Flag: `--seed`
- Valid inputs: any string with length >= 32

Specify a UTF-8 seed string that will be used to generate the key pair.
Key generation is deterministic, so the same seed will always
result in the same key.

If no seed is specified, the key pair will be generated using
a random 32 byte seed.

⚠️ Using seed with production keys can be dangerous if seed was not generated 
by using safe random generators.

### Signature Algorithm

- Flag: `--sig-algo`
- Valid inputs: `"ECDSA_P256", "ECDSA_secp256k1"`

Specify the ECDSA signature algorithm for the key pair.

Flow supports the secp256k1 and P-256 curves.

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
