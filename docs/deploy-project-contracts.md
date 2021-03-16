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
# Deploy project contracts to all Testnet targets
> flow project deploy --network=testnet

NonFungibleToken -> 0xf8d6e0586b0a20c7
KittyItems -> 0xf8d6e0586b0a20c7

✅ All contracts deployed successfully
```

In the example above, your `flow.json` file might look something like this:

```json
{
  "contracts": {
    "NonFungibleToken": "./cadence/contracts/NonFungibleToken.cdc",
    "KittyItems": "./cadence/contracts/KittyItems.cdc"
  },
  "deployments": {
    "testnet": {
      "my-testnet-account": ["KittyItems", "NonFungibleToken"]
    }
  }
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

## Security

⚠️ Warning: Please be careful when using private keys in configuration files. We suggest you 
to separate private keys in another configuration file, put that file in `.gitignore` and then 
reference that account in configuration with `fromeFile` property. 

### Private Account Configuration File
`flow.json` Main configuration file example:
```json
{
  "contracts": {
    "NonFungibleToken": "./cadence/contracts/NonFungibleToken.cdc",
    "KittyItems": "./cadence/contracts/KittyItems.cdc"
  },
  "deployments": {
    "testnet": {
      "my-testnet-account": ["KittyItems", "NonFungibleToken"]
    }
  },
  "accounts": {
    "my-testnet-account": { "fromFile": "./flow.testnet.json" }
  }
}
```

`flow.testnet.json` Private configuration file. **Put this file in `.gitignore`**
```json
{
  "accounts": {
    "my-testnet-account": {
      "address": "3ae53cb6e3f42a79",
      "keys": "334232967f52bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad1111"
    }
  }
}
```

### Private Account Configuration Environment Variable

Use environment variable for any values that should be kept private (private keys, addresses...). 
See example bellow:

`flow.json` Main configuration file. Set environment variable when running flow cli like so:
```shell
PRIVATE_KEY=key flow project deploy
```
```json
{
  ...
  "accounts": {
    "my-testnet-account": {
      "address": "3ae53cb6e3f42a79",
      "keys": "$PRIVATE_KEY"
    }
  }
  ...
}
```

### Composing Multiple Configuration Files
You can use composition of configuration files like so:
```shell
flow project deploy -f main.json -f private.json
```

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

## Options

### Network

- Flag: `--network`
- Valid inputs: the name of a network configured in `flow.json`
- Default: `emulator`

Specify which network you want to deploy the project contracts to.

### Allow Updates

- Flag: `--update`
- Valid inputs: `true`, `false`
- Default: `false`

Indicate whether to overwrite and upgrade existing contracts.

⚠️ _Warning: contract upgrades are a dangerous experimental feature._
