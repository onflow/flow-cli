package services

import (
	"fmt"
	"strings"

	"github.com/onflow/flow-cli/flow/util"

	"github.com/onflow/flow-cli/flow/config"
	"github.com/onflow/flow-cli/flow/gateway"
	"github.com/onflow/flow-cli/flow/lib"
	"github.com/onflow/flow-go-sdk"
)

// Scripts service handles all interactions for transactions
type Transactions struct {
	gateway gateway.Gateway
	project *lib.Project
	logger  util.Logger
}

// NewTransactions create new transaction service
func NewTransactions(
	gateway gateway.Gateway,
	project *lib.Project,
	logger util.Logger,
) *Transactions {
	return &Transactions{
		gateway: gateway,
		project: project,
		logger:  logger,
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
		err := lib.GcloudApplicationSignin(signer)
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

	t.logger.Info(fmt.Sprintf("Sending transaction with ID %s", tx.ID()))

	tx, err = t.gateway.SendTransaction(tx, signer)
	if err != nil {
		return nil, nil, err
	}

	t.logger.StartProgress("Waiting for transaction to be sealed...")

	res, err := t.gateway.GetTransactionResult(tx)

	t.logger.StartProgress("")

	return tx, res, err
}

// Send transaction
func (t *Transactions) GetStatus(
	transactionID string,
	waitSeal bool,
) (*flow.Transaction, *flow.TransactionResult, error) {
	txID := flow.HexToID(
		strings.ReplaceAll(transactionID, "0x", ""),
	)

	tx, err := t.gateway.GetTransaction(txID)
	if err != nil {
		return nil, nil, err
	}

	t.logger.StartProgress("Waiting for transaction to be sealed...")

	result, err := t.gateway.GetTransactionResult(tx)

	t.logger.StopProgress("")

	return tx, result, err
}
