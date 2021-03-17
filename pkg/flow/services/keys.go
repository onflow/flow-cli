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

package services

import (
	"fmt"

	"github.com/onflow/flow-cli/pkg/flow"

	"github.com/onflow/flow-cli/pkg/flow/util"

	"github.com/onflow/flow-cli/pkg/flow/gateway"
	"github.com/onflow/flow-go-sdk/crypto"
)

// Keys service handles all interactions for keys
type Keys struct {
	gateway gateway.Gateway
	project *flow.Project
	logger  util.Logger
}

// NewTransactions create new transaction service
func NewKeys(
	gateway gateway.Gateway,
	project *flow.Project,
	logger util.Logger,
) *Keys {
	return &Keys{
		gateway: gateway,
		project: project,
		logger:  logger,
	}
}

func (k *Keys) Generate(inputSeed string, signatureAlgo string) (*crypto.PrivateKey, error) {
	var seed []byte
	var err error

	if inputSeed == "" {
		seed, err = flow.RandomSeed(crypto.MinSeedLength)
		if err != nil {
			return nil, err
		}
	} else {
		seed = []byte(inputSeed)
	}

	sigAlgo := crypto.StringToSignatureAlgorithm(signatureAlgo)
	if sigAlgo == crypto.UnknownSignatureAlgorithm {
		return nil, fmt.Errorf("invalid signature algorithm: %s", signatureAlgo)
	}

	privateKey, err := crypto.GeneratePrivateKey(sigAlgo, seed)
	return &privateKey, err
}
