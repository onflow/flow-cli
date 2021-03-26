---
title: Flow CLI security
sidebar_title: Security
description: How to securely use CLI
---

Handling accounts, and their private keys is intrinsically dangerous. We must take extra
precautions to secure private keys and not expose them to third parties. Flow CLI enables 
many options to secure the private keys.

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

### Google Cloud Key Managment System
tbd