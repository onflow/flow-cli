---
title: Execute a Script with the Flow CLI
sidebar_title: Execute a Script
description: How to execute a Cadence script on Flow from the command line
---

The Flow CLI provides a command to execute a Cadence script on
the Flow execution state with any Flow Access API.

`flow scripts execute`

## Example Usage

```shell
# Submit a transaction to Flow Testnet
> flow scripts execute MyScript.cdc \
    --host access.testnet.nodes.onflow.org:9000
```


## Arguments

### Script Code

The first argument is the path to the Cadence file containing the 
script to be executed.

## Options

### Host

- Flag: `--host`
- Valid inputs: an IP address or hostname.
- Default: `localhost:3569` (Flow Emulator)

Specify the hostname of the Access API that will be
used to execute the script.
