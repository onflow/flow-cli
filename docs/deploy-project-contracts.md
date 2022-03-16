---
title: Deploy a Flow Project
sidebar_title: Deploy a Project
description: How to deploy Flow project contracts with the CLI
---

```shell
flow project deploy
```

This command automatically deploys your project's contracts based on the 
configuration defined in your `flow.json` file.

Before using this command, read about how to 
[configure project contracts and deployment targets](project-contracts.md).

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

## Initialization Arguments
Deploying contracts that take initialization arguments 
can be achieved with adding those arguments to the configuration. 

Each deployment can be specified as an object containing 
`name` and `args` key specifying arguments to be 
used during the deployment. Example:

```
...
  "deployments": {
    "testnet": {
      "my-testnet-account": [
        "NonFungibleToken", {
            "name": "Foo", 
            "args": [
                { "type": "String", "value": "Hello World" },
                { "type": "UInt32", "value": "10" }
            ]
        }]
    }
  }
...
```


⚠️ Warning: before proceeding, 
we recommend reading the [Flow CLI security guidelines](security.md) 
to learn about the best practices for private key storage.

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

In the example above, the `KittyItems` contract would be rewritten like this:

```cadence:title=KittyItems.cdc
import NonFungibleToken from 0xf8d6e0586b0a20c7

pub contract KittyItems { 
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
used to execute the command. This flag overrides
any host defined by the `--network` flag.

### Network Key

- Flag: `--network-key`
- Valid inputs: A valid network public key of the host in hex string format

Specify the network public key of the Access API that will be
used to create a secure GRPC client when executing the command.

### Network

- Flag: `--network`
- Short Flag: `-n`
- Valid inputs: the name of a network defined in the configuration (`flow.json`)
- Default: `emulator`

Specify which network you want the command to use for execution.

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
