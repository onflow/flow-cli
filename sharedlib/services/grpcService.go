package services

import (
	"context"
	"fmt"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
	"strings"
)

type RpcServices struct {
	client *client.Client
	ctx    context.Context
}

func NewRpcService(host string) (*RpcServices, error) {
	client, err := client.New(host, grpc.WithInsecure())
	ctx := context.Background()

	if err != nil || client == nil {
		return nil, fmt.Errorf("failed to connect to host %s", host)
	}

	return &RpcServices{
		client: client,
		ctx:    ctx,
	}, nil
}

func (s *RpcServices) GetAccount(address string) (*flow.Account, error) {
	flowAddress := flow.HexToAddress(
		strings.ReplaceAll(address, "0x", ""),
	)

	account, err := s.client.GetAccount(s.ctx, flowAddress)
	if err != nil {
		return nil, fmt.Errorf("Failed to get account with address %s: %s", address, err)
	}

	return account, nil
}

/* WIP
func (s *RpcServices) CreateAccount(
	signerAccount *cli.Account,
	keys []string,
	signatureAlgorithm string,
	hashingAlgorithm string,
	contracts []string,
) (*flow.Account, error) {

	accountKeys := make([]*flow.AccountKey, len(keys))

	sigAlgo := crypto.StringToSignatureAlgorithm(signatureAlgorithm)
	if sigAlgo == crypto.UnknownSignatureAlgorithm {
		return nil, fmt.Errorf("Failed to determine signature algorithm from %s", sigAlgo)
	}
	hashAlgo := crypto.StringToHashAlgorithm(hashingAlgorithm)
	if hashAlgo == crypto.UnknownHashAlgorithm {
		return nil, fmt.Errorf("Failed to determine hash algorithm from %s", hashAlgo)
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
			return nil, fmt.Errorf( "Failed to read contract name and path from flag. Ensure you're providing a contract name and a file path. %s", contract)
		}
		contractName := contractFlagContent[0]
		contractPath := contractFlagContent[1]
		contractSource, err := ioutil.ReadFile(contractPath)
		if err != nil {
			return nil, fmt.Errorf("Failed to read contract from source file %s", contractPath)
		}
		contractTemplates = append(contractTemplates,
			templates.Contract{
				Name:   contractName,
				Source: string(contractSource),
			},
		)
	}

	f.CreateTransaction()
	tx := templates.CreateAccount(accountKeys, contractTemplates, signerAccount.Address())

	result := flow.SendTransaction(s.client, s.ctx, signerAccount, tx)

	return nil, nil
}
*/
