---
title: Decode Account Keys with the Flow CLI
sidebar_title: Decode Public Keys
description: How to decode Flow public keys from the command line
---

The Flow CLI provides a command to decode encoded public account keys.

```shell
flow keys decode <rlp|pem> <encoded public key>
```

## Example Usage

```shell
> flow keys decode rlp f847b84084d716c14b051ad6b001624f738f5d302636e6b07cc75e4530af7776a4368a2b586dbefc0564ee28384c2696f178cbed52e62811bcc9ecb59568c996d342db2402038203e8

Public Key 		 84d716c14b051ad6b001624f738f5d302636e6b07cc75e4530af7776a4368a2b586dbefc0564ee28384c2696f178cbed52e62811bcc9ecb59568c996d342db24 
Signature algorithm 	 ECDSA_P256
Hash algorithm 		 SHA3_256
Weight 			 1000
Revoked 		 false
```

## Flags

### Filter

- Flag: `--filter`
- Short Flag: `-x`
- Valid inputs: a case-sensitive name of the result property.

Specify any property name from the result you want to return as the only value.

### Output

- Flag: `--output`
- Short Flag: `-o`
- Valid inputs: `json`, `inline`

Specify the format of the command results.

### Save

- Flag: `--save`
- Short Flag: `-s`
- Valid inputs: a path in the current filesystem.

Specify the filename where you want the result to be saved
