---
title: Add Contracts to a Flow Project
sidebar_title: Add Project Contracts
description: How to define the Cadence contracts for Flow project
---

## Add a Contract

To add a contract to your project, update the `"contracts"` section of your `flow.json` file.

Contracts are specified as key-value pairs, where the key is the contract name,
and the value is the location of the Cadence source code.

For example, the configuration below will register the 
contract `Foo` from the `FooContract.cdc` file.

```json
{
  "contracts": {
    "Foo": "./cadence/contracts/FooContract.cdc"
  }
}
```

## Define Contract Deployment Targets

Once a contract is added, it can then be assigned to one or more deployment targets.

A deployment target is an account to which the contract will be deployed.
In a typical project, a contract has one deployment target per network (e.g. Emulator, Testnet, Mainnet).

Deployment targets are defined in the `"deployments"` section of your `flow.json` file.

Targets are grouped by their network, where each network is a mapping from target account to contract list. 
Multiple contracts can be deployed to the same target account.

For example, here's how we'd deploy contracts `Foo` and `Bar` to the account `my-testnet-account`:

```json
{
  "contracts": {
    "Foo": "./cadence/contracts/FooContract.cdc",
    "Bar": "./cadence/contracts/BarContract.cdc"
  },
  "deployments": {
    "testnet": {
      "my-testnet-account": ["Foo", "Bar"]
    }
  }
}
```
