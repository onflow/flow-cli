## ⬆️ Install or Upgrade

Follow the [Flow CLI installation guide](../install.md) for instructions on how to install or upgrade the CLI.

## ⭐ Features

### Project Deployment with Contract Initialization Arguments
Project deployment was improved, and it now supports providing initialization arguments during the deployment of contracts. It is easy to specify all the arguments in the configuration like so:

```js
// flow.json
{
  // ...
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
  // ...
}
```

### Network Status Command
The network status command allows you to query the status of each network and see if the network is available.

Example:
```
> flow status --network testnet

Status:		 🟢 ONLINE
Network:	 testnet
Access Node:	 access.devnet.nodes.onflow.org:9000
```

### Global Configuration
Flow CLI now supports global configuration which is a `flow.json` file saved in your home directory and loaded as the first configuration file wherever you execute the CLI command.

You can generate a global configuration using the `--global` flag.

Command example: `flow init --global`.

Global flow configuration is saved as:
- macOS: `~/flow.json`
- Linux: `~/flow.json`
- Windows: `C:\Users\$USER\flow.json`

You can read more about it in [the docs](../initialize-configuration.md).

### Environment File Support

The CLI will load environment variables defined in the `.env` file in the active directory, if one exists. These variables can be substituted inside the `flow.json`, just like any other environment variable.

Example `.env` file:
```bash
PRIVATE_KEY=123
```

```js
// flow.json
{
  // ...
  "accounts": {
    "my-testnet-account": {
      "address": "3ae53cb6e3f42a79",
      "keys": "${PRIVATE_KEY}"
    }
  }
  // ...
}
```

## 🎉 Improvements

### Default Network Without Configuration
Default network is provided even if no configuration is present which allows you to use the CLI on even more commands without the requirement of having a configuration pre-initialized.

### Chain ID Removed
Chain ID property was removed from the configuration as it is not needed anymore.
With this improvement, the new configuration is less complex and shorter.

### Send Signed Progress
Send signed transaction now includes progress output the same way as sending transaction command does.

## 🐞 Bug Fixes

### Keys Generate JSON output
Keys generation output in JSON format was fixed and it now shows correctly private and public keys.

### Account Key Index When Sending Transactions
Account key index is now fetched from the configuration and it doesn't default to 0 anymore.

### Transaction Boolean Argument
The transaction boolean argument wasn't parsed correctly when passed in comma split format.

### JSON Outputs Fixes
JSON output format was not working properly for some commands.
