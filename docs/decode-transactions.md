---
title: Decode a Transaction with the Flow CLI
sidebar_title: Decode a Transaction
description: How to decode a Flow transaction from the command line
---

The Flow CLI provides a command to decode a transaction
from RLP in a file. It uses same transaction format as get command

```shell
flow transactions decode <file>
```

## Example Usage

```shell
> flow transactions decode ./rlp-file.rlp 

ID		c1a52308fb906358d4a33c1f1d5fc458d3cfea0d570a51a9dea915b90d678346
Payer		83de1a7075f190a1
Authorizers	[83de1a7075f190a1]

Proposal Key:	
    Address	    83de1a7075f190a1
    Index	    1
    Sequence	1

No Payload Signatures

Envelope Signature 0: 83de1a7075f190a1
Signatures (minimized, use --include signatures)

Code (hidden, use --include code)

Payload (hidden, use --include payload)
```

## Arguments

### Filename

- Name: `<file_name>`
- Valid Input: file name.

The first argument is the filename containing the transaction RLP.

## Flags
    
### Include Fields

- Flag: `--include`
- Valid inputs: `code`, `payload`, `signatures`

Specify fields to include in the result output. Applies only to the text output.

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

### Version Check

- Flag: `--skip-version-check`
- Default: `false`

Skip version check during start up to speed up process for slow connections.