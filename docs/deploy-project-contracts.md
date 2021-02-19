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
FungibleToken -> 0xf8d6e0586b0a20c7
Kibble -> 0xf8d6e0586b0a20c7

✅ All contracts deployed successfully
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


## Configuration

Project uses configuration file you can generate by using `flow project init` command or by hand. After you generate configuration file you must add contracts and accounts you wish to use for deployment. 
Bellow is an example configuration file and we will document each configuration option.

```
{
  "contracts": {
    "NonFungibleToken": "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc",
    "Kibble": "./cadence/kibble/contracts/Kibble.cdc",
    "KittyItems": "./cadence/kittyItems/contracts/KittyItems.cdc",
    "KittyItemsMarket": "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc",
    "FungibleToken": {
      "testnet": "0x123",
      "emulator": "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc"
    }
  },

  "deploy": {
    "testnet": {
      "admin-account": ["NonFungibleToken"],
      "user-account": ["Kibble", "KittyItems", "KittyItemsMarket"]
    }, 
    "emulator": {
      "emulator-account": ["NonFungibleToken", "FungibleToken", "Kibble", "KittyItems", "KittyItemsMarket"]
    }
  },

  "accounts": {
    "admin-account": {
      "address": "0x2244224422",
      "keys": "22232967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
    },
    "user-account": {
      "address": "0x123123123",
      "keys": "22232967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
    },
    "emulator-account": {
      "address": "f8d6e0586b0a20c7",
      "keys": "2eae2f31cb5b756151fa11d82949c634b8f28796a711d7eb1e52cc301edfb3e0",
      "chain": "flow-emulator"
    }
  },

  "networks": {
    "emulator": {
      "host": "127.0.0.1:3569",
      "serviceAccount": "emulator-service"
    },
    "testnet": "access.testnet.nodes.onflow.org:9000"
  }
}
```

### Contracts

Contract section defines contract names and their code location or address. We can also use advance format where we can define code location for each network.

#### Simple format 

```
...
 "contracts": {
    "NonFungibleToken": "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc"
 }
 ...
```

#### Advanced format 

Advanced format allows us to define `aliases`. Aliases define different source of the contract for that specific network. In the simple scnario bellow the contract FungibleToken would be imported from the address 0x123 when deploying to testnet network. We can specify alias for each network we have defined. When deploying to testnet it is always a good idea to specify aliases for all the common contracts that are already deployed to testnet and can be find here: https://docs.onflow.org/core-contracts

```
...
"FungibleToken": {
  "source": "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc",
  "aliases": {
    "testnet": "0x123"
  }
}
...
```

### Accounts

Account section is used to define account properties such as keys and addresses. Each account must include a name which then gets referenced throught the config file.

#### Simple format

All we need to define in simple format is `address` for the account and its private key under `keys`.

```
...
"accounts": {
  "admin-account": {
    "address": "0x2244224422",
    "keys": "22232967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
  }
}
...
```

#### Advanced format

Advanced format allows us to define any paramter possible. You can check all the values for paramters (here). Please note we can use `service` for address in case the account is used on `emulator` network 
as this is a special value that is defined on the run time to the default service address on the emulator network.

```
...
"accounts": {
  "admin-account": {
    "address": "service",
    "chain": "flow-emulator",
    "keys": [
      {
        "type": "hex",
        "index": 0,
        "signatureAlgorithm": "ECDSA_P256",
        "hashAlgorithm": "SHA3_256",
        "context": {
          "privateKey": "1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
        }
      },
      {
        "type": "hex",
        "index": 1,
        "signatureAlgorithm": "ECDSA_P256",
        "hashAlgorithm": "SHA3_256",
        "context": {
          "privateKey": "2372967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
        }
      }
    ]
  }
}
...
```

### Deploys

Deploys section defines where the `deploy-contracts` command will deploy specified contracts. This configuration property is like glue that ties together accounts, 
contracts and networks all referenced by their name. 

#### Simple format

```
"deploy": {
  "emulator": {
    "emulator-account": ["NonFungibleToken", "FungibleToken", "Kibble", "KittyItems", "KittyItemsMarket"]
  },
  "testnet": {
    "admin-account": ["NonFungibleToken"],
    "user-account": ["Kibble", "KittyItems", "KittyItemsMarket"]
  }, 
}
...
```

### Networks

Use this section to define networks and their attributes. 

#### Simple format

```
...
"networks": {
  "emulator": {
    "host": "127.0.0.1:3569",
    "serviceAccount": "emulator-service"
  },
  "testnet": "access.testnet.nodes.onflow.org:9000"
}
...
```