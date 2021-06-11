package tests

import "github.com/spf13/afero"

type resource struct {
	Name   string
	Source []byte
}

var ContractHelloString = resource{
	Name: "contractHello.cdc",
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

var TransactionArgString = resource{
	Name: "transactionArg.cdc",
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

var resources = []resource{
	ContractHelloString,
	TransactionArgString,
}

func ReaderWriter() afero.Afero {
	var mockFS = afero.NewMemMapFs()

	for _, c := range resources {
		_ = afero.WriteFile(mockFS, c.Name, c.Source, 0644)
	}

	return afero.Afero{mockFS}
}
