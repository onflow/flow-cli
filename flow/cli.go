// Package cli defines constants, configurations, and utilities that are used across the Flow CLI.
package cli

import (
	"crypto/rand"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/onflow/flow-go-sdk/crypto"
)

const (
	EnvPrefix           = "FLOW"
	DefaultSigAlgo      = crypto.ECDSA_P256
	DefaultHashAlgo     = crypto.SHA3_256
	DefaultHost         = "127.0.0.1:3569"
	UFix64DecimalPlaces = 8
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

// FixedPointToString converts the given amount to a string with the given number of decimal places.
func FixedPointToString(amount uint64, decimalPlaces int) string {
	amountStr := strconv.Itoa(int(amount))
	if len(amountStr) < decimalPlaces {
		padding := strings.Repeat("0", decimalPlaces-len(amountStr))
		return fmt.Sprintf("0.%s%s", padding, amountStr)
	} else if len(amountStr) == decimalPlaces {
		return fmt.Sprintf("0.%s", amountStr)
	}
	return fmt.Sprintf("%s.%s", amountStr[:len(amountStr)-decimalPlaces], amountStr[len(amountStr)-decimalPlaces:])
}

func FormatUFix64(flow uint64) string {
	return FixedPointToString(flow, UFix64DecimalPlaces)
}
