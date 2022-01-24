## ⬆️ Install or Upgrade

Follow the [Flow CLI installation guide](../install.md) for instructions on how to install or upgrade the CLI.

## 💥 Breaking Changes

### Configuration Format

The default configuration format (i.e. the contents of `flow.json`) has been updated.
It is now unified to work with all CLI commands. 
The new format is not backwards compatible with the old format.

If needed, you can generate a new configuration file with the `flow init` command.

Read more about the new configuration format in the [documentation](../configuration.md).

### Updated: `flow blocks get`

The `--latest`, `--id` and `--height` have been removed.

Instead, use the new argument syntax:

```sh
# get latest block
flow blocks get latest

# get a block by ID
flow blocks get 6bb0e0fceef9225a3cf9ceb6df9a31bd0063e6ee8e8dd7fdd93b831783243cd3

# get a block by height
flow blocks get 28329914
```

Read more about this change in the [documentation](../get-blocks.md).

### Removed: `flow keys decode`

The `flow keys decode` command has been temporarily removed due to a bug that requires further investigation.

### Removed: `flow keys save`

The `flow keys save` command has been removed in favour of an upcoming `flow accounts add` command. 

## ⚠️ Deprecation Warnings

The following functionality has been deprecated and will be removed in an upcoming release.

**`flow accounts create`, `flow accounts add-contract`, `flow accounts remove-contract`, `flow accounts update-contract`**

- Flag `--results` is deprecated, results are displayed by default.

**`flow accounts get`**

- Flag `--code` is deprecated, use `--contracts` flag instead.

**`flow events get`**

- Flag `--verbose` is deprecated.

**`flow keys generate`**

- Flag `--algo` is deprecated, use flag `--sig-algo`.

**`flow transactions send`**

- Flag `--code` is deprecated, use filename argument instead.
- Flag `--args` is deprecated, use `--arg` or `--args-json` instead.
- Flag `--results` is deprecated, results are displayed by default.

**`flow scripts execute`**

- Flag `--code` is deprecated, use filename argument instead.
- Flag `--args` is deprecated, use `--arg` or `--args-json` instead.

**`flow transactions status`**

- This command has been deprecated in favour of `flow transactions get`.

**`flow project init`**

- This command has been deprecated in favour of `flow init`.

**`flow project start-emulator`**

- This command has been deprecated in favour of `flow emulator`.

**`flow emulator start`**

- This command has been deprecated in favour of `flow emulator`.

## ⭐ Features

### Output

Output format was changed, so it stays consistent between commands. New flags were introduced 
that control the output. Let's take a quick look at the new flags, but make sure to read 
more about them in the documentation on each command:

- Output: `--output` specify the format of the command results (JSON, inline...),
- Save: `--save` specify the filename where you want the result to be saved,
- Log: `--log` control how much output you want to see during command execution,
- Filter: `--filter` Specify any property name from the result you want to return as the only value.

All the flags and their allowed values are specified 
for each command in the [documentation](../index.md).

Changed output for fetching account.
```
Address  179b6b1cb6755e31
Balance  0
Keys     2

Key 0   Public Key               c8a2a318b9099cc6...a0fe320dba7
        Weight                   1000
        Signature Algorithm      ECDSA_P256
        Hash Algorithm           SHA3_256

Code             
         pub contract Foo {
                pub var bar: String
         
                init() {
                        self.bar = "Hello, World!"
                }
         }
```

Output account result as JSON.

```
{"address":"179b6b1cb6755e31","balance":0,"code":"CnB1YiBj...SIKCX0KfQo=","keys":[{"index":0,"publicKey":{},"sigAlgo":2,"hashAlgo":3,"weight":1000,"sequenceNumber":0,"revoked":false}],"Contracts":null}
```

Improved progress feedback with loaders.
```
Loading 0x1fd892083b3e2a4c...⠼
```

### Shared Library

You can import Flow CLI shared library from the `flowcli` package and use the functionality 
from the service layer in your own software. Codebase was divided into two components, first 
is the CLI interaction layer, and the second is the shared library component which is meant 
to be reused.

### Account Staking Info Command

New command to fetch staking info from the account was added. Read more about it in the
[documentation](../account-staking-info.md).

```shell
> accounts staking-info 535b975637fb6bee --host access.testnet.nodes.onflow.org:9000

Account Staking Info:
        ID: 			 "ca00101101010100001011010101010101010101010101011010101010101010"
        Initial Weight: 	 100
        Networking Address: 	 "ca00101101010100001011010101010101010101010101011010101010101010"
        Networking Key: 	 "ca00101101010100001011010101010101010101010101011010101010101010ca00101101010100001011010101010101010101010101011010101010101010"
        Role: 			 1
        Staking Key: 		 "ca00101101010100001011010101010101010101010101011010101010101010ca00101101010100001011010101010101010101010101011010101010101010ca00101101010100001011010101010101010101010101011010101010101010"
        Tokens Committed: 	 0.00000000
        Tokens To Unstake: 	 0.00000000
        Tokens Rewarded: 	 82627.77000000
        Tokens Staked: 		 250000.00000000
        Tokens Unstaked: 	 0.00000000
        Tokens Unstaking: 	 0.00000000
        Node Total Stake (including delegators):    250000.00000000


Account Delegation Info:
        ID: 			 7
        Tokens Committed: 	 0.00000000
        Tokens To Unstake: 	 0.00000000
        Tokens Rewarded: 	 30397.81936000
        Tokens Staked: 		 100000.00000000
        Tokens Unstaked: 	 0.00000000
        Tokens Unstaking: 	 0.00000000

```

## 🐞 Bug Fixes

### Address 0x prefix

Addresses are not required to be prefixed with `0x` anymore. You can use either format, but 
due to consistency we advise using `0x` prefix with addresses represented in `hex` format.

### Project deploy error

Deploying contract provides improved error handling in case something goes wrong you 
can now read what the error was right from the commandline. 

Example of error output:
```
Deploying 2 contracts for accounts: emulator-account

❌  contract Kibble is already deployed to this account. Use the --update flag to force update
❌  contract KittyItemsMarket is already deployed to this account. Use the --update flag to force update
❌  failed to deploy contracts

❌ Command Error: failed to deploy contracts
```
