package services

import (
	"fmt"
	"strings"

	"github.com/onflow/flow-cli/sharedlib/util"

	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/sharedlib/gateway"
	"github.com/onflow/flow-cli/sharedlib/lib"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/templates"
)

// Accounts service handles all interactions for accounts
type Accounts struct {
	gateway gateway.Gateway
	project cli.Project
}

// NewAccounts create new account service
func NewAccounts(gateway gateway.Gateway, project cli.Project) *Accounts {
	return &Accounts{
		gateway: gateway,
		project: project,
	}
}

// Get gets an account based on address
func (a *Accounts) Get(address string) (*flow.Account, error) {
	flowAddress := flow.HexToAddress(
		strings.ReplaceAll(address, "0x", ""),
	)

	return a.gateway.GetAccount(flowAddress)
}

// Create creates an account with signer name, keys, algorithms, contracts and returns the new account
func (a *Accounts) Create(
	signerName string,
	keys []string,
	signatureAlgorithm string,
	hashingAlgorithm string,
	contracts []string,
) (*flow.Account, error) {

	signer := a.project.GetAccountByName(signerName)
	if signer == nil {
		return nil, fmt.Errorf("signer account: [%s] doesn't exists in configuration", signerName)
	}

	accountKeys := make([]*flow.AccountKey, len(keys))

	sigAlgo := crypto.StringToSignatureAlgorithm(signatureAlgorithm)
	if sigAlgo == crypto.UnknownSignatureAlgorithm {
		return nil, fmt.Errorf("failed to determine signature algorithm from %s", sigAlgo)
	}
	hashAlgo := crypto.StringToHashAlgorithm(hashingAlgorithm)
	if hashAlgo == crypto.UnknownHashAlgorithm {
		return nil, fmt.Errorf("failed to determine hash algorithm from %s", hashAlgo)
	}

	for i, publicKeyHex := range keys {
		publicKey := cli.MustDecodePublicKeyHex(cli.DefaultSigAlgo, publicKeyHex)
		accountKeys[i] = &flow.AccountKey{
			PublicKey: publicKey,
			SigAlgo:   sigAlgo,
			HashAlgo:  hashAlgo,
			Weight:    flow.AccountKeyWeightThreshold,
		}
	}

	contractTemplates := []templates.Contract{}

	for _, contract := range contracts {
		contractFlagContent := strings.SplitN(contract, ":", 2)
		if len(contractFlagContent) != 2 {
			return nil, fmt.Errorf("failed to read contract name and path from flag. Ensure you're providing a contract name and a file path %s", contract)
		}
		contractName := contractFlagContent[0]
		contractPath := contractFlagContent[1]

		contractSource, err := util.LoadFile(contractPath)
		if err != nil {
			return nil, err
		}

		contractTemplates = append(contractTemplates,
			templates.Contract{
				Name:   contractName,
				Source: string(contractSource),
			},
		)
	}

	tx := templates.CreateAccount(accountKeys, contractTemplates, signer.Address())
	tx, err := a.gateway.SendTransaction(tx, signer)
	if err != nil {
		return nil, err
	}

	result, err := a.gateway.GetTransactionResult(tx)
	if err != nil {
		return nil, err
	}

	events := lib.EventsFromTransaction(result)
	newAccountAddress := events.GetAddress()

	if newAccountAddress == nil {
		return nil, fmt.Errorf("New account address couldn't be fetched")
	}

	return a.gateway.GetAccount(*newAccountAddress)
}

// AddContract adds new contract to the account and returns the updated account
func (a *Accounts) AddContract(
	accountName string,
	contractName string,
	contractFilename string,
	updateExisting bool,
) (*flow.Account, error) {
	contractSource, err := util.LoadFile(contractFilename)
	if err != nil {
		return nil, err
	}

	account := a.project.GetAccountByName(accountName)
	if account == nil {
		return nil, fmt.Errorf("account: [%s] doesn't exists in configuration", accountName)
	}

	tx := templates.AddAccountContract(
		account.Address(),
		templates.Contract{
			Name:   contractName,
			Source: string(contractSource),
		},
	)

	// if we are updating contract
	if updateExisting {
		tx = templates.UpdateAccountContract(
			account.Address(),
			templates.Contract{
				Name:   contractName,
				Source: string(contractSource),
			},
		)
	}

	tx, err = a.gateway.SendTransaction(tx, account)
	if err != nil {
		return nil, err
	}

	_, err = a.gateway.GetTransactionResult(tx)
	if err != nil {
		return nil, err
	}

	return a.gateway.GetAccount(account.Address())
}

// RemoveContracts removes a contract from the account
func (a *Accounts) RemoveContract(
	contractName string,
	accountName string,
) (*flow.Account, error) {
	account := a.project.GetAccountByName(accountName)
	if account == nil {
		return nil, fmt.Errorf("account: [%s] doesn't exists in configuration", accountName)
	}

	tx := templates.RemoveAccountContract(account.Address(), contractName)
	tx, err := a.gateway.SendTransaction(tx, account)
	if err != nil {
		return nil, err
	}

	return a.gateway.GetAccount(account.Address())
}
