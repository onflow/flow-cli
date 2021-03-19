---
title: Get an Account with the Flow CLI
sidebar_title: Get an Account
description: How to get a Flow account from the command line
---

The Flow CLI provides a command to fetch any account by its address from the Flow network.

`flow accounts get <address>`


## Example Usage

```shell
flow accounts get 0xf8d6e0586b0a20c7
```

### Example response
```shell
Address	 0xf8d6e0586b0a20c7
Balance	 9999999999970000000
Keys	 1

Key 0	Public Key		 858a7d978b25d61f348841a343f79131f4b9fab341dd8a476a6f4367c25510570bf69b795fc9c3d2b7191327d869bcf848508526a3c1cafd1af34f71c7765117
	Weight			 1000
	Signature Algorithm	 ECDSA_P256
	Hash Algorithm		 SHA3_256

Contracts Deployed: 2
Contract: 'FlowServiceAccount'
Contract: 'FlowStorageFees'


```

## Flags

### Code

- Flag: `--code`

Display contract code deployed to the account.


