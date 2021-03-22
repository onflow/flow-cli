---
title: Get Block with the Flow CLI
sidebar_title: Get Collection
description: How to get a collection from the command line
---

The Flow CLI provides a command to fetch any collection from the Flow network.

```bash
flow collections get 3e694588e789a72489667a36dd73104dea4579bcd400959d47aedccd7f930eeb \
--host access.mainnet.nodes.onflow.org:9000
```

{optional warning}

## Example Usage

```shell
{usage example}
```

### Example response

```shell
Collection ID 3e694588e789a72489667a36dd73104dea4579bcd400959d47aedccd7f930eeb:
acc2ae1ff6deb2f4d7663d24af6ab1baf797ec264fd76a745a30792f6882093b
ae8bfbc85ce994899a3f942072bfd3455823b1f7652106ac102d161c17fcb55c
70c4d39d34e654173c5c2746e7bb3a6cdf1f5e6963538d62bad2156fc02ea1b2
2466237b5eafb469c01e2e5f929a05866de459df3bd768cde748e068c81c57bf

```

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

### Network

- Flag: `--network`
- Short Flag: `-n`
- Valid inputs: the name of a network defined in the configuration (`flow.json`)

Specify which network you want the command to use for execution.

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

- Flag: `--conf`
- Short Flag: `-f`
- Valid inputs: valid filename

Specify a filename for the configuration files, you can provide multiple configuration
files by using `-f` flag multiple times.





