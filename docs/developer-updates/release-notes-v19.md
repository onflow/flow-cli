## ⬆️ Install or Upgrade

Follow the [Flow CLI installation guide](https://docs.onflow.org/flow-cli/install/) for instructions on how to install or upgrade the CLI.

## ⭐ Features

### Project Deployment with Contract Initialization Arguments
Project deployment was enhanced and it now supports providing initialization 
arguments during the deployment of contracts. It is easy to specify all 
the arguments in the configuration like so:

```
...
  "deployments": {
    "testnet": {
      "my-testnet-account": [
        "NonFungibleToken", {
            "name": "KittyItems", 
            "args": [{
                "name": "supply",
                "type": "UInt",
                "value": "10",
            }]
        }]
    }
  }
...
```

### Network Status Command
Network status command allows you to query the status of each network and 
see if the network is available.

Example:
```
> flow status --network testnet

Status:		 🟢 ONLINE
Network:	 testnet
Access Node:	 access.devnet.nodes.onflow.org:9000
```

### Global Configuration
Flow CLI now supports global configuration which is a `flow.json` file saved in your home
directory and loaded as the first configuration file wherever you execute the CLI command.

You can generate a global configuration using `--global` flag.

Command example: `flow init --global`.

Global flow configuration is saved as:
- MacOs: `~/flow.json`
- Linux: `~/flow.json`
- Windows: `C:\Users\$USER\flow.json`

You can read more about it in [the docs](https://docs.onflow.org/flow-cli/initialize-configuration/).

### Environment File Support

The CLI will load environment variables defined in the 
`.env` file in the active directory, if one exists. 
These variables can be substituted inside the `flow.json`, 
just like any other environment variable.

Example `.env` file:
```bash
PRIVATE_KEY=123
```

```json
// flow.json
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

## 🎉 Improvements

### Default Network Without Configuration
Default network is provided even if no configuration is present which
allows you to use the CLI on even more commands without the requirement of
having a configuration pre-initialized.

### Chain ID Removed
Chain ID property was removed from the configuration as it is not needed anymore. 
With this improvement, the new configuration is less complex and shorter.

### Multiple Keys Removed
Multiple keys were removed from the account configuration making it 
less complex. Should you have a rare need of specifying multiple keys 
on account you can specify multiple accounts with the same address and 
different keys. This new functionality is backward compatible.

## 🐞 Bug Fixes

### Keys Generate JSON output
Keys generation output in JSON format was fixed and it now shows correctly 
private and public keys.