package flowkit

import (
	"bytes"
	"context"
	"fmt"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/access/grpc"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"strings"
	"sync"

	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/project"
)

var _ Services = &Flowkit{}

type Flowkit struct {
	state   *State
	network *config.Network
	gateway gateway.Gateway
	logger  output.Logger
}

func (f *Flowkit) Network() *config.Network {
	return f.network
}

func (f *Flowkit) Ping() (*config.Network, error) {
	err := f.gateway.Ping()
	if err != nil {
		return nil, err
	}

	return f.network, nil
}

// GetAccount fetches account on the Flow network.
func (f *Flowkit) GetAccount(ctx context.Context, address flow.Address) (*flow.Account, error) {
	f.logger.StartProgress(fmt.Sprintf("Loading %s...", address))

	account, err := f.gateway.GetAccount(address)
	f.logger.StopProgress()

	return account, err
}

// CreateAccount on the Flow network with the provided keys and using the signer for creation transaction.
// Returns the newly created account as well as the ID of the transaction that created the account.
//
// Keys is a slice but only one can be passed as well. If the transaction fails or there are other issues an error is returned.
func (f *Flowkit) CreateAccount(ctx context.Context, signer *Account, keys []Key) (*flow.Account, flow.Identifier, error) {
	if f.state == nil {
		return nil, flow.EmptyID, config.ErrDoesNotExist
	}

	var accKeys []*flow.AccountKey
	for _, k := range keys {
		if k.weight == 0 { // if key weight is specified
			k.weight = flow.AccountKeyWeightThreshold
		}

		accKey := &flow.AccountKey{
			PublicKey: k.public,
			SigAlgo:   k.sigAlgo,
			HashAlgo:  k.hashAlgo,
			Weight:    k.weight,
		}

		err := accKey.Validate()
		if err != nil {
			return nil, flow.EmptyID, fmt.Errorf("invalid account key: %w", err)
		}

		accKeys = append(accKeys, accKey)
	}

	tx, err := NewCreateAccountTransaction(signer, accKeys, nil)
	if err != nil {
		return nil, flow.EmptyID, err
	}

	tx, err = f.prepareTransaction(tx, signer)
	if err != nil {
		return nil, flow.EmptyID, err
	}

	f.logger.Info(fmt.Sprintf("Transaction ID: %s", tx.FlowTransaction().ID()))
	f.logger.StartProgress("Creating account...")
	defer f.logger.StopProgress()

	sentTx, err := f.gateway.SendSignedTransaction(tx)
	if err != nil {
		return nil, flow.EmptyID, errors.Wrap(err, "account creation transaction failed")
	}

	f.logger.StartProgress("Waiting for transaction to be sealed...")

	result, err := f.gateway.GetTransactionResult(sentTx.ID(), true)
	if err != nil {
		return nil, flow.EmptyID, err
	}

	if result.Error != nil {
		return nil, flow.EmptyID, result.Error
	}

	events := EventsFromTransaction(result)
	newAccountAddress := events.GetCreatedAddresses()
	if len(newAccountAddress) == 0 {
		return nil, flow.EmptyID, fmt.Errorf("new account address couldn't be fetched")
	}

	f.logger.StopProgress()

	account, err := f.gateway.GetAccount(*newAccountAddress[0])
	if err != nil {
		return nil, flow.EmptyID, err
	}

	return account, sentTx.ID(), nil // we know it's the only and first event
}

// prepareTransaction prepares transaction for sending with data from network
func (f *Flowkit) prepareTransaction(
	tx *Transaction,
	account *Account,
) (*Transaction, error) {

	block, err := f.gateway.GetLatestBlock()
	if err != nil {
		return nil, err
	}

	proposer, err := f.gateway.GetAccount(account.Address())
	if err != nil {
		return nil, err
	}

	tx.SetBlockReference(block)
	if err = tx.SetProposer(proposer, account.Key().Index()); err != nil {
		return nil, err
	}

	tx, err = tx.Sign()
	if err != nil {
		return nil, err
	}

	return tx, nil
}

var errUpdateNoDiff = errors.New("contract already exists and is the same as the contract provided for update")

// AddContract to the Flow account provided and return the transaction ID.
//
// If the contract already exists on the account the operation will fail and error will be returned.
// Use UpdateContract method for such usage.
func (f *Flowkit) AddContract(
	ctx context.Context,
	account *Account,
	contract *Script,
	updateExisting bool,
) (flow.Identifier, bool, error) {
	program, err := project.NewProgram(contract)
	if err != nil {
		return flow.EmptyID, false, err
	}

	if program.HasImports() {
		contracts, err := f.state.DeploymentContractsByNetwork(f.network)
		if err != nil {
			return flow.EmptyID, false, err
		}

		importReplacer := project.NewImportReplacer(
			contracts,
			f.state.AliasesForNetwork(f.network),
		)

		program, err = importReplacer.Replace(program)
		if err != nil {
			return flow.EmptyID, false, err
		}
	}

	name, err := program.Name()
	if err != nil {
		return flow.EmptyID, false, err
	}

	tx, err := NewAddAccountContractTransaction(
		account,
		name,
		program.Code(),
		contract.Args,
	)
	if err != nil {
		return flow.EmptyID, false, err
	}

	f.logger.StartProgress(
		fmt.Sprintf(
			"%s contract '%s' on account '%s'...",
			map[bool]string{true: "Updating", false: "Creating"}[updateExisting],
			name,
			account.Address(),
		),
	)
	defer f.logger.StopProgress()

	// check if contract exists on account
	flowAccount, err := f.gateway.GetAccount(account.Address())
	if err != nil {
		return flow.EmptyID, false, err
	}
	existingContract, exists := flowAccount.Contracts[name]
	noDiffInContract := bytes.Equal(program.Code(), existingContract)
	if exists && noDiffInContract {
		return flow.EmptyID, false, errUpdateNoDiff
	}
	if exists && !updateExisting {
		return flow.EmptyID, false, fmt.Errorf(
			fmt.Sprintf("contract %s exists in account %s", name, account.Name()),
		)
	}

	// if we are updating contract
	if exists && updateExisting {
		tx, err = NewUpdateAccountContractTransaction(
			account,
			name,
			contract.Code(),
		)
		if err != nil {
			return flow.EmptyID, false, err
		}
	}

	tx, err = f.prepareTransaction(tx, account)
	if err != nil {
		return flow.EmptyID, false, err
	}

	f.logger.Info(fmt.Sprintf("Transaction ID: %s", tx.FlowTransaction().ID()))

	// send transaction with contract
	sentTx, err := f.gateway.SendSignedTransaction(tx)
	if err != nil {
		return flow.EmptyID, false, fmt.Errorf("failed to send transaction to deploy a contract: %w", err)
	}

	// we wait for transaction to be sealed
	trx, err := f.gateway.GetTransactionResult(sentTx.ID(), true)
	if err != nil {
		return flow.EmptyID, false, err
	}
	if trx.Error != nil {
		return flow.EmptyID, false, trx.Error
	}

	f.logger.StopProgress()
	f.logger.Info(fmt.Sprintf(
		"Contract '%s' %s on the account '%s'.",
		name,
		map[bool]string{true: "updated", false: "created"}[updateExisting],
		account.Address(),
	))

	return sentTx.ID(), updateExisting, err
}

// RemoveContract from the provided account by its name.
//
// If removal is successful transaction ID is returned.
func (f *Flowkit) RemoveContract(ctx context.Context, account *Account, contractName string) (flow.Identifier, error) {
	// check if contracts exists on the account
	flowAcc, err := f.gateway.GetAccount(account.Address())
	if err != nil {
		return flow.EmptyID, err
	}

	existingContracts := maps.Keys(flowAcc.Contracts)
	if !slices.Contains(existingContracts, contractName) {
		return flow.EmptyID, fmt.Errorf(
			"can not remove a non-existing contract named '%s'. Account only contains the contracts: %v",
			contractName,
			strings.Join(existingContracts, ", "),
		)
	}

	tx, err := NewRemoveAccountContractTransaction(account, contractName)
	if err != nil {
		return flow.EmptyID, err
	}

	tx, err = f.prepareTransaction(tx, account)
	if err != nil {
		return flow.EmptyID, err
	}

	f.logger.Info(fmt.Sprintf("Transaction ID: %s", tx.FlowTransaction().ID().String()))
	f.logger.StartProgress(
		fmt.Sprintf("Removing Contract %s from %s...", contractName, account.Address()),
	)
	defer f.logger.StopProgress()

	sentTx, err := f.gateway.SendSignedTransaction(tx)
	if err != nil {
		return flow.EmptyID, err
	}

	txr, err := f.gateway.GetTransactionResult(sentTx.ID(), true)
	if err != nil {
		return flow.EmptyID, err
	}
	if txr != nil && txr.Error != nil {
		return flow.EmptyID, txr.Error
	}

	f.logger.StopProgress()
	f.logger.Info(fmt.Sprintf(
		"Contract %s removed from account %s.",
		contractName,
		account.Address(),
	))

	return sentTx.ID(), nil
}

func (f *Flowkit) GetBlock(ctx context.Context, query BlockQuery) (*flow.Block, error) {
	f.logger.StartProgress("Fetching Block...")
	defer f.logger.StopProgress()

	// smart parsing of query
	var err error
	var block *flow.Block
	if query.Latest {
		block, err = f.gateway.GetLatestBlock()
	} else if query.Height > 0 {
		block, err = f.gateway.GetBlockByHeight(query.Height)
	} else if query.ID != nil {
		block, err = f.gateway.GetBlockByID(*query.ID)
	} else {
		return nil, fmt.Errorf("invalid query, valid are: \"latest\", block height or block ID")
	}

	if err != nil {
		return nil, fmt.Errorf("error fetching block: %s", err.Error())
	}

	if block == nil {
		return nil, fmt.Errorf("block not found")
	}

	f.logger.StopProgress()

	return block, err
}

func (f *Flowkit) GetCollection(ctx context.Context, ID flow.Identifier) (*flow.Collection, error) {
	return f.gateway.GetCollection(ID)
}

func (f *Flowkit) GetEvents(ctx context.Context, names []string, startHeight uint64, endHeight uint64, worker *EventWorker) ([]flow.BlockEvents, error) {
	if endHeight < startHeight {
		return nil, fmt.Errorf("cannot have end height (%d) of block range less that start height (%d)", endHeight, startHeight)
	}

	f.logger.StartProgress("Fetching events...")
	defer f.logger.StopProgress()

	queries := makeEventQueries(names, startHeight, endHeight, worker.blocksPerWorker)

	jobChan := make(chan grpc.EventRangeQuery, worker.count)
	results := make(chan EventWorkerResult)

	var wg sync.WaitGroup

	for i := 0; i < worker.count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f.eventWorker(jobChan, results)
		}()
	}

	// wait on the workers to finish and close the result channel
	// to signal downstream that all work is done
	go func() {
		defer close(results)
		wg.Wait()
	}()

	go func() {
		defer close(jobChan)
		for _, query := range queries {
			jobChan <- query
		}
	}()

	var resultEvents []flow.BlockEvents
	for eventResult := range results {
		if eventResult.Error != nil {
			return nil, eventResult.Error
		}

		resultEvents = append(resultEvents, eventResult.Events...)
	}

	return resultEvents, nil
}

func (f *Flowkit) eventWorker(jobChan <-chan grpc.EventRangeQuery, results chan<- EventWorkerResult) {
	for q := range jobChan {
		blockEvents, err := f.gateway.GetEvents(q.Type, q.StartHeight, q.EndHeight)
		if err != nil {
			results <- EventWorkerResult{nil, err}
		}
		results <- EventWorkerResult{blockEvents, nil}
	}
}

type EventWorkerResult struct {
	Events []flow.BlockEvents
	Error  error
}

func makeEventQueries(events []string, startHeight uint64, endHeight uint64, blockCount uint64) []grpc.EventRangeQuery {
	var queries []grpc.EventRangeQuery
	for startHeight <= endHeight {
		suggestedEndHeight := startHeight + blockCount - 1 //since we are inclusive
		end := endHeight
		if suggestedEndHeight < endHeight {
			end = suggestedEndHeight
		}
		for _, event := range events {
			queries = append(queries, grpc.EventRangeQuery{
				Type:        event,
				StartHeight: startHeight,
				EndHeight:   end,
			})
		}
		startHeight = suggestedEndHeight + 1
	}
	return queries

}

func (f *Flowkit) GenerateKey(ctx context.Context, inputSeed string, sigAlgo crypto.SignatureAlgorithm) (crypto.PrivateKey, error) {
	//TODO implement me
	panic("implement me")
}

func (f *Flowkit) GenerateMnemonicKey(ctx context.Context, derivationPath string, sigAlgo crypto.SignatureAlgorithm) (crypto.PrivateKey, string, error) {
	//TODO implement me
	panic("implement me")
}

func (f *Flowkit) DeployProject(ctx context.Context, update bool) ([]*project.Contract, error) {
	//TODO implement me
	panic("implement me")
}

func (f *Flowkit) ExecuteScript(ctx context.Context, script *Script) (cadence.Value, error) {
	//TODO implement me
	panic("implement me")
}

func (f *Flowkit) GetTransactionByID(ctx context.Context, ID flow.Identifier, waitSeal bool) (*flow.Transaction, *flow.TransactionResult, error) {
	//TODO implement me
	panic("implement me")
}

func (f *Flowkit) GetTransactionsByBlockID(ctx context.Context, blockID flow.Identifier, waitSeal bool) ([]*flow.Transaction, []*flow.TransactionResult, error) {
	//TODO implement me
	panic("implement me")
}

func (f *Flowkit) BuildTransaction(addresses *transactionAddresses, proposerKeyIndex int, script *Script, gasLimit uint64) (*Transaction, error) {
	//TODO implement me
	panic("implement me")
}

func (f *Flowkit) SignTransactionPayload(signer *Account, payload []byte) (*Transaction, error) {
	//TODO implement me
	panic("implement me")
}

func (f *Flowkit) SendSignedTransaction(tx *Transaction) (*flow.Transaction, *flow.TransactionResult, error) {
	//TODO implement me
	panic("implement me")
}

func (f *Flowkit) SendTransaction(accounts *transactionAccountRoles, script *Script, gasLimit uint64) (*flow.Transaction, *flow.TransactionResult, error) {
	//TODO implement me
	panic("implement me")
}
