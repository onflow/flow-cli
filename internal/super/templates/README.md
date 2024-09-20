## 👋 Welcome Flow Developer!
Welcome to your new {{ .ProjectType }} project. This project is a starting point for you to develop smart contracts on the Flow blockchain. It comes with a few example contracts and scripts to help you get started.

## 🔨 Getting started
Getting started can feel overwhelming, but we are here for you. Depending on how accustomed you are to Flow here's a list of resources you might find useful:
- **[Cadence documentation](https://developers.flow.com/cadence/language)**: here you will find language reference for Cadence, which will be the language in which you develop your smart contracts,
- **[Visual Studio Code](https://code.visualstudio.com/?wt.mc_id=DX_841432)** and **[Cadence extension](https://marketplace.visualstudio.com/items?itemName=onflow.cadence)**: we suggest using Visual Studio Code IDE for writing Cadence with the Cadence extension installed, that will give you nice syntax highlitning and additional smart features,
- **[SDKs](https://developers.flow.com/tools#sdks)**: here you will find a list of SDKs you can use to ease the interaction with Flow network (sending transactions, fetching accounts etc),
- **[Tools](https://developers.flow.com/tools#development-tools)**: development tools you can use to make your development easier, [Flowser](https://docs.flowser.dev/) can be super handy to see what's going on the blockchain while you develop
{{ if .IsNFTProject }}

NFT Resources:
- **[flow-nft](https://github.com/onflow/flow-nft)**: home of the Flow NFT standards, contains utility contracts, examples, and documentation
- **[nft-storefront](https://github.com/onflow/nft-storefront/)**: NFT Storefront is an open marketplace contract used by most Flow NFT marketplaces
{{ end }}
{{ if .IsFungibleProject }}

Fungible Token Resources:
- **[flow-ft](https://github.com/onflow/flow-ft)**: home of the Flow Fungible Token standards, contains utility contracts, examples, and documentation
{{ end }}


## 📦 Project Structure
Your project comes with some standard folders which have a special purpose:
- `flow.json` configuration file for your project, you can think of it as package.json.  It has been initialized with a basic configuration{{ if len .Dependecies }} and some Core Contract dependencies{{ end }}.
- `/cadence` inside here is where your Cadence smart contracts code lives

Your project is set up with the following dependencies:
{{ range .Dependecies }}
  - `{{ .Name }}`
{{ end }}

Inside `cadence` folder you will find:
- `/contracts` location for Cadence contracts go in this folder
{{ range .Contracts }}
  - `{{ .Name }}.cdc` is the file for the `{{ .Name }}` contract
{{ end }}
- `/scripts` location for Cadence scripts goes here
{{ range .Scripts }}
  - `{{ .Name }}.cdc` is the file for the `{{ .Name }}` script
{{ end }}
- `/transactions` location for Cadence transactions goes in this folder
{{ range .Transactions }}
  - `{{ .Name }}.cdc` is the file for the `{{ .Name }}` transaction
{{ end }}
- `/tests` location for Cadence tests goes in this folder
{{ range .Tests }}
  - `{{ .Name }}.cdc` is the file for the `{{ .Name }}` test
{{ end }}

## 👨‍💻 Start Developing

### Creating a new contract
To add a new contract to your project you can use the following command:

```shell
> flow generate contract
```

This command will create a new contract file and add it to the `flow.json` configuration file.

### Creating a new script

To add a new script to your project you can use the following command:

```shell
> flow generate script
```

This command will create a new script file and add it to the `flow.json` configuration file.

You can import any of your own contracts or installed dependencies in your script file using the `import` keyword.  For example:

```cadence
import Counter from "Counter"
```

### Creating a new transaction

To add a new transaction to your project you can use the following command:

```shell
> flow generate transaction
```

This command will create a new transaction file and add it to the `flow.json` configuration file.  You can import any dependencies as you would in a script file.

### Installing external dependencies

If you want to use external contract dependencies (like NonFungibleToken, FlowToken, FungibleToken,..) you can install them using Cadence dependency manager: https://developers.flow.com/tools/flow-cli/dependency-manager

Use [ContractBrowser](https://contractbrowser.com/) to explore available 3rd party contracts in the Flow ecosystem.

## 🧪 Testing
To verify that your project is working as expected you can run the tests using the following command:
```shell
> flow test
```

This command will run all the tests in the `cadence/tests` folder. You can add more tests to this folder as you develop your project.

To learn more about testing in Cadence, check out the [Cadence testing documentation](https://cadence-lang.org/docs/testing-framework).

## 🚀 Deploying

### Deploying to the Flow Emulator

To deploy your project to the Flow Emulator you can use the following command:
```shell
> flow emulator --start
```

This command will start the Flow Emulator and deploy your project to it. You can then interact with your project using the Flow Emulator.

### Deploying to the Flow Testnet

To deploy your project to the Flow Testnet you can use the following command:
```shell
> flow project deploy --network=testnet
```

This command will deploy your project to the Flow Testnet. You can then interact with your project using the Flow Testnet.

### Deploying to the Flow Mainnet

To deploy your project to the Flow Mainnet you can use the following command:
```shell
> flow project deploy --network=mainnet
```

This command will deploy your project to the Flow Mainnet. You can then interact with your project using the Flow Mainnet.

## Further Reading

- Cadence Language Reference https://cadence-lang.org/docs/language
- Flow Smart Contract Project Development Standards https://cadence-lang.org/docs/project-development-tips
- Cadence anti-patterns https://cadence-lang.org/docs/anti-patterns