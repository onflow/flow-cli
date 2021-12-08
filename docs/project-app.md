---
title: Create Flow app from scaffold
sidebar_title: Create App
description: How to create a Flow app with recommended structure
---

The Flow CLI provides a command to create an app from examples
and templates.

```shell
flow app create
```

## Example Usage

```shell
> flow app create my-app

Enumerating objects: 349, done.
Counting objects: 100% (349/349), done.
Compressing objects: 100% (261/261), done.
Total 349 (delta 55), reused 347 (delta 53), pack-reused 0
? Which do you want to start with? Template
? Which API template you want to start with? express
? Which Cadence template you want to start with? default
? Which Web template you want to start with? react

Created  /path/to/my-app
Api      express
Cadence  default
Web      react
```

## Arguments

### Path

- Name: `path`
- Valid Input: Path

A relative path to the app location. Can be a new folder name or existing empty folder.  
You can start with a full featured example or a custom template.

If you start with an example, there are some exmples included (may not up-to-date):

- `kitty-items`: An app based on CryptoKitties. ([Repo](https://github.com/onflow/kitty-items))

If you start with a template, there will be 3 folder `api`, `cadence`, and `web`.

- `api`: Backend service will be put in here, you can provide functions which cannot be
  implemented in client side, such as draw, whitelist registration, KYC, or some works can be
  offloaded from client side, such as DEX price, NFT lists.
- `cadence`: Contracts will be put in here.
- `web`: Frontend app will be put in here, users will use the app to interact with your Cadence
  contracts. Check out [fcl.js](https://github.com/onflow/fcl-js) to get more details.

## Flags

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
