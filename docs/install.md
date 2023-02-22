---
title: Install the Flow CLI
description: How to install the Flow command-line interface (CLI)
---

The Flow CLI can be installed on macOS, Windows (7 or greater) and most Linux systems.

## macOS

### Homebrew

```sh
brew install flow-cli
```

### From a pre-built binary

_This installation method only works on x86-64._

This script downloads and installs the appropriate binary for your system:

```sh
sh -ci "$(curl -fsSL https://raw.githubusercontent.com/onflow/flow-cli/master/install.sh)"
```

To update, simply re-run the installation command above.

It is currently not possible to install earlier versions of the Flow CLI with Homebrew.
## Linux

### From a pre-built binary

_This installation method only works on x86-64._

This script downloads and installs the appropriate binary for your system:

```sh
sh -ci "$(curl -fsSL https://raw.githubusercontent.com/onflow/flow-cli/master/install.sh)"
```

To update, simply re-run the installation command above.

### Install a specific version

To install a specific version of Flow CLI newer than v0.42.0, append the version tag to the command (e.g. the command below installs CLI version v0.44.0).

```sh
sh -ci "$(curl -fsSL https://raw.githubusercontent.com/onflow/flow-cli/master/install.sh)" -- v0.44.0
```

To install a version older than v0.42.0, refer to [Installing versions before 0.42.0](https://developers.flow.com/tools/flow-cli/install#installing-versions-before-0420) below.

## Windows

### From a pre-built binary

_This installation method only works on Windows 10, 8.1, or 7 (SP1, with [PowerShell 3.0](https://www.microsoft.com/en-ca/download/details.aspx?id=34595)), on x86-64._

1. Open PowerShell ([Instructions](https://docs.microsoft.com/en-us/powershell/scripting/install/installing-windows-powershell?view=powershell-7#finding-powershell-in-windows-10-81-80-and-7))
2. In PowerShell, run:

    ```powershell
    iex "& { $(irm 'https://raw.githubusercontent.com/onflow/flow-cli/master/install.ps1') }"
    ```

To update, simply re-run the installation command above.

# Upgrade the Flow CLI

## macOS

### Homebrew

```sh
brew upgrade flow-cli
```

### From a pre-built binary

_This update method only works on x86-64._

This script downloads and updates the appropriate binary for your system:

```sh
sh -ci "$(curl -fsSL https://raw.githubusercontent.com/onflow/flow-cli/master/install.sh)"
```

## Linux

### From a pre-built binary

_This update method only works on x86-64._

This script downloads and updates the appropriate binary for your system:

```sh
sh -ci "$(curl -fsSL https://raw.githubusercontent.com/onflow/flow-cli/master/install.sh)"
```

## Windows

### From a pre-built binary

_This update method only works on Windows 10, 8.1, or 7 (SP1, with [PowerShell 3.0](https://www.microsoft.com/en-ca/download/details.aspx?id=34595)), on x86-64._

1. Open PowerShell ([Instructions](https://docs.microsoft.com/en-us/powershell/scripting/install/installing-windows-powershell?view=powershell-7#finding-powershell-in-windows-10-81-80-and-7))
2. In PowerShell, run:

    ```powershell
    iex "& { $(irm 'https://raw.githubusercontent.com/onflow/flow-cli/master/install.ps1') }"
    ```

# Uninstalling Flow CLI
To remove the flow CLI you can run the following command if it was previously installed using a pre-built binary. 

- macOS: `rm /usr/local/bin/flow`
- Linux: `rm ~/.local/bin/flow`
- Windows: `rm ~/Users/{user}/AppData/Flow/flow.exe`

If you installed it using Hombrew you can remove it using: `brew uninstall flow-cli`.

## Installing versions before 0.42.0
If you want to install versions before v0.42.0 you have to use a different install command. 

**Linux/macOS**
```
https://raw.githubusercontent.com/onflow/flow-cli/v0.41.3/install.ps1

sh -ci "$(curl -fsSL https://raw.githubusercontent.com/onflow/flow-cli/v0.41.3/install.sh)" -- v0.41.2
```

**Windows**
```
iex "& { $(irm 'https://raw.githubusercontent.com/onflow/flow-cli/master/install.ps1') }"
```
