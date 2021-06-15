---
title: Decode Public Keys with the Flow CLI
sidebar_title: Decode Public Keys
description: How to decode Flow public keys from the command line
---

The Flow CLI provides a command to decode encoded public account keys.

```shell
flow keys decode <rlp|pem> <encoded public key>
```

## Example Usage

### Decode RLP Encoded Public Key
```shell
> flow keys decode rlp f847b84084d716c14b051ad6b001624f738f5d302636e6b07cc75e4530af7776a4368a2b586dbefc0564ee28384c2696f178cbed52e62811bcc9ecb59568c996d342db2402038203e8

Public Key 		 84d716c1...bcc9ecb59568c996d342db24 
Signature algorithm 	 ECDSA_P256
Hash algorithm 		 SHA3_256
Weight 			 1000
Revoked 		 false
```

### Decode PEM Encoded Public Key From File
```shell
> flow keys decode pem --from-file key.pem

Public Key 		 d479b3c...c4615360039a6660a366a95f 
Signature algorithm 	 ECDSA_P256
Hash algorithm 		 UNKNOWN
Revoked 		 false

```

## Arguments

### Encoding
- Valid inputs: `rlp`, `pem` 

First argument specifies a valid encoding of the public key provided.

### Optional: Public Key
- Name: `encoded public key`
- Valid inputs: valid encoded key content

Optional second argument provides content of the encoded public key. 
If this argument is omitted the `--from-file` must be used instead.  

## Flags

### From File

- Flag: `--from-file`
- Valid inputs: valid filepath

Provide file with the encoded public key. 

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
