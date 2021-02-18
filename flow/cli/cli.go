/*
 * Flow CLI
 *
 * Copyright 2019-2020 Dapper Labs, Inc.
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

// Package cli defines constants, configurations, and utilities that are used across the Flow CLI.
package cli

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/onflow/flow-go-sdk/crypto"
)

const (
	EnvPrefix           = "FLOW"
	DefaultSigAlgo      = crypto.ECDSA_P256
	DefaultHashAlgo     = crypto.SHA3_256
	DefaultHost         = "127.0.0.1:3569"
	UFix64DecimalPlaces = 8
)

func Exit(code int, msg string) {
	fmt.Println(msg)
	os.Exit(code)
}

func Exitf(code int, msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
	os.Exit(code)
}

// Colors
var Yellow = color.New(color.FgYellow, color.Bold).SprintfFunc()
var Green = color.New(color.FgGreen, color.Bold).SprintfFunc()
var Red = color.New(color.FgRed, color.Bold).SprintfFunc()
var Bold = color.New(color.Bold).SprintfFunc()

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

var squareBracketRegex = regexp.MustCompile(`(?s)\[(.*)\]`)

// GcloudApplicationSignin signs in as an application user using gcloud command line tool
// currently assumes gcloud is already installed on the machine
// will by default pop a browser window to sign in
func GcloudApplicationSignin(project string) {
	// os.Exec(fmt.Sprintf("gcloud auth application-default login --%s"))
	loginCmd := exec.Command("gcloud", "auth", "application-default", "login", fmt.Sprintf("--project=%s", project))

	output, err := loginCmd.CombinedOutput()
	if err != nil {
		Exitf(1, "Failed to run %q: %s\n", loginCmd.String(), err)
	}
	regexResult := squareBracketRegex.FindAllStringSubmatch(string(output), -1)
	// Should only be one value. Second index since first index contains the square brackets
	googleApplicationCreds := regexResult[0][1]
	fmt.Printf("Saving credentials and setting GOOGLE_APPLICATION_CREDENTIALS to file: %s\n", googleApplicationCreds)

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", googleApplicationCreds)
}
