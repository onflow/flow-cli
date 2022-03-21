---
title: Start Development Tools with the Flow CLI
sidebar_title: Development Tools
description: How to start development tools using the Flow CLI
---

The Flow CLI integrates different development tools, which can now be easily started 
and managed from a single place. 

Currently the CLI supports starting:  
- [FCL Development Wallet](https://github.com/onflow/fcl-dev-wallet)


## FCL Development Wallet

The FCL dev wallet is a mock Flow wallet that simulates the protocols used by FCL to interact with the Flow blockchain on behalf of simulated user accounts.

**Be sure you have the emulator running before starting this command**
_You can start it using the `flow emulator` command_.

```shell
flow dev-wallet
```
_⚠️ This project implements an FCL compatible
interface, but should **not** be used as a reference for
building a production grade wallet._

After starting dev-wallet, you can set your fcl config to use it like below:

```javascript
import * as fcl from "@onflow/fcl"

fcl.config()
  // Point App at Emulator
  .put("accessNode.api", "http://localhost:8080") 
  // Point FCL at dev-wallet (default port)
  .put("discovery.wallet", "http://localhost:8701/fcl/authn") 
```
You can read more about setting up dev-wallet at [FCL Dev Wallet Project](https://github.com/onflow/fcl-dev-wallet)


## Flags

### Port

- Flag: `--port`
- Valid inputs: Number
- Default: `8701`

Port on which the dev wallet server will listen on. 

### Emulator Port

- Flag: `--emulator-port`
- Valid inputs: Number

Port on which the emulator is listening on.

### Configuration

- Flag: `--config-path`
- Short Flag: `-f`
- Valid inputs: valid filename

Specify a filename for the configuration files, you can provide multiple configuration
files by using `-f` flag multiple times.





