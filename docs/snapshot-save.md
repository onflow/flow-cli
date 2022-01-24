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
- OutputPath: `path to save the snapshot file`
- Valid Input: any valid string path

This is where the protocol snapshot JSON file will be saved.

## Flags

### Signer

- Flag: `--signer`
- Valid inputs: the name of an account defined in the configuration (`flow.json`)

Specify the name of the account that will be used to sign the transaction.

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






