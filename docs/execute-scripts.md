---
title: Execute a Script with the Flow CLI
sidebar_title: Execute a Script
description: How to execute a Cadence script on Flow from the command line
---

The Flow CLI provides a command to execute a Cadence script on
the Flow execution state with any Flow Access API.

`flow scripts execute <filename>`

## Example Usage

```shell
# Submit a transaction to Flow Testnet
> flow scripts execute script.cdc --arg String:"Hello" --arg String:"World"

"Hello World"
```

## Arguments

### Filename
- Name: `filename`
- Valid inputs: Any filename and path valid on the system.

The first argument is a path to a Cadence file containing the 
script to be executed.

## Flags

### Arguments
- Flag: `--arg`
- Valid inputs: Argument in `Type:Value` format.

Arguments passed to the Cadence script in `Type:Value` format. 
The `Type` must be the same as type in the script source code for that argument.  

### Arguments JSON
- Flag: `--argsJSON`
- Valid inputs: Arguments in JSON-Cadence format.

Arguments passed to the Cadence script in `Type:Value` format.
The `Type` must be the same as type in the script source code for that argument.

### Code
⚠️  DEPRECATED: use filename argument.

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





