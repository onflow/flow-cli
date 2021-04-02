---
title: Sign a Transaction with the Flow CLI
sidebar_title: Sign a Transaction
description: How to sign a Flow transaction from the command line
---

The Flow CLI provides a command to sign transactions with options to specify 
authorizator accounts, payer accounts and proposer accounts.

`flow transctions sign`

## Example Usage

```shell
> flow transactions sign ./tests/transaction.cdc --arg String:"Meow"

Hash		b03b18a8d9d30ff7c9f0fdaa80fcaab242c2f36eedb687dd9b368326311fe376
Payer		f8d6e0586b0a20c7
Authorizers	[f8d6e0586b0a20c7]

Proposal Key:	
    Address	f8d6e0586b0a20c7
    Index	0
    Sequence	6

No Envelope Signatures

Payload Signature 0:
    Address	f8d6e0586b0a20c7
    Signature	b5b1dfed2a899037...164e1b224a7ac924018e7033b68b0df86769dd54
    Key Index	0


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
f90184f...a199bfd9b837a11a0885f9104b54014750f5e3e5bfe4a5795968b0df86769dd54c0
```


### Example Sign and Send

```shell

> flow transactions sign ./tests/transaction.cdc --arg String:"Meow" --filter Payload --save payload1
💾 result saved to: payload1 


> flow transactions send --payload payload1

Status		✅ SEALED
Hash		9a38fb25c9fedc20b008aaed7a5ff00169a178411238573d6a6a55982a645129
Payer		f8d6e0586b0a20c7
Authorizers	[f8d6e0586b0a20c7]

Proposal Key:	
    Address	f8d6e0586b0a20c7
    Index	0
    Sequence	6

Payload Signature 0:
    Address	f8d6e0586b0a20c7
    Signature	c7ede88cfd45c7c01b2...097e4d5040f730f640c1d4ac
    Key Index	0

Envelope Signature 0:
    Address	f8d6e0586b0a20c7
    Signature	fb8226d6c61ddd58a2b...798e7a9be084eb622cf40f4f
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
f901cbf90138b......4649a4088797523c02d53f0c798e7a9be084eb622cf40f4f


```

### Example Different Payer and Signer

```shell
> flow transactions sign ./tests/transaction.cdc --arg String:"Meow" --payer-address 179b6b1cb6755e31  --filter Payload --save payload1
💾 result saved to: payload1 


> flow transactions sign --payload payload1 --role payer --signer alice --filter Payload --save payload2
💾 result saved to: payload2


> transactions send --payload payload2

Status		✅ SEALED
Hash		379ccb941b956c760f19307bbc961cc1e6bcdd7334b5941e9f55ab2151f52d43
Payer		179b6b1cb6755e31
Authorizers	[f8d6e0586b0a20c7]

Proposal Key:	
    Address	f8d6e0586b0a20c7
    Index	0
    Sequence	7

Payload Signature 0:
    Address	f8d6e0586b0a20c7
    Signature	f940766ce...3d2001e0b4b795e4f35345a18a04abc87466f6d0f1617b949b
    Key Index	0

Envelope Signature 0:
    Address	f8d6e0586b0a20c7
    Signature	000621...6ff9cdef2d62fb5e98ebbc7a7bacdb505ab799522c76f1d6e5aa3
    Key Index	0

Envelope Signature 1:
    Address	179b6b1cb6755e31
    Signature	c423ebae6...a855564dc927dcd3f764b81c67048d7936af131abe1952dc80
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
f90211f901...7a6520417574684163636f756e7429207b0a2020202073656c662e67756573

```


## Arguments

### Filename (optional)
- Name: `filename`
- Valid inputs: Any filename and path valid on the system.

The first argument is a path to a Cadence file containing the
transaction to be executed. This argument is optional as we can 
use `--payload` flag if we already have created a transaction and
we would just like to sign it.

## Flags

### Signer

- Flag: `--signer`
- Valid inputs: the name of an account defined in the configuration (`flow.json`)

Specify the name of the account that will be used to sign the transaction.

### Payload

- Flag: `--payload`
- Valid inputs: any filename and path valid on the system.

Specify the filename containing valid transaction payload that will be used for signing.

### Proposer

- Flag: `--proposer`
- Valid inputs: account from configuration.

Specify a name of the account that is proposing the transaction. 
Account must be defined in flow configuration.

### Role

- Flag: `--role`
- Valid inputs: authorizer, proposer, payer.

Specify a role of the signer. 
Read more about signer roles [here](https://docs.onflow.org/concepts/accounts-and-keys/).

### Add Authorizers

- Flag: `--add-authorizer`
- Valid Input: Flow account address.

Additional authorizer addresses to add to the transaction.
Read more about authorizers [here](https://docs.onflow.org/concepts/accounts-and-keys/).

### Payer Address

- Flag: `--payer-address`
- Valid Input: Flow account address.

Specify account address that will be paying for the transaction.
Read more about payers [here](https://docs.onflow.org/concepts/accounts-and-keys/).

### Arguments

- Flag: `--arg`
- Valid inputs: argument in `Type:Value` format.

Arguments passed to the Cadence transaction in `Type:Value` format.
The `Type` must be the same as type in the transaction source code for that argument.

### Arguments JSON

- Flag: `--argsJSON`
- Valid inputs: arguments in JSON-Cadence form.

Arguments passed to the Cadence transaction in `Type:Value` format.
The `Type` must be the same type as the corresponding parameter
in the Cadence transaction code.

### Host
- Flag: `--host`
- Valid inputs: an IP address or hostname.
- Default: `127.0.0.1:3569` (Flow Emulator)

Specify the hostname of the Access API that will be
used to execute the commands.

### Network

- Flag: `--network`
- Short Flag: `-n`
- Valid inputs: the name of a network defined in the configuration (`flow.json`)

Specify which network you want the command to use for execution.

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

### Configuration

- Flag: `--conf`
- Short Flag: `-f`
- Valid inputs: valid filename

Specify a filename for the configuration files, you can provide multiple configuration
files by using `-f` flag multiple times.




































```bash
Authorizers	[f8d6e0586b0a20c7]

Proposal Key:	
    Address	f8d6e0586b0a20c7
    Index	0
    Sequence	16

Payload Signature 0:
    Address	f8d6e0586b0a20c7...e888b1a5840e8881caa78208a45ac1ce9f77d98cb85ac982c4c8ca6b3510
    Key Index	0


Transaction Payload:
f90183f90137b8d17472616e73616374696f6e286772656574696e673a20537472696e6729207b0a20206c657420677
56573743a20416464726573730a0a20207072657061726528617574686f72697a65723a20417574684163636f756e74
29207b0a2020202073656c662e6775657374203d20617574686f72697a65722e616464726573730a20207d0a0a20206
5786563757465207b0a202020206c6f67286772656574696e672e636f6e63617428222c22292e636f6e636174287365
6c662e67756573742e746f537472696e67282929290a20207d0a7de1a07b2274797065223a22537472696e67222c227


```


Example:

Simple two step sign send
```
> transactions sign ./tests/transaction.cdc --arg String:"Foo"

Authorizers	[f8d6e0586b0a20c7]
Payer		f8d6e0586b0a20c7

Proposal Key:	
    Address	f8d6e0586b0a20c7
    Index	0
    Sequence	4

Payload Signature 0:
    Address	f8d6e0586b0a20c7
    Signature	bccac2e89d0407300a1a24f900757cdf15a49ef5e9ca92cc6cc54ea56ddda841bc8f47b803e5a97768f155d105376d62d40c2dbeaa822944289b92ad7eb33b9b
    Key Index	0


Transaction Payload:
f90183f90137b8d17472616e73616374696f6e286772656574696e673a20537472696e6729207b0a20206c65742067756573743a20416464726573730a0a20207072657061726528617574686f72697a65723a20417574684163636f756e7429207b0a2020202073656c662e6775657374203d20617574686f72697a65722e616464726573730a20207d0a0a202065786563757465207b0a202020206c6f67286772656574696e672e636f6e63617428222c22292e636f6e6361742873656c662e67756573742e746f537472696e67282929290a20207d0a7de1a07b2274797065223a22537472696e67222c2276616c7565223a22466f6f227d0aa076a251f4028186c2a934efa598e2cd41a6a33700893ae7098475cd05cc6c37fb8203e888f8d6e0586b0a20c7800488f8d6e0586b0a20c7c988f8d6e0586b0a20c7f846f8448080b840bccac2e89d0407300a1a24f900757cdf15a49ef5e9ca92cc6cc54ea56ddda841bc8f47b803e5a97768f155d105376d62d40c2dbeaa822944289b92ad7eb33b9bc0


> transactions send --payload payload1

Hash	 3130447e195587ef7a01ad40effd281680a02e7411b204391c837d451e246d42
Status	 SEALED
Payer	 f8d6e0586b0a20c7
Events	


```


Different Payer:
```
> keys generate
🔴️ Store Private Key safely and don't share with anyone! 
Private Key 	 8fea7a92c85aa1b653eb5fe407842886b76a2c009b800e82c570767cf010f384 
Public Key 	 ad26f02fbdd3f372e2fbcf94bf0374c9502e6fdf2446a3009772642195b811be528143217139d8111dda7a7b30f7a57ec798cc12d86d2e850f5e0cccbb195da2 

> accounts create --key ad26f02fbdd3f372e2fbcf94bf0374c9502e6fdf2446a3009772642195b811be528143217139d8111dda7a7b30f7a57ec798cc12d86d2e850f5e0cccbb195da2
Address	 0x179b6b1cb6755e31
Balance	 10000000
Keys	 1

Key 0	Public Key		 ad26f02fbdd3f372e2fbcf94bf0374c9502e6fdf2446a3009772642195b811be528143217139d8111dda7a7b30f7a57ec798cc12d86d2e850f5e0cccbb195da2
	Weight			 1000
	Signature Algorithm	 ECDSA_P256
	Hash Algorithm		 SHA3_256

Contracts Deployed: 0


> transactions sign ./tests/transaction.cdc --arg String:"Foo" --payer-address 0x179b6b1cb6755e31 --filter Payload --save payload2

> cmd/flow/main.go transactions sign --payload payload2 --role payer --signer payer-account 

Authorizers	[f8d6e0586b0a20c7]
Payer		179b6b1cb6755e31

Proposal Key:	
    Address	f8d6e0586b0a20c7
    Index	0
    Sequence	3

Payload Signature 0:
    Address	f8d6e0586b0a20c7
    Signature	ce858a8bd8a58d29f2cf37e4de1aad2d6c577b342ef285e02762cf224614d90bac8de3fe0ea742e0c5d78e4dc981c0e84ba1d208fc4ec4554bd9e732290567ee
    Key Index	0

Envelope Signature 0:
    Address	179b6b1cb6755e31
    Signature	4ba3a88be452a684f54b29a9ddc72469a98f3863992d173ed0ab3c52efe7f0c0c0a4fc03d343c971086929f097e5c1032655b61efc3a5e87ea3ab22755aad409
    Key Index	0


Transaction Payload:
f901caf90137b8d17472616e73616374696f6e286772656574696e673a20537472696e6729207b0a20206c65742067756573743a20416464726573730a0a20207072657061726528617574686f72697a65723a20417574684163636f756e7429207b0a2020202073656c662e6775657374203d20617574686f72697a65722e616464726573730a20207d0a0a202065786563757465207b0a202020206c6f67286772656574696e672e636f6e63617428222c22292e636f6e6361742873656c662e67756573742e746f537472696e67282929290a20207d0a7de1a07b2274797065223a22537472696e67222c2276616c7565223a22466f6f227d0aa0baaeb14e65391f51236b203665764aae224171aefc60bbfd1c2899a56e604c128203e888f8d6e0586b0a20c7800388179b6b1cb6755e31c988f8d6e0586b0a20c7f846f8448080b840ce858a8bd8a58d29f2cf37e4de1aad2d6c577b342ef285e02762cf224614d90bac8de3fe0ea742e0c5d78e4dc981c0e84ba1d208fc4ec4554bd9e732290567eef846f8440180b8404ba3a88be452a684f54b29a9ddc72469a98f3863992d173ed0ab3c52efe7f0c0c0a4fc03d343c971086929f097e5c1032655b61efc3a5e87ea3ab22755aad409





> go run cmd/flow/main.go transactions send --payload payload3

Hash	 047b9fab1ff28fd7fed35672c611dcef40ace745ed535417dae714062497dec4
Status	 SEALED
Payer	 179b6b1cb6755e31
Events	 

```

Different Authorizer

```

> transactions sign ./tests/transaction.cdc --arg String:"Foo" --add-authorizer 179b6b1cb6755e31

Authorizers	[f8d6e0586b0a20c7 179b6b1cb6755e31]
Payer		0000000000000000

Proposal Key:	
    Address	f8d6e0586b0a20c7
    Index	0
    Sequence	4

Payload Signature 0:
    Address	f8d6e0586b0a20c7
    Signature	64c0a6cb39c55364cff5c6d73622d22165efa994deb9a5b8b53ff80f0931db74380311ff9765bafcbce666d9b407638d9b0e26ad97b7a9dad963337d623b0de3
    Key Index	0


Transaction Payload:
f9018cf90140b8d17472616e73616374696f6e286772656574696e673a20537472696e6729207b0a20206c65742067756573743a20416464726573730a0a20207072657061726528617574686f72697a65723a20417574684163636f756e7429207b0a2020202073656c662e6775657374203d20617574686f72697a65722e616464726573730a20207d0a0a202065786563757465207b0a202020206c6f67286772656574696e672e636f6e63617428222c22292e636f6e6361742873656c662e67756573742e746f537472696e67282929290a20207d0a7de1a07b2274797065223a22537472696e67222c2276616c7565223a22466f6f227d0aa076a251f4028186c2a934efa598e2cd41a6a33700893ae7098475cd05cc6c37fb8203e888f8d6e0586b0a20c78004880000000000000000d288f8d6e0586b0a20c788179b6b1cb6755e31f846f8448080b84064c0a6cb39c55364cff5c6d73622d22165efa994deb9a5b8b53ff80f0931db74380311ff9765bafcbce666d9b407638d9b0e26ad97b7a9dad963337d623b0de3c0

// saved to payload 1

> transactions sign --payload payload1 --signer payer-account

Hash		9f6b7f270c1471991890935f3eb82d9913e9ad172e0bb5f0d445a8a511e5e4df
Payer		f8d6e0586b0a20c7
Authorizers	[f8d6e0586b0a20c7 179b6b1cb6755e31]

Proposal Key:	
    Address	f8d6e0586b0a20c7
    Index	0
    Sequence	5

Payload Signature 0:
    Address	f8d6e0586b0a20c7
    Signature	5c56f2e465b13f289f341733a95b8673aeb9bfa8e3ed9021ca6d51a0b59a4ed278ef95491efdfb7c30e6d599d52358b4b698431782096e7e38a499130dfaeb1e
    Key Index	0

Payload Signature 1:
    Address	179b6b1cb6755e31
    Signature	ca04dafea2a5438ba6d5ac7811bc2656035a23001eee3c96ddf581d015683006c42943b8475d3c61e6131578cc3fe80470ae8706e489b186bc2b333ff12f7282
    Key Index	0


Payload:
f901d2f90140b8d17472616e73616374696f6e286772656574696e673a20537472696e6729207b0a20206c65742067756573743a20416464726573730a0a20207072657061726528617574686f72697a65723a20417574684163636f756e7429207b0a2020202073656c662e6775657374203d20617574686f72697a65722e616464726573730a20207d0a0a202065786563757465207b0a202020206c6f67286772656574696e672e636f6e63617428222c22292e636f6e6361742873656c662e67756573742e746f537472696e67282929290a20207d0a7de1a07b2274797065223a22537472696e67222c2276616c7565223a22466f6f227d0aa0ff25bc88a84e5989cafda9da5e55baa9308e511a12fd4d2788c0c359c3f1e6738203e888f8d6e0586b0a20c7800588f8d6e0586b0a20c7d288f8d6e0586b0a20c788179b6b1cb6755e31f88cf8448080b8405c56f2e465b13f289f341733a95b8673aeb9bfa8e3ed9021ca6d51a0b59a4ed278ef95491efdfb7c30e6d599d52358b4b698431782096e7e38a499130dfaeb1ef8440180b840ca04dafea2a5438ba6d5ac7811bc2656035a23001eee3c96ddf581d015683006c42943b8475d3c61e6131578cc3fe80470ae8706e489b186bc2b333ff12f7282c0

// saved to payload 2

> transactions send --payload payload2

❌ Transaction Error 
execution error code 100: Execution failed:
error: authorizer count mismatch for transaction: expected 1, got 2
--> 5fde5868a7f23b64335fea2c92eee97272bdc3d6bd5389a26ba25c496fe141e8



Status		✅ SEALED
Hash		5fde5868a7f23b64335fea2c92eee97272bdc3d6bd5389a26ba25c496fe141e8
Payer		f8d6e0586b0a20c7
Authorizers	[f8d6e0586b0a20c7 179b6b1cb6755e31]

Proposal Key:	
    Address	f8d6e0586b0a20c7
    Index	0
    Sequence	5

Payload Signature 0:
    Address	f8d6e0586b0a20c7
    Signature	5c56f2e465b13f289f341733a95b8673aeb9bfa8e3ed9021ca6d51a0b59a4ed278ef95491efdfb7c30e6d599d52358b4b698431782096e7e38a499130dfaeb1e
    Key Index	0

Payload Signature 1:
    Address	179b6b1cb6755e31
    Signature	d02d792d0157ce357b87a1a8381e7a8a599386782d5ac0ae0738d2c1f1e96d495aad6fc058f0f81b6dfecc0c72b2358c55cfe6f68aaf13882cf177f48d44c337
    Key Index	0

Envelope Signature 0:
    Address	f8d6e0586b0a20c7
    Signature	2fce0b37e3cc7c44b0e709c7d2a8f145bf9af9aab8be7b1e4e76e20a1e6abc67958dd1be5c4a19f2283c9cddc9ace243688a3c2198858a502551dfd70b32790a
    Key Index	0


Events:	 
    None

Payload:
f90219f90140b8d17472616e73616374696f6e286772656574696e673a20537472696e6729207b0a20206c65742067756573743a20416464726573730a0a20207072657061726528617574686f72697a65723a20417574684163636f756e7429207b0a2020202073656c662e6775657374203d20617574686f72697a65722e616464726573730a20207d0a0a202065786563757465207b0a202020206c6f67286772656574696e672e636f6e63617428222c22292e636f6e6361742873656c662e67756573742e746f537472696e67282929290a20207d0a7de1a07b2274797065223a22537472696e67222c2276616c7565223a22466f6f227d0aa0ff25bc88a84e5989cafda9da5e55baa9308e511a12fd4d2788c0c359c3f1e6738203e888f8d6e0586b0a20c7800588f8d6e0586b0a20c7d288f8d6e0586b0a20c788179b6b1cb6755e31f88cf8448080b8405c56f2e465b13f289f341733a95b8673aeb9bfa8e3ed9021ca6d51a0b59a4ed278ef95491efdfb7c30e6d599d52358b4b698431782096e7e38a499130dfaeb1ef8440180b840d02d792d0157ce357b87a1a8381e7a8a599386782d5ac0ae0738d2c1f1e96d495aad6fc058f0f81b6dfecc0c72b2358c55cfe6f68aaf13882cf177f48d44c337f846f8448080b8402fce0b37e3cc7c44b0e709c7d2a8f145bf9af9aab8be7b1e4e76e20a1e6abc67958dd1be5c4a19f2283c9cddc9ace243688a3c2198858a502551dfd70b32790a


```





Kan Implementation - Multiple authorizers

```
go run cmd/flow/main.go transactions sign --payload payload11 --role authorizer --signer payer-account 
^C
MacBook-Pro:flow-cli Dapper$ go run cmd/flow/main.go transactions sign --payload payload11 --role authorizer --signer payer-account --output payload12
⚠️ You are using a new experimental configuration format. Support for this format is not yet available across all CLI commands.
⚠️ You are using a new experimental configuration format. Support for this format is not yet available across all CLI commands.
Authorizers (2):
  Authorizer 0: f8d6e0586b0a20c7
  Authorizer 1: 179b6b1cb6755e31

Arguments (1):
  Argument 0: {"type":"String","value":"Hello"}


Script:
transaction(greeting: String) {
  let guest: Address

  prepare(authorizer: AuthAccount) {
    self.guest = authorizer.address
  }

  execute {
    log(greeting.concat(",").concat(self.guest.toString()))
  }
}
===
Proposer Address: f8d6e0586b0a20c7
Proposer Key Index: 0
Payer Address: f8d6e0586b0a20c7
===
Payload Signatures (1):
  Signature 0 Signer Address: f8d6e0586b0a20c7
  Signature 0 Signer Key Index: 0
===
Does this look correct? (y/n)
> y
Payload contents verified
hexrlp encoded transaction written to payload12
MacBook-Pro:flow-cli Dapper$ cat payload12 
f901d2f90140b8d17472616e73616374696f6e286772656574696e673a20537472696e6729207b0a20206c65742067756573743a20416464726573730a0a20207072657061726528617574686f72697a65723a20417574684163636f756e7429207b0a2020202073656c662e6775657374203d20617574686f72697a65722e616464726573730a20207d0a0a202065786563757465207b0a202020206c6f67286772656574696e672e636f6e63617428222c22292e636f6e6361742873656c662e67756573742e746f537472696e67282929290a20207d0a7de3a27b2274797065223a22537472696e67222c2276616c7565223a2248656c6c6f227d0aa085710ebc28deba306bba85d19cf09aa525f1d4badd37843acc677f01fbcfc1f18088f8d6e0586b0a20c7800588f8d6e0586b0a20c7d288f8d6e0586b0a20c788179b6b1cb6755e31f88cf8448080b840ad1ea2b2ea309d875f4a54c49449bd891b45d76eaf3aafa0597d8e9b9e9822b8f86885c23d4eacc99ca3e4f4b0b321a375aebe972e6404d906fa77c0e674daa4f8440180b840267801d484a57a9be7136786807ef2f27059c4635f8412bb594e49b08b8b1af458ec93804d98ae24163987443ffb7847b57e188762aa483dc792ce902df67459c0MacBook-Pro:flow-cli Dapper$ 
MacBook-Pro:flow-cli Dapper$ 
MacBook-Pro:flow-cli Dapper$ 
MacBook-Pro:flow-cli Dapper$ go run cmd/flow/main.go transactions send --payload payload12 --signer emulator-account --results
⚠️ You are using a new experimental configuration format. Support for this format is not yet available across all CLI commands.
⚠️ You are using a new experimental configuration format. Support for this format is not yet available across all CLI commands.
Authorizers (2):
  Authorizer 0: f8d6e0586b0a20c7
  Authorizer 1: 179b6b1cb6755e31

Arguments (1):
  Argument 0: {"type":"String","value":"Hello"}


Script:
transaction(greeting: String) {
  let guest: Address

  prepare(authorizer: AuthAccount) {
    self.guest = authorizer.address
  }

  execute {
    log(greeting.concat(",").concat(self.guest.toString()))
  }
}
===
Proposer Address: f8d6e0586b0a20c7
Proposer Key Index: 0
Payer Address: f8d6e0586b0a20c7
===
Payload Signatures (2):
  Signature 0 Signer Address: f8d6e0586b0a20c7
  Signature 0 Signer Key Index: 0
  Signature 1 Signer Address: 179b6b1cb6755e31
  Signature 1 Signer Key Index: 0
===
Does this look correct? (y/n)
> y
Payload contents verified
Submitting transaction with ID b0aea61a31b3e872c2dac826d37a9667f252bc82f3e1d2be98f25485a0c09bd7 ...
Successfully submitted transaction with ID b0aea61a31b3e872c2dac826d37a9667f252bc82f3e1d2be98f25485a0c09bd7
Waiting for transaction b0aea61a31b3e872c2dac826d37a9667f252bc82f3e1d2be98f25485a0c09bd7 to be sealed...

Transaction b0aea61a31b3e872c2dac826d37a9667f252bc82f3e1d2be98f25485a0c09bd7 sealed

Status: SEALED
Execution Error: execution error code 100: Execution failed:
error: authorizer count mismatch for transaction: expected 1, got 2
--> b0aea61a31b3e872c2dac826d37a9667f252bc82f3e1d2be98f25485a0c09bd7

Code: 
transaction(greeting: String) {
  let guest: Address

  prepare(authorizer: AuthAccount) {
    self.guest = authorizer.address
  }

  execute {
    log(greeting.concat(",").concat(self.guest.toString()))
  }
}
Events:
  None

```