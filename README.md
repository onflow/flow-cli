# Flow CLI

The Flow CLI is a command-line interface that provides useful utilities for building Flow applications.

## Installation

// TODO: Link

## Development 

### Releasing

- Tag a new release and push it
- Build the binaries: `make versioned-binaries`
- Upload the binaries: `make publish`
- Update the Homebrew formula: e.g. `brew bump-formula-pr flow-cli --version=0.12.3`

To make the new version the default version that is installed 

- Change `version.txt` and commit it
- Upload the version file: `gsutil cp version.txt gs://flow-cli`
