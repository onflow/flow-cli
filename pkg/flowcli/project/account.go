package project

import (
	"fmt"

	"github.com/onflow/flow-cli/pkg/flowcli/config"
	"github.com/onflow/flow-cli/pkg/flowcli/util"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

type Account struct {
	name    string
	address flow.Address
	chainID flow.ChainID
	keys    []AccountKey
}

func (a *Account) Address() flow.Address {
	return a.address
}

func (a *Account) Name() string {
	return a.name
}

func (a *Account) DefaultKey() AccountKey {
	return a.keys[0]
}

func (a *Account) SetDefaultKey(key AccountKey) {
	a.keys[0] = key
}

func accountsFromConfig(conf *config.Config) ([]*Account, error) {
	accounts := make([]*Account, 0, len(conf.Accounts))

	for _, accountConf := range conf.Accounts {
		account, err := AccountFromConfig(accountConf)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

func AccountFromAddressAndKey(address flow.Address, privateKey crypto.PrivateKey) *Account {
	key := NewHexAccountKeyFromPrivateKey(0, crypto.SHA3_256, privateKey)
	chainID, _ := util.GetAddressNetwork(address)

	return &Account{
		name:    "",
		address: address,
		chainID: chainID,
		keys:    []AccountKey{key},
	}
}

func AccountFromConfig(accountConf config.Account) (*Account, error) {
	accountKeys := make([]AccountKey, 0, len(accountConf.Keys))

	for _, key := range accountConf.Keys {
		accountKey, err := NewAccountKey(key)
		if err != nil {
			return nil, err
		}

		accountKeys = append(accountKeys, accountKey)
	}

	return &Account{
		name:    accountConf.Name,
		address: accountConf.Address,
		chainID: accountConf.ChainID,
		keys:    accountKeys,
	}, nil
}

func accountsToConfig(accounts []*Account) config.Accounts {
	accountConfs := make([]config.Account, 0)

	for _, account := range accounts {
		accountConfs = append(accountConfs, accountToConfig(account))
	}

	return accountConfs
}

func accountToConfig(account *Account) config.Account {
	keyConfigs := make([]config.AccountKey, 0, len(account.keys))

	for _, key := range account.keys {
		keyConfigs = append(keyConfigs, key.ToConfig())
	}

	return config.Account{
		Name:    account.name,
		Address: account.address,
		ChainID: account.chainID,
		Keys:    keyConfigs,
	}
}

func generateEmulatorServiceAccount(sigAlgo crypto.SignatureAlgorithm, hashAlgo crypto.HashAlgorithm) (*Account, error) {
	seed, err := util.RandomSeed(crypto.MinSeedLength)
	if err != nil {
		return nil, err
	}

	privateKey, err := crypto.GeneratePrivateKey(sigAlgo, seed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate emulator service key: %v", err)
	}

	serviceAccountKey := NewHexAccountKeyFromPrivateKey(0, hashAlgo, privateKey)

	return &Account{
		name:    DefaultEmulatorServiceAccountName,
		address: flow.ServiceAddress(flow.Emulator),
		chainID: flow.Emulator,
		keys: []AccountKey{
			serviceAccountKey,
		},
	}, nil
}
