---
title: Create Emulator Snapshot with the Flow CLI
sidebar_title: Create Emulator Snapshot
description: How to start create emulator snapshot from the command line
---

The Flow CLI provides a command to create emulator snapshots, which are points in blockchain 
history you can later jump to and reset the state to that moment. This can be useful for testing where you 
establish a begining state, run tests and after revert back to the initial state.

The command syntax is:
```shell
flow emulator snapshot create|load|list {name}
```

## Example Usage

### Create a new snapshot
Create a new emulator snapshot at the current block with a name of `myInitialState`. 
```shell
> flow emulator snapshot create myInitialState
```

### Load an existing snapshot
To jump to a previously created snapshot we use the load command in combination with the name.
```shell
> flow emulator snapshot load myInitialState
```

### List all existing snapshots
To list all the existing snapshots we previously created and can load to we run the following command:
```shell
> flow emulator list
```


To learn more about using the Emulator, have a look at the [README of the repository](https://github.com/onflow/flow-emulator).

## Flags

### Emulator Flags
You can specify any [emulator flags found here](https://github.com/onflow/flow-emulator#configuration) and they will be applied to the emulator service.

### Configuration

- Flag: `--config-path`
- Short Flag: `-f`
- Valid inputs: valid filename

Specify a filename for the configuration files, you can provide multiple configuration
files by using `-f` flag multiple times.

### Version Check

- Flag: `--skip-version-check`
- Default: `false`

Skip version check during start up to speed up process for slow connections.
