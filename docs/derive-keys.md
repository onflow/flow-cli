---
title: Derive Public Key with the Flow CLI
sidebar_title: Derive Public Key
description: How to derive Flow public key from a private key from the command line
---

The Flow CLI provides a command to derive Public Key from a Private Key.

```shell
flow keys derive <private key>
```

## Example Usage

### Derive Public Key from a Private Key
```shell
> flow keys derive c778170793026a9a7a3815dabed68ded445bde7f40a8c66889908197412be89f 
```

### Example response

```shell
> flow keys generate

🔴️ Store Private Key safely and don't share with anyone! 
Private Key     c778170793026a9a7a3815dabed68ded445bde7f40a8c66889908197412be89f 
Public Key 	    584245c57e5316d6606c53b1ce46dae29f5c9bd26e9e8...aaa5091b2eebcb2ac71c75cf70842878878a2d650f7 
```

## Arguments

### Private Key
- Name: `private key`
- Valid inputs: valid private key content

## Flags

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
