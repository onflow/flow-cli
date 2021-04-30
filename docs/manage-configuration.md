---
title: Manage Flow Configuration
sidebar_title: Manage Configuration
description: How to configure the Flow CLI
---

Configuration should be managed by using `config add` 
and `config remove` commands. Using add command will also 
validate values that will be added to the configuration.

```shell
flow config add <account|contract|network|deployment>
flow config remove <account|contract|network|deployment>
```

## Example Usage

```shell
flow config add account

Name: Admin
Address: f8d6e0586b0a20c7
✔ ECDSA_P256
✔ SHA3_256
Private key: e382a0e494...9285809356
Key index (Default: 0): 0
```

### Configuration

- Flag: `--config-path`
- Short Flag: `-f`
- Valid inputs: valid filename

Specify a filename for the configuration files, you can provide multiple configuration
files by using `-f` flag multiple times.





