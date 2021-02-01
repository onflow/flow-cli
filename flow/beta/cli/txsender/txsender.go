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

package txsender

import (
	"context"
	"time"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"

	"github.com/onflow/flow-cli/flow/beta/cli"
)

const (
	defaultGasLimit           = 1000
	defaultResultPollInterval = time.Second
)

type Sender struct {
	client *client.Client
}

func NewSender(c *client.Client) *Sender {
	return &Sender{
		client: c,
	}
}

func (s *Sender) Send(
	ctx context.Context,
	tx *flow.Transaction,
	signer *cli.Account,
) <-chan Result {

	result := make(chan Result)

	go func() {
		txResult, err := s.send(ctx, tx, signer)
		result <- Result{
			txResult: txResult,
			err:      err,
		}
	}()

	return result
}

func (s *Sender) send(
	ctx context.Context,
	tx *flow.Transaction,
	signer *cli.Account,
) (*flow.TransactionResult, error) {

	latestSealedBlock, err := s.client.GetLatestBlockHeader(ctx, true)
	if err != nil {
		return nil, err
	}

	seqNo, err := s.getSequenceNumber(ctx, signer)
	if err != nil {
		return nil, err
	}

	tx.SetGasLimit(defaultGasLimit).
		SetPayer(signer.Address()).
		SetProposalKey(signer.Address(), signer.DefaultKey().Index(), seqNo).
		SetReferenceBlockID(latestSealedBlock.ID)

	err = tx.SignEnvelope(signer.Address(), signer.DefaultKey().Index(), signer.DefaultKey().Signer())
	if err != nil {
		return nil, err
	}

	err = s.client.SendTransaction(ctx, *tx)
	if err != nil {
		return nil, err
	}

	result, err := s.waitForSeal(ctx, tx.ID())
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Sender) waitForSeal(
	ctx context.Context,
	id flow.Identifier,
) (*flow.TransactionResult, error) {
	result, err := s.client.GetTransactionResult(ctx, id)
	if err != nil {
		return nil, err
	}

	for result.Status != flow.TransactionStatusSealed {
		time.Sleep(defaultResultPollInterval)

		result, err = s.client.GetTransactionResult(ctx, id)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (s *Sender) getSequenceNumber(ctx context.Context, account *cli.Account) (uint64, error) {
	accountResult, err := s.client.GetAccount(ctx, account.Address())
	if err != nil {
		return 0, err
	}

	keyIndex := account.DefaultKey().Index()

	accountKey := accountResult.Keys[keyIndex]

	return accountKey.SequenceNumber, nil
}

type Result struct {
	txResult *flow.TransactionResult
	err      error
}

func (r *Result) Error() error {
	if r.err != nil {
		return r.err
	}

	if r.txResult.Error != nil {
		return r.txResult.Error
	}

	return nil
}

func (r *Result) TransactionResult() *flow.TransactionResult {
	return r.txResult
}
