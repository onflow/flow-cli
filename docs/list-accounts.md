---
title: List all Account names
sidebar_title: List Accounts
description: How to list all Flow accounts from the command line
---

The Flow CLI provides a command to list all account names in flow configuration file.

```shell
flow accounts list
```

## Example Usage

```shell
flow accounts list
```

### Example response
```shell
account1,account2,flow-account3

```

## Arguments
none

## Flags

### Sort Names

- Flag: `--sort`

Sorts account names the result output. 


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

