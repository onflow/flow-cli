package services

import (
	"fmt"

	"github.com/onflow/flow-cli/sharedlib/util"

	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/flow/config"
	"github.com/onflow/flow-cli/sharedlib/gateway"
	"github.com/onflow/flow-cli/sharedlib/lib"
	"github.com/onflow/flow-go-sdk"
)

// Scripts service handles all interactions for transactions
type Transactions struct {
	gateway gateway.Gateway
	project *cli.Project
}

// NewTransactions create new transaction service
func NewTransactions(gateway gateway.Gateway, project *cli.Project) *Transactions {
	return &Transactions{
		gateway: gateway,
		project: project,
	}
}

// Send transaction
func (t *Transactions) Send(
	transactionFilename string,
	signerName string,
	args []string,
	argsJSON string,
) (*flow.Transaction, *flow.TransactionResult, error) {

	signer := t.project.GetAccountByName(signerName)
	if signer == nil {
		return nil, nil, fmt.Errorf("signer account: [%s] doesn't exists in configuration", signerName)
	}

	// if google kms account then sign in
	if signer.DefaultKey().Type() == config.KeyTypeGoogleKMS {
		err := cli.GcloudApplicationSignin(signer)
		if err != nil {
			return nil, nil, err
		}
	}

	code, err := util.LoadFile(transactionFilename)
	if err != nil {
		return nil, nil, err
	}

	tx := flow.NewTransaction().
		SetScript(code).
		AddAuthorizer(signer.Address())

	transactionArguments, err := lib.ParseArguments(args, argsJSON)
	if err != nil {
		return nil, nil, err
	}

	for _, arg := range transactionArguments {
		err := tx.AddArgument(arg)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to add %s argument to a transaction", transactionFilename)
		}
	}

	tx, err = t.gateway.SendTransaction(tx, signer)
	if err != nil {
		return nil, nil, err
	}

	res, err := t.gateway.GetTransactionResult(tx)

	return tx, res, err
}

// Send transaction
func (t *Transactions) GetStatus(
	transactionID string,
	waitSeal bool,
) (*flow.Transaction, *flow.TransactionResult, error) {
	txID := flow.HexToID(transactionID)

	tx, err := t.gateway.GetTransaction(txID)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.gateway.GetTransactionResult(tx)

	return tx, result, err
}
