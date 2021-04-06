---
title: Flow CLI Configuration
sidebar_title: Configuration
description: What is Flow CLI Configuration
---

Flow CLI uses a state called configuration which is stored in a file (usually `flow.json`). 

Flow configuration (`flow.json`) file will contain the following properties:

- A `networks` list pre-populated with the Flow emulator, testnet and mainnet connection configuration.
- An `accounts` list pre-populated with the Flow Emulator service account.
- An `emulators` list pre-populated with Flow Emulator configuration.
- A `deployments` empty object where all [deployment targets](https://docs.onflow.org/flow-cli/project-contracts/) can be defined. 
- A `contracts` empty object where you [define contracts](https://docs.onflow.org/flow-cli/project-contracts/) you wish to deploy.

## Example Project Configuration

```json
{
  "emulators": {
    "default": {
      "port": 3569,
      "serviceAccount": "emulator-account"
    }
  },
  "networks": {
    "emulator": {
      "host": "127.0.0.1:3569",
      "chain": "flow-emulator"
    }
  },
  "accounts": {
    "emulator-account": {
      "address": "f8d6e0586b0a20c7",
      "keys": "ae1b44c0f5e8f6992ef2348898a35e50a8b0b9684000da8b1dade1b3bcd6ebee",
      "chain": "flow-emulator"
    }
  },
  "deployments": {},
  "contracts": {}
}
```

## Configuration

Below is an example of a configuration file for a complete Flow project.
We'll walk through each property one by one.

```json
{
  "contracts": {
    "NonFungibleToken": "./cadence/contracts/NonFungibleToken.cdc",
    "Kibble": "./cadence/contracts/Kibble.cdc",
    "KittyItems": "./cadence/contracts/KittyItems.cdc",
    "KittyItemsMarket": "./cadence/contracts/KittyItemsMarket.cdc",
    "FungibleToken": {
      "source": "./cadence/contracts/FungibleToken.cdc",
      "aliases": {
        "testnet": "9a0766d93b6608b7",
        "emulator": "ee82856bf20e2aa6"
      }
    }
  },

  "deployments": {
    "testnet": {
      "admin-account": ["NonFungibleToken"],
      "user-account": ["Kibble", "KittyItems", "KittyItemsMarket"]
    }, 
    "emulator": {
      "emulator-account": [
        "NonFungibleToken",
        "Kibble",
        "KittyItems",
        "KittyItemsMarket"
      ]
    }
  },

  "accounts": {
    "admin-account": {
      "address": "3ae53cb6e3f42a79",
      "keys": "12332967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad1111"
    },
    "user-account": {
      "address": "e2a8b7f23e8b548f",
      "keys": "22232967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad1111"
    },
    "emulator-account": {
      "address": "f8d6e0586b0a20c7",
      "keys": "2eae2f31cb5b756151fa11d82949c634b8f28796a711d7eb1e52cc301ed11111",
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

Contracts are specified as key-value pairs, where the key is the contract name, 
and the value is the location of the Cadence source code.

The advanced format allows us to specify aliases for each network.

#### Simple Format 

```json
...

"contracts": {
  "NonFungibleToken": "./cadence/contracts/NonFungibleToken.cdc"
}

...
```

#### Advanced Format 

Using advanced format we can define `aliases`. Aliases define an address where the contract is already deployed for that specific network. 
In the example scenario below the contract `FungibleToken` would be imported from the address `9a0766d93b6608b7` when deploying to testnet network 
and address `ee82856bf20e2aa6` when deploying to testnet. 
We can specify aliases for each network we have defined. When deploying to testnet it is always a good idea to specify aliases for all the [common 
contracts](https://docs.onflow.org/core-contracts) that have already been deployed to the testnet. 

⚠️ If we use an alias for the contract we should not specify it in the `deployment` section for that network. 

Our example bellow should not include `FungibleToken` in  `deployment` section for testnet and emulator network.

```json
...

"FungibleToken": {
  "source": "./cadence/contracts/FungibleToken.cdc",
  "aliases": {
    "testnet": "9a0766d93b6608b7",
    "emulator": "ee82856bf20e2aa6"
  }
}

...
```

### Accounts

The accounts section is used to define account properties such as keys and addresses. 
Each account must include a name, which is then referenced throughout the configuration file.

#### Simple Format

When using the simple format, simply specify the address for the account, and a single hex-encoded
private key.

```json
...

"accounts": {
  "admin-account": {
    "address": "3ae53cb6e3f42a79",
    "keys": "12332967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad1111"
  }
}

...
```

#### Advanced format

The advanced format allows us to define more properties for the account. 
We can define the signature algorithm and hashing algorithm, as well as custom key formats.

Please note that we can use `service` for address in case the account is used on `emulator` network as this is a special 
value that is defined on the run time to the default service address on the emulator network.

```json
...

"accounts": {
  "admin-account-multiple-keys": {
    "address": "service",
    "chain": "flow-emulator",
    "keys": [
      {
        "type": "hex",
        "index": 0,
        "signatureAlgorithm": "ECDSA_P256",
        "hashAlgorithm": "SHA3_256",
        "context": {
          "privateKey": "12332967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad1111"
        }
      },
      {
        "type": "hex",
        "index": 1,
        "signatureAlgorithm": "ECDSA_P256",
        "hashAlgorithm": "SHA3_256",
        "context": {
          "privateKey": "333332967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad1111"
        }
      }
    ]
  }
}

...
```

### Deployments

The deployments section defines where the `project deploy` command will deploy specified contracts. 
This configuration property acts as the glue that ties together accounts, 
contracts and networks, all of which are referenced by name.

```json
...

"deployments": {
  "emulator": {
    "emulator-account": [
      "NonFungibleToken",
      "Kibble",
      "KittyItems",
      "KittyItemsMarket"
    ]
  },
  "testnet": {
    "admin-account": ["NonFungibleToken"],
    "user-account": [
      "Kibble",
      "KittyItems",
      "KittyItemsMarket"
    ]
  }
}

...
```

### Networks

Use this section to define networks and connection parameters for that specific network. 

#### Simple format

```json
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
