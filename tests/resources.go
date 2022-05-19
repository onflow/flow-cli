/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tests

import (
	"fmt"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/afero"

	"github.com/onflow/flow-cli/pkg/flowkit"
)

type resource struct {
	Name     string
	Filename string
	Source   []byte
}

var ContractHelloString = resource{
	Name:     "Hello",
	Filename: "contractHello.cdc",
	Source: []byte(`
		pub contract Hello {
			pub let greeting: String
			init() {
				self.greeting = "Hello, World!"
			}
			pub fun hello(): String {
				return self.greeting
			}
		}
	`),
}

var ContractSimple = resource{
	Name:     "Simple",
	Filename: "contractSimple.cdc",
	Source: []byte(`
		pub contract Simple {}
	`),
}

var ContractSimpleUpdated = resource{
	Name:     "Simple",
	Filename: "contractSimpleUpdated.cdc",
	Source: []byte(`
		pub contract Simple {
			pub let greeting: String
			init() {
				self.greeting = "Foo"
			}
		}
	`),
}

var ContractEvents = resource{
	Name:     "ContractEvents",
	Filename: "contractEvents.cdc",
	Source: []byte(`
		pub contract ContractEvents {
			pub struct S {
				pub var x: Int
				pub var y: String
				
				init(x: Int, y: String) {
					self.x = x
					self.y = y
				}
			}

			pub event EventA(x: Int)
			pub event EventB(x: Int, y: Int)
			pub event EventC(x: UInt8)
			pub event EventD(x: String)
			pub event EventE(x: UFix64) 
			pub event EventF(x: Address)
			pub event EventG(x: [UInt8])
			pub event EventH(x: [[UInt8]])
			pub event EventI(x: {String: Int})
			pub event EventJ(x: S)
			
			init() {
				emit EventA(x: 1)				
				emit EventB(x: 4, y: 2)	
				emit EventC(x: 1)
				emit EventD(x: "hello")
				emit EventE(x: 10.2)
				emit EventF(x: 0x436164656E636521)
				emit EventG(x: "hello".utf8)
				emit EventH(x: ["hello".utf8, "world".utf8])
				emit EventI(x: { "hello": 1337 })
				emit EventJ(x: S(x: 1, y: "hello world"))
			}
		}
	`),
}

var ContractA = resource{
	Name:     "ContractA",
	Filename: "contractA.cdc",
	Source:   []byte(`pub contract ContractA {}`),
}

var ContractB = resource{
	Name:     "ContractB",
	Filename: "contractB.cdc",
	Source: []byte(`
		import ContractA from "./contractA.cdc"
		pub contract ContractB {}
	`),
}

var ContractC = resource{
	Name:     "ContractC",
	Filename: "contractC.cdc",
	Source: []byte(`
		import ContractB from "./contractB.cdc"
		import ContractA from "./contractA.cdc"

		pub contract ContractC {
			pub let x: String
			init(x: String) {
				self.x = x
			}
		}
	`),
}

var TransactionArgString = resource{
	Filename: "transactionArg.cdc",
	Source: []byte(`
		transaction(greeting: String) {
			let guest: Address
			
			prepare(authorizer: AuthAccount) {
				self.guest = authorizer.address
			}
			
			execute {
				log(greeting.concat(",").concat(self.guest.toString()))
			}
		}
	`),
}

var TransactionImports = resource{
	Filename: "transactionImport.cdc",
	Source: []byte(`
		import Hello from "./contractHello.cdc"
		
		transaction() {
			prepare(authorizer: AuthAccount) {}
			execute {
				Hello.hello()
			}
		}
	`),
}

var TransactionSimple = resource{
	Filename: "transactionSimple.cdc",
	Source: []byte(`
		transaction() {}
	`),
}

var TransactionSingleAuth = resource{
	Filename: "transactionAuth1.cdc",
	Source: []byte(`
		transaction() {
			prepare(authorizer: AuthAccount) {}
		}
	`),
}

var TransactionTwoAuth = resource{
	Filename: "transactionAuth2.cdc",
	Source: []byte(`
		transaction() {
			prepare(auth1: AuthAccount, auth2: AuthAccount) {}
		}
	`),
}

var TransactionMultipleDeclarations = resource{
	Filename: "transactionMultipleDec.cdc",
	Source: []byte(`
		pub fun dummy(_ address: Address): Void {}

		transaction() {
			prepare(authorizer: AuthAccount) {}
		}
	`),
}

var ScriptWithError = resource{
	Filename: "scriptError.cdc",
	Source: []byte(`
	    	pub fun main(name: String): Strin {
		  return "Hello ".concat(name)
		}
	`),
}

var ScriptArgString = resource{
	Filename: "scriptArg.cdc",
	Source: []byte(`
		pub fun main(name: String): String {
		  return "Hello ".concat(name)
		}
	`),
}

var ScriptImport = resource{
	Filename: "scriptImport.cdc",
	Source: []byte(`
		import Hello from "./contractHello.cdc"

		pub fun main(): String {
		  return "Hello ".concat(Hello.greeting)
		}
	`),
}

var resources = []resource{
	ContractHelloString,
	TransactionArgString,
	ScriptArgString,
	ContractSimple,
	ContractSimpleUpdated,
	TransactionSimple,
	ScriptImport,
	ContractA,
	ContractB,
	ContractC,
}

func ReaderWriter() afero.Afero {
	var mockFS = afero.NewMemMapFs()

	for _, c := range resources {
		_ = afero.WriteFile(mockFS, c.Filename, c.Source, 0644)
	}

	return afero.Afero{Fs: mockFS}
}

func Alice() *flowkit.Account {
	return newAccount("Alice", "0x1", "seedseedseedseedseedseedseedseedseedseedseedseedAlice")
}

func Bob() *flowkit.Account {
	return newAccount("Bob", "0x2", "seedseedseedseedseedseedseedseedseedseedseedseedBob")
}

func Charlie() *flowkit.Account {
	return newAccount("Charlie", "0x3", "seedseedseedseedseedseedseedseedseedseedseedseedCharlie")
}
func Donald() *flowkit.Account {
	return newAccount("Donald", "0x3", "seedseedseedseedseedseedseedseedseedseedseedseedDonald")
}
func newAccount(name string, address string, seed string) *flowkit.Account {
	a := &flowkit.Account{}
	a.SetAddress(flow.HexToAddress(address))
	a.SetName(name)
	pk, _ := crypto.GeneratePrivateKey(crypto.ECDSA_P256, []byte(seed))

	a.SetKey(flowkit.NewHexAccountKeyFromPrivateKey(0, crypto.SHA3_256, pk))

	return a
}

func PubKeys() []crypto.PublicKey {
	var pubKeys []crypto.PublicKey
	privKeys := PrivKeys()
	for _, priv := range privKeys {
		pubKeys = append(pubKeys, priv.PublicKey())
	}
	return pubKeys
}

func PrivKeys() []crypto.PrivateKey {
	var privKeys []crypto.PrivateKey
	for x := 0; x < 5; x++ {
		pk, _ := crypto.GeneratePrivateKey(
			crypto.ECDSA_P256,
			[]byte(fmt.Sprintf("seedseedseedseedseedseedseedseedseedseedseedseed%d", x)),
		)
		privKeys = append(privKeys, pk)
	}
	return privKeys
}

func SigAlgos() []crypto.SignatureAlgorithm {
	var sigAlgos []crypto.SignatureAlgorithm
	privKeys := PrivKeys()
	for _, priv := range privKeys {
		sigAlgos = append(sigAlgos, priv.Algorithm())
	}
	return sigAlgos
}

func HashAlgos() []crypto.HashAlgorithm {
	var hashAlgos []crypto.HashAlgorithm
	for x := 0; x < 5; x++ {
		hashAlgos = append(hashAlgos, crypto.SHA3_256)
	}
	return hashAlgos
}
