# Flow CLI

The Flow CLI is a command-line interface that provides useful utilities for building Flow applications.

## Installation

The Flow CLI can be installed in one of three ways:

### macOS

#### Homebrew

```sh
brew install flow-cli
```

#### From a pre-built binary

_This installation method only works on x86-64._

This script downloads and installs the appropriate binary for your system:

```sh
sh -ci "$(curl -fsSL https://storage.googleapis.com/flow-cli/install.sh)"
```

To update, simply re-run the installation command above.

### Linux

#### From a pre-built binary

_This installation method only works on x86-64._

This script downloads and installs the appropriate binary for your system:

```sh
sh -ci "$(curl -fsSL https://storage.googleapis.com/flow-cli/install.sh)"
```

To update, simply re-run the installation command above.

### Windows

#### From a pre-built binary

_This installation method only works on Windows 10, 8.1, or 7 (SP1, with [PowerShell 3.0](https://www.microsoft.com/en-ca/download/details.aspx?id=34595)), on x86-64._

1. Open PowerShell ([Instructions](https://docs.microsoft.com/en-us/powershell/scripting/install/installing-windows-powershell?view=powershell-7#finding-powershell-in-windows-10-81-80-and-7))
2. In PowerShell, run:

    ```powershell
    iex "& { $(irm 'https://storage.googleapis.com/flow-cli/install.ps1') }"
    ```

To update, simply re-run the installation command above.


## Development 

### Releasing

- Tag a new release and push it
- Build the binaries: `make versioned-binaries`
- Upload the binaries: `make publish`
- Update the Homebrew formula: e.g. `brew bump-formula-pr flow-cli --version=0.12.3`

To make the new version the default version that is installed 

- Change `version.txt` and commit it
- Upload the version file: `gsutil cp version.txt gs://flow-cli`
