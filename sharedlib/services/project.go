package services

import (
	"fmt"
	"strings"

	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flow-go-sdk/templates"

	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/flow/project/contracts"
	"github.com/onflow/flow-cli/sharedlib/gateway"
	"github.com/onflow/flow-go-sdk"
)

// Project service handles all interactions for project
type Project struct {
	gateway gateway.Gateway
	project *cli.Project
}

// NewProject create new project service
func NewProject(gateway gateway.Gateway, project *cli.Project) *Project {
	return &Project{
		gateway: gateway,
		project: project,
	}
}

func (p *Project) Init(reset bool, serviceKeySigAlgo string, serviceKeyHashAlgo string, servicePrivateKey string) (*cli.Project, error) {
	if !cli.ProjectExists(cli.DefaultConfigPath) || reset {
		serviceKeySigAlgo := crypto.StringToSignatureAlgorithm(serviceKeySigAlgo)
		serviceKeyHashAlgo := crypto.StringToHashAlgorithm(serviceKeyHashAlgo)

		project := cli.InitProject(serviceKeySigAlgo, serviceKeyHashAlgo)

		if len(servicePrivateKey) > 0 {
			serviceKey := cli.MustDecodePrivateKeyHex(serviceKeySigAlgo, servicePrivateKey)
			project.SetEmulatorServiceKey(serviceKey)
		}

		project.Save(cli.DefaultConfigPath)

		return project, nil
	} else {
		return nil, fmt.Errorf("configuration already exists at: %s, if you want to reset configuration use the reset flag", cli.DefaultConfigPath)
	}
}

func (p *Project) Deploy(network string, update bool) ([]*contracts.Contract, error) {
	// check there are not multiple accounts with same contract
	// TODO: specify which contract by name is a problem
	if p.project.ContractConflictExists(network) {
		return nil, fmt.Errorf("currently it is not possible to deploy same contract with multiple accounts, please check Deployments in config and make sure a contract is only present in one account")
	}

	processor := contracts.NewPreprocessor(
		contracts.FilesystemLoader{},
		p.project.GetAliases(network),
	)

	for _, contract := range p.project.GetContractsByNetwork(network) {
		err := processor.AddContractSource(
			contract.Name,
			contract.Source,
			contract.Target,
		)
		if err != nil {
			return nil, err
		}
	}

	err := processor.ResolveImports()
	if err != nil {
		return nil, err
	}

	contracts, err := processor.ContractDeploymentOrder()
	if err != nil {
		return nil, err
	}

	// TODO: change to log
	fmt.Printf(
		"Deploying %v contracts for accounts: %s\n",
		len(contracts),
		strings.Join(p.project.GetAllAccountNames(), ","),
	)

	var errs []error

	for _, contract := range contracts {
		targetAccount := p.project.GetAccountByAddress(contract.Target().String())

		targetAccountInfo, err := p.gateway.GetAccount(targetAccount.Address())
		if err != nil {
			return nil, fmt.Errorf("failed to fetch information for account %s", targetAccount.Address())
		}

		var tx *flow.Transaction

		_, exists := targetAccountInfo.Contracts[contract.Name()]
		if exists {
			if !update {
				continue
			}

			tx = prepareUpdateContractTransaction(targetAccount.Address(), contract)
		} else {
			tx = prepareAddContractTransaction(targetAccount.Address(), contract)
		}

		tx, err = p.gateway.SendTransaction(tx, targetAccount)

		// TODO: part of verbose logging
		spinner := cli.NewSpinner(
			fmt.Sprintf("%s ", cli.Bold(contract.Name())),
			" deploying...",
		)
		spinner.Start()

		result, err := p.gateway.GetTransactionResult(tx)

		// TODO: part of verbose logging
		if result.Error == nil {
			spinner.Stop(fmt.Sprintf("%s -> 0x%s", cli.Green(contract.Name()), contract.Target()))
		} else {
			spinner.Stop(fmt.Sprintf("%s error", cli.Red(contract.Name())))
			errs = append(errs, result.Error)
		}
	}

	// TODO: part of logging
	if len(errs) == 0 {
		fmt.Println("\n✅ All contracts deployed successfully")
	} else {
		// REF: better output when errors
		fmt.Println("\n❌ Failed to deploy all contracts")
	}

	return contracts, nil
}

func prepareUpdateContractTransaction(
	targetAccount flow.Address,
	contract *contracts.Contract,
) *flow.Transaction {
	return templates.UpdateAccountContract(
		targetAccount,
		templates.Contract{
			Name:   contract.Name(),
			Source: contract.TranspiledCode(),
		},
	)
}

func prepareAddContractTransaction(
	targetAccount flow.Address,
	contract *contracts.Contract,
) *flow.Transaction {
	return templates.AddAccountContract(
		targetAccount,
		templates.Contract{
			Name:   contract.Name(),
			Source: contract.TranspiledCode(),
		},
	)
}
