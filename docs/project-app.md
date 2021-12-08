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

✔ Example
✔ kitty-items
Enumerating objects: 5632, done.
Counting objects: 100% (1342/1342), done.
Compressing objects: 100% (692/692), done.
Total 5632 (delta 768), reused 708 (delta 649), pack-reused 4290

Created	 /Users/dapper/Dev/flyinglimao/flow-cli/test4
Example	 kitty-items

```

## Arguments

### Path

- Name: `path`
- Valid Input: Path

A relative path to the app location. Can be a new folder name or existing empty folder.  
You can start with a fully featured example or a custom template.

If you choose example, we will automatically clone a project for you from the list of possible examples: (may not up-to-date):

- `kitty-items`: An app based on CryptoKitties. ([Repo](https://github.com/onflow/kitty-items))

If you start with a template, three folders will be generated for you `api`, `cadence`, and `web`.

- `api`: A place to put your files implementing the backend functionality.
- `cadence`: Cadence contracts, transactions and scripts should be located in this folder.
- `web`: Save your frontend files in this folder. You can use [fcl.js](https://github.com/onflow/fcl-js) to implement a frontend.

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
