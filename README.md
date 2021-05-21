![](./logo-cli.svg)

# Flow CLI

The Flow CLI is a command-line interface that provides useful utilities for building Flow applications.

## Installation

To install the Flow CLI, follow the [installation instructions](https://docs.onflow.org/flow-cli/install) on the Flow documentation website.

## Documentation

You can find the CLI documentation on the [Flow documentation website](https://docs.onflow.org/flow-cli).

## Development 

### Releasing

- Tag a new release and push it
- Build the binaries: `make versioned-binaries`
- Test built binaries locally
- Upload the binaries: `make publish`
- Update the Homebrew formula: e.g. `brew bump-formula-pr flow-cli --version=0.12.3`

To make the new version the default version that is installed 

- Change `version.txt` and commit it
- Upload the version file: `gsutil cp version.txt gs://flow-cli`
