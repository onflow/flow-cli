package services

import (
	"fmt"

	"github.com/onflow/flow-cli/flow/lib"

	"github.com/onflow/flow-cli/flow/util"

	"github.com/onflow/flow-cli/flow/gateway"
	"github.com/onflow/flow-go-sdk/crypto"
)

// Keys service handles all interactions for keys
type Keys struct {
	gateway gateway.Gateway
	project *lib.Project
	logger  util.Logger
}

// NewTransactions create new transaction service
func NewKeys(
	gateway gateway.Gateway,
	project *lib.Project,
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
		seed, err = lib.RandomSeed(crypto.MinSeedLength)
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
