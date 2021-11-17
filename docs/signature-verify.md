---
title: Verify a Signature with the Flow CLI
sidebar_title: Verify Signature
description: How to verify a signature from the command line 
---

Verify validity of a signature based on provided message and public key of the signature creator.

```shell
flow signatures verify <message> <signature> <public key>
```

## Example Usage

```shell
> flow signatures verify 
  'The quick brown fox jumps over the lazy dog' 
  b1c9eff5d829fdeaf2dad6308fc8033e3b8875bc185ef804ce5d0d980545ef5be0f98b47afc979d12272d257ce13c4b490e431bfcada485cb1d2e3f209be8d07 
  0xc92a7c72a78f8f046a79f8a5fe1ef72424258a55eb869f13e6133301d64ad025d3362d5df9e7c82289637af1431042c4025d241fd430242368ce662d39636987

Valid 			 true
Message 		 The quick brown fox jumps over the lazy dog
Signature 		 b1c9eff5d829fdeaf2...7ce13c4b490eada485cb1d2e3f209be8d07
Public Key 		 c92a7c72a78...1431042c4025d241fd430242368ce662d39636987
Hash Algorithm 		 SHA3_256
Signature Algorithm 	 ECDSA_P256
```

## Arguments

### Message
- Name: `message`

Message data used for creating the signature.

### Signature
- Name: `signature`

Message signature that will be verified.

### Public Key
- Name: `public key`

Public key of the private key used for creating the signature. 

## Flags

### Public Key Signature Algorithm

- Flag: `--sig-algo`
- Valid inputs: `"ECDSA_P256", "ECDSA_secp256k1"`

Specify the ECDSA signature algorithm of the key pair used for signing.

Flow supports the secp256k1 and P-256 curves.

### Public Key Hash Algorithm

- Flag: `--hash-algo`
- Valid inputs: `"SHA2_256", "SHA3_256"`
- Default: `"SHA3_256"`

Specify the hash algorithm of the key pair used for signing. 

### Filter

- Flag: `--filter`
- Short Flag: `-x`
- Valid inputs: case-sensitive name of the result property.

Specify any property name from the result you want to return as the only value.

### Output

- Flag: `--output`
- Short Flag: `-o`
- Valid inputs: `json`, `inline`

Specify in which format you want to display the result.

### Save

- Flag: `--save`
- Short Flag: `-s`
- Valid inputs: valid filename

Specify the filename where you want the result to be saved.

### Log

- Flag: `--log`
- Short Flag: `-l`
- Valid inputs: `none`, `error`, `debug`
- Default: `info`

Specify the log level. Control how much output you want to see while command execution.




