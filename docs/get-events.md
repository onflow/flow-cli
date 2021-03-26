---
title: Get Events with the Flow CLI
sidebar_title: Get Event
description: How to get an event from the command line
---

The Flow CLI provides a command to fetch any block from the Flow network.

Events can be requested for a specific sealed block range via the 
start and end block height fields and further filtered by event name.

```shell
flow events get <event_name> <block_height_range_start> <optional:block_height_range_end|latest>`
```

## Example Usage

```shell
flow events get A.0b2a3299cc857e29.TopShot.Deposit 12913388 12913389 \
 --host access.mainnet.nodes.onflow.org:9000
```

### Example response

```shell
Events Block #12913388:
	 Index	 2
	 Type	 A.0b2a3299cc857e29.TopShot.Deposit
	 Tx ID	 0a1e6cdc4eeda0e23402193d7ad5ba01a175df4c08f48fa7ac8d53e811c5357c
	 Values
		 id (UInt64)	3102159
		 to ({}?)	24214cf0faa7844d

	 Index	 2
	 Type	 A.0b2a3299cc857e29.TopShot.Deposit
	 Tx ID	 1fa5e64dcdc8ed5dad87ba58207ee4c058feb38fa271fff659ab992dc2ec2645
	 Values
		 id (UInt64)	5178448
		 to ({}?)	26c96b6c2c31e419

	 Index	 9
	 Type	 A.0b2a3299cc857e29.TopShot.Deposit
	 Tx ID	 262ab3996bdf98f5f15804c12b4e5d4e89c0fa9b71d57be4d7c6e8288c507c4a
	 Values
		 id (UInt64)	1530408
		 to ({}?)	2da5c6d1a541971b

...
```

## Arguments

### Event Name

- Name: `event_name`
- Valid Input: String

Fully-qualified identifier for the events.

### Block Height Range Start

- Name: `block_height_range_start`
- Valid Input: Number (lower than `block_height_range_end` value)

Height of the block in the chain.

### Block Height Range End (optional)

- Name: `block_height_range_end`
- Valid Input: Number (higher than `block_height_range_end` value) or value `latest`

Height of the block in the chain. Use `latest` for latest block.

## Flags

### Host

- Flag: `--host`
- Valid inputs: an IP address or hostname.
- Default: `127.0.0.1:3569` (Flow Emulator)

Specify the hostname of the Access API that will be
used to execute the command. This flag overrides
any host defined by the `--network` flag.

### Network

- Flag: `--network`
- Short Flag: `-n`
- Valid inputs: the name of a network defined in the configuration (`flow.json`)

Specify which network you want the command to use for execution.

### Filter

- Flag: `--filter`
- Short Flag: `-x`
- Valid inputs: a case-sensitive name of the result property.

Specify any property name from the result you want to return as the only value.

### Output

- Flag: `--output`
- Short Flag: `-o`
- Valid inputs: `json`, `inline`

Specify the format of the command results.

### Save

- Flag: `--save`
- Short Flag: `-s`
- Valid inputs: a path in the current filesystem.

Specify the filename where you want the result to be saved

### Log

- Flag: `--log`
- Short Flag: `-l`
- Valid inputs: `none`, `error`, `debug`
- Default: `info`

Specify the log level. Control how much output you want to see during command execution.

### Configuration

- Flag: `--conf`
- Short Flag: `-f`
- Valid inputs: a path in the current filesystem.
- Default: `flow.json`

Specify the path to the `flow.json` configuration file.
You can use the `-f` flag multiple times to merge
several configuration files.
