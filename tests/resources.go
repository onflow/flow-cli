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

var Contracts = []resource{
	ContractHelloString,
}

func ReaderWriter() afero.Afero {
	var mockFS = afero.NewMemMapFs()

	for _, c := range Contracts {
		_ = afero.WriteFile(mockFS, c.Name, c.Source, 0644)
	}

	return afero.Afero{mockFS}
}
