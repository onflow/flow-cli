package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

const serviceAccountName = "service"

type Account struct {
	Address    flow.Address
	PrivateKey crypto.PrivateKey
	SigAlgo    crypto.SignatureAlgorithm
	HashAlgo   crypto.HashAlgorithm
}

// An internal utility struct that defines how Account is converted to JSON.
type accountJSON struct {
	Address    string `json:"address"`
	PrivateKey string `json:"privateKey"`
	SigAlgo    string `json:"sigAlgorithm"`
	HashAlgo   string `json:"hashAlgorithm"`
}

func (acct *Account) MarshalJSON() ([]byte, error) {
	prKeyBytes := acct.PrivateKey.Encode()
	prKeyHex := hex.EncodeToString(prKeyBytes)

	return json.Marshal(accountJSON{
		Address:    acct.Address.Hex(),
		PrivateKey: prKeyHex,
		SigAlgo:    acct.SigAlgo.String(),
		HashAlgo:   acct.HashAlgo.String(),
	})
}

func (acct *Account) UnmarshalJSON(data []byte) (err error) {
	var alias accountJSON
	if err = json.Unmarshal(data, &alias); err != nil {
		return
	}

	acct.SigAlgo = crypto.StringToSignatureAlgorithm(alias.SigAlgo)
	acct.HashAlgo = crypto.StringToHashAlgorithm(alias.HashAlgo)

	var prKeyBytes []byte
	prKeyBytes, err = hex.DecodeString(alias.PrivateKey)
	if err != nil {
		return
	}

	acct.Address = flow.HexToAddress(alias.Address)
	acct.PrivateKey, err = crypto.DecodePrivateKey(acct.SigAlgo, prKeyBytes)
	return
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

func SaveConfig(conf *Config) error {
	data, err := json.MarshalIndent(conf, "", "\t")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(ConfigPath, data, 0777)
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
			fmt.Printf("Network config file %s does not exist. Please initialize first\n", ConfigPath)
		} else {
			fmt.Printf("Failed to open project configuration in %s\n", ConfigPath)
		}

		os.Exit(1)
	}

	d := json.NewDecoder(f)

	var conf Config

	if err := d.Decode(&conf); err != nil {
		fmt.Printf("%s contains invalid json: %s\n", ConfigPath, err.Error())
		os.Exit(1)
	}
	fmt.Println(conf.Host)

	return &conf
}

func ConfigExists() bool {
	info, err := os.Stat(ConfigPath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
