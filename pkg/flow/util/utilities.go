/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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

package util

import (
	"fmt"
	"io/ioutil"
	"strings"

	flow2 "github.com/onflow/flow-cli/pkg/flow"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

func LoadFile(filename string) ([]byte, error) {
	var code []byte
	var err error

	if filename != "" {
		code, err = ioutil.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("Failed to load file: %s", filename)
		}
	}

	return code, nil
}

func IsByteSlice(v interface{}) bool {
	slice, isSlice := v.([]interface{})
	if !isSlice {
		return false
	}
	_, isBytes := slice[0].(byte)
	return isBytes
}

// AccountFromAddressAndKey get account from address and private key
func AccountFromAddressAndKey(accountAddress string, accountPrivateKey string) (*flow2.Account, error) {
	address := flow.HexToAddress(
		strings.ReplaceAll(accountAddress, "0x", ""),
	)

	privateKey, err := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, accountPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("private key is not correct")
	}

	account := flow2.AccountFromAddressAndKey(address, privateKey)
	return account, nil
}
