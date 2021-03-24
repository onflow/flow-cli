---
title: Deploy a Flow Project
sidebar_title: Deploy a Project
description: How to deploy Flow project contracts with the CLI
---

⚠️ _Warning: CLI projects are an experimental feature. Functionality is subject to change._

```shell
flow project deploy
```

This command automatically deploys your project's contracts based on the 
configuration defined in your `flow.json` file.

Before using this command, read about how to 
[configure project contracts and deployment targets](https://docs.onflow.org/flow-cli/project-contracts/).

## Example Usage

```shell
> flow project deploy --network=testnet

Deploying 2 contracts for accounts: my-testnet-account

NonFungibleToken -> 0x8910590293346ec4
KittyItems -> 0x8910590293346ec4

✨  All contracts deployed successfully
```

In the example above, your `flow.json` file might look something like this:

```json
{
  ...
  "contracts": {
    "NonFungibleToken": "./cadence/contracts/NonFungibleToken.cdc",
    "KittyItems": "./cadence/contracts/KittyItems.cdc"
  },
  "deployments": {
    "testnet": {
      "my-testnet-account": ["KittyItems", "NonFungibleToken"]
    }
  },
  ...
}
```

Here's a sketch of the contract source files:

```cadence:title=NonFungibleToken.cdc
pub contract NonFungibleToken { 
  // ...
}
```

```cadence:title=KittyItems.cdc
import NonFungibleToken from "./NonFungibleToken.cdc"

pub contract KittyItems { 
  // ...
}
```

⚠️ Warning: Before proceeding you must read [Security document](https://docs.onflow.org/flow-cli/security/) so you will know how to secure 
your private keys.


This way you can keep your private accounts in the `private.json` file and add that file to `.gitignore`.

## Dependency Resolution

The `deploy` command attempts to resolve the import statements in all contracts being deployed.

After the dependencies are found, the CLI will deploy the contracts in a deterministic order
such that no contract is deployed until all of its dependencies are deployed.
The command will return an error if no such ordering exists due to one or more cyclic dependencies.

In the example above, `Foo` will always be deployed before `Bar`.

## Address Replacement

After resolving all dependencies, the `deploy` command rewrites each contract so 
that its dependencies are imported from their _target addresses_ rather than their 
source file location.

The rewritten versions are then deployed to their respective targets,
leaving the original contract files unchanged.

In the example above, the `Bar` contract would be rewritten like this:

```cadence:title=Bar.cdc
import Foo from 0xf8d6e0586b0a20c7

pub contract Bar { 
  // ...
}
```

## Flags

### Allow Updates

- Flag: `--update`
- Valid inputs: `true`, `false`
- Default: `false`

Indicate whether to overwrite and upgrade existing contracts.

⚠️ _Warning: contract upgrades are a dangerous experimental feature._


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

