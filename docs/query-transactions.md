---
title: Query a Transaction Result with the Flow CLI
sidebar_title: Query a Transaction Result
description: How to get the result of a Flow transaction from the command line
---

The Flow CLI provides a command to query the result of a transaction
that was previously submitted to an Access API.

`flow transactions status`

## Example Usage

```shell
# Query a transaction on Flow Testnet
> flow transactions status \
    --host access.testnet.nodes.onflow.org:9000 \
    74c02a67297458dbed26273d3b407eedcb42957bdc8d2deb3c6939145bf2b240
```

## Arguments

### Transaction ID

The first argument is the ID (hash) of the transaction.

## Options
    
### Display Transaction Code

- Flag: `--code`
- Valid inputs: `true`, `false`
- Default: `false`

Indicate whether to print the transaction Cadence code.

### Wait for Seal

- Flag: `--sealed`
- Valid inputs: `true`, `false`
- Default: `false`

Indicate whether to wait for the transaction to be sealed.
If true, the CLI will block until the transaction has been sealed or
a timeout is reached.

### Host

- Flag: `--host`
- Valid inputs: an IP address or hostname.
- Default: `localhost:3569` (Flow Emulator)

Specify the hostname of the Access API that will be
used to submit the transaction query.
