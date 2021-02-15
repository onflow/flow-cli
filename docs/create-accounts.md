---
title: Create an Account with the Flow CLI
sidebar_title: Create an Account
description: How to create a Flow account from the command line
---

The Flow CLI provides a command to submit an account creation 
transaction to any Flow Access API.

`flow accounts create`

⚠️ _This command requires an existing Testnet or Mainnet account._

## Example Usage

```shell
# Create an account on Flow Testnet
> flow accounts create \
    --key a69c6986e69fa1eadcd3bcb4aa51ee8aed74fc9430004af6b96f9e7d0e4891e84cfb99171846ba6d0354d195571397f5904cd319c3e01e96375d5777f1a47010 \
    --sig-algo ECDSA_secp256k1 \
    --hash-algo SHA3_256 \
    --host access.testnet.nodes.onflow.org:9000 \
    --signer my-testnet-account \
    --results
```

In the above example, the `flow.json` file would look something like this:

```json
{
  "accounts": {
    "my-testnet-account": {
      "address": "f8d6e0586b0a20c7",
      "privateKey": "xxxxxxxx",
      "sigAlgorithm": "ECDSA_P256",
      "hashAlgorithm": "SHA3_256"
    }
  }
}
```

## Options
    
### Public Key

- Flag: `--key,-k`
- Valid inputs: a hex-encoded public key in raw form.

Specify the public key that will be added to the new account
upon creation.

### Public Key Signature Algorithm
    
- Flag: `--sig-algo`
- Valid inputs: `"ECDSA_P256", "ECDSA_secp256k1"`
- Default: `"ECDSA_P256"`

Specify the ECDSA signature algorithm for the provided public key.
This option can only be used together with the `--key` flag.

Flow supports the secp256k1 and P-256 curves.

### Public Key Hash Algorithm

- Flag: `--hash-algo`
- Valid inputs: `"SHA2_256", "SHA3_256"`
- Default: `"SHA3_256"`

Specify the hashing algorithm that will be paired with the public key
upon account creation.

### Wait for Sealed

- Flag: `--sealed`
- Valid inputs: `true`, `false`
- Default: `false`

Indicate whether to wait for the transaction to be sealed.
If true, the CLI will block until the transaction has been sealed, or
a timeout is reached.

### Signer

- Flag: `--signer,s`
- Valid inputs: the name of an account defined in `flow.json`

Specify the name of the account that will be used to sign the transaction
and pay the account creation fee.

### Host

- Flag: `--host`
- Valid inputs: an IP address or hostname.
- Default: `localhost:3569` (Flow Emulator)

Specify the hostname of the Access API that will be
used to query the transaction.

### Results

- Flag: `--results`
- Valid inputs: `true`, `false`
- Default: `false`

Indicate whether to wait for the transaction to be sealed
and display the result, including the new account address.

If false, the command returns immediately after sending the transaction
to the Access API. You can later use the `transactions status` command
to fetch the result.
