---
title: Deploy a Flow Project
sidebar_title: Deploy a Project
description: How to deploy Flow project contracts with the CLI
---

⚠️ _Warning: CLI projects are an experimental feature._

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

Deploying Foo to my-testnet-account (0xf8d6e0586b0a20c7)...
Deploying Bar to my-testnet-account (0xf8d6e0586b0a20c7)...

All contracts deployed successfully.
```

In the example above, your `flow.json` file might look something like this:

```json
{
  "contracts": {
    "Foo": "./cadence/contracts/FooContract.cdc",
    "Bar": "./cadence/contracts/BarContract.cdc"
  },
  "deploy": {
    "testnet": {
      "my-testnet-account": ["Foo", "Bar"]
    }
  }
}
```

Here's a sketch of the contract source files:

```cadence:title=Foo.cdc
pub contract Foo { 
  // ...
}
```

```cadence:title=Bar.cdc
import Foo from Foo.cdc

pub contract Bar { 
  // ...
}
```

## Dependency Resolution

The `deploy` command attempts to resolve the import statements in all contracts being deployed.

After the dependencies are found, the CLI will deploy the contracts in a deterministic ordering
such that no contract is deployed until all of its dependencies are deployed.
The command will return an error if no such ordering exists due to one or more cyclic dependencies.

In the example above, `Foo` will always be deployed before `Bar`.

## Address Replacement

After resolving all dependencies, the `deploy` command rewrites each contract so 
that its dependencies are imported from their _target addresses_ rather than their source files.

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
