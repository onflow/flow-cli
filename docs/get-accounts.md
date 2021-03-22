---
title: Get an Account with the Flow CLI
sidebar_title: Get an Account
description: How to get a Flow account from the command line
---

The Flow CLI provides a command to fetch any account by its address from the Flow network.

`flow accounts get <address>`


## Example Usage

```shell
flow accounts get 0xf8d6e0586b0a20c7
```

### Example response
```shell
Address	 0xf8d6e0586b0a20c7
Balance	 9999999999970000000
Keys	 1

Key 0	Public Key		 858a7d978b25d61f348841a343f79131f4b9fab341dd8a476a6f4367c25510570bf69b795fc9c3d2b7191327d869bcf848508526a3c1cafd1af34f71c7765117
	Weight			 1000
	Signature Algorithm	 ECDSA_P256
	Hash Algorithm		 SHA3_256

Contracts Deployed: 2
Contract: 'FlowServiceAccount'
Contract: 'FlowStorageFees'


```

## Flags

### Contracts

- Flag: `--contracts`

Display contracts deployed to the account.

### Code 
⚠️  DEPRECATED: use contracts flag instead.


### Host
- Flag: `--host`
- Valid inputs: an IP address or hostname.
- Default: `127.0.0.1:3569` (Flow Emulator)

Specify the hostname of the Access API that will be
used to execute the commands.

### Network

- Flag: `--network`
- Short Flag: `-n`
- Valid inputs: the name of a network defined in the configuration (`flow.json`)

Specify which network you want the command to use for execution.

### Filter

- Flag: `--filter`
- Short Flag: `-x`
- Valid inputs: case-sensitive name of the result property.

Specify any property name from the result you want to return as the only value.

### Output

- Flag: `--output`
- Short Flag: `-o`
- Valid inputs: `json`, `inline`

Specify in which format you want to display the result.

### Save

- Flag: `--save`
- Short Flag: `-s`
- Valid inputs: valid filename

Specify the filename where you want the result to be saved.

### Log

- Flag: `--log`
- Short Flag: `-l`
- Valid inputs: `none`, `error`, `debug`
- Default: `info`

Specify the log level. Control how much output you want to see while command execution.

### Configuration

- Flag: `--conf`
- Short Flag: `-f`
- Valid inputs: valid filename

Specify a filename for the configuration files, you can provide multiple configuration
files by using `-f` flag multiple times.





