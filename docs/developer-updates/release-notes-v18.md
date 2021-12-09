## ⬆️ Install or Upgrade

Follow the [Flow CLI installation guide](../install.md) for instructions on how to install or upgrade the CLI.

## ⭐ Features

### Resolve Contract Imports in Scripts and Transactions
This is a new feature that allows you to send transactions and scripts that reference 
contracts deployed using the `project deploy` command. Imports are resolved 
by matching contract source paths to their declarations in `flow.json`.

For example:

Example script: `get_supply.cdc`
```
import Kibble from "../../contracts/Kibble.cdc"

pub fun main(): UFix64 {
    let supply = Kibble.totalSupply
    return supply
}

```

Example command:
```
flow scripts execute ./get_supply.cdc
```

Note: Please make sure you first deploy the contract using `flow project deploy` 
command and that contracts are correctly added to the flow configuration.


### Build, Sign and Send Transactions
New functionality allows you to build a transaction, sign it 
and send signed to the network in separated steps. 

#### Build Transaction
Build new transaction and specify who will be the proposer, signer and payer account 
or address. Example:

```
flow transactions build ./transaction.cdc --proposer alice --payer bob --authorizer bob --filter payload --save payload1.rlp
```

Check more about [this functionality in docs](../build-transactions.md).

#### Sign Transaction
After using build command and saving payload to a file you should sign the transaction 
with each account. Example:

```
flow transactions sign ./payload1.rlp --signer alice --filter payload --save payload2.rlp 
```

#### Send Signed Transaction
When authorizer, payer and proposer sign the transaction it is ready to be 
sent to the network. Anyone can execute the `send-signed` command. Example:

```
flow transactions send-signed ./payload3.rlp
```

### Version Check
Automatically checks if a new version exists and outputs a warning in case there 
is a newer version. Example:
```
⚠️  Version Warning: New version v0.18.0 of Flow CLI is available.
Follow the Flow [CLI installation guide](../install.md) for instructions on how to install or upgrade the CLI
```


### Create Account With Multiple Keys and Weights
Account creation can be done using multiple keys (`--key`) and new `--key-weight` 
flag. Flag enables you to set key weight for each of the keys. Command example: 
```
accounts create \
    --key ca8cc7...76f67 --key-weight 500 \
    --key da8123...043ce --key-weight 500

Address	 0x179b6b1cb6755e31
Balance	 0.10000000
Keys	 2

Key 0	Public Key		 ca8cc7...76f67
	Weight			 500
	Signature Algorithm	 ECDSA_P256
	Hash Algorithm		 SHA3_256
	Revoked 		 false
	Sequence Number 	 0
	Index 			 0


Key 1	Public Key		 da8123...043ce
	Weight			 500
	Signature Algorithm	 ECDSA_P256
	Hash Algorithm		 SHA3_256
	Revoked 		 false
	Sequence Number 	 0
	Index 			 1

```

## 🎉 Improvements

### Account Response Improved
Account response includes two new fields in key section: `Sequence Number` and `Index`.

### Transaction Result Improved
Transaction result displays more information about the transaction. New format example:

```
Status		✅ SEALED
ID		b6430b35ba23849a8acb4fa1a4a1d5cce3ed4589111ecbb3984de1b6bd1ba39e
Payer		a2c4941b5f3c7151
Authorizers	[a2c4941b5f3c7151]

Proposal Key:	
    Address	a2c4941b5f3c7151
    Index	0
    Sequence	9

No Payload Signatures

Envelope Signature 0:
    Address	a2c4941b5f3c7151
    Signature	5391a6fed0fe...2742048166f9d5c925a8dcb78a6d8c710921d67
    Key Index	0


Events:	 None


Arguments (1):
    - Argument 0: {"type":"String","value":"Meow"}


Code

transaction(greeting: String) {
  let guest: Address

  prepare(authorizer: AuthAccount) {
    self.guest = authorizer.address
  }

  execute {
    log(greeting.concat(",").concat(self.guest.toString()))
  }
}


Payload:
f90184f90138...8a9462751237da2742048166f9d5c925a8dcb78a6d8c710921d67
```

Transaction error is now shown at the top of the result.
```
Transaction 0dd6294a7614bc0fbeb39b44a6e9f68e821225caa4baf4104a17dc1193d4f011 sealed

Status: SEALED
Execution Error: Execution failed:
error: invalid move operation for non-resource
  --> 0dd6294a7614bc0fbeb39b44a6e9f68e821225caa4baf4104a17dc1193d4f011:15:15
   |
15 |         useRes(<-result)
   |                ^^^ unexpected `<-`

error: mismatched types
  --> 0dd6294a7614bc0fbeb39b44a6e9f68e821225caa4baf4104a17dc1193d4f011:15:15
   |
15 |         useRes(<-result)
   |                ^^^^^^^^ expected `AnyResource`, got `&AnyResource`

Events:
  None
```

## 🐞 Bug Fixes

### New Transaction ID Log
While sending transaction was in progress output displayed wrong transaction ID.

### Init Reset Fix
Old configuration format caused an error saying to reset the 
configuration using reset flag, but when ran it produced the same error again. 
This bug was fixed.

### Emulator Config Path
When running emulator command `flow emulator` config flag `-f` was ignored. 
This has been fixed, so you can provide a custom path to the config while running 
the start emulator command.
