// Package cli defines constants, configurations, and utilities that are used across the Flow CLI.
package cli

import (
	"crypto/rand"
	"fmt"
	"os"

	"github.com/onflow/flow-go-sdk/crypto"
)

const (
	EnvPrefix       = "FLOW"
	DefaultSigAlgo  = crypto.ECDSA_P256
	DefaultHashAlgo = crypto.SHA3_256
)

var ConfigPath = "flow.json"

func Exit(code int, msg string) {
	fmt.Println(msg)
	os.Exit(code)
}

func Exitf(code int, msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
	os.Exit(code)
}

func MustDecodePrivateKeyHex(sigAlgo crypto.SignatureAlgorithm, prKeyHex string) crypto.PrivateKey {
	prKey, err := crypto.DecodePrivateKeyHex(sigAlgo, prKeyHex)
	if err != nil {
		Exitf(1, "Failed to decode private key: %v", err)
	}
	return prKey
}

func MustDecodePublicKeyHex(sigAlgo crypto.SignatureAlgorithm, pubKeyHex string) crypto.PublicKey {
	pubKey, err := crypto.DecodePublicKeyHex(sigAlgo, pubKeyHex)
	if err != nil {
		Exitf(1, "Failed to decode public key: %v", err)
	}
	return pubKey
}

func RandomSeed(n int) []byte {
	seed := make([]byte, n)

	_, err := rand.Read(seed)
	if err != nil {
		Exitf(1, "Failed to generate random seed: %v", err)
	}

	return seed
}
