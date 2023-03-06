package services

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/onflow/cadence"
	tmpl "github.com/onflow/flow-core-contracts/lib/go/templates"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/util"
)

// CLIServices interface defines functions that are used within the Flow CLI but are not useful to be
// part of the flowkit interface. Reasons for deciding the location should be to think if this is
// something that is specialised to usage in CLI or is something others will want to use as part of
// the flowkit SDK. If the function seems a bit specialised then define it here.
type CLIServices interface {
	StakingInfo(address flow.Address) ([]map[string]interface{}, []map[string]interface{}, error)
	NodeTotalStake(nodeId string, chain flow.ChainID) (*cadence.Value, error)
	DecodePEMKey(key string, sigAlgo crypto.SignatureAlgorithm) (*flow.AccountKey, error)
	DecodeRLPKey(publicKey string) (*flow.AccountKey, error)
	CheckForStandardContractUsageOnMainnet() error
	GetLatestProtocolStateSnapshot() ([]byte, error)
	GetRLPTransaction(rlpUrl string) ([]byte, error)
	PostRLPTransaction(rlpUrl string, tx *flow.Transaction) error
}

var _ CLIServices = &Services{}

func NewInternal(state *flowkit.State, gateway gateway.Gateway, logger output.Logger) *Services {
	return &Services{
		state, gateway, logger,
	}
}

type Services struct {
	state   *flowkit.State
	gateway gateway.Gateway
	logger  output.Logger
}

func (s *Services) StakingInfo(address flow.Address) ([]map[string]interface{}, []map[string]interface{}, error) {
	s.logger.StartProgress(fmt.Sprintf("Fetching info for %s...", address.String()))
	defer s.logger.StopProgress()

	cadenceAddress := []cadence.Value{cadence.NewAddress(address)}

	chain, err := util.GetAddressNetwork(address)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to determine network from address, check the address and network",
		)
	}

	if chain == flow.Emulator {
		return nil, nil, fmt.Errorf("emulator chain not supported")
	}

	env := util.EnvFromNetwork(chain)

	stakingInfoScript := tmpl.GenerateCollectionGetAllNodeInfoScript(env)
	delegationInfoScript := tmpl.GenerateCollectionGetAllDelegatorInfoScript(env)

	stakingValue, err := s.gateway.ExecuteScript(stakingInfoScript, cadenceAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting staking info: %s", err.Error())
	}

	delegationValue, err := s.gateway.ExecuteScript(delegationInfoScript, cadenceAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting delegation info: %s", err.Error())
	}

	// get staking infos and delegation infos
	stakingInfos, err := flowkit.NewStakingInfoFromValue(stakingValue)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing staking info: %s", err.Error())
	}
	delegationInfos, err := flowkit.NewStakingInfoFromValue(delegationValue)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing delegation info: %s", err.Error())
	}

	// get a set of node ids from all staking infos
	nodeStakes := make(map[string]cadence.Value)
	for _, stakingInfo := range stakingInfos {
		nodeID, ok := stakingInfo["id"]
		if ok {
			nodeStakes[nodeIDToString(nodeID)] = nil
		}
	}
	totalCommitmentScript := tmpl.GenerateGetTotalCommitmentBalanceScript(env)

	// foreach node id, get the node total stake
	for nodeID := range nodeStakes {
		stake, err := s.gateway.ExecuteScript(totalCommitmentScript, []cadence.Value{cadence.String(nodeID)})
		if err != nil {
			return nil, nil, fmt.Errorf("error getting total stake for node: %s", err.Error())
		}

		nodeStakes[nodeID] = stake
	}

	// foreach staking info, add the node total stake
	for _, stakingInfo := range stakingInfos {
		nodeID, ok := stakingInfo["id"]
		if ok {
			stakingInfo["nodeTotalStake"] = nodeStakes[nodeIDToString(nodeID)].(cadence.UFix64)
		}
	}

	s.logger.StopProgress()

	return stakingInfos, delegationInfos, nil
}

func nodeIDToString(value interface{}) string {
	return value.(cadence.String).ToGoValue().(string)
}

func (s *Services) NodeTotalStake(nodeId string, chain flow.ChainID) (*cadence.Value, error) {
	s.logger.StartProgress(fmt.Sprintf("Fetching total stake for node id %s...", nodeId))
	defer s.logger.StopProgress()

	if chain == flow.Emulator {
		return nil, fmt.Errorf("emulator chain not supported")
	}

	env := util.EnvFromNetwork(chain)

	stakingInfoScript := tmpl.GenerateGetTotalCommitmentBalanceScript(env)
	stakingValue, err := s.gateway.ExecuteScript(stakingInfoScript, []cadence.Value{cadence.String(nodeId)})
	if err != nil {
		return nil, fmt.Errorf("error getting total stake for node: %s", err.Error())
	}

	s.logger.StopProgress()

	return &stakingValue, nil
}

func (s *Services) DecodePEMKey(key string, sigAlgo crypto.SignatureAlgorithm) (*flow.AccountKey, error) {
	pk, err := crypto.DecodePublicKeyPEM(sigAlgo, key)
	if err != nil {
		return nil, err
	}

	return &flow.AccountKey{
		PublicKey: pk,
		SigAlgo:   sigAlgo,
		Weight:    -1,
	}, nil
}

func (s *Services) DecodeRLPKey(publicKey string) (*flow.AccountKey, error) {
	publicKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	accountKey, err := flow.DecodeAccountKey(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %w", err)
	}

	return accountKey, nil
}

func (s *Services) CheckForStandardContractUsageOnMainnet() error {
	mainnetContracts := map[string]standardContract{
		"FungibleToken": {
			name:     "FungibleToken",
			address:  flow.HexToAddress("0xf233dcee88fe0abe"),
			infoLink: "https://developers.flow.com/flow/core-contracts/fungible-token",
		},
		"FlowToken": {
			name:     "FlowToken",
			address:  flow.HexToAddress("0x1654653399040a61"),
			infoLink: "https://developers.flow.com/flow/core-contracts/flow-token",
		},
		"FlowFees": {
			name:     "FlowFees",
			address:  flow.HexToAddress("0xf919ee77447b7497"),
			infoLink: "https://developers.flow.com/flow/core-contracts/flow-fees",
		},
		"FlowServiceAccount": {
			name:     "FlowServiceAccount",
			address:  flow.HexToAddress("0xe467b9dd11fa00df"),
			infoLink: "https://developers.flow.com/flow/core-contracts/service-account",
		},
		"FlowStorageFees": {
			name:     "FlowStorageFees",
			address:  flow.HexToAddress("0xe467b9dd11fa00df"),
			infoLink: "https://developers.flow.com/flow/core-contracts/service-account",
		},
		"FlowIDTableStaking": {
			name:     "FlowIDTableStaking",
			address:  flow.HexToAddress("0x8624b52f9ddcd04a"),
			infoLink: "https://developers.flow.com/flow/core-contracts/staking-contract-reference",
		},
		"FlowEpoch": {
			name:     "FlowEpoch",
			address:  flow.HexToAddress("0x8624b52f9ddcd04a"),
			infoLink: "https://developers.flow.com/flow/core-contracts/epoch-contract-reference",
		},
		"FlowClusterQC": {
			name:     "FlowClusterQC",
			address:  flow.HexToAddress("0x8624b52f9ddcd04a"),
			infoLink: "https://developers.flow.com/flow/core-contracts/epoch-contract-reference",
		},
		"FlowDKG": {
			name:     "FlowDKG",
			address:  flow.HexToAddress("0x8624b52f9ddcd04a"),
			infoLink: "https://developers.flow.com/flow/core-contracts/epoch-contract-reference",
		},
		"NonFungibleToken": {
			name:     "NonFungibleToken",
			address:  flow.HexToAddress("0x1d7e57aa55817448"),
			infoLink: "https://developers.flow.com/flow/core-contracts/non-fungible-token",
		},
		"MetadataViews": {
			name:     "MetadataViews",
			address:  flow.HexToAddress("0x1d7e57aa55817448"),
			infoLink: "https://developers.flow.com/flow/core-contracts/nft-metadata",
		},
	}

	contracts, err := s.state.DeploymentContractsByNetwork(config.MainnetNetwork)
	if err != nil {
		return err
	}

	for _, contract := range contracts {
		standardContract, ok := mainnetContracts[contract.Name]
		if !ok {
			continue
		}

		s.logger.Info(fmt.Sprintf("It seems like you are trying to deploy %s to Mainnet \n", contract.Name))
		s.logger.Info(fmt.Sprintf("It is a standard contract already deployed at address 0x%s \n", standardContract.address.String()))
		s.logger.Info(fmt.Sprintf("You can read more about it here: %s \n", standardContract.infoLink))

		if output.WantToUseMainnetVersionPrompt() {
			err := s.replaceStandardContractReferenceToAlias(standardContract)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type standardContract struct {
	name     string
	address  flow.Address
	infoLink string
}

func (s *Services) replaceStandardContractReferenceToAlias(standardContract standardContract) error {
	//replace contract with alias
	contract := s.state.Config().Contracts.ByName(standardContract.name)
	if contract == nil {
		return fmt.Errorf("contract not found") // shouldn't occur
	}
	contract.Aliases.Add(config.MainnetNetwork.Name, standardContract.address)

	//remove from deploy
	for di, d := range s.state.Config().Deployments {
		if d.Network != config.MainnetNetwork.Name {
			continue
		}
		for ci, c := range d.Contracts {
			if c.Name == standardContract.name {
				s.state.Config().Deployments[di].Contracts = append((d.Contracts)[0:ci], (d.Contracts)[ci+1:]...)
				break
			}
		}
	}
	return nil
}

func (s *Services) GetLatestProtocolStateSnapshot() ([]byte, error) {
	s.logger.StartProgress("Downloading protocol snapshot...")

	if !s.gateway.SecureConnection() {
		s.logger.Info(fmt.Sprintf("%s warning: using insecure client connection to download snapshot, you should use a secure network configuration...", output.WarningEmoji()))
	}

	b, err := s.gateway.GetLatestProtocolStateSnapshot()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest finalized protocol snapshot from gateway: %w", err)
	}

	s.logger.StopProgress()

	return b, nil
}

func (s *Services) GetRLPTransaction(rlpUrl string) ([]byte, error) {
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	resp, err := client.Get(rlpUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error downloading RLP identifier")
	}

	return io.ReadAll(resp.Body)
}

func (s *Services) PostRLPTransaction(rlpUrl string, tx *flow.Transaction) error {
	signedRlp := hex.EncodeToString(tx.Encode())
	resp, err := http.Post(rlpUrl, "application/text", bytes.NewBufferString(signedRlp))

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error posting signed RLP")
	}

	return nil
}
