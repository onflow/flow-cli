# flow-cli — The Flow Command-Line Interface

[![License](https://img.shields.io/github/license/onflow/flow-cli)](https://github.com/onflow/flow-cli/blob/master/LICENSE)
[![Release](https://img.shields.io/github/v/release/onflow/flow-cli)](https://github.com/onflow/flow-cli/releases)
[![Discord](https://img.shields.io/discord/613813861610684416?label=Discord&logo=discord)](https://discord.gg/flow)
[![Built on Flow](https://img.shields.io/badge/Built%20on-Flow-00EF8B)](https://flow.com)
[![Go Reference](https://pkg.go.dev/badge/github.com/onflow/flow-cli.svg)](https://pkg.go.dev/github.com/onflow/flow-cli)

<p align="center">
  <a href="https://developers.flow.com/tools/flow-cli/install">
    <img src="./cli-banner.svg" alt="Logo" width="270" height="auto">
  </a>

  <p align="center">
    <i>Flow CLI brings Flow to your terminal. Easily interact with the network and build your dapps.</i>
    <br />
    <a href="https://developers.flow.com/tools/flow-cli/install"><strong>Read the docs»</strong></a>
    <br />
    <br />
    <a href="https://github.com/onflow/flow-cli/issues">Report Bug</a>
    ·
    <a href="https://github.com/onflow/flow-cli/blob/master/CONTRIBUTING.md">Contribute</a>
    ·
    <a href="https://github.com/onflow/flow-cli/blob/master/CONTRIBUTING.md#cli-guidelines">Read Guidelines</a>
  </p>
</p>

## TL;DR

- **What:** Official command-line interface for the Flow network. Deploy contracts, run transactions and scripts, manage accounts and keys, and run a local emulator from your terminal.
- **Who it's for:** Developers building dapps, smart contracts, or tooling on Flow.
- **Why use it:** A single tool covering project scaffolding, local development, contract deployment, and network interaction across emulator, testnet, and mainnet.
- **Status:** see [Releases](https://github.com/onflow/flow-cli/releases) for the latest version.
- **License:** Apache-2.0.
- **Related repos:** [onflow/cadence](https://github.com/onflow/cadence) · [onflow/flow-go](https://github.com/onflow/flow-go) · [onflow/fcl-js](https://github.com/onflow/fcl-js)
- The reference command-line interface for the Flow network, open-sourced since 2019.

## Installation

To install the Flow CLI, follow the [installation instructions](https://developers.flow.com/tools/flow-cli/install) on the Flow documentation website.

## Documentation

You can find the CLI documentation on the [CLI documentation website](https://developers.flow.com/tools/flow-cli).

## Features
The Flow CLI is a command line tool that allows you to interact with the Flow network.
Read about supported commands in the [CLI documentation website](https://developers.flow.com/tools/flow-cli).

```
Usage:
  flow [command]

👋 Welcome Flow developer!
   If you are starting a new flow project use our super commands, start by running 'flow init'. 

🔥 Super Commands
  generate     Generate template files for common Cadence code
  init         Start a new Flow project

📦 Flow Entities
  accounts     Create and retrieve accounts and deploy contracts
  blocks       Retrieve blocks
  collections  Retrieve collections
  events       Retrieve events

💬 Flow Interactions
  scripts      Execute Cadence scripts
  transactions Build, sign, send and retrieve transactions

🔨 Flow Tools
  cadence      Execute Cadence code
  dev-wallet   Run a development wallet
  emulator     Run Flow network for development
  flix         execute, generate, package
  flowser      Run Flowser project explorer
  test         Run Cadence tests

🏄 Flow Project
  deploy       Deploy all project contracts
  project      Manage your Cadence project
  run          Start emulator and deploy all project contracts

🔒 Flow Security
  keys         Generate and decode Flow keys
  signatures   Signature verification and creation

🔗 Dependency Manager
  dependencies Manage contracts and dependencies
```

The Flow CLI includes several commands to interact with Flow networks, such as querying account information, or sending transactions. It also includes the [Flow Emulator](https://developers.flow.com/tools/emulator).


![Alt Text](./cli.gif)

## Contributing 

Read [contributing](./CONTRIBUTING.md) document.

## FAQ

**What is the Flow CLI?**
The Flow CLI is a command-line tool for interacting with the Flow network. It supports deploying contracts, sending transactions, running scripts, managing accounts and keys, and running a local emulator.

**How do I install the Flow CLI?**
Follow the installation instructions at [developers.flow.com/tools/flow-cli/install](https://developers.flow.com/tools/flow-cli/install). Installers are provided for macOS, Linux, and Windows.

**How do I start a new Flow project?**
Run `flow init` to scaffold a new project. This uses the super commands described in the Features section above.

**Does the Flow CLI include a local emulator?**
Yes. The CLI bundles the [Flow Emulator](https://developers.flow.com/tools/emulator) so you can develop and test locally before deploying to testnet or mainnet.

**Where can I find the reference for `flow.json`?**
See [docs/configuration.md](./docs/configuration.md) in this repo and the [configuration reference](https://developers.flow.com/tools/flow-cli) on the docs site.

**How do I report a bug or request a feature?**
Open an issue at [github.com/onflow/flow-cli/issues](https://github.com/onflow/flow-cli/issues). For security issues, follow [SECURITY.md](./SECURITY.md).

**What license is the Flow CLI released under?**
Apache-2.0. See [LICENSE](./LICENSE).

## About Flow

This repo is part of the [Flow network](https://flow.com), a Layer 1 blockchain built for consumer applications, AI Agents, and DeFi at scale.

- Developer docs: https://developers.flow.com
- Cadence language: https://cadence-lang.org
- Community: [Flow Discord](https://discord.gg/flow) · [Flow Forum](https://forum.flow.com)
- Governance: [Flow Improvement Proposals](https://github.com/onflow/flips)
