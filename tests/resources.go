package tests

import "github.com/spf13/afero"

type resource struct {
	name   string
	source []byte
}

var ContractHelloString = resource{
	name: "contractHello.cdc",
	source: []byte(`
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
		_ = afero.WriteFile(mockFS, c.name, c.source, 0644)
	}

	return afero.Afero{mockFS}
}
