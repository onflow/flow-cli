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

package cli

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/crypto/cloudkms"
)

type KeyType string

const (
	serviceAccountName = "service"

	KeyTypeHex   KeyType = "hex"   // Hex private key with in memory signer
	KeyTypeKMS   KeyType = "kms"   // Google KMS signer
	KeyTypeShell KeyType = "shell" // Exec out to a shell script

	defaultKeyType = KeyTypeHex
)

type Account struct {
	KeyType    KeyType
	KeyIndex   int
	KeyContext map[string]string
	Address    flow.Address
	PrivateKey crypto.PrivateKey
	SigAlgo    crypto.SignatureAlgorithm
	HashAlgo   crypto.HashAlgorithm
	Signer     crypto.Signer
}

// An internal utility struct that defines how Account is converted to JSON.
type accountJSON struct {
	Address    string            `json:"address"`
	PrivateKey string            `json:"privateKey"`
	SigAlgo    string            `json:"sigAlgorithm"`
	HashAlgo   string            `json:"hashAlgorithm"`
	KeyType    string            `json:"keyType"`
	KeyIndex   int               `json:"keyIndex"`
	KeyContext map[string]string `json:"keyContext"`
}

func (acct *Account) MarshalJSON() ([]byte, error) {
	prKeyHex := "deprecated"
	keyContext := acct.KeyContext
	if keyContext == nil {
		keyContext = make(map[string]string)
	}

	switch acct.KeyType {
	case KeyTypeHex:
		prKeyBytes := acct.PrivateKey.Encode()
		keyContext["privateKey"] = hex.EncodeToString(prKeyBytes)
		// Deprecated, but keep for now
		prKeyHex = keyContext["privateKey"]
	case KeyTypeKMS:
		// Key context should be filled, do nothing
		// TODO: Could validate contents
	default:
		return nil, fmt.Errorf("unknown key type %s", acct.KeyType)
	}

	return json.Marshal(accountJSON{
		Address:    acct.Address.Hex(),
		PrivateKey: prKeyHex,
		SigAlgo:    acct.SigAlgo.String(),
		HashAlgo:   acct.HashAlgo.String(),
		KeyType:    string(acct.KeyType),
		KeyIndex:   acct.KeyIndex,
		KeyContext: keyContext,
	})
}

func (acct *Account) UnmarshalJSON(data []byte) (err error) {
	var alias accountJSON
	if err = json.Unmarshal(data, &alias); err != nil {
		return
	}

	acct.Address = flow.HexToAddress(alias.Address)
	acct.SigAlgo = crypto.StringToSignatureAlgorithm(alias.SigAlgo)
	acct.HashAlgo = crypto.StringToHashAlgorithm(alias.HashAlgo)
	acct.KeyIndex = alias.KeyIndex
	acct.KeyContext = alias.KeyContext
	if alias.KeyType == "" {
		acct.KeyType = defaultKeyType
	} else {
		acct.KeyType = KeyType(alias.KeyType)
	}

	if acct.KeyType == KeyTypeHex {
		var prKeyBytes []byte
		prKeyBytes, err = hex.DecodeString(alias.PrivateKey)
		if err != nil {
			return
		}

		acct.PrivateKey, err = crypto.DecodePrivateKey(acct.SigAlgo, prKeyBytes)
		if err != nil {
			return
		}
	}

	return
}

func (account *Account) LoadSigner() error {
	switch account.KeyType {
	case KeyTypeHex:
		account.Signer = crypto.NewNaiveSigner(
			account.PrivateKey,
			account.HashAlgo,
		)
	case KeyTypeKMS:
		ctx := context.Background()
		accountKMSKey, err := kmsKeyFromKeyContext(account.KeyContext)
		if err != nil {
			return err
		}
		kmsClient, err := cloudkms.NewClient(ctx)
		if err != nil {
			return err
		}

		accountKMSSigner, err := kmsClient.SignerForKey(
			ctx,
			account.Address,
			accountKMSKey,
		)
		if err != nil {
			return err
		}
		account.Signer = accountKMSSigner
	default:
		return fmt.Errorf("Could not load signer with type %s", account.KeyType)
	}
	return nil
}

type Config struct {
	Host     string              `json:"host"`
	Accounts map[string]*Account `json:"accounts"`
}

func NewConfig() *Config {
	return &Config{
		Accounts: make(map[string]*Account),
	}
}

func (c *Config) ServiceAccount() *Account {
	serviceAcct, ok := c.Accounts[serviceAccountName]
	if !ok {
		Exit(1, "Missing service account!")
	}
	return serviceAcct
}

func (c *Config) SetServiceAccountKey(privateKey crypto.PrivateKey, hashAlgo crypto.HashAlgorithm) {
	c.Accounts[serviceAccountName] = &Account{
		Address:    flow.ServiceAddress(flow.Emulator),
		PrivateKey: privateKey,
		SigAlgo:    privateKey.Algorithm(),
		HashAlgo:   hashAlgo,
	}
}

func (c *Config) HostWithOverride(overrideIfNotEmpty string) string {
	if len(strings.TrimSpace(overrideIfNotEmpty)) > 0 {
		return overrideIfNotEmpty
	}
	if len(strings.TrimSpace(c.Host)) > 0 {
		return c.Host
	}
	return DefaultHost
}

func (c *Config) LoadSigners() error {
	for configName, account := range c.Accounts {
		err := account.LoadSigner()
		if err != nil {
			return fmt.Errorf("Could not load signer for config %s: %w", configName, err)
		}
	}
	return nil
}

func SaveConfig(conf *Config) error {
	data, err := json.MarshalIndent(conf, "", "\t")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(ConfigPath, data, 0777)
	if err != nil {
		return err
	}
	fmt.Printf("💾 Config file %s saved\n", ConfigPath)
	return nil
}

func MustSaveConfig(conf *Config) {
	if err := SaveConfig(conf); err != nil {
		Exitf(1, "Failed to save config err: %v", err)
	}
}

func LoadConfig() *Config {
	f, err := os.Open(ConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Project config file %s does not exist. Please initialize first\n", ConfigPath)
		} else {
			fmt.Printf("Failed to open project configuration in %s\n", ConfigPath)
		}

		os.Exit(1)
	}

	d := json.NewDecoder(f)

	conf := new(Config)

	if err := d.Decode(conf); err != nil {
		fmt.Printf("%s contains invalid json: %s\n", ConfigPath, err.Error())
		os.Exit(1)
	}
	fmt.Println(conf.Host)

	err = conf.LoadSigners()
	if err != nil {
		fmt.Printf("could not load signers for %s: %s\n", ConfigPath, err.Error())
		os.Exit(1)
	}
	return conf
}

func ConfigExists() bool {
	info, err := os.Stat(ConfigPath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

var kmsKeyContextFields = []string{
	"projectId",
	"locationId",
	"keyRingId",
	"keyId",
	"keyVersion",
}

func kmsKeyFromKeyContext(keyContext map[string]string) (cloudkms.Key, error) {

	for _, field := range kmsKeyContextFields {
		if val, ok := keyContext[field]; !ok || val == "" {
			return cloudkms.Key{}, fmt.Errorf("Could not generate KMS key from Context. Invalid value for %s", field)
		}
	}

	return cloudkms.Key{
		ProjectID:  keyContext[kmsKeyContextFields[0]],
		LocationID: keyContext[kmsKeyContextFields[1]],
		KeyRingID:  keyContext[kmsKeyContextFields[2]],
		KeyID:      keyContext[kmsKeyContextFields[3]],
		KeyVersion: keyContext[kmsKeyContextFields[4]],
	}, nil
}

// Regex that matches the resource name of GCP KMS keys, and parses out the values
var resourceRegexp = regexp.MustCompile(`projects/(?P<projectId>[^/]*)/locations/(?P<location>[^/]*)/keyRings/(?P<keyringId>[^/]*)/cryptoKeys/(?P<keyId>[^/]*)/cryptoKeyVersions/(?P<keyVersion>[^/]*)`)

func KeyContextFromKMSResourceID(resourceID string) (map[string]string, error) {
	match := resourceRegexp.FindStringSubmatch(resourceID)
	keyContext := make(map[string]string)
	for i, name := range resourceRegexp.SubexpNames() {
		if i != 0 && name != "" {
			keyContext[name] = match[i]
		}
	}

	return keyContext, nil
}
