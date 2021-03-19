---
title: Generate Keys with the Flow CLI
sidebar_title: Generate a Key
description: How to generate a Flow account key-pair from the command line
---

The Flow CLI provides a command to generate ECDSA key pairs
that can be [attached to new or existing Flow accounts](https://docs.onflow.org/concepts/accounts-and-keys).

`flow keys generate`

## Example Usage

```shell
> flow keys generate

Generating key pair with signature algorithm:                 ECDSA_P256
...
🔐 Private key (do not share with anyone):                    xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
🕊 Encoded public key (share freely):                         a69c6986e69fa1eadcd3bcb4aa51ee8aed74fc9430004af6b96f9e7d0e4891e84cfb99171846ba6d0354d195571397f5904cd319c3e01e96375d5777f1a47010
```

## Options

### Signature Algorithm

- Flag: `--algo,-a`
- Valid inputs: `"ECDSA_P256", "ECDSA_secp256k1"`
- Default: `"ECDSA_P256"`

Specify the ECDSA signature algorithm for the key pair.

Flow supports the secp256k1 and P-256 curves.

### Seed

- Flag: `--seed,s`
- Valid inputs: any string with length >= 32

Specify a UTF-8 seed string that will be used to generate the key pair.
Key generation is deterministic, so the same seed will always
result in the same key.

If no seed is specified, the key pair will be generated using
a random 32 byte seed.

### Host
- Flag: `--host`
- Valid inputs: an IP address or hostname.
- Default: `127.0.0.1:3569` (Flow Emulator)

Specify the hostname of the Access API that will be
used to execute the commands.