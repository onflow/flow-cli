---
title: Get Network Status
sidebar_title: Network Status
description: How to get access node status from the command line
---

The Flow CLI provides a command to get network status of specified Flow Access Node

```shell
flow status
```

## Example Usage
```shell
> flow status --network mainnet

Network is: ONLINE
```

## Options

### Network

- Flag: `--network`
- Short Flag: `-n`
- Valid inputs: the name of a network defined in the configuration (`flow.json`)

Specify which network you want the command to use for execution.

