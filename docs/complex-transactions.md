---
title: Send Complex Transaction with the Flow CLI
sidebar_title: Build a Complex Transaction
description: How to build and send a complex Flow transaction from the command line
---

**Simple Transactions**

Sending a transaction using the Flow CLI can simply be 
achieved by using the [send command documented here](/flow-cli/send-transactions/).

**Complex Transactions**

If you would like to build more complex transactions the Flow CLI provides 
commands to build, sign and send transactions allowing you to specify different 
authorizers, signers and proposers.  

The process of sending a complex transactions includes three steps:
1. [build a transaction](/flow-cli/build-transactions/)
2. [sign the built transaction](/flow-cli/sign-transaction/)
3. [send signed transaction](/flow-cli/send-signed-transactions/)

Read more about each command flags and arguments in the above links.

## Examples
We will describe common examples for complex transactions. All examples are using an [example configuration](#configuration).

### Single payer, proposer and authorizer
The simplest Flow transaction declares a single account as the proposer, payer and authorizer.

Build the transaction:
```shell
> flow transactions build tx.cdc 
  --proposer alice 
  --payer alice 
  --authorizer alice 
  --filter payload --save tx1
```
Sign the transaction:
```shell
> flow transactions sign tx1 --signer alice 
  --filter payload --save tx2
```
Submit the signed transaction:
```shell
> flow transactions send-signed tx2
```
Transaction content (`tx.cdc`):
```
transaction {
    preapre(signer: AuthAccount) {}
    execute { ... }
}
```

### Single payer and proposer, multiple authorizers
A transaction that declares same payer and proposer but multiple authorizers each required to sign the transaction. Please note that the order of signing is important, and [the payer must sign last](https://docs.onflow.org/concepts/transaction-signing/#payer-signs-last).

Build the transaction:
```shell
> flow transactions build tx.cdc 
  --proposer alice
  --payer alice
  --authorizer bob
  --authorizer charlie 
  --filter payload --save tx1
```
Sign the transaction with authorizers:
```shell
> flow transactions sign tx1 --signer bob
  --filter payload --save tx2
```
```shell
> flow transactions sign tx2 --signer charlie
  --filter payload --save tx3
```
Sign the transaction with payer:
```shell
> flow transactions sign tx3 --signer alice
  --filter payload --save tx4
```
Submit the signed transaction:
```shell
> flow transactions send-signed tx4
```
Transaction content (`tx.cdc`):
```
transaction {
    preapre(bob: AuthAccount, charlie: AuthAccount) {}
    execute { ... }
}
```

### Different payer, proposer and authorizer
A transaction that declares different payer, proposer and authorizer each signing separately. 
Please note that the order of signing is important, and [the payer must sign last](https://docs.onflow.org/concepts/transaction-signing/#payer-signs-last).  

Build the transaction:
```shell
> flow transactions build tx.cdc 
  --proposer alice 
  --payer bob 
  --authorizer charlie 
  --filter payload --save tx1
```
Sign the transaction with proposer:
```shell
> flow transactions sign tx1 --signer alice 
  --filter payload --save tx2
```
Sign the transaction with authorizer:
```shell
> flow transactions sign tx2 --signer charlie 
  --filter payload --save tx3
```
Sign the transaction with payer:
```shell
> flow transactions sign tx3 --signer bob 
  --filter payload --save tx4
```
Submit the signed transaction:
```shell
> flow transactions send-signed tx4
```
Transaction content (`tx.cdc`):
```
transaction {
    preapre(charlie: AuthAccount) {}
    execute { ... }
}
```

### Single payer, proposer and authorizer but multiple keys
A transaction that declares same payer, proposer and authorizer but the signer account has two keys with half weight, required to sign with both.


Build the transaction:
```shell
> flow transactions build tx.cdc 
  --proposer dylan1 
  --payer dylan1
  --authorizer dylan1 
  --filter payload --save tx1
```
Sign the transaction with the first key:
```shell
> flow transactions sign tx1 --signer dylan1 
  --filter payload --save tx2
```
Sign the transaction with the second key:
```shell
> flow transactions sign tx2 --signer dylan2 
  --filter payload --save tx3
```
Submit the signed transaction:
```shell
> flow transactions send-signed tx3
```
Transaction content (`tx.cdc`):
```
transaction {
    preapre(signer: AuthAccount) {}
    execute { ... }
}
```

### Configuration
This is an example configuration using mock values:
```json
{
    ... 
    "accounts": {
        "alice": {
            "address": "0x1",
            "key": "111...111"
        },
        "bob": {
            "address": "0x2",
            "key": "222...222"
        },
        "charlie": {
            "address": "0x3",
            "key": "333...333"
        },
        "dylan1": {
            "address": "0x4",
            "key": "444...444"
        },
        "dylan2": {
            "address": "0x4",
            "key": "555...555"
        }
    }
    ...
}
```
