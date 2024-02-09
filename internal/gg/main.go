package main

import (
	"fmt"

	"github.com/onflow/cadence"
	cadencejson "github.com/onflow/cadence/encoding/json"
)

func main() {
	fmt.Println("hello")

	vals := cadence.NewArray(
		[]cadence.Value{
			cadence.NewBool(true), cadence.NewBool(false),
		},
	)

	fmt.Println("vals =", vals)

	b := cadencejson.MustEncode(vals)
	fmt.Println("b =", string(b))
}
