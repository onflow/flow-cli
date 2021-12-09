## ⬆️ Install or Upgrade

Follow the [Flow CLI installation guide](../install.md) for instructions on how to install or upgrade the CLI.


## 💥 Breaking Changes

### Flow Go SDK Update
Update Flow Go SDK version from v0.19.0 to v0.20.0. 
Read more about new version in the [release notes](https://github.com/onflow/flow-go-sdk/releases/tag/v0.20.0).

## ⭐ Features

### New Command To Manage Configuration
Add or remove resources from the configuration using the new `flow config` command. 
Usage is possible via interactive prompt or by using flags. Command syntax is as follows:
```js
flow config <add|remove> <account|contract|deployment|network>
```

Example for adding an account to the config via interactive prompt:

```bash
Name: Foo
Address: f8d6e0586b0a20c7
✔ ECDSA_P256
✔ SHA3_256
Private key: 1286...01afc
Key index (Default: 0): 0

Account Foo added to the configuration
```

Example for adding an account to the config without interactive prompt:

```bash
./main config add account --address f8d6e0586b0a20c7 --name Foo --private-key 1286...01afc

Account Foo added to the configuration
```

We recommend using manage command to do any changes in the configuration as it will also 
validate input values for you and will abstract any changes in the configuration format.

### Decode Keys
Command for decoding public keys in the RLP encoded format.

Example of using the command: 

```bash
> flow keys decode f847b84084d716c14b051ad6b001624f738f5d302636e6b07cc75e4530af7776a4368a2b586dbefc0564ee28384c2696f178cbed52e62811bcc9ecb59568c996d342db2402038203e8

Public Key 		 84d716c14b051ad6b001624f738f5d302636e6b07cc75e4530af7776a4368a2b586dbefc0564ee28384c2696f178cbed52e62811bcc9ecb59568c996d342db24 
Signature algorithm 	 ECDSA_P256
Hash algorithm 		 SHA3_256
Weight 			 1000
Revoked 		 false
```

## 🎉 Improvements

### Include And Exclude Flags
Include and Exclude flags were added to the transaction and account resource thus 
allowing you to further specify verbosity of the output.

### Documentation Changes
Multiple reported documentation fixes.

## 🐞 Bug Fixes

### Import Detection Fix
Fix for a reported bug: An error occurs when executing a script that imports a built-in contract (Crypto contract) with Flow CLI command.
