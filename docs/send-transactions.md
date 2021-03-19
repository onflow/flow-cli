---
title: Send a Transaction with the Flow CLI
sidebar_title: Send a Transaction
description: How to send a Flow transaction from the command line
---

The Flow CLI provides a command to sign and send transactions to
any Flow Access API.

`flow transactions send`

## Example Usage

```shell
# Submit a transaction to Flow Testnet
> flow transactions send \
    --code MyTransaction.cdc \
    --signer my-testnet-account \
    --host access.testnet.nodes.onflow.org:9000
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
    
### Transaction Code

- Flag: `--code,-c`

Specify a path to a Cadence file containing the transaction script.

### Signer

- Flag: `--signer,s`
- Valid inputs: the name of an account defined in `flow.json`

Specify the name of the account that will be used to sign the transaction.

### Host
- Flag: `--host`
- Valid inputs: an IP address or hostname.
- Default: `127.0.0.1:3569` (Flow Emulator)

Specify the hostname of the Access API that will be
used to execute the commands.

### Results

- Flag: `--results`
- Valid inputs: `true`, `false`
- Default: `false`

Indicate whether to wait for the transaction to be sealed
and display the result.

If false, the command returns immediately after sending the transaction
to the Access API. You can later use the `transactions status` command 
to fetch the result.
