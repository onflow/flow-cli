---
title: Flow CLI Configuration
sidebar_title: Configuration
description: What is Flow CLI Configuration
---

Flow CLI uses a state called configuration which is stored in a file (usually `flow.json`). 

Flow configuration (`flow.json`) file will contain the following properties:

- A `networks` list pre-populated with the Flow emulator, testnet and mainnet connection configuration.
- An `accounts` list pre-populated with the Flow Emulator service account.
- A `deployments` empty object where all [deployment targets](/tools/flow-cli/project-contracts#define-contract-deployment-targets) can be defined. 
- A `contracts` empty object where you [define contracts](/tools/flow-cli/project-contracts#add-a-contract) you wish to deploy.

## Example Project Configuration

```json
{
  "networks": {
    "emulator": "127.0.0.1:3569",
    "mainnet": "access.mainnet.nodes.onflow.org:9000",
    "testnet": "access.devnet.nodes.onflow.org:9000"
  },
  "accounts": {
    "emulator-account": {
      "address": "f8d6e0586b0a20c7",
      "key": "ae1b44c0f5e8f6992ef2348898a35e50a8b0b9684000da8b1dade1b3bcd6ebee",
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
      "key": "12332967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad1111"
    },
    "user-account": {
      "address": "e2a8b7f23e8b548f",
      "key": "22232967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad1111"
    },
    "emulator-account": {
      "address": "f8d6e0586b0a20c7",
      "key": "2eae2f31cb5b756151fa11d82949c634b8f28796a711d7eb1e52cc301ed11111",
    }
  },

  "networks": {
    "emulator": "127.0.0.1:3569",
    "mainnet": "access.mainnet.nodes.onflow.org:9000",
    "testnet": "access.devnet.nodes.onflow.org:9000",
    "testnetSecure": {
      "Host": "access-001.devnet30.nodes.onflow.org:9001",
      "NetworkKey": "ba69f7d2e82b9edf25b103c195cd371cf0cc047ef8884a9bbe331e62982d46daeebf836f7445a2ac16741013b192959d8ad26998aff12f2adc67a99e1eb2988d"
    }
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
and address `ee82856bf20e2aa6` when deploying to the Flow emulator. 
We can specify aliases for each network we have defined. When deploying to testnet it is always a good idea to specify aliases for all the [common 
contracts](https://docs.onflow.org/core-contracts) that have already been deployed to the testnet. 

⚠️ If we use an alias for the contract we should not specify it in the `deployment` section for that network. 

Our example below should not include `FungibleToken` in  `deployment` section for testnet and emulator network.

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

Format used to specify advanced contracts is:
```json
"CONTRACT NAME": {
    "source": "CONTRACT SOURCE FILE LOCATION",
    "aliases": {
        "NETWORK NAME": "ADDRESS ON SPECIFIED NETWORK WITH DEPLOYED CONTRACT"
        ...
    }
}
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
    "key": "12332967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad1111"
  }
}

...
```

#### Advanced format

The advanced format allows us to define more properties for the account. 
We can define the signature algorithm and hashing algorithm, as well as custom key formats.

Please note that we can use `service` for address in case the account is used on `emulator` network as this is a special 
value that is defined on the run time to the default service address on the emulator network.

**Example for advanced hex format:**
```json
...

"accounts": {
  "admin-account": {
    "address": "service",
    "key":{
        "type": "hex",
        "index": 0,
        "signatureAlgorithm": "ECDSA_P256",
        "hashAlgorithm": "SHA3_256",
        "privateKey": "12332967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad1111"
      }
  }
}

...
```

You can also use BIP44 to derive keys from a mnemonic. For more details please see the [FLIP](https://github.com/onflow/flips/blob/main/flips/20201125-bip-44-multi-account.md)

**Example for BIP44 format:**
```json
...

"accounts": {
  "admin-account": {
    "address": "service",
    "key":{
        "type": "bip44",
        "index": 0,
        "signatureAlgorithm": "ECDSA_P256",
        "hashAlgorithm": "SHA3_256",
        "mnemonic": "skull design wagon top faith actor valley crystal subject volcano access join",
        "derivationPath": "m/44'/539'/0'/0/0"
      }
  }
}

...
```

Note: Default value for `derivationPath` is `m/44'/539'/0'/0/0` if omitted. 



You can also use a key management system (KMS) to sign the transactions. Currently, we only support Google KMS.

**Example for Google KMS format:**
```json
...
"accounts": {
  "admin-account": {
    "address": "service",
    "key": {
        "type": "google-kms",
        "index": 0,
        "signatureAlgorithm": "ECDSA_P256",
        "hashAlgorithm": "SHA3_256",
        "resourceID": "projects/flow/locations/us/keyRings/foo/bar/cryptoKeyVersions/1"
    }
  }
}
...
```

You can store the account key to a separate file and provide the file location as part of the key configuration.

**Example for separate key file:**
```json
...
"accounts": {
  "admin-account": {
    "address": "service",
    "key": {
        "type": "file",
        "location": "./test.key"
    }
  }
}
...
```
Inside the `test.key` file you should only put the hex key content (e.g. `12332967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad1111`)

### Deployments

The deployments section defines where the `project deploy` command will deploy specified contracts. 
This configuration property acts as the glue that ties together accounts, 
contracts and networks, all of which are referenced by name.

In the deployments section we specify the network, account name and list of contracts to be deployed to that account.

Format specifying the deployment is:
```json
...
"deployments": {
  "NETWORK": {
    "ACCOUNT NAME": ["CONTRACT NAME"]
  }
}

...
```


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

Format for networks is:

```json
...
"networks": {
  "NETWORK NAME": "ADDRESS"
}
...
```

```json
...
"networks": {
  "NETWORK NAME": {
    "host": "ADDRESS",
    "key": "ACCESS NODE NETWORK KEY"    
  }
}
...
```

```json
...

"networks": {
    "emulator": "127.0.0.1:3569",
    "mainnet": "access.mainnet.nodes.onflow.org:9000",
    "testnet": "access.devnet.nodes.onflow.org:9000",
    "testnetSecure": {
        "host": "access-001.devnet30.nodes.onflow.org:9001",
        "key": "ba69f7d2e82b9edf25b103c195cd371cf0cc047ef8884a9bbe331e62982d46daeebf836f7445a2ac16741013b192959d8ad26998aff12f2adc67a99e1eb2988d"
    },
}

...
```
### Emulators

The default emulator CLI is automatically configured with name being `"default"` and values of 
`serviceAccount`: `"emulator-account"` and `port`: `"3569"`. The default emulator configuration will not show up on 
flow.json.

To customize emulator values, add emulator section like the example below: 

```json
...

"emulators": {
    "custom-emulator": {
        "port": 3600,
        "serviceAccount": "emulator-account"
    }
}

...
```
