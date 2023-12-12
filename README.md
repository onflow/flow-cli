<br />
<p align="center">
  <a href="https://docs.onflow.org/flow-cli/install/">
    <img src="./cli-banner.svg" alt="Logo" width="270" height="auto">
  </a>

  <p align="center">
    <i>Flow CLI brings Flow to your terminal. Easily interact with the network and build your dapps.</i>
    <br />
    <a href="https://developers.flow.com/tooling/flow-cli/install"><strong>Read the docs»</strong></a>
    <br />
    <br />
    <a href="https://github.com/onflow/flow-cli/issues">Report Bug</a>
    ·
    <a href="https://github.com/onflow/flow-cli/blob/master/CONTRIBUTING.md">Contribute</a>
    ·
    <a href="https://github.com/onflow/flow-cli/blob/master/CONTRIBUTING.md#cli-guidelines">Read Guidelines</a>
  </p>
</p>
<br />
<br />

## Installation

To install the Flow CLI, follow the [installation instructions](https://developers.flow.com/tools/flow-cli/install) on the Flow documentation website.

## Documentation

You can find the CLI documentation on the [CLI documentation website](https://developers.flow.com/tools/flow-cli).

## Features
The Flow CLI is a command line tool that allows you to interact with the Flow blockchain. 
Read about supported commands in the [CLI documentation website](https://developers.flow.com/tools/flow-cli).

```
Usage:
  flow [command]

👋 Welcome Flow developer!
   If you are starting a new flow project try running 'flow setup <project_name>'. 

🔥 Super Commands
  dev          Build your Flow project
  flix         Execute FLIX template with a given id, name, or local filename
  generate     Generate new boilerplate files
  setup        Start a new Flow project

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
  flowser      Run Flowser project explorer
  test         Run Cadence tests

🏄 Flow Project
  deploy       Deploy all project contracts
  init         Initialize a new configuration
  project      Manage your Cadence project
  run          Start emulator and deploy all project contracts

🔒 Flow Security
  keys         Generate and decode Flow keys
  signatures   Signature verification and creation

```

The Flow CLI includes several commands to interact with Flow networks, such as querying account information, or sending transactions. It also includes the [Flow Emulator](https://developers.flow.com/tools/emulator).


![Alt Text](./cli.gif)

## Contributing 

Read [contributing](./CONTRIBUTING.md) document.


## Running Flow EVM

⚠️ **This feature is experimental in nature, and its stability cannot be guaranteed. It does not reflect the intended direction for utilizing the EVM in a production environment.** ⚠️

In order to run Flow EVM you have to start the emulator with the evm flag enabled:
```
flow emulator --evm-enabled
```

You can then proceed by creating an account using the:

```
flow evm create-account {funding-amount}
```

This will create an account inside the EVM, fund it with the funding amount provided and create a bridged 
account resource which will be saved inside the Flow account that was used as a signer in the above command
(default signer is emulator account if not provided with `--signer` flag). Please be aware you can't create 
multiple accounts on a single Flow account right now as the bridged account resource is always stored 
in same place and can't be overwritten.

After creating an account you can deploy a contract to the EVM by running:

```
flow evm deploy {compiled binary}
```

You need to provide location of the file that contains the compiled EVM binary. The response will include the 
EVM address of the contract that you will later need to interact with it.

You can interact with deploy contract by calling functions using:

```
flow evm run {Flow caller address} {EVM contract address} {contract function name} --ABI {abi location}
```

The Flow caller address is the account that contains EVM bridged account resource, that will be used to 
execute the call inside the EVM. So you must first create such a resource by using the `create-account` command. 
The EVM contract address will be provided to you when `deploy` command is executed. You also need to specify the 
name of the contract function you want to call and the file containing ABI specification, which is normally produced 
in the compile process. 
