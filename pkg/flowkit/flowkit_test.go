package flowkit

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/cadence/runtime/stdlib"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/project"
	"github.com/onflow/flow-cli/pkg/flowkit/tests"
	emulator "github.com/onflow/flow-emulator"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/access/grpc"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"os"
	"strings"
	"testing"
)

func setup() (*State, Flowkit, *tests.TestGateway) {
	readerWriter, _ := tests.ReaderWriter()
	state, err := Init(readerWriter, crypto.ECDSA_P256, crypto.SHA3_256)
	if err != nil {
		panic(err)
	}

	gw := tests.DefaultMockGateway()
	flowkit := Flowkit{
		state:   state,
		network: config.DefaultEmulatorNetwork(),
		gateway: gw,
		logger:  output.NewStdoutLogger(output.NoneLog),
	}

	return state, flowkit, gw
}

func resourceToContract(res tests.Resource) *Script {
	return NewScript(res.Source, nil, res.Filename)
}

var ctx = context.Background()

func TestAccounts(t *testing.T) {
	state, _, _ := setup()
	pubKey, _ := crypto.DecodePublicKeyHex(crypto.ECDSA_P256, "858a7d978b25d61f348841a343f79131f4b9fab341dd8a476a6f4367c25510570bf69b795fc9c3d2b7191327d869bcf848508526a3c1cafd1af34f71c7765117")
	serviceAcc, _ := state.EmulatorServiceAccount()
	serviceAddress := serviceAcc.Address()

	t.Run("Get an Account", func(t *testing.T) {
		_, flowkit, gw := setup()
		account, err := flowkit.GetAccount(ctx, serviceAddress)

		gw.Mock.AssertCalled(t, "GetAccount", serviceAddress)
		assert.NoError(t, err)
		assert.Equal(t, serviceAddress, account.Address)
	})

	t.Run("Create an Account", func(t *testing.T) {
		_, flowkit, gw := setup()
		newAddress := flow.HexToAddress("192440c99cb17282")

		gw.SendSignedTransaction.Run(func(args mock.Arguments) {
			tx := args.Get(0).(*Transaction)
			assert.Equal(t, serviceAddress, tx.FlowTransaction().Authorizers[0])
			assert.Equal(t, serviceAddress, tx.Signer().Address())

			gw.SendSignedTransaction.Return(tests.NewTransaction(), nil)
		})

		compareAddress := serviceAddress
		gw.GetAccount.Run(func(args mock.Arguments) {
			address := args.Get(0).(flow.Address)
			assert.Equal(t, address, compareAddress)
			compareAddress = newAddress
			gw.GetAccount.Return(
				tests.NewAccountWithAddress(address.String()), nil,
			)
		})

		gw.GetTransactionResult.Return(
			tests.NewAccountCreateResult(newAddress), nil,
		)

		account, ID, err := flowkit.CreateAccount(
			ctx,
			serviceAcc,
			[]Key{{
				pubKey,
				flow.AccountKeyWeightThreshold,
				crypto.ECDSA_P256,
				crypto.SHA3_256,
			}},
		)

		gw.Mock.AssertCalled(t, tests.GetAccountFunc, serviceAddress)
		gw.Mock.AssertCalled(t, tests.GetAccountFunc, newAddress)
		gw.Mock.AssertNumberOfCalls(t, tests.GetAccountFunc, 2)
		gw.Mock.AssertNumberOfCalls(t, tests.GetTransactionResultFunc, 1)
		gw.Mock.AssertNumberOfCalls(t, tests.SendSignedTransactionFunc, 1)
		assert.NotNil(t, account)
		assert.NotNil(t, ID)
		assert.Equal(t, account.Address, newAddress)
		assert.NoError(t, err)
	})

	t.Run("Contract Add for Account", func(t *testing.T) {
		_, flowkit, gw := setup()
		gw.SendSignedTransaction.Run(func(args mock.Arguments) {
			tx := args.Get(0).(*Transaction)
			assert.Equal(t, tx.Signer().Address(), serviceAddress)
			assert.True(t, strings.Contains(string(tx.FlowTransaction().Script), "signer.contracts.add"))

			gw.SendSignedTransaction.Return(tests.NewTransaction(), nil)
		})

		ID, _, err := flowkit.AddContract(
			ctx,
			serviceAcc,
			resourceToContract(tests.ContractHelloString),
			false,
		)

		gw.Mock.AssertCalled(t, tests.GetAccountFunc, serviceAddress)
		gw.Mock.AssertNumberOfCalls(t, tests.GetAccountFunc, 2)
		gw.Mock.AssertNumberOfCalls(t, tests.GetTransactionResultFunc, 1)
		gw.Mock.AssertNumberOfCalls(t, tests.SendSignedTransactionFunc, 1)
		assert.NotNil(t, ID)
		assert.NoError(t, err)
	})

	t.Run("Contract Remove for Account", func(t *testing.T) {
		_, flowkit, gw := setup()
		gw.SendSignedTransaction.Run(func(args mock.Arguments) {
			tx := args.Get(0).(*Transaction)
			assert.Equal(t, tx.Signer().Address(), serviceAddress)
			assert.True(t, strings.Contains(string(tx.FlowTransaction().Script), "signer.contracts.remove"))

			gw.SendSignedTransaction.Return(tests.NewTransaction(), nil)
		})

		gw.GetAccount.Run(func(args mock.Arguments) {
			addr := args.Get(0).(flow.Address)
			assert.Equal(t, addr.String(), serviceAcc.Address().String())
			racc := tests.NewAccountWithAddress(addr.String())
			racc.Contracts = map[string][]byte{
				tests.ContractHelloString.Name: tests.ContractHelloString.Source,
			}

			gw.GetAccount.Return(racc, nil)
		})

		account, err := flowkit.RemoveContract(
			ctx,
			serviceAcc,
			tests.ContractHelloString.Name,
		)

		gw.Mock.AssertCalled(t, tests.GetAccountFunc, serviceAddress)
		gw.Mock.AssertNumberOfCalls(t, tests.GetAccountFunc, 2)
		gw.Mock.AssertNumberOfCalls(t, tests.GetTransactionResultFunc, 1)
		gw.Mock.AssertNumberOfCalls(t, tests.SendSignedTransactionFunc, 1)
		assert.NotNil(t, account)
		assert.NoError(t, err)
	})

	t.Run("Staking Info for Account", func(t *testing.T) {
		_, flowkit, gw := setup()

		count := 0
		gw.ExecuteScript.Run(func(args mock.Arguments) {
			count++
			assert.True(t, strings.Contains(string(args.Get(0).([]byte)), "import FlowIDTableStaking from 0x9eca2b38b18b5dfe"))
			gw.ExecuteScript.Return(cadence.NewArray([]cadence.Value{}), nil)
		})

		val1, val2, err := flowkit.Accounts.StakingInfo(flow.HexToAddress("df9c30eb2252f1fa"))
		assert.NoError(t, err)
		assert.NotNil(t, val1)
		assert.NotNil(t, val2)
		assert.Equal(t, 2, count)
	})
	t.Run("Staking Info for Account fetches node total", func(t *testing.T) {
		_, flowkit, gw := setup()

		count := 0
		gw.ExecuteScript.Run(func(args mock.Arguments) {
			assert.True(t, strings.Contains(string(args.Get(0).([]byte)), "import FlowIDTableStaking from 0x9eca2b38b18b5dfe"))
			if count < 2 {
				gw.ExecuteScript.Return(cadence.NewArray(
					[]cadence.Value{
						cadence.Struct{
							StructType: &cadence.StructType{
								Fields: []cadence.Field{
									{
										Identifier: "id",
									},
								},
							},
							Fields: []cadence.Value{
								cadence.String("8f4d09dae7918afbf62c48fa968a9e8b0891cee8442065fa47cc05f4bc9a8a91"),
							},
						},
					}), nil)
			} else {
				assert.True(t, strings.Contains(args.Get(1).([]cadence.Value)[0].String(), "8f4d09dae7918afbf62c48fa968a9e8b0891cee8442065fa47cc05f4bc9a8a91"))
				gw.ExecuteScript.Return(cadence.NewUFix64("1.0"))
			}
			count++
		})

		val1, val2, err := s.Accounts.StakingInfo(flow.HexToAddress("df9c30eb2252f1fa"))
		assert.NoError(t, err)
		assert.NotNil(t, val1)
		assert.NotNil(t, val2)
		assert.Equal(t, 3, count)
	})
}

func setupIntegration() (*State, Flowkit) {
	readerWriter, _ := tests.ReaderWriter()
	state, err := Init(readerWriter, crypto.ECDSA_P256, crypto.SHA3_256)
	if err != nil {
		panic(err)
	}

	acc, _ := state.EmulatorServiceAccount()
	gw := gateway.NewEmulatorGatewayWithOpts(acc, gateway.WithEmulatorOptions(
		emulator.WithTransactionExpiry(10),
	))

	flowkit := Flowkit{
		state:   state,
		network: config.DefaultEmulatorNetwork(),
		gateway: gw,
		logger:  output.NewStdoutLogger(output.NoneLog),
	}

	return state, flowkit
}

func TestAccountsCreate_Integration(t *testing.T) {
	t.Parallel()

	type accountsIn struct {
		account  *Account
		pubKeys  []crypto.PublicKey
		weights  []int
		sigAlgo  []crypto.SignatureAlgorithm
		hashAlgo []crypto.HashAlgorithm
	}

	type accountsOut struct {
		address string
		code    map[string][]byte
		balance uint64
		pubKeys []crypto.PublicKey
		weights []int
	}

	t.Run("Create", func(t *testing.T) {
		t.Parallel()

		state, flowkit := setupIntegration()
		srvAcc, _ := state.EmulatorServiceAccount()

		accIn := []accountsIn{{
			account: srvAcc,
			sigAlgo: []crypto.SignatureAlgorithm{
				tests.SigAlgos()[0],
			},
			hashAlgo: []crypto.HashAlgorithm{
				tests.HashAlgos()[0],
			},
			pubKeys: []crypto.PublicKey{
				tests.PubKeys()[0],
			},
			weights: []int{flow.AccountKeyWeightThreshold},
		}, {
			account: srvAcc,
			sigAlgo: []crypto.SignatureAlgorithm{
				tests.SigAlgos()[0],
				tests.SigAlgos()[1],
			},
			hashAlgo: []crypto.HashAlgorithm{
				tests.HashAlgos()[0],
				tests.HashAlgos()[1],
			},
			pubKeys: []crypto.PublicKey{
				tests.PubKeys()[0],
				tests.PubKeys()[1],
			},
			weights: []int{500, 500},
		}, {
			account: srvAcc,
			sigAlgo: []crypto.SignatureAlgorithm{
				tests.SigAlgos()[0],
			},
			hashAlgo: []crypto.HashAlgorithm{
				tests.HashAlgos()[0],
			},
			pubKeys: []crypto.PublicKey{
				tests.PubKeys()[0],
			},
			weights: []int{flow.AccountKeyWeightThreshold},
		}}

		accOut := []accountsOut{{
			address: "01cf0e2f2f715450",
			code:    map[string][]byte{},
			balance: uint64(100000),
			pubKeys: []crypto.PublicKey{
				tests.PubKeys()[0],
			},
			weights: []int{flow.AccountKeyWeightThreshold},
		}, {
			address: "179b6b1cb6755e31",
			code:    map[string][]byte{},
			balance: uint64(100000),
			pubKeys: []crypto.PublicKey{
				tests.PubKeys()[0],
				tests.PubKeys()[1],
			},
			weights: []int{500, 500},
		}, {
			address: "f3fcd2c1a78f5eee",
			code: map[string][]byte{
				tests.ContractSimple.Name: tests.ContractSimple.Source,
			},
			balance: uint64(100000),
			pubKeys: []crypto.PublicKey{
				tests.PubKeys()[0],
			},
			weights: []int{flow.AccountKeyWeightThreshold},
		}}

		for i, a := range accIn {
			keys := make([]Key, 0)
			for j := range a.pubKeys {
				keys = append(keys, Key{
					public:   a.pubKeys[j],
					weight:   a.weights[j],
					sigAlgo:  a.sigAlgo[j],
					hashAlgo: a.hashAlgo[j],
				})
			}

			acc, ID, err := flowkit.CreateAccount(ctx, a.account, keys)
			c := accOut[i]

			assert.NoError(t, err)
			assert.NotNil(t, acc)
			assert.NotNil(t, ID)
			assert.Equal(t, acc.Address.String(), c.address)
			assert.Equal(t, acc.Contracts, c.code)
			assert.Equal(t, acc.Balance, c.balance)
			assert.Len(t, acc.Keys, len(c.pubKeys))

			for x, k := range acc.Keys {
				assert.Equal(t, k.PublicKey, a.pubKeys[x])
				assert.Equal(t, k.Weight, c.weights[x])
				assert.Equal(t, k.SigAlgo, a.sigAlgo[x])
				assert.Equal(t, k.HashAlgo, a.hashAlgo[x])
			}

		}

	})

	t.Run("Create Invalid", func(t *testing.T) {
		t.Parallel()

		state, flowkit := setupIntegration()
		srvAcc, _ := state.EmulatorServiceAccount()

		errOut := []string{
			"invalid account key: signing algorithm (UNKNOWN) and hashing algorithm (SHA3_256) are not a valid pair for a Flow account key",
			"invalid account key: signing algorithm (UNKNOWN) and hashing algorithm (UNKNOWN) are not a valid pair for a Flow account key",
			"number of keys and weights provided must match, number of provided keys: 2, number of provided key weights: 1",
			"number of keys and weights provided must match, number of provided keys: 1, number of provided key weights: 2",
		}

		accIn := []accountsIn{
			{
				account:  srvAcc,
				sigAlgo:  []crypto.SignatureAlgorithm{crypto.UnknownSignatureAlgorithm},
				hashAlgo: []crypto.HashAlgorithm{crypto.SHA3_256},
				pubKeys: []crypto.PublicKey{
					tests.PubKeys()[0],
				},
				weights: []int{1000},
			}, {
				account:  srvAcc,
				sigAlgo:  []crypto.SignatureAlgorithm{crypto.UnknownSignatureAlgorithm},
				hashAlgo: []crypto.HashAlgorithm{crypto.UnknownHashAlgorithm},
				pubKeys: []crypto.PublicKey{
					tests.PubKeys()[0],
				},
				weights: []int{1000},
			}, {
				account:  srvAcc,
				sigAlgo:  []crypto.SignatureAlgorithm{crypto.ECDSA_P256},
				hashAlgo: []crypto.HashAlgorithm{crypto.SHA3_256},
				pubKeys: []crypto.PublicKey{
					tests.PubKeys()[0],
					tests.PubKeys()[1],
				},
				weights: []int{1000},
			}, {
				account:  srvAcc,
				sigAlgo:  []crypto.SignatureAlgorithm{crypto.ECDSA_P256},
				hashAlgo: []crypto.HashAlgorithm{crypto.SHA3_256},
				pubKeys: []crypto.PublicKey{
					tests.PubKeys()[0],
				},
				weights: []int{1000, 1000},
			},
			/*{
			 	TODO(sideninja): uncomment this test case after https://github.com/onflow/flow-go-sdk/pull/199 is released
				account:  srvAcc,
				sigAlgo:  crypto.ECDSA_P256,
				hashAlgo: crypto.SHA3_256,
				args:     nil,
				pubKeys: []crypto.PublicKey{
					tests.PubKeys()[0],
				},
				weights: []int{-1},
			}*/
		}

		for i, a := range accIn {
			keys := make([]Key, 0)
			for i := range a.pubKeys {
				keys = append(keys, Key{
					public:   a.pubKeys[i],
					weight:   a.weights[i],
					sigAlgo:  a.sigAlgo[i],
					hashAlgo: a.hashAlgo[i],
				})
			}
			acc, ID, err := flowkit.CreateAccount(ctx, a.account, keys)
			errMsg := errOut[i]

			assert.Nil(t, acc)
			assert.Nil(t, ID)
			assert.Error(t, err)
			assert.Equal(t, errMsg, err.Error())
		}
	})

}

func TestAccountsAddContract_Integration(t *testing.T) {
	t.Parallel()

	t.Run("Update Contract", func(t *testing.T) {
		t.Parallel()

		state, flowkit := setupIntegration()
		srvAcc, _ := state.EmulatorServiceAccount()

		ID, _, err := flowkit.AddContract(
			ctx,
			srvAcc,
			resourceToContract(tests.ContractSimple),
			false,
		)
		require.NoError(t, err)
		require.NotNil(t, ID)

		acc, err := flowkit.GetAccount(ctx, srvAcc.Address())
		require.NoError(t, err)
		require.NotNil(t, acc)
		assert.Equal(t, acc.Contracts["Simple"], tests.ContractSimple.Source)

		ID, _, err = flowkit.AddContract(
			ctx,
			srvAcc,
			resourceToContract(tests.ContractSimpleUpdated),
			true,
		)
		require.NoError(t, err)

		acc, err = flowkit.GetAccount(ctx, srvAcc.Address())
		require.NoError(t, err)
		assert.Equal(t, acc.Contracts["Simple"], tests.ContractSimpleUpdated.Source)
	})

	t.Run("Add Contract Invalid", func(t *testing.T) {
		t.Parallel()

		state, flowkit := setupIntegration()
		srvAcc, _ := state.EmulatorServiceAccount()

		// prepare existing contract
		_, _, err := flowkit.AddContract(
			ctx,
			srvAcc,
			resourceToContract(tests.ContractSimple),
			false,
		)
		assert.NoError(t, err)

		_, _, err = flowkit.AddContract(
			ctx,
			srvAcc,
			resourceToContract(tests.ContractSimple),
			false,
		)

		require.Error(t, err)
		assert.Error(t, err, "cannot overwrite existing contract with name \"Simple\"")
	})
}

func TestAccountsAddContractWithArgs(t *testing.T) {
	state, flowkit := setupIntegration()
	srvAcc, _ := state.EmulatorServiceAccount()

	//adding contract without argument should return an error
	_, _, err := flowkit.AddContract(
		ctx,
		srvAcc,
		resourceToContract(tests.ContractSimpleWithArgs),
		false,
	)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "invalid argument count, too few arguments: expected 1, got 0"))

	c := resourceToContract(tests.ContractSimpleWithArgs)
	c.Args = []cadence.Value{cadence.UInt64(4)}

	_, _, err = flowkit.AddContract(ctx, srvAcc, c, false)
	assert.NoError(t, err)

	acc, err := flowkit.GetAccount(ctx, srvAcc.Address())
	require.NoError(t, err)
	assert.NotNil(t, acc)
	assert.Equal(t, acc.Contracts["Simple"], tests.ContractSimpleWithArgs.Source)
}

func TestAccountsRemoveContract_Integration(t *testing.T) {
	t.Parallel()

	state, flowkit := setupIntegration()
	srvAcc, _ := state.EmulatorServiceAccount()

	c := tests.ContractSimple
	// prepare existing contract
	_, _, err := flowkit.AddContract(
		ctx,
		srvAcc,
		NewScript(c.Source, nil, c.Filename),
		false,
	)
	assert.NoError(t, err)

	t.Run("Remove Contract", func(t *testing.T) {
		t.Parallel()

		_, err := flowkit.RemoveContract(ctx, srvAcc, tests.ContractSimple.Name)
		require.NoError(t, err)

		acc, err := flowkit.GetAccount(ctx, srvAcc.Address())
		require.NoError(t, err)
		assert.Equal(t, acc.Contracts[tests.ContractSimple.Name], []byte(nil))
	})
}

func TestAccountsGet_Integration(t *testing.T) {
	t.Parallel()

	state, flowkit := setupIntegration()
	srvAcc, _ := state.EmulatorServiceAccount()

	t.Run("Get Account", func(t *testing.T) {
		t.Parallel()
		acc, err := flowkit.GetAccount(ctx, srvAcc.Address())

		assert.NoError(t, err)
		assert.NotNil(t, acc)
		assert.Equal(t, acc.Address, srvAcc.Address())
	})

	t.Run("Get Account Invalid", func(t *testing.T) {
		t.Parallel()

		acc, err := flowkit.GetAccount(ctx, flow.HexToAddress("0x1"))
		assert.Nil(t, acc)
		assert.Equal(t, err.Error(), "could not find account with address 0000000000000001")
	})
}

func TestAccountsStakingInfo_Integration(t *testing.T) {
	t.Parallel()
	state, flowkit := setupIntegration()
	srvAcc, _ := state.EmulatorServiceAccount()

	t.Run("Get Staking Info", func(t *testing.T) {
		_, _, err := s.Accounts.StakingInfo(srvAcc.Address()) // unfortunately can't do integration test
		assert.Equal(t, err.Error(), "emulator chain not supported")
	})
}

func TestBlocks(t *testing.T) {
	t.Parallel()

	t.Run("Get Latest Block", func(t *testing.T) {
		t.Parallel()

		_, flowkit, gw := setup()

		_, err := flowkit.GetBlock(ctx, BlockQuery{Latest: true})

		gw.Mock.AssertCalled(t, tests.GetLatestBlockFunc)
		gw.Mock.AssertNotCalled(t, tests.GetBlockByHeightFunc)
		gw.Mock.AssertNotCalled(t, tests.GetBlockByIDFunc)
		assert.NoError(t, err)
	})

	t.Run("Get Block by Height", func(t *testing.T) {
		t.Parallel()

		_, flowkit, gw := setup()

		block := tests.NewBlock()
		block.Height = 10
		gw.GetBlockByHeight.Return(block, nil)

		_, err := flowkit.GetBlock(ctx, BlockQuery{Height: 10})

		gw.Mock.AssertCalled(t, tests.GetBlockByHeightFunc, uint64(10))
		gw.Mock.AssertNotCalled(t, tests.GetLatestBlockFunc)
		gw.Mock.AssertNotCalled(t, tests.GetBlockByIDFunc)
		assert.NoError(t, err)
	})

	t.Run("Get Block by ID", func(t *testing.T) {
		t.Parallel()
		_, flowkit, gw := setup()

		ID := flow.HexToID("a310685082f0b09f2a148b2e8905f08ea458ed873596b53b200699e8e1f6536f")
		_, err := flowkit.GetBlock(ctx, BlockQuery{ID: &ID})

		assert.NoError(t, err)
		gw.Mock.AssertCalled(t, tests.GetBlockByIDFunc, ID)
		gw.Mock.AssertNotCalled(t, tests.GetBlockByHeightFunc)
		gw.Mock.AssertNotCalled(t, tests.GetLatestBlockFunc)
	})

}

func TestBlocksGet_Integration(t *testing.T) {
	t.Parallel()

	t.Run("Get Block", func(t *testing.T) {
		t.Parallel()
		_, flowkit := setupIntegration()

		block, err := flowkit.GetBlock(ctx, BlockQuery{Latest: true})

		assert.NoError(t, err)
		assert.Equal(t, block.Height, uint64(0))
		assert.Equal(t, block.ID.String(), "03d40910037d575d52831647b39814f445bc8cc7ba8653286c0eb1473778c34f")
	})
}

func TestCollections(t *testing.T) {
	t.Parallel()

	t.Run("Get Collection", func(t *testing.T) {
		_, flowkit, gw := setup()
		ID := flow.HexToID("a310685082f0b09f2a148b2e8905f08ea458ed873596b53b200699e8e1f6536f")

		_, err := flowkit.GetCollection(ctx, ID)

		assert.NoError(t, err)
		gw.Mock.AssertCalled(t, "GetCollection", ID)
	})
}

func TestEvents(t *testing.T) {
	t.Parallel()

	t.Run("Get Events", func(t *testing.T) {
		t.Parallel()

		_, flowkit, gw := setup()
		_, err := flowkit.GetEvents(ctx, []string{"flow.CreateAccount"}, 0, 0, nil)

		assert.NoError(t, err)
		gw.Mock.AssertCalled(t, tests.GetEventsFunc, "flow.CreateAccount", uint64(0), uint64(0))
	})

	t.Run("Should have larger endHeight then startHeight", func(t *testing.T) {
		t.Parallel()

		_, flowkit, _ := setup()
		_, err := flowkit.GetEvents(ctx, []string{"flow.CreateAccount"}, 10, 0, nil)
		assert.EqualError(t, err, "cannot have end height (0) of block range less that start height (10)")
	})

	t.Run("Test create queries", func(t *testing.T) {

		names := []string{"first", "second"}
		queries := makeEventQueries(names, 0, 400, 250)
		expected := []grpc.EventRangeQuery{
			{Type: "first", StartHeight: 0, EndHeight: 249},
			{Type: "second", StartHeight: 0, EndHeight: 249},
			{Type: "first", StartHeight: 250, EndHeight: 400},
			{Type: "second", StartHeight: 250, EndHeight: 400},
		}
		assert.Equal(t, expected, queries)
	})

	t.Run("Should handle error from get events in goroutine", func(t *testing.T) {
		t.Parallel()

		_, flowkit, gw := setup()

		gw.GetEvents.Return([]flow.BlockEvents{}, errors.New("failed getting event"))

		_, err := flowkit.GetEvents(ctx, []string{"flow.CreateAccount"}, 0, 1, nil)

		assert.EqualError(t, err, "failed getting event")
	})

}

func TestEvents_Integration(t *testing.T) {
	t.Parallel()

	t.Run("Get Events for non existent event", func(t *testing.T) {
		t.Parallel()

		_, flowkit := setupIntegration()

		events, err := flowkit.GetEvents(ctx, []string{"nonexisting"}, 0, 0, nil)
		assert.NoError(t, err)
		assert.Len(t, events, 1)
		assert.Len(t, events[0].Events, 0)
	})

	t.Run("Get Events while adding contracts", func(t *testing.T) {
		t.Parallel()

		state, flowkit := setupIntegration()
		srvAcc, _ := state.EmulatorServiceAccount()

		// create events
		_, _, err := flowkit.AddContract(
			ctx,
			srvAcc,
			resourceToContract(tests.ContractEvents),
			false,
		)
		assert.NoError(t, err)
		assert.NoError(t, err)
		for x := 'A'; x <= 'J'; x++ { // test contract emits events named from A to J
			eName := fmt.Sprintf("A.%s.ContractEvents.Event%c", srvAcc.Address().String(), x)
			events, err := flowkit.GetEvents(ctx, []string{eName}, 0, 1, nil)
			assert.NoError(t, err)
			assert.Len(t, events, 2)
			assert.Len(t, events[1].Events, 1)

		}
	})

	t.Run("Get Events while adding contracts in parallel", func(t *testing.T) {
		t.Parallel()

		state, flowkit := setupIntegration()
		srvAcc, _ := state.EmulatorServiceAccount()

		// create events
		_, _, err := flowkit.AddContract(
			ctx,
			srvAcc,
			resourceToContract(tests.ContractEvents),
			false,
		)
		assert.NoError(t, err)

		assert.NoError(t, err)
		var eventNames []string
		for x := 'A'; x <= 'J'; x++ { // test contract emits events named from A to J
			eName := fmt.Sprintf("A.%s.ContractEvents.Event%c", srvAcc.Address().String(), x)
			eventNames = append(eventNames, eName)
		}

		events, err := flowkit.GetEvents(ctx, eventNames, 0, 1, &EventWorker{
			count:           5,
			blocksPerWorker: 250,
		})
		assert.NoError(t, err)
		assert.Len(t, events, 20)
		assert.Len(t, events[1].Events, 1)
	})
}

func TestKeys(t *testing.T) {
	t.Parallel()

	t.Run("Generate Keys", func(t *testing.T) {
		t.Parallel()

		_, s, _ := setup()
		key, err := s.Keys.Generate("", crypto.ECDSA_P256)
		assert.NoError(t, err)

		assert.Equal(t, len(key.String()), 66)
	})

	t.Run("Generate Keys with seed", func(t *testing.T) {
		t.Parallel()

		_, s, _ := setup()
		key, err := s.Keys.Generate("aaaaaaaaaaaaaaaaaaaaaaannndddddd_its_gone", crypto.ECDSA_P256)

		assert.NoError(t, err)
		assert.Equal(t, key.String(), "0x134f702d0872dba9c7aea15498aab9b2ffedd5aeebfd8ac3cf47c591f0d7ce52")
	})

	t.Run("Test Vector SLIP-0010", func(t *testing.T) {
		// test against SLIP-0010 test vector. All data are taken from:
		//  https://github.com/satoshilabs/slips/blob/master/slip-0010.md#test-vectors
		t.Parallel()
		_, s, _ := setup()

		type testEntry struct {
			sigAlgo    crypto.SignatureAlgorithm
			seed       string
			path       string
			privateKey string
		}

		testVector := []testEntry{
			testEntry{
				sigAlgo:    crypto.ECDSA_secp256k1,
				seed:       "000102030405060708090a0b0c0d0e0f",
				path:       "m/0'/1/2'/2/1000000000",
				privateKey: "0x471b76e389e528d6de6d816857e012c5455051cad6660850e58372a6c3e6e7c8",
			},
			testEntry{
				sigAlgo:    crypto.ECDSA_P256,
				seed:       "000102030405060708090a0b0c0d0e0f",
				path:       "m/0'/1/2'/2/1000000000",
				privateKey: "0x21c4f269ef0a5fd1badf47eeacebeeaa3de22eb8e5b0adcd0f27dd99d34d0119",
			},
			testEntry{
				sigAlgo:    crypto.ECDSA_secp256k1,
				seed:       "fffcf9f6f3f0edeae7e4e1dedbd8d5d2cfccc9c6c3c0bdbab7b4b1aeaba8a5a29f9c999693908d8a8784817e7b7875726f6c696663605d5a5754514e4b484542",
				path:       "m/0/2147483647'/1/2147483646'/2",
				privateKey: "0xbb7d39bdb83ecf58f2fd82b6d918341cbef428661ef01ab97c28a4842125ac23",
			},
			testEntry{
				sigAlgo:    crypto.ECDSA_P256,
				seed:       "fffcf9f6f3f0edeae7e4e1dedbd8d5d2cfccc9c6c3c0bdbab7b4b1aeaba8a5a29f9c999693908d8a8784817e7b7875726f6c696663605d5a5754514e4b484542",
				path:       "m/0/2147483647'/1/2147483646'/2",
				privateKey: "0xbb0a77ba01cc31d77205d51d08bd313b979a71ef4de9b062f8958297e746bd67",
			},
		}

		for _, test := range testVector {
			seed, err := hex.DecodeString(test.seed)
			assert.NoError(t, err)
			// use derivePrivateKeyFromSeed to test instead of DerivePrivateKeyFromMnemonic
			// because the test vector provides seeds, while it's not possible to derive mnemonics
			// corresponding to seeds.
			privateKey, err := s.Keys.derivePrivateKeyFromSeed(seed, test.sigAlgo, test.path)
			assert.NoError(t, err)
			assert.Equal(t, test.privateKey, privateKey.String())
		}
	})

	t.Run("Generate Keys with mnemonic (default path)", func(t *testing.T) {
		t.Parallel()

		_, s, _ := setup()
		key, err := s.Keys.DerivePrivateKeyFromMnemonic("normal dune pole key case cradle unfold require tornado mercy hospital buyer", crypto.ECDSA_P256, "")

		assert.NoError(t, err)
		assert.Equal(t, key.String(), "0x638dc9ad0eee91d09249f0fd7c5323a11600e20d5b9105b66b782a96236e74cf")
	})

	//https://github.com/onflow/ledger-app-flow/blob/dc61213a9c3d73152b78b7391d04165d07f1ad89/tests_speculos/test-basic-show-address-expert.js#L28
	t.Run("Generate Keys with mnemonic (custom path - Ledger)", func(t *testing.T) {
		t.Parallel()

		_, s, _ := setup()
		//ledger test mnemonic: https://github.com/onflow/ledger-app-flow#using-a-real-device-for-integration-tests-nano-s-and-nano-s-plus
		key, err := s.Keys.DerivePrivateKeyFromMnemonic("equip will roof matter pink blind book anxiety banner elbow sun young", crypto.ECDSA_secp256k1, "m/44'/539'/513'/0/0")

		assert.NoError(t, err)
		assert.Equal(t, key.String(), "0xd18d051afca7150781fef111f3387d132d31c4a6250268db0f61f836a205e0b8")

		assert.Equal(t, hex.EncodeToString(key.PublicKey().Encode()), "d7482bbaff7827035d5b238df318b10604673dc613808723efbd23fbc4b9fad34a415828d924ec7b83ac0eddf22ef115b7c203ee39fb080572d7e51775ee54be")
	})

	t.Run("Generate Keys with private key", func(t *testing.T) {
		t.Parallel()

		_, s, _ := setup()
		key, err := s.Keys.ParsePrivateKey("af232020ea7a7256eebdcebd609457d0dea51436a4377d2b577a3cf1f6d45c44", crypto.ECDSA_P256)

		assert.NoError(t, err)
		assert.Equal(t, key.String(), "0xaf232020ea7a7256eebdcebd609457d0dea51436a4377d2b577a3cf1f6d45c44")
		assert.Equal(t, key.PublicKey().String(), "0x3da1d2eb3d9f1a0f57b434dca6bac2068216ccc5c69221a70f5c060152a39296ad28ad260536977f88eea45da9064b81a18c17f5cdc30e638752767359f0b496")
	})

	t.Run("Generate Keys Invalid", func(t *testing.T) {
		t.Parallel()

		_, s, _ := setup()
		inputs := []map[string]crypto.SignatureAlgorithm{
			{"im_short": crypto.ECDSA_P256},
			{"": crypto.StringToSignatureAlgorithm("JUSTNO")},
		}

		errs := []string{
			"failed to generate private key: crypto: insufficient seed length 8, must be at least 32 bytes for ECDSA_P256",
			"failed to generate private key: crypto: Go SDK does not support key generation for UNKNOWN algorithm",
		}

		for x, in := range inputs {
			for k, v := range in {
				_, err := s.Keys.Generate(k, v)
				assert.Equal(t, err.Error(), errs[x])
				x++
			}
		}

	})

	t.Run("Decode RLP Key", func(t *testing.T) {
		t.Parallel()

		_, s, _ := setup()
		dkey, err := s.Keys.DecodeRLP("f847b84084d716c14b051ad6b001624f738f5d302636e6b07cc75e4530af7776a4368a2b586dbefc0564ee28384c2696f178cbed52e62811bcc9ecb59568c996d342db2402038203e8")

		assert.NoError(t, err)
		assert.Equal(t, dkey.PublicKey.String(), "0x84d716c14b051ad6b001624f738f5d302636e6b07cc75e4530af7776a4368a2b586dbefc0564ee28384c2696f178cbed52e62811bcc9ecb59568c996d342db24")
		assert.Equal(t, dkey.SigAlgo.String(), "ECDSA_P256")
	})

	t.Run("Decode RLP Key Invalid", func(t *testing.T) {
		t.Parallel()

		_, s, _ := setup()
		_, err := s.Keys.DecodeRLP("aaa")
		assert.Equal(t, err.Error(), "failed to decode public key: encoding/hex: odd length hex string")
	})

	t.Run("Decode PEM Key", func(t *testing.T) {
		t.Parallel()

		_, s, _ := setup()
		dkey, err := s.Keys.DecodePEM("-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE1HmzzcntvdsZXLErNRYa3oJrAypk\nvdQGLMh/s7p+ccnPZG/yOZC7RTLKRcRFx+kIzvJ4ssRhU2ADmmZgo2apXw==\n-----END PUBLIC KEY-----", crypto.ECDSA_P256)

		assert.NoError(t, err)
		assert.Equal(t, dkey.PublicKey.String(), "0xd479b3cdc9edbddb195cb12b35161ade826b032a64bdd4062cc87fb3ba7e71c9cf646ff23990bb4532ca45c445c7e908cef278b2c4615360039a6660a366a95f")
		assert.Equal(t, dkey.SigAlgo.String(), "ECDSA_P256")
	})

	t.Run("Decode PEM Key Invalid", func(t *testing.T) {
		t.Parallel()

		_, s, _ := setup()
		_, err := s.Keys.DecodePEM("nope", crypto.ECDSA_P256)
		assert.Equal(t, err.Error(), "crypto: failed to parse PEM string, not all bytes in PEM key were decoded: 6e6f7065")
	})
}

func TestProject(t *testing.T) {
	t.Parallel()

	t.Run("Init Project", func(t *testing.T) {
		t.Parallel()

		st, s, _ := setup()
		pkey := tests.PrivKeys()[0]
		init, err := s.Project.Init(st.ReaderWriter(), false, false, crypto.ECDSA_P256, crypto.SHA3_256, pkey)
		assert.NoError(t, err)

		sacc, err := init.EmulatorServiceAccount()
		assert.NotNil(t, sacc)
		assert.NoError(t, err)
		assert.Equal(t, sacc.Name(), config.DefaultEmulatorServiceAccountName)
		assert.Equal(t, sacc.Address().String(), "f8d6e0586b0a20c7")

		p, err := sacc.Key().PrivateKey()
		assert.NoError(t, err)
		assert.Equal(t, (*p).String(), pkey.String())

		init, err = s.Project.Init(st.ReaderWriter(), false, false, crypto.ECDSA_P256, crypto.SHA3_256, nil)
		assert.NoError(t, err)
		em, err := init.EmulatorServiceAccount()
		assert.NoError(t, err)
		k, err := em.Key().PrivateKey()
		assert.NoError(t, err)
		assert.NotNil(t, (*k).String())
	})

	t.Run("Deploy Project", func(t *testing.T) {
		t.Parallel()

		state, flowkit, gw := setup()

		c := config.Contract{
			Name:     "Hello",
			Location: tests.ContractHelloString.Filename,
		}
		state.Contracts().AddOrUpdate(c)

		n := config.Network{
			Name: "emulator",
			Host: "127.0.0.1:3569",
		}
		state.Networks().AddOrUpdate(n.Name, n)

		acct2 := tests.Donald()
		state.Accounts().AddOrUpdate(acct2)

		d := config.Deployment{
			Network: n.Name,
			Account: acct2.Name(),
			Contracts: []config.ContractDeployment{{
				Name: c.Name,
				Args: nil,
			}},
		}
		state.Deployments().AddOrUpdate(d)

		gw.SendSignedTransaction.Run(func(args mock.Arguments) {
			tx := args.Get(0).(*Transaction)
			assert.Equal(t, tx.FlowTransaction().Payer, acct2.Address())
			assert.True(t, strings.Contains(string(tx.FlowTransaction().Script), "signer.contracts.add"))

			gw.SendSignedTransaction.Return(tests.NewTransaction(), nil)
		})

		contracts, err := s.Project.Deploy("emulator", false)

		assert.NoError(t, err)
		assert.Equal(t, len(contracts), 1)
		assert.Equal(t, contracts[0].AccountAddress, acct2.Address())
	})

	t.Run("Deploy Project Using LocationAliases", func(t *testing.T) {
		t.Parallel()

		emulator := config.DefaultEmulatorNetwork().Name
		state, flowkit, gw := setup()

		c1 := config.Contract{
			Name:     "ContractB",
			Location: tests.ContractB.Filename,
		}
		state.Contracts().AddOrUpdate(c1)

		c2 := config.Contract{
			Name:     "Aliased",
			Location: tests.ContractA.Filename,
			Aliases: []config.Alias{{
				Network: emulator,
				Address: tests.Donald().Address(),
			}},
		}
		state.Contracts().AddOrUpdate(c2)

		c3 := config.Contract{
			Name:     "ContractC",
			Location: tests.ContractC.Filename,
		}
		state.Contracts().AddOrUpdate(c3)

		state.Networks().AddOrUpdate(emulator, config.DefaultEmulatorNetwork())

		a := tests.Alice()
		state.Accounts().AddOrUpdate(a)

		d := config.Deployment{
			Network: emulator,
			Account: a.Name(),
			Contracts: []config.ContractDeployment{
				{Name: c1.Name}, {Name: c3.Name},
			},
		}
		state.Deployments().AddOrUpdate(d)

		// for checking imports are correctly resolved
		resolved := map[string]string{
			tests.ContractB.Name: fmt.Sprintf(`import ContractA from 0x%s`, tests.Donald().Address().Hex()),
			tests.ContractC.Name: fmt.Sprintf(`
		import ContractB from 0x%s
		import ContractA from 0x%s`, a.Address().Hex(), tests.Donald().Address().Hex()),
		} // don't change formatting of the above code since it compares the strings with included formatting

		gw.SendSignedTransaction.Run(func(args mock.Arguments) {
			tx := args.Get(0).(*Transaction)
			assert.Equal(t, tx.FlowTransaction().Payer, a.Address())
			assert.True(t, strings.Contains(string(tx.FlowTransaction().Script), "signer.contracts.add"))

			argCode := tx.FlowTransaction().Arguments[1]
			decodeCode, _ := jsoncdc.Decode(nil, argCode)
			code, _ := hex.DecodeString(decodeCode.ToGoValue().(string))

			argName := tx.FlowTransaction().Arguments[0]
			decodeName, _ := jsoncdc.Decode(nil, argName)

			testCode, found := resolved[decodeName.ToGoValue().(string)]
			require.True(t, found)
			assert.True(t, strings.Contains(string(code), testCode))

			gw.SendSignedTransaction.Return(tests.NewTransaction(), nil)
		})

		contracts, err := s.Project.Deploy(emulator, false)

		assert.NoError(t, err)
		assert.Equal(t, len(contracts), 2)
		gw.Mock.AssertCalled(t, tests.GetLatestBlockFunc)
		gw.Mock.AssertCalled(t, tests.GetAccountFunc, a.Address())
		gw.Mock.AssertNumberOfCalls(t, tests.GetTransactionResultFunc, 2)
	})

	t.Run("Deploy Project New Import Schema and LocationAliases", func(t *testing.T) {
		t.Parallel()

		emulator := config.DefaultEmulatorNetwork().Name
		state, flowkit, gw := setup()

		c1 := config.Contract{
			Name:     "ContractBB",
			Location: tests.ContractBB.Filename,
		}
		state.Contracts().AddOrUpdate(c1)

		c2 := config.Contract{
			Name:     "ContractAA",
			Location: tests.ContractAA.Filename,
			Aliases: []config.Alias{{
				Network: emulator,
				Address: tests.Donald().Address(),
			}},
		}
		state.Contracts().AddOrUpdate(c2)

		c3 := config.Contract{
			Name:     "ContractCC",
			Location: tests.ContractCC.Filename,
		}
		state.Contracts().AddOrUpdate(c3)

		state.Networks().AddOrUpdate(emulator, config.DefaultEmulatorNetwork())

		a := tests.Alice()
		state.Accounts().AddOrUpdate(a)

		d := config.Deployment{
			Network: emulator,
			Account: a.Name(),
			Contracts: []config.ContractDeployment{
				{Name: c1.Name}, {Name: c3.Name},
			},
		}
		state.Deployments().AddOrUpdate(d)

		// for checking imports are correctly resolved
		resolved := map[string]string{
			tests.ContractB.Name: fmt.Sprintf(`import ContractAA from 0x%s`, tests.Donald().Address().Hex()),
			tests.ContractC.Name: fmt.Sprintf(`
		import ContractBB from 0x%s
		import ContractAA from 0x%s`, a.Address().Hex(), tests.Donald().Address().Hex()),
		} // don't change formatting of the above code since it compares the strings with included formatting

		gw.SendSignedTransaction.Run(func(args mock.Arguments) {
			tx := args.Get(0).(*Transaction)
			assert.Equal(t, tx.FlowTransaction().Payer, a.Address())
			assert.True(t, strings.Contains(string(tx.FlowTransaction().Script), "signer.contracts.add"))

			argCode := tx.FlowTransaction().Arguments[1]
			decodeCode, _ := jsoncdc.Decode(nil, argCode)
			code, _ := hex.DecodeString(decodeCode.ToGoValue().(string))

			argName := tx.FlowTransaction().Arguments[0]
			decodeName, _ := jsoncdc.Decode(nil, argName)

			testCode, found := resolved[decodeName.ToGoValue().(string)]
			require.True(t, found)
			assert.True(t, strings.Contains(string(code), testCode))

			gw.SendSignedTransaction.Return(tests.NewTransaction(), nil)
		})

		contracts, err := s.Project.Deploy(emulator, false)

		assert.NoError(t, err)
		assert.Equal(t, len(contracts), 2)
		gw.Mock.AssertCalled(t, tests.GetLatestBlockFunc)
		gw.Mock.AssertCalled(t, tests.GetAccountFunc, a.Address())
		gw.Mock.AssertNumberOfCalls(t, tests.GetTransactionResultFunc, 2)
	})

	t.Run("Deploy Project Duplicate Address", func(t *testing.T) {
		t.Parallel()

		state, flowkit, gw := setup()

		c := config.Contract{
			Name:     "Hello",
			Location: tests.ContractHelloString.Filename,
		}
		state.Contracts().AddOrUpdate(c)

		n := config.Network{
			Name: "emulator",
			Host: "127.0.0.1:3569",
		}
		state.Networks().AddOrUpdate(n.Name, n)

		acct1 := tests.Charlie()
		state.Accounts().AddOrUpdate(acct1)

		acct2 := tests.Donald()
		state.Accounts().AddOrUpdate(acct2)

		d := config.Deployment{
			Network: n.Name,
			Account: acct2.Name(),
			Contracts: []config.ContractDeployment{{
				Name: c.Name,
				Args: nil,
			}},
		}
		state.Deployments().AddOrUpdate(d)

		gw.SendSignedTransaction.Run(func(args mock.Arguments) {
			tx := args.Get(0).(*Transaction)
			assert.Equal(t, tx.FlowTransaction().Payer, acct2.Address())
			assert.True(t, strings.Contains(string(tx.FlowTransaction().Script), "signer.contracts.add"))

			gw.SendSignedTransaction.Return(tests.NewTransaction(), nil)
		})

		contracts, err := s.Project.Deploy("emulator", false)

		assert.NoError(t, err)
		assert.Equal(t, len(contracts), 1)
		assert.Equal(t, contracts[0].AccountAddress, acct2.Address())
	})

}

// used for integration tests
func simpleDeploy(state *State, s *Services, update bool) ([]*project.Contract, error) {
	srvAcc, _ := state.EmulatorServiceAccount()

	c := config.Contract{
		Name:     tests.ContractHelloString.Name,
		Location: tests.ContractHelloString.Filename,
	}
	state.Contracts().AddOrUpdate(c)

	n := config.Network{
		Name: "emulator",
		Host: "127.0.0.1:3569",
	}
	state.Networks().AddOrUpdate(n.Name, n)

	d := config.Deployment{
		Network: n.Name,
		Account: srvAcc.Name(),
		Contracts: []config.ContractDeployment{{
			Name: c.Name,
			Args: nil,
		}},
	}
	state.Deployments().AddOrUpdate(d)

	return s.Project.Deploy(n.Name, update)
}

func TestProject_Integration(t *testing.T) {
	t.Parallel()

	t.Run("Deploy Project", func(t *testing.T) {
		t.Parallel()

		state, flowkit := setupIntegration()
		contracts, err := simpleDeploy(state, s, false)
		assert.NoError(t, err)
		assert.Len(t, contracts, 1)
		assert.Equal(t, contracts[0].Name, tests.ContractHelloString.Name)
		assert.Equal(t, string(contracts[0].Code()), string(tests.ContractHelloString.Source))
	})

	t.Run("Deploy Complex Project", func(t *testing.T) {
		t.Parallel()

		state, flowkit := setupIntegration()
		srvAcc, _ := state.EmulatorServiceAccount()

		n := config.DefaultEmulatorNetwork()
		state.Networks().AddOrUpdate(n.Name, n)

		contractFixtures := []tests.Resource{
			tests.ContractB, tests.ContractC,
		}

		testContracts := make([]config.Contract, len(contractFixtures))
		for i, c := range contractFixtures {
			testContracts[i] = config.Contract{
				Name:     c.Name,
				Location: c.Filename,
			}
			state.Contracts().AddOrUpdate(testContracts[i])
		}

		cA := tests.ContractA
		state.Contracts().AddOrUpdate(config.Contract{
			Name:     cA.Name,
			Location: cA.Filename,
			Aliases: []config.Alias{{
				Network: n.Name,
				Address: srvAcc.Address(),
			}},
		})

		state.Deployments().AddOrUpdate(config.Deployment{
			Network: n.Name,
			Account: srvAcc.Name(),
			Contracts: []config.ContractDeployment{{
				Name: testContracts[0].Name,
				Args: nil,
			}, {
				Name: testContracts[1].Name,
				Args: []cadence.Value{
					cadence.String("foo"),
				},
			}},
		})

		// deploy contract imported as alias
		_, _, err := s.Accounts.AddContract(
			srvAcc,
			NewScript(tests.ContractA.Source, nil, tests.ContractA.Filename),
			n.Name,
			false,
		)
		require.NoError(t, err)

		// replace imports manually to assert that replacing worked in deploy service
		addr := fmt.Sprintf("0x%s", srvAcc.Address())
		replacedContracts := make([]string, len(contractFixtures))
		for i, c := range contractFixtures {
			replacedContracts[i] = strings.ReplaceAll(string(c.Source), `"./contractA.cdc"`, addr)
			replacedContracts[i] = strings.ReplaceAll(replacedContracts[i], `"./contractB.cdc"`, addr)
		}

		contracts, err := s.Project.Deploy(n.Name, false)
		assert.NoError(t, err)
		assert.Len(t, contracts, 2)

		account, err := s.Accounts.Get(srvAcc.Address())

		for i, c := range testContracts {
			code, exists := account.Contracts[c.Name]
			assert.True(t, exists)
			assert.Equal(t, replacedContracts[i], string(code))
		}
	})

	t.Run("Deploy Project Update", func(t *testing.T) {
		t.Parallel()

		// setup
		state, flowkit := setupIntegration()
		_, err := simpleDeploy(state, s, false)
		assert.NoError(t, err)

		_, err = simpleDeploy(state, s, true)
		assert.NoError(t, err)
	})

}

func TestScripts(t *testing.T) {
	t.Parallel()

	t.Run("Execute Script", func(t *testing.T) {
		_, flowkit, gw := setup()

		gw.ExecuteScript.Run(func(args mock.Arguments) {
			assert.Len(t, string(args.Get(0).([]byte)), 78)
			assert.Equal(t, "\"Foo\"", args.Get(1).([]cadence.Value)[0].String())
			gw.ExecuteScript.Return(cadence.MustConvertValue(""), nil)
		})

		args := []cadence.Value{cadence.String("Foo")}
		_, err := s.Scripts.Execute(
			NewScript(tests.ScriptArgString.Source, args, ""),
			"",
		)

		assert.NoError(t, err)
	})

}

func TestScripts_Integration(t *testing.T) {
	t.Parallel()

	t.Run("Execute", func(t *testing.T) {
		t.Parallel()
		_, flowkit := setupIntegration()

		args := []cadence.Value{cadence.String("Foo")}
		res, err := s.Scripts.Execute(
			NewScript(tests.ScriptArgString.Source, args, ""),
			"",
		)

		assert.NoError(t, err)
		assert.Equal(t, "\"Hello Foo\"", res.String())
	})

	t.Run("Execute report error", func(t *testing.T) {
		t.Parallel()
		_, flowkit := setupIntegration()
		args := []cadence.Value{cadence.String("Foo")}
		res, err := s.Scripts.Execute(
			NewScript(tests.ScriptWithError.Source, args, ""),
			"",
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot find type in this scope")
		assert.Nil(t, res)

	})

	t.Run("Execute With Imports", func(t *testing.T) {
		t.Parallel()
		state, flowkit := setupIntegration()
		srvAcc, _ := state.EmulatorServiceAccount()

		// setup
		c := config.Contract{
			Name:     tests.ContractHelloString.Name,
			Location: tests.ContractHelloString.Filename,
		}
		state.Contracts().AddOrUpdate(c)

		n := config.Network{
			Name: "emulator",
			Host: "127.0.0.1:3569",
		}
		state.Networks().AddOrUpdate(n.Name, n)

		d := config.Deployment{
			Network: n.Name,
			Account: srvAcc.Name(),
			Contracts: []config.ContractDeployment{{
				Name: c.Name,
				Args: nil,
			}},
		}
		state.Deployments().AddOrUpdate(d)
		_, _, _ = s.Accounts.AddContract(
			srvAcc,
			resourceToContract(tests.ContractHelloString),
			"",
			false,
		)

		res, err := s.Scripts.Execute(
			NewScript(tests.ScriptImport.Source, nil, tests.ScriptImport.Filename),
			n.Name,
		)
		assert.NoError(t, err)
		assert.Equal(t, res.String(), "\"Hello Hello, World!\"")
	})

	t.Run("Execute Script Invalid", func(t *testing.T) {
		t.Parallel()
		_, flowkit := setupIntegration()
		in := [][]string{
			{tests.ScriptImport.Filename, ""},
			{"", "emulator"},
			{tests.ScriptImport.Filename, "foo"},
		}

		out := []string{
			"missing network, specify which network to use to resolve imports in script code",
			"resolving imports in scripts not supported",
			"import ./contractHello.cdc could not be resolved from provided contracts",
		}

		for x, i := range in {
			_, err := s.Scripts.Execute(
				NewScript(tests.ScriptImport.Source, nil, i[0]),
				i[1],
			)
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), out[x])
		}

	})
}

func TestExecutingTests(t *testing.T) {
	t.Parallel()

	t.Run("simple", func(t *testing.T) {
		t.Parallel()

		st, s, _ := setup()

		script := tests.TestScriptSimple
		results, err := s.Tests.Execute(script.Source, script.Filename, st.ReaderWriter())

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.NoError(t, results[0].Error)
	})

	t.Run("simple failing", func(t *testing.T) {
		t.Parallel()

		st, s, _ := setup()

		script := tests.TestScriptSimpleFailing
		results, err := s.Tests.Execute(script.Source, script.Filename, st.ReaderWriter())

		require.NoError(t, err)
		require.Len(t, results, 1)

		err = results[0].Error
		require.Error(t, err)
		assert.ErrorAs(t, err, &stdlib.AssertionError{})
	})

	t.Run("with import", func(t *testing.T) {
		t.Parallel()

		// Setup
		st, s, _ := setup()

		c := config.Contract{
			Name:     tests.ContractHelloString.Name,
			Location: tests.ContractHelloString.Filename,
		}
		st.Contracts().AddOrUpdate(c)

		// Execute script
		script := tests.TestScriptWithImport
		results, err := s.Tests.Execute(script.Source, script.Filename, st.ReaderWriter())

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.NoError(t, results[0].Error)
	})

	t.Run("with file read", func(t *testing.T) {
		t.Parallel()

		// Setup
		st, s, _ := setup()
		readerWriter := st.ReaderWriter()
		readerWriter.WriteFile(
			tests.SomeFile.Filename,
			tests.SomeFile.Source,
			os.ModeTemporary,
		)

		// Execute script
		script := tests.TestScriptWithFileRead
		results, err := s.Tests.Execute(script.Source, script.Filename, readerWriter)

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.NoError(t, results[0].Error)
	})
}

const gasLimit = 1000

func TestTransactions(t *testing.T) {
	t.Parallel()

	state, _, _ := setup()
	serviceAcc, _ := state.EmulatorServiceAccount()
	serviceAddress := serviceAcc.Address()

	t.Run("Get Transaction", func(t *testing.T) {
		t.Parallel()
		_, flowkit, gw := setup()
		txs := tests.NewTransaction()

		_, _, err := s.Transactions.GetStatus(txs.ID(), true)

		assert.NoError(t, err)
		gw.Mock.AssertNumberOfCalls(t, tests.GetTransactionResultFunc, 1)
		gw.Mock.AssertCalled(t, tests.GetTransactionFunc, txs.ID())
	})

	t.Run("Send Transaction args", func(t *testing.T) {
		t.Parallel()
		_, flowkit, gw := setup()

		var txID flow.Identifier
		gw.SendSignedTransaction.Run(func(args mock.Arguments) {
			tx := args.Get(0).(*Transaction)
			arg, err := tx.FlowTransaction().Argument(0)
			assert.NoError(t, err)
			assert.Equal(t, "\"Bar\"", arg.String())
			assert.Equal(t, serviceAddress, tx.Signer().Address())
			assert.Len(t, string(tx.FlowTransaction().Script), 227)

			t := tests.NewTransaction()
			txID = t.ID()
			gw.SendSignedTransaction.Return(t, nil)
		})

		gw.GetTransactionResult.Run(func(args mock.Arguments) {
			assert.Equal(t, txID, args.Get(0).(flow.Identifier))
			gw.GetTransactionResult.Return(tests.NewTransactionResult(nil), nil)
		})

		_, _, err := s.Transactions.Send(
			NewSingleTransactionAccount(serviceAcc),
			NewScript(
				tests.TransactionArgString.Source,
				[]cadence.Value{cadence.String("Bar")},
				"",
			),
			gasLimit,
			"",
		)

		assert.NoError(t, err)
		gw.Mock.AssertNumberOfCalls(t, tests.SendSignedTransactionFunc, 1)
		gw.Mock.AssertNumberOfCalls(t, tests.GetTransactionResultFunc, 1)
	})

}

func setupAccounts(state *State, s *Services) {
	setupAccount(state, s, tests.Alice())
	setupAccount(state, s, tests.Bob())
	setupAccount(state, s, tests.Charlie())
}

func setupAccount(state *State, s *Services, account *Account) {
	srv, _ := state.EmulatorServiceAccount()

	key := account.Key()
	pk, _ := key.PrivateKey()
	acc, _ := s.Accounts.Create(srv,
		[]crypto.PublicKey{(*pk).PublicKey()},
		[]int{flow.AccountKeyWeightThreshold},
		[]crypto.SignatureAlgorithm{key.SigAlgo()},
		[]crypto.HashAlgorithm{key.HashAlgo()},
		nil,
	)

	newAcc :=
		NewAccount(account.Name()).
			SetAddress(acc.Address).
			SetKey(key)

	state.Accounts().AddOrUpdate(newAcc)
}

func Test_TransactionRoles(t *testing.T) {
	t.Run("Building Signers", func(t *testing.T) {
		state, flowkit := setupIntegration()
		setupAccounts(state, s)
		a, _ := state.Accounts().ByName("Alice")
		b, _ := state.Accounts().ByName("Bob")
		c, _ := state.Accounts().ByName("Charlie")

		// we make copies with diffrerent names but same addresses for testing building signers,
		// since if same addresses are present that's should be treated as same account
		aCopy1 := *a
		aCopy2 := *a
		aCopy1.SetName("Boo")
		aCopy2.SetName("Zoo")

		tests := []struct {
			*transactionAccountRoles
			signerAddresses []flow.Address
		}{{
			transactionAccountRoles: &transactionAccountRoles{
				proposer:    a,
				authorizers: []*Account{b},
				payer:       c,
			},
			signerAddresses: []flow.Address{
				a.Address(),
				b.Address(),
				c.Address(),
			},
		}, {
			transactionAccountRoles: &transactionAccountRoles{
				proposer:    a,
				authorizers: []*Account{a},
				payer:       a,
			},
			signerAddresses: []flow.Address{
				a.Address(),
			},
		}, {
			transactionAccountRoles: &transactionAccountRoles{
				proposer:    a,
				payer:       b,
				authorizers: []*Account{a},
			},
			signerAddresses: []flow.Address{
				a.Address(), b.Address(),
			},
		}, {
			transactionAccountRoles: &transactionAccountRoles{
				proposer: a,
				payer:    a,
			},
			signerAddresses: []flow.Address{
				a.Address(),
			},
		}, {
			transactionAccountRoles: &transactionAccountRoles{
				proposer:    &aCopy1,
				payer:       &aCopy2,
				authorizers: []*Account{a},
			},
			signerAddresses: []flow.Address{
				a.Address(),
			},
		}}

		for i, test := range tests {
			signerAccs := test.getSigners()
			signerAddrs := make([]flow.Address, len(signerAccs))
			for i, sig := range signerAccs {
				signerAddrs[i] = sig.Address()
			}

			assert.Equal(t, test.signerAddresses, signerAddrs, fmt.Sprintf("test %d failed", i))
		}
	})

	t.Run("Building Addresses", func(t *testing.T) {
		state, flowkit := setupIntegration()
		setupAccounts(state, s)
		a, _ := state.Accounts().ByName("Alice")
		b, _ := state.Accounts().ByName("Bob")
		c, _ := state.Accounts().ByName("Charlie")

		roles := &transactionAccountRoles{
			proposer:    a,
			authorizers: []*Account{b, c},
			payer:       c,
		}

		addresses := roles.toAddresses()

		assert.Equal(t, a.Address(), addresses.proposer)
		assert.Equal(t, c.Address(), addresses.payer)
		assert.Equal(t, []flow.Address{b.Address(), c.Address()}, addresses.authorizers)
	})
}

func TestTransactions_Integration(t *testing.T) {
	t.Parallel()

	t.Run("Build Transaction", func(t *testing.T) {
		t.Parallel()
		state, flowkit := setupIntegration()
		setupAccounts(state, s)

		type txIn struct {
			prop    flow.Address
			auth    []flow.Address
			payer   flow.Address
			index   int
			code    []byte
			file    string
			gas     uint64
			args    []cadence.Value
			network string
			yes     bool
		}

		a, _ := state.Accounts().ByName("Alice")
		b, _ := state.Accounts().ByName("Bob")
		c, _ := state.Accounts().ByName("Charlie")

		txIns := []txIn{{
			a.Address(),
			[]flow.Address{a.Address()},
			a.Address(),
			0,
			tests.TransactionSimple.Source,
			tests.TransactionSimple.Filename,
			flow.DefaultTransactionGasLimit,
			nil,
			"",
			true,
		}, {
			c.Address(),
			[]flow.Address{a.Address(), b.Address()},
			c.Address(),
			0,
			tests.TransactionSimple.Source,
			tests.TransactionSimple.Filename,
			flow.DefaultTransactionGasLimit,
			nil,
			"",
			true,
		}}

		for i, txIn := range txIns {
			tx, err := s.Transactions.Build(
				NewTransactionAddresses(txIn.prop, txIn.payer, txIn.auth),
				txIn.index,
				NewScript(txIn.code, txIn.args, txIn.file),
				txIn.gas,
				txIn.network,
			)

			require.NoError(t, err, fmt.Sprintf("test vector %d", i))
			ftx := tx.FlowTransaction()
			assert.Equal(t, ftx.Script, txIn.code)
			assert.Equal(t, ftx.Payer, txIn.payer)
			assert.Equal(t, len(ftx.Authorizers), 0) // make sure authorizers weren't added as tx input doesn't require them
			assert.Equal(t, ftx.ProposalKey.Address, txIn.prop)
			assert.Equal(t, ftx.ProposalKey.KeyIndex, txIn.index)
		}

	})

	t.Run("Build Transaction with Imports", func(t *testing.T) {
		t.Parallel()
		state, flowkit := setupIntegration()
		setupAccounts(state, s)

		srvAcc, _ := state.EmulatorServiceAccount()
		signer := srvAcc.Address()

		// setup
		c := config.Contract{
			Name:     tests.ContractHelloString.Name,
			Location: tests.ContractHelloString.Filename,
		}
		state.Contracts().AddOrUpdate(c)

		n := config.Network{
			Name: "emulator",
			Host: "127.0.0.1:3569",
		}
		state.Networks().AddOrUpdate(n.Name, n)

		d := config.Deployment{
			Network: n.Name,
			Account: srvAcc.Name(),
			Contracts: []config.ContractDeployment{{
				Name: c.Name,
				Args: nil,
			}},
		}
		state.Deployments().AddOrUpdate(d)
		_, _, _ = s.Accounts.AddContract(
			srvAcc,
			resourceToContract(tests.ContractHelloString),
			"",
			false,
		)

		tx, err := s.Transactions.Build(
			NewTransactionAddresses(signer, signer, []flow.Address{signer}),
			srvAcc.Key().Index(),
			NewScript(tests.TransactionImports.Source, nil, tests.TransactionImports.Filename),
			flow.DefaultTransactionGasLimit,
			n.Name,
		)

		assert.NoError(t, err)
		ftx := tx.FlowTransaction()
		assert.Equal(t,
			string(ftx.Script),
			strings.ReplaceAll(
				string(tests.TransactionImports.Source),
				"import Hello from \"./contractHello.cdc\"",
				fmt.Sprintf("import Hello from 0x%s", srvAcc.Address()),
			),
		)
	})

	t.Run("Sign transaction", func(t *testing.T) {
		t.Parallel()
		state, flowkit := setupIntegration()
		setupAccounts(state, s)

		a, _ := state.Accounts().ByName("Alice")

		tx, err := s.Transactions.Build(
			NewTransactionAddresses(a.Address(), a.Address(), nil),
			0,
			NewScript(tests.TransactionSimple.Source, nil, tests.TransactionSimple.Filename),
			flow.DefaultTransactionGasLimit,
			"",
		)

		assert.Nil(t, err)
		assert.NotNil(t, tx)

		txSigned, err := s.Transactions.Sign(
			a,
			[]byte(fmt.Sprintf("%x", tx.FlowTransaction().Encode())),
		)
		assert.Nil(t, err)
		assert.NotNil(t, txSigned)
		assert.Equal(t, len(txSigned.FlowTransaction().Authorizers), 0)
		assert.Equal(t, txSigned.FlowTransaction().Payer, a.Address())
		assert.Equal(t, txSigned.FlowTransaction().ProposalKey.Address, a.Address())
		assert.Equal(t, txSigned.FlowTransaction().ProposalKey.KeyIndex, 0)
		assert.Equal(t, txSigned.FlowTransaction().Script, tests.TransactionSimple.Source)
	})

	t.Run("Build, Sign and Send Transaction", func(t *testing.T) {
		t.Parallel()
		state, flowkit := setupIntegration()
		setupAccounts(state, s)

		a, _ := state.Accounts().ByName("Alice")

		tx, err := s.Transactions.Build(
			NewTransactionAddresses(a.Address(), a.Address(), []flow.Address{a.Address()}),
			0,
			NewScript(tests.TransactionSingleAuth.Source, nil, tests.TransactionSingleAuth.Filename),
			flow.DefaultTransactionGasLimit,
			"",
		)

		assert.Nil(t, err)
		assert.NotNil(t, tx)

		txSigned, err := s.Transactions.Sign(
			a,
			[]byte(fmt.Sprintf("%x", tx.FlowTransaction().Encode())),
		)
		assert.Nil(t, err)
		assert.NotNil(t, txSigned)

		txSent, txResult, err := s.Transactions.SendSigned(txSigned)
		assert.Nil(t, err)
		assert.Equal(t, txResult.Status, flow.TransactionStatusSealed)
		assert.NotNil(t, txSent.ID())

	})

	t.Run("Fails signing transaction, wrong account", func(t *testing.T) {
		t.Parallel()
		state, flowkit := setupIntegration()
		setupAccounts(state, s)

		a, _ := state.Accounts().ByName("Alice")

		tx, err := s.Transactions.Build(
			NewTransactionAddresses(a.Address(), a.Address(), []flow.Address{a.Address()}),
			0,
			NewScript(tests.TransactionSingleAuth.Source, nil, tests.TransactionSingleAuth.Filename),
			flow.DefaultTransactionGasLimit,
			"",
		)

		assert.Nil(t, err)
		assert.NotNil(t, tx)

		// sign with wrong account
		a, _ = state.Accounts().ByName("Bob")

		txSigned, err := s.Transactions.Sign(
			a,
			[]byte(fmt.Sprintf("%x", tx.FlowTransaction().Encode())),
		)
		assert.EqualError(t, err, "not a valid signer 179b6b1cb6755e31, proposer: 01cf0e2f2f715450, payer: 01cf0e2f2f715450, authorizers: [01cf0e2f2f715450]")
		assert.Nil(t, txSigned)
	})

	t.Run("Fails building, authorizers mismatch", func(t *testing.T) {
		t.Parallel()
		state, flowkit := setupIntegration()
		setupAccounts(state, s)

		a, _ := state.Accounts().ByName("Alice")

		tx, err := s.Transactions.Build(
			NewTransactionAddresses(a.Address(), a.Address(), []flow.Address{a.Address()}),
			0,
			NewScript(tests.TransactionTwoAuth.Source, nil, tests.TransactionTwoAuth.Filename),
			flow.DefaultTransactionGasLimit,
			"",
		)

		assert.EqualError(t, err, "provided authorizers length mismatch, required authorizers 2, but provided 1")
		assert.Nil(t, tx)
	})

	// todo(sideninja) we should convert different variations of sending transaction to table tests

	t.Run("Send Transaction No Auths", func(t *testing.T) {
		t.Parallel()
		state, flowkit := setupIntegration()
		setupAccounts(state, s)

		a, _ := state.Accounts().ByName("Alice")

		tx, txr, err := s.Transactions.Send(
			&transactionAccountRoles{
				proposer: a,
				payer:    a,
			},
			NewScript(tests.TransactionSimple.Source, nil, tests.TransactionSimple.Filename),
			flow.DefaultTransactionGasLimit,
			"",
		)
		assert.NoError(t, err)
		assert.Equal(t, tx.Payer.String(), a.Address().String())
		assert.Equal(t, tx.ProposalKey.KeyIndex, a.Key().Index())
		assert.Nil(t, txr.Error)
		assert.Equal(t, txr.Status, flow.TransactionStatusSealed)
	})

	t.Run("Send Transaction With Auths", func(t *testing.T) {
		t.Parallel()
		state, flowkit := setupIntegration()
		setupAccounts(state, s)

		a, _ := state.Accounts().ByName("Alice")

		tx, txr, err := s.Transactions.Send(
			NewSingleTransactionAccount(a),
			NewScript(tests.TransactionSingleAuth.Source, nil, tests.TransactionSingleAuth.Filename),
			flow.DefaultTransactionGasLimit,
			"",
		)
		assert.NoError(t, err)
		assert.Equal(t, tx.Payer.String(), a.Address().String())
		assert.Equal(t, tx.ProposalKey.KeyIndex, a.Key().Index())
		assert.Nil(t, txr.Error)
		assert.Equal(t, txr.Status, flow.TransactionStatusSealed)
	})

	t.Run("Send Transaction multiple account roles", func(t *testing.T) {
		t.Parallel()
		state, flowkit := setupIntegration()
		setupAccounts(state, s)

		a, _ := state.Accounts().ByName("Alice")
		b, _ := state.Accounts().ByName("Bob")
		c, _ := state.Accounts().ByName("Charlie")

		roles, err := NewTransactionAccountRoles(a, b, []*Account{c})
		require.NoError(t, err)

		tx, txr, err := s.Transactions.Send(
			roles,
			NewScript(tests.TransactionSingleAuth.Source, nil, tests.TransactionSingleAuth.Filename),
			flow.DefaultTransactionGasLimit,
			"",
		)
		assert.NoError(t, err)
		assert.Equal(t, tx.Payer.String(), b.Address().String())
		assert.Equal(t, tx.Authorizers[0].String(), c.Address().String())
		assert.Equal(t, tx.ProposalKey.KeyIndex, a.Key().Index())
		assert.Nil(t, txr.Error)
		assert.Equal(t, txr.Status, flow.TransactionStatusSealed)
	})

	t.Run("Send Transaction two account roles", func(t *testing.T) {
		t.Parallel()
		state, flowkit := setupIntegration()
		setupAccounts(state, s)

		a, _ := state.Accounts().ByName("Alice")
		b, _ := state.Accounts().ByName("Bob")

		roles, err := NewTransactionAccountRoles(a, b, []*Account{a})
		require.NoError(t, err)

		tx, txr, err := s.Transactions.Send(
			roles,
			NewScript(tests.TransactionSingleAuth.Source, nil, tests.TransactionSingleAuth.Filename),
			flow.DefaultTransactionGasLimit,
			"",
		)
		assert.NoError(t, err)
		assert.Equal(t, tx.Payer.String(), b.Address().String())
		assert.Equal(t, tx.Authorizers[0].String(), a.Address().String())
		assert.Equal(t, tx.ProposalKey.KeyIndex, a.Key().Index())
		assert.Nil(t, txr.Error)
		assert.Equal(t, txr.Status, flow.TransactionStatusSealed)
	})

	t.Run("Send Transaction with arguments", func(t *testing.T) {
		t.Parallel()
		state, flowkit := setupIntegration()
		setupAccounts(state, s)

		a, _ := state.Accounts().ByName("Alice")

		tx, txr, err := s.Transactions.Send(
			NewSingleTransactionAccount(a),
			NewScript(
				tests.TransactionArgString.Source,
				[]cadence.Value{
					cadence.String("Bar"),
				},
				tests.TransactionArgString.Filename,
			),
			flow.DefaultTransactionGasLimit,
			"",
		)
		assert.NoError(t, err)
		assert.Equal(t, tx.Payer.String(), a.Address().String())
		assert.Equal(t, tx.ProposalKey.KeyIndex, a.Key().Index())
		assert.Equal(t, fmt.Sprintf("%x", tx.Arguments), "[7b2276616c7565223a22426172222c2274797065223a22537472696e67227d]")
		assert.Nil(t, txr.Error)
		assert.Equal(t, txr.Status, flow.TransactionStatusSealed)
	})

	t.Run("Send transaction with multiple func declaration", func(t *testing.T) {
		t.Parallel()
		state, flowkit := setupIntegration()
		setupAccounts(state, s)

		a, _ := state.Accounts().ByName("Alice")

		tx, txr, err := s.Transactions.Send(
			NewSingleTransactionAccount(a),
			NewScript(
				tests.TransactionMultipleDeclarations.Source,
				nil,
				tests.TransactionMultipleDeclarations.Filename,
			),
			flow.DefaultTransactionGasLimit,
			"",
		)
		assert.NoError(t, err)
		assert.Equal(t, tx.Payer.String(), a.Address().String())
		assert.Equal(t, tx.ProposalKey.KeyIndex, a.Key().Index())
		assert.Nil(t, txr.Error)
		assert.Equal(t, txr.Status, flow.TransactionStatusSealed)
	})
}
