---
title: Save a protocol snapshot with the FLOW CLI
sidebar_title: Snapshot Save
description: How to save a protocol snapshot from the command line
---

The FLOW CLI provides a command to save the latest finalized protocol state snapshot

```shell
flow snapshot save <output path>
```

## Example Usage

```shell
flow snapshot save  /tmp/snapshot.json --network testnet
```

### Example response
```shell
snapshot saved: /tmp/snapshot.json
```

## Arguments

### Output Path
- Name: `output path`
- Valid Input: any valid string path

Output path where the protocol snapshot JSON file will be saved.

## Flags


### Host
- Flag: `--host`
- Valid inputs: an IP address or hostname.
- Default: `127.0.0.1:3569` (Flow Emulator)

Specify the hostname of the Access API that will be
used to execute the commands.

### Network Key

- Flag: `--network-key`
- Valid inputs: A valid network public key of the host in hex string format

Specify the network public key of the Access API that will be
used to create a secure GRPC client when executing the command.

### Network

- Flag: `--network`
- Short Flag: `-n`
- Valid inputs: the name of a network defined in the configuration (`flow.json`)
- Default: `emulator`

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






