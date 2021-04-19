## ⬆️ Install or Upgrade

Follow the [Flow CLI installation guide](https://docs.onflow.org/flow-cli/install/) for instructions on how to install or upgrade the CLI.

## 💥 Breaking Changes

### Updated: 
### Removed:

## ⚠️ Deprecation Warnings

## ⭐ Features

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

### 