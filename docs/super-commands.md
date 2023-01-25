---
title: Using Super Commands with the Flow CLI
sidebar_title: Super Commands
description: How Flow super commands work
---

Flow CLI Super commands are set of commands that can be used during development 
of your dApp to greatly simplify the workflow. The result is you can focus on writing the contracts 
and the commands will take care of the rest. 


## Setup
The initial command to start your new Flow project is `flow setup` . By default, it will create 
basic folder structure and a flow.json configuration. 

Running the command:
```
> flow setup
```

Will create the following folders and files:
- `/contracts` folder should contain all your Cadence contracts,
- `/scripts` folder should contain all your Cadence scripts,
- `/transactions` folder should contain all your Cadence transactions,
- `/tests` folder should contain all your Cadence tests,
- `flow.json` is a configuration file for your project, which will be automatically maintained.

### Using Scaffolds
Based on the purpose of your project you can select from a list of available scaffolds. 
You can access the scaffolds by simply using the `--scaffold` flag like so:
```
> flow setup --scaffold
```

The list of scaffolds will continuously grow, and you are welcome to contribute to that. 
You can contribute by creating your own scaffold repository which can then be added to the scaffold 
list by [following instructions here](https://github.com/onflow/flow-cli/blob/master/CONTRIBUTING.md#adding-a-scaffold).


## Develop
After creating the project using the flow setup command you can start the emulator in 
the project directory by running `flow emulator`. After emulator is started up you can continue by 
running the flow develop command like so:
```
> flow dev
```
This will continuously watch for your projects Cadence files for changes and keep them in sync with the deployed contracts on the emulator. 

The output will look something like:
```
[15:53:38] Syncing all the contracts...

😜 charlie
  |- MyContract contracts/charlie/MyContract.cdc

😏 emulator-account
  |- HelloWorld contracts/HelloWorld.cdc
```

After the command is started it will automatically watch any changes you make to 
Cadence files and make sure to continuously sync those changes on the emulator network. 
If you make any mistakes it will report the errors as well. 

It is recommended that you use VSCode as the IDE and run the command in the terminal window of the IDE.
The latest VSCode extension also supports resolution of the improved import syntax, more on that later.


⚠️ Please note that this command only works on the emulator network. It's meant for development only and hence 
it doesn't allow interacting with testnet or mainnet network. After your project is completed you will 
be soon able to migrate it using a migration super command. Also, please note the command requires a 
running emulator which you have to start. If you restart the emulator the command needs to be restarted as well. 
This command is meant to be used during development, and it updates the contracts by removing and redeploying 
them, which means that if you manually interacted with those contracts and stored resources in accounts 
storage that stored items might no longer be valid after contract is updated. Our advise is to first focus 
on development and use automated tests to assert correct functionality and interact with contracts manually 
after this cycle is complete. Also note that this is still a very experimental feature, so it might 
undergo a lot of changes and improvements as we learn from the usage. 

**Deploying Contracts**

When adding the contracts to the `/contracts` folder it will automatically deploy them to the `default` 
account, which is also created for you at startup of running `flow dev`. 

If you want to add the contracts to a separate account all you have to do is create a folder inside the 
`/contracts` folder and add the contract there, that will first automatically create the account with 
the same name as the folder name and then deploy all the contracts inside 
that folder to that newly created account.

Example:
If I want to have a contract named `A.cdc` deployed to a default account and a contract named `B.cdc` 
deployed to account called `Bob` my folder structure inside contracts folder will look like:
```
/contracts
    A.cdc
    bob/
        B.cdc
```

## Import Schema
In order to simplify your workflow we have introduced a new way to import your contracts during development. 
You can simply import them by their name.

The new import schema format looks like:
```
import "{name of the contract}"
```
Example:
```
import "HelloWorld"
```
This will automatically import the contract you have created in your project with the same name and 
save the configuration in flow.json. It doesn't matter if the contract is deployed on a non-default account.





