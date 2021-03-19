---
title: 
sidebar_title: 
description: 
---

{short description}

`{command}`

{optional warning}

## Example Usage

```shell
{usage example}
```

### Example response

## Arguments

### {Argument 1}
- Name: `{argument}`
- Valid Input: `{input}`

{argument general description}

## Arguments

### Address
- Name: `address`
- Valid Input: Flow account address

Flow [account address](https://docs.onflow.org/concepts/accounts-and-keys/) (prefixed with `0x` or not).


## Flags

### {Option 1}

- Flag: `{flag value}`
- Valid inputs: {input description}

{flag general description}

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