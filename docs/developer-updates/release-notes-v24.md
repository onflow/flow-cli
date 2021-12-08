## ⬆️ Install or Upgrade

Follow the [Flow CLI installation guide](../install.md) for instructions on how to install or upgrade the CLI.

## 💥 Breaking Changes

### Cadence and Language Server Update
Cadence was updated to version v0.18.0 and language server to version v0.18.2 which implements the breaking changes from flowkit library.

### FlowKit API Changes
CLI implements a flowkit utility library which can be reused in other services. This is early stage development and the API for this library was refactored and improved. 

## ⭐ Features

### Decode PEM Public Key
New command for decoding PEM encoded public key. You can use the decoding command like so:
```bash
flow keys decode pem --from-file key.pem

Public Key 		 d479b3cdc9edbddb195cb12b35161ade826b032a64bdd4062cc87fb3ba7e71c9cf646ff23990bb4532ca45c445c7e908cef278b2c4615360039a6660a366a95f 
Signature algorithm 	 ECDSA_P256
Revoked 		 false

```

## 🎉 Improvements

### Validate Configuration
Configuration validation has been improved and will provide better feedback when there are wrong values set in the `flow.json`.

### Updated Cobra
Cobra library was updated to the latest version.

### Refactored Testing
Testing suite was completely refactored and improved which will provide better code coverage and more reliable codebase.

## 🐞 Bug Fixes

### Refactored Event Display
Events output on the transaction command was refactored, so it better handles special values in the events.

### Flow Init Warning
Flow init command incorrectly displayed a warning which is now removed.

### Transaction IDs Output
All commands that send transactions to the network now display that transaction ID for better visibility.

