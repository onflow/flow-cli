---
title: Flow CLI security
sidebar_title: Security
description: How to securely use CLI
---

The managing of accounts and private keys is intrinsically dangerous. 
We must take extra precautions to not expose private key data when using
the CLI.

The Flow CLI provides several options to secure private account data.

⚠️ Warning: please be careful when using private keys in configuration files. 
Never commit private key data to source control.
If private key data must be kept in text, we suggest using a separate file
that is not checked into source control (e.g. excluded with `.gitignore`).

### Private Account Configuration File

#### Main configuration file
```json
// flow.json
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

#### Private configuration file

⚠️ Put this file in `.gitignore`:

```json
// flow.testnet.json
{
  "accounts": {
    "my-testnet-account": {
      "address": "3ae53cb6e3f42a79",
      "key": "334232967f52bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad1111"
    }
  }
}
```

### Store Configuration in Environment Variables

You can use environment variables for values that should be kept private (e.g. private keys, addresses).

See example below:

```shell
PRIVATE_KEY=key flow project deploy
```

```json
// flow.json
{
  ...
  "accounts": {
    "my-testnet-account": {
      "address": "3ae53cb6e3f42a79",
      "key": "$PRIVATE_KEY"
    }
  }
  ...
}
```

### Private Dotenv File

The CLI will load environment variables defined in the `.env` file in the active directory, if one exists. 
These variables can be substituted inside the `flow.json`, 
just like any other environment variable.

⚠️ You should never commit `.env` to source control, 
especially if it contains sensitive information 
like a private key.

Example `.env` file:
```bash
PRIVATE_KEY=123
```

### Composing Multiple Configuration Files

You can merge multiple configuration files like so:

```shell
flow project deploy -f main.json -f private.json
```
