package project

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/onflow/flow-go-sdk/templates"

	"github.com/onflow/flow-cli/pkg/flowcli"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"

	"github.com/onflow/flow-cli/pkg/flowcli/util"
)

type signerRole string

const (
	SignerRoleAuthorizer signerRole = "authorizer"
	SignerRoleProposer   signerRole = "proposer"
	SignerRolePayer      signerRole = "payer"
)

func NewTransaction() *Transaction {
	return &Transaction{
		tx:        flow.NewTransaction(),
		contracts: []templates.Contract{},
	}
}

type Transaction struct {
	signer    *Account
	role      signerRole
	proposer  *Account
	payer     flow.Address
	tx        *flow.Transaction
	contracts []templates.Contract
}

func (t *Transaction) Signer() *Account {
	return t.signer
}

func (t *Transaction) FlowTransaction() *flow.Transaction {
	return t.tx
}

func (t *Transaction) SetPayloadFromFile(filename string) error {
	partialTxHex, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read partial transaction from %s: %v", filename, err)
	}
	partialTxBytes, err := hex.DecodeString(string(partialTxHex))
	if err != nil {
		return fmt.Errorf("failed to decode partial transaction from %s: %v", filename, err)
	}
	tx, err := flow.DecodeTransaction(partialTxBytes)
	if err != nil {
		return fmt.Errorf("failed to decode transaction from %s: %v", filename, err)
	}

	t.tx = tx
	return nil
}

func (t *Transaction) SetScriptWithArgsFromFile(filepath string, args []string, argsJSON string) error {
	script, err := util.LoadFile(filepath)
	if err != nil {
		return err
	}

	return t.SetScriptWithArgs(script, args, argsJSON)
}

func (t *Transaction) SetScriptWithArgs(script []byte, args []string, argsJSON string) error {
	t.tx.SetScript(script)
	return t.AddRawArguments(args, argsJSON)
}

func (t *Transaction) SetSigner(account *Account) error {
	err := account.ValidateKey()
	if err != nil {
		return err
	}

	t.signer = account
	return nil
}

func (t *Transaction) SetProposer(account *Account) error {
	err := account.ValidateKey()
	if err != nil {
		return err
	}

	t.proposer = account
	return nil
}

func (t *Transaction) SetPayer(address flow.Address) {
	t.payer = address
}

func (t *Transaction) AddRawArguments(args []string, argsJSON string) error {
	txArguments, err := flowcli.ParseArguments(args, argsJSON)
	if err != nil {
		return err
	}

	return t.AddArguments(txArguments)
}

func (t *Transaction) AddArguments(args []cadence.Value) error {
	for _, arg := range args {
		err := t.AddArgument(arg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Transaction) AddArgument(arg cadence.Value) error {
	return t.tx.AddArgument(arg)
}

func (t *Transaction) AddAuthorizers(addresses []string) error {
	for _, address := range addresses {
		err := t.AddAuthorizer(address)
		if err != nil { // return error even if one breaks
			return err
		}
	}

	return nil
}

func (t *Transaction) AddAuthorizer(address string) error {
	authorizerAddress := flow.HexToAddress(address)
	if authorizerAddress == flow.EmptyAddress {
		return fmt.Errorf("invalid authorizer address provided %s", address)
	}

	t.tx.AddAuthorizer(authorizerAddress)
	return nil
}

func (t *Transaction) SignerRole(role string) error {
	t.role = signerRole(role)

	switch t.role {
	case SignerRoleAuthorizer: // Ignored if we're loading from a tx payload
		err := t.AddAuthorizer(t.signer.Address().String())
		if err != nil {
			return err
		}
	case SignerRolePayer:
		if t.payer != t.signer.Address() {
			return fmt.Errorf("role specified as Payer, but Payer address also provided, and different: %s != %s", t.payer, t.signer.Address())
		}
	case SignerRoleProposer: // Just sign payload, no special actions needed
	default:
		return fmt.Errorf("unknown role %s", role)
	}

	return nil
}

func (t *Transaction) Sign() (*Transaction, error) {
	keyIndex := t.signer.DefaultKey().Index()
	signerAddress := t.signer.Address()
	signer, err := t.signer.DefaultKey().Signer(context.Background())
	if err != nil {
		return nil, err
	}

	switch t.role {
	case SignerRoleAuthorizer, SignerRoleProposer:
		err := t.tx.SignPayload(signerAddress, keyIndex, signer)
		if err != nil {
			return nil, fmt.Errorf("failed to sign transaction: %s", err)
		}
	case SignerRolePayer:
		err := t.tx.SignEnvelope(signerAddress, keyIndex, signer)
		if err != nil {
			return nil, fmt.Errorf("failed to sign transaction: %s", err)
		}
	}

	return t, nil
}

func (t *Transaction) AddContractsFromArgs(contractArgs []string) error {
	for _, contract := range contractArgs {
		contractFlagContent := strings.SplitN(contract, ":", 2)
		if len(contractFlagContent) != 2 {
			return fmt.Errorf("wrong format for contract. Correct format is name:path, but got: %s", contract)
		}
		contractName := contractFlagContent[0]
		contractPath := contractFlagContent[1]

		contractSource, err := util.LoadFile(contractPath)
		if err != nil {
			return err
		}

		t.AddContract(contractName, string(contractSource))
	}

	return nil
}

func (t *Transaction) AddContract(name string, source string) {
	t.contracts = append(t.contracts,
		templates.Contract{
			Name:   name,
			Source: source,
		},
	)
}

func (t *Transaction) SetCreateAccount(keys []*flow.AccountKey) {
	t.tx = templates.CreateAccount(keys, t.contracts, t.signer.Address())
}

func (t *Transaction) SetUpdateContract(name string, source string) {
	t.AddContract(name, source)
	t.tx = templates.UpdateAccountContract(t.signer.Address(), t.contracts[0])
}

func (t *Transaction) SetDeployContract(name string, source string) {
	t.AddContract(name, source)
	t.tx = templates.AddAccountContract(t.signer.Address(), t.contracts[0])
}

func (t *Transaction) SetRemoveContract(name string) {
	t.tx = templates.RemoveAccountContract(t.signer.Address(), name)
}
