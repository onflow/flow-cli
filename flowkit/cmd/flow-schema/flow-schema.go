package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	configJson "github.com/onflow/flow-cli/flowkit/config/json"
)

func main() {
	var verify bool
	flag.BoolVar(&verify, "verify", false, "Verify the schema")

	flag.Parse()
	path := flag.Arg(0)

	if path == "" {
		fmt.Println("Path is required")
		os.Exit(1)
	}

	schema := configJson.GenerateSchema()
	json, err := json.MarshalIndent(schema, "", "  ")

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if verify {
		fileContents, err := os.ReadFile(path)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if string(fileContents) != string(json) {
			fmt.Println("Schema is out of date - have you run `make generate-schema`?")
			os.Exit(1)
		}
	} else {
		os.WriteFile(path, json, 0644)
	}
}
