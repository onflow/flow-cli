---
title: Get Events with the Flow CLI
sidebar_title: Get Event
description: How to get an event from the command line
---

Use the event command to fetch a single or multiple events in a specific range of blocks. 
You can provide start and end block height range, but also specify number of the latest blocks to 
be used to search for specified event. Events are fetched concurrently by using multiple workers which 
optionally you can also control by specifying the flags.

```shell
flow events get <event_name>
```

## Example Usage

Get the event by name `A.0b2a3299cc857e29.TopShot.Deposit` from the last 20 blocks on mainnet.
```shell
> flow events get A.0b2a3299cc857e29.TopShot.Deposit --last 20 --network mainnet

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

Get two events `A.1654653399040a61.FlowToken.TokensDeposited` 
and `A.1654653399040a61.FlowToken.TokensWithdrawn` in the block height range on mainnet. 
```shell
> flow events get \
  A.1654653399040a61.FlowToken.TokensDeposited \
  A.1654653399040a61.FlowToken.TokensWithdrawn \ 
  --start 11559500 --end 11559600 --network mainnet
  
  Events Block #17015045:
    Index	0
    Type	A.1654653399040a61.FlowToken.TokensWithdrawn
    Tx ID	6dcf60d54036acb52b2e01e69890ce34c3146849998d64364200e4b21e9ac7f1
    Values
		- amount (UFix64): 0.00100000 
		- from (Address?): 0x9e06eebf494e2d78 

    Index	1
    Type	A.1654653399040a61.FlowToken.TokensWithdrawn
    Tx ID	6dcf60d54036acb52b2e01e69890ce34c3146849998d64364200e4b21e9ac7f1
    Values
		- amount (UFix64): 0.00100000 
		- from (Never?): nil 

  Events Block #17015047:
    Index	0
    Type	A.1654653399040a61.FlowToken.TokensWithdrawn
    Tx ID	24979a3c0203f514f7f5822cc8ae7046e24f25d4a775bef697a654898fb7673e
    Values
		- amount (UFix64): 0.00100000 
		- from (Address?): 0x18eb4ee6b3c026d2 

    Index	1
    Type	A.1654653399040a61.FlowToken.TokensWithdrawn
    Tx ID	24979a3c0203f514f7f5822cc8ae7046e24f25d4a775bef697a654898fb7673e
    Values
		- amount (UFix64): 0.00100000 
		- from (Never?): nil 
```

## Arguments

### Event Name

- Name: `event_name`
- Valid Input: String

Fully-qualified identifier for the events.
You can provide multiple event names separated by a space.

## Flags

### Start

- Flag: `--start`
- Valid inputs: valid block height

Specify the start block height used alongside the end flag. 
This will define the lower boundary of the block range.

### End

- Flag: `--end`
- Valid inputs: valid block height

Specify the end block height used alongside the start flag.
This will define the upper boundary of the block range.

### Last

- Flag: `--last`
- Valid inputs: number
- Default: `10`

Specify the number of blocks relative to the last block. Ignored if the 
start flag is set. Used as a default if no flags are provided.

### Batch

- Flag: `--batch`
- Valid inputs: number
- Default: `25`

Number of blocks each worker will fetch.

### Workers

- Flag: `--workers`
- Valid inputs: number
- Default: `10`

Number of workers to use when fetching events concurrently.


### Host

- Flag: `--host`
- Valid inputs: an IP address or hostname.
- Default: `127.0.0.1:3569` (Flow Emulator)

Specify the hostname of the Access API that will be
used to execute the command. This flag overrides
any host defined by the `--network` flag.

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

- Flag: `--config-path`
- Short Flag: `-f`
- Valid inputs: a path in the current filesystem.
- Default: `flow.json`

Specify the path to the `flow.json` configuration file.
You can use the `-f` flag multiple times to merge
several configuration files.
