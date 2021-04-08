---
title: Initialize Flow Configuration
sidebar_title: Initialize Configuration
description: How to initialize Flow configuration using CLI
---

Flow CLI uses a state to operate which is called configuration (usually `flow.json` file). 
Before using commands that require this configuration we must initialize the project by 
using the init command. Read more about [state configuration here](https://docs.onflow.org/flow-cli/configuration/).

`flow init`

## Example Usage

```shell
> flow init

Configuration initialized
Service account: 0xf8d6e0586b0a20c7

Start emulator by running: 'flow emulator' 
Reset configuration using: 'flow init --reset'

```

### Error Handling

Existing configuration will cause the error bellow. 
You should initialize in an empty folder or reset configuration using `--reset` flag 
or by removing the configuration file first.
```shell
❌ Command Error: configuration already exists at: flow.json, if you want to reset configuration use the reset flag
```


## Flags

### Reset

- Flag: `reset`

Using this flag will reset the existing configuration and create a new one.

### Service Private Key

- Flag: `service-private-key`
- Valid inputs: a hex-encoded private key in raw form.

Private key used on the default service account.


### Service Key Signature Algorithm

- Flag: `--sig-algo`
- Valid inputs: `"ECDSA_P256", "ECDSA_secp256k1"`
- Default: `"ECDSA_P256"`

Specify the ECDSA signature algorithm for the provided public key.

Flow supports the secp256k1 and P-256 curves.

### Service Key Hash Algorithm

- Flag: `--hash-algo`
- Valid inputs: `"SHA2_256", "SHA3_256"`
- Default: `"SHA3_256"`

Specify the hashing algorithm that will be paired with the public key
upon account creation.

### Log

- Flag: `--log`
- Short Flag: `-l`
- Valid inputs: `none`, `error`, `debug`
- Default: `info`

Specify the log level. Control how much output you want to see while command execution.






