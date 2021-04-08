---
title: Get Network Status 
sidebar_title: Network Status 
description: How to get access node status from the command line
---

The Flow CLI provides a command to get network status of specified Flow Access Node

`flow status`

## Example Usage

```shell
> flow status --network testnet

Status:		 🟢 ONLINE
Network:	 testnet
Access Node:	 access.devnet.nodes.onflow.org:9000
```

## Options

### Network

- Flag: `--network`
- Short Flag: `-n`
- Valid inputs: the name of a network defined in the configuration (`flow.json`)

Specify which network you want the command to use for execution.

