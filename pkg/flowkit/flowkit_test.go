package flowkit

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/cadence/runtime/stdlib"
	emulator "github.com/onflow/flow-emulator"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/access/grpc"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/project"
	"github.com/onflow/flow-cli/pkg/flowkit/tests"
)

func Alice() *Account {
	return newAccount("Alice", "0x1", "seedseedseedseedseedseedseedseedseedseedseedseedAlice")
}

func Bob() *Account {
	return newAccount("Bob", "0x2", "seedseedseedseedseedseedseedseedseedseedseedseedBob")
}

func Charlie() *Account {
	return newAccount("Charlie", "0x3", "seedseedseedseedseedseedseedseedseedseedseedseedCharlie")
}

func Donald() *Account {
	return newAccount("Donald", "0x3", "seedseedseedseedseedseedseedseedseedseedseedseedDonald")
}

func newAccount(name string, address string, seed string) *Account {
	privateKey, _ := crypto.GeneratePrivateKey(crypto.ECDSA_P256, []byte(seed))

	account := NewAccount(name).
		SetAddress(flow.HexToAddress(address)).
		SetKey(
			NewHexAccountKeyFromPrivateKey(0, crypto.SHA3_256, privateKey),
		)

	return account
}

func setup() (*State, Flowkit, *tests.TestGateway) {
	readerWriter, _ := tests.ReaderWriter()
	state, err := Init(readerWriter, crypto.ECDSA_P256, crypto.SHA3_256)
	if err != nil {
		panic(err)
	}

	gw := tests.DefaultMockGateway()
	flowkit := Flowkit{
		state:   state,
		network: config.EmulatorNetwork,
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
}

func setupIntegration() (*State, Flowkit) {
	readerWriter, _ := tests.ReaderWriter()
	state, err := Init(readerWriter, crypto.ECDSA_P256, crypto.SHA3_256)
	if err != nil {
		panic(err)
	}

	acc, _ := state.EmulatorServiceAccount()
	pk, _ := acc.Key().PrivateKey()
	gw := gateway.NewEmulatorGatewayWithOpts(&gateway.EmulatorKey{
		PublicKey: (*pk).PublicKey(),
		SigAlgo:   acc.Key().SigAlgo(),
		HashAlgo:  acc.Key().HashAlgo(),
	}, gateway.WithEmulatorOptions(
		emulator.WithTransactionExpiry(10),
	))

	flowkit := Flowkit{
		state:   state,
		network: config.EmulatorNetwork,
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

		_, flowkit, _ := setup()
		key, err := flowkit.GenerateKey(ctx, crypto.ECDSA_P256, "")
		assert.NoError(t, err)
		assert.Equal(t, len(key.String()), 66)
	})

	t.Run("Generate Keys with seed", func(t *testing.T) {
		t.Parallel()

		_, flowkit, _ := setup()
		key, err := flowkit.GenerateKey(ctx, crypto.ECDSA_P256, "aaaaaaaaaaaaaaaaaaaaaaannndddddd_its_gone")

		assert.NoError(t, err)
		assert.Equal(t, key.String(), "0x134f702d0872dba9c7aea15498aab9b2ffedd5aeebfd8ac3cf47c591f0d7ce52")
	})

	t.Run("Test Vector SLIP-0010", func(t *testing.T) {
		// test against SLIP-0010 test vector. All data are taken from:
		//  https://github.com/satoshilabs/slips/blob/master/slip-0010.md#test-vectors
		t.Parallel()
		_, flowkit, _ := setup()

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
			privateKey, err := flowkit.derivePrivateKeyFromSeed(seed, test.sigAlgo, test.path)
			assert.NoError(t, err)
			assert.Equal(t, test.privateKey, privateKey.String())
		}
	})

	t.Run("Generate Keys with mnemonic (default path)", func(t *testing.T) {
		t.Parallel()

		_, flowkit, _ := setup()
		key, err := flowkit.DerivePrivateKeyFromMnemonic(ctx, "normal dune pole key case cradle unfold require tornado mercy hospital buyer", crypto.ECDSA_P256, "")

		assert.NoError(t, err)
		assert.Equal(t, key.String(), "0x638dc9ad0eee91d09249f0fd7c5323a11600e20d5b9105b66b782a96236e74cf")
	})

	//https://github.com/onflow/ledger-app-flow/blob/dc61213a9c3d73152b78b7391d04165d07f1ad89/tests_speculos/test-basic-show-address-expert.js#L28
	t.Run("Generate Keys with mnemonic (custom path - Ledger)", func(t *testing.T) {
		t.Parallel()

		_, flowkit, _ := setup()
		//ledger test mnemonic: https://github.com/onflow/ledger-app-flow#using-a-real-device-for-integration-tests-nano-s-and-nano-s-plus
		key, err := flowkit.DerivePrivateKeyFromMnemonic(ctx, "equip will roof matter pink blind book anxiety banner elbow sun young", crypto.ECDSA_secp256k1, "m/44'/539'/513'/0/0")

		assert.NoError(t, err)
		assert.Equal(t, key.String(), "0xd18d051afca7150781fef111f3387d132d31c4a6250268db0f61f836a205e0b8")

		assert.Equal(t, hex.EncodeToString(key.PublicKey().Encode()), "d7482bbaff7827035d5b238df318b10604673dc613808723efbd23fbc4b9fad34a415828d924ec7b83ac0eddf22ef115b7c203ee39fb080572d7e51775ee54be")
	})

	t.Run("Generate Keys Invalid", func(t *testing.T) {
		t.Parallel()

		_, flowkit, _ := setup()
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
				_, err := flowkit.GenerateKey(ctx, v, k)
				assert.Equal(t, err.Error(), errs[x])
				x++
			}
		}

	})
}

func TestProject(t *testing.T) {
	t.Parallel()

	t.Run("Deploy Project", func(t *testing.T) {
		t.Parallel()

		state, flowkit, gw := setup()

		c := config.Contract{
			Name:     "Hello",
			Location: tests.ContractHelloString.Filename,
		}
		state.Contracts().AddOrUpdate(c)
		state.Networks().AddOrUpdate(config.EmulatorNetwork)

		acct2 := Donald()
		state.Accounts().AddOrUpdate(acct2)

		d := config.Deployment{
			Network: config.EmulatorNetwork.Name,
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

		contracts, err := flowkit.DeployProject(ctx, false)

		assert.NoError(t, err)
		assert.Equal(t, len(contracts), 1)
		assert.Equal(t, contracts[0].AccountAddress, acct2.Address())
	})

	t.Run("Deploy Project Using LocationAliases", func(t *testing.T) {
		t.Parallel()

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
				Network: config.EmulatorNetwork.Name,
				Address: Donald().Address(),
			}},
		}
		state.Contracts().AddOrUpdate(c2)

		c3 := config.Contract{
			Name:     "ContractC",
			Location: tests.ContractC.Filename,
		}
		state.Contracts().AddOrUpdate(c3)

		state.Networks().AddOrUpdate(config.EmulatorNetwork)

		a := Alice()
		state.Accounts().AddOrUpdate(a)

		d := config.Deployment{
			Network: config.EmulatorNetwork.Name,
			Account: a.Name(),
			Contracts: []config.ContractDeployment{
				{Name: c1.Name}, {Name: c3.Name},
			},
		}
		state.Deployments().AddOrUpdate(d)

		// for checking imports are correctly resolved
		resolved := map[string]string{
			tests.ContractB.Name: fmt.Sprintf(`import ContractA from 0x%s`, Donald().Address().Hex()),
			tests.ContractC.Name: fmt.Sprintf(`
		import ContractB from 0x%s
		import ContractA from 0x%s`, a.Address().Hex(), Donald().Address().Hex()),
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

		contracts, err := flowkit.DeployProject(ctx, false)

		assert.NoError(t, err)
		assert.Equal(t, len(contracts), 2)
		gw.Mock.AssertCalled(t, tests.GetLatestBlockFunc)
		gw.Mock.AssertCalled(t, tests.GetAccountFunc, a.Address())
		gw.Mock.AssertNumberOfCalls(t, tests.GetTransactionResultFunc, 2)
	})

	t.Run("Deploy Project New Import Schema and LocationAliases", func(t *testing.T) {
		t.Parallel()

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
				Network: config.EmulatorNetwork.Name,
				Address: Donald().Address(),
			}},
		}
		state.Contracts().AddOrUpdate(c2)

		c3 := config.Contract{
			Name:     "ContractCC",
			Location: tests.ContractCC.Filename,
		}
		state.Contracts().AddOrUpdate(c3)

		state.Networks().AddOrUpdate(config.EmulatorNetwork)

		a := Alice()
		state.Accounts().AddOrUpdate(a)

		d := config.Deployment{
			Network: config.EmulatorNetwork.Name,
			Account: a.Name(),
			Contracts: []config.ContractDeployment{
				{Name: c1.Name}, {Name: c3.Name},
			},
		}
		state.Deployments().AddOrUpdate(d)

		// for checking imports are correctly resolved
		resolved := map[string]string{
			tests.ContractB.Name: fmt.Sprintf(`import ContractAA from 0x%s`, Donald().Address().Hex()),
			tests.ContractC.Name: fmt.Sprintf(`
		import ContractBB from 0x%s
		import ContractAA from 0x%s`, a.Address().Hex(), Donald().Address().Hex()),
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

		contracts, err := flowkit.DeployProject(ctx, false)

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
		state.Networks().AddOrUpdate(config.EmulatorNetwork)

		acct1 := Charlie()
		state.Accounts().AddOrUpdate(acct1)

		acct2 := Donald()
		state.Accounts().AddOrUpdate(acct2)

		d := config.Deployment{
			Network: config.EmulatorNetwork.Name,
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

		contracts, err := flowkit.DeployProject(ctx, false)

		assert.NoError(t, err)
		assert.Equal(t, len(contracts), 1)
		assert.Equal(t, contracts[0].AccountAddress, acct2.Address())
	})

}

// used for integration tests
func simpleDeploy(state *State, flowkit Flowkit, update bool) ([]*project.Contract, error) {
	srvAcc, _ := state.EmulatorServiceAccount()

	c := config.Contract{
		Name:     tests.ContractHelloString.Name,
		Location: tests.ContractHelloString.Filename,
	}
	state.Contracts().AddOrUpdate(c)

	state.Networks().AddOrUpdate(config.EmulatorNetwork)

	d := config.Deployment{
		Network: config.EmulatorNetwork.Name,
		Account: srvAcc.Name(),
		Contracts: []config.ContractDeployment{{
			Name: c.Name,
			Args: nil,
		}},
	}
	state.Deployments().AddOrUpdate(d)

	return flowkit.DeployProject(ctx, update)
}

func TestProject_Integration(t *testing.T) {
	t.Parallel()

	t.Run("Deploy Project", func(t *testing.T) {
		t.Parallel()

		state, flowkit := setupIntegration()
		contracts, err := simpleDeploy(state, flowkit, false)
		assert.NoError(t, err)
		assert.Len(t, contracts, 1)
		assert.Equal(t, contracts[0].Name, tests.ContractHelloString.Name)
		assert.Equal(t, string(contracts[0].Code()), string(tests.ContractHelloString.Source))
	})

	t.Run("Deploy Complex Project", func(t *testing.T) {
		t.Parallel()

		state, flowkit := setupIntegration()
		srvAcc, _ := state.EmulatorServiceAccount()
		state.Networks().AddOrUpdate(config.EmulatorNetwork)

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
				Network: config.EmulatorNetwork.Name,
				Address: srvAcc.Address(),
			}},
		})

		state.Deployments().AddOrUpdate(config.Deployment{
			Network: config.EmulatorNetwork.Name,
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
		_, _, err := flowkit.AddContract(
			ctx,
			srvAcc,
			NewScript(tests.ContractA.Source, nil, tests.ContractA.Filename),
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

		contracts, err := flowkit.DeployProject(ctx, false)
		assert.NoError(t, err)
		assert.Len(t, contracts, 2)

		account, err := flowkit.GetAccount(ctx, srvAcc.Address())

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
		_, err := simpleDeploy(state, flowkit, false)
		assert.NoError(t, err)

		_, err = simpleDeploy(state, flowkit, true)
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
		_, err := flowkit.ExecuteScript(
			ctx,
			NewScript(tests.ScriptArgString.Source, args, ""),
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
		res, err := flowkit.ExecuteScript(
			ctx,
			NewScript(tests.ScriptArgString.Source, args, ""),
		)

		assert.NoError(t, err)
		assert.Equal(t, "\"Hello Foo\"", res.String())
	})

	t.Run("Execute report error", func(t *testing.T) {
		t.Parallel()
		_, flowkit := setupIntegration()
		args := []cadence.Value{cadence.String("Foo")}
		res, err := flowkit.ExecuteScript(
			ctx,
			NewScript(tests.ScriptWithError.Source, args, ""),
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
		state.Networks().AddOrUpdate(config.EmulatorNetwork)

		d := config.Deployment{
			Network: config.EmulatorNetwork.Name,
			Account: srvAcc.Name(),
			Contracts: []config.ContractDeployment{{
				Name: c.Name,
				Args: nil,
			}},
		}
		state.Deployments().AddOrUpdate(d)
		_, _, _ = flowkit.AddContract(
			ctx,
			srvAcc,
			resourceToContract(tests.ContractHelloString),
			false,
		)

		res, err := flowkit.ExecuteScript(
			ctx,
			NewScript(tests.ScriptImport.Source, nil, tests.ScriptImport.Filename),
		)
		assert.NoError(t, err)
		assert.Equal(t, res.String(), "\"Hello Hello, World!\"")
	})

	t.Run("Execute Script Invalid", func(t *testing.T) {
		t.Parallel()
		_, flowkit := setupIntegration()
		in := []string{"", tests.ScriptImport.Filename}

		out := []string{
			"resolving imports in scripts not supported",
			"import ./contractHello.cdc could not be resolved from provided contracts",
		}

		for x, i := range in {
			_, err := flowkit.ExecuteScript(
				ctx,
				NewScript(tests.ScriptImport.Source, nil, i),
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
		_, flowkit, _ := setup()

		script := tests.TestScriptSimple
		results, err := flowkit.Test(ctx, script.Source, script.Filename)

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.NoError(t, results[0].Error)
	})

	t.Run("simple failing", func(t *testing.T) {
		t.Parallel()
		_, flowkit, _ := setup()

		script := tests.TestScriptSimpleFailing
		results, err := flowkit.Test(ctx, script.Source, script.Filename)

		require.NoError(t, err)
		require.Len(t, results, 1)

		err = results[0].Error
		require.Error(t, err)
		assert.ErrorAs(t, err, &stdlib.AssertionError{})
	})

	t.Run("with import", func(t *testing.T) {
		t.Parallel()
		st, flowkit, _ := setup()

		c := config.Contract{
			Name:     tests.ContractHelloString.Name,
			Location: tests.ContractHelloString.Filename,
		}
		st.Contracts().AddOrUpdate(c)

		// Execute script
		script := tests.TestScriptWithImport
		results, err := flowkit.Test(ctx, script.Source, script.Filename)

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.NoError(t, results[0].Error)
	})

	t.Run("with file read", func(t *testing.T) {
		t.Parallel()
		st, flowkit, _ := setup()

		readerWriter := st.ReaderWriter()
		readerWriter.WriteFile(
			tests.SomeFile.Filename,
			tests.SomeFile.Source,
			os.ModeTemporary,
		)

		// Execute script
		script := tests.TestScriptWithFileRead
		results, err := flowkit.Test(ctx, script.Source, script.Filename)

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

		_, _, err := flowkit.GetTransactionByID(ctx, txs.ID(), true)

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

		_, _, err := flowkit.SendTransaction(
			ctx,
			NewSingleTransactionAccount(serviceAcc),
			NewScript(
				tests.TransactionArgString.Source,
				[]cadence.Value{cadence.String("Bar")},
				"",
			),
			gasLimit,
		)

		assert.NoError(t, err)
		gw.Mock.AssertNumberOfCalls(t, tests.SendSignedTransactionFunc, 1)
		gw.Mock.AssertNumberOfCalls(t, tests.GetTransactionResultFunc, 1)
	})

}

func setupAccounts(state *State, flowkit Flowkit) {
	setupAccount(state, flowkit, Alice())
	setupAccount(state, flowkit, Bob())
	setupAccount(state, flowkit, Charlie())
}

func setupAccount(state *State, flowkit Flowkit, account *Account) {
	srv, _ := state.EmulatorServiceAccount()

	key := account.Key()
	pk, _ := key.PrivateKey()
	acc, _, _ := flowkit.CreateAccount(
		ctx,
		srv,
		[]Key{{
			(*pk).PublicKey(),
			flow.AccountKeyWeightThreshold,
			key.SigAlgo(),
			key.HashAlgo(),
		}},
	)

	newAcc := NewAccount(account.Name()).
		SetAddress(acc.Address).
		SetKey(key)

	state.Accounts().AddOrUpdate(newAcc)
}

func Test_TransactionRoles(t *testing.T) {
	t.Run("Building Signers", func(t *testing.T) {
		state, flowkit := setupIntegration()
		setupAccounts(state, flowkit)
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
		setupAccounts(state, flowkit)
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
		state, f := setupIntegration()
		setupAccounts(state, f)

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
			tx, err := f.BuildTransaction(
				ctx,
				&TransactionAddressesRoles{
					proposer:    txIn.prop,
					authorizers: txIn.auth,
					payer:       txIn.payer,
				},
				txIn.index,
				NewScript(txIn.code, txIn.args, txIn.file),
				txIn.gas,
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
		setupAccounts(state, flowkit)

		srvAcc, _ := state.EmulatorServiceAccount()
		signer := srvAcc.Address()

		// setup
		c := config.Contract{
			Name:     tests.ContractHelloString.Name,
			Location: tests.ContractHelloString.Filename,
		}
		state.Contracts().AddOrUpdate(c)
		state.Networks().AddOrUpdate(config.EmulatorNetwork)

		d := config.Deployment{
			Network: config.EmulatorNetwork.Name,
			Account: srvAcc.Name(),
			Contracts: []config.ContractDeployment{{
				Name: c.Name,
				Args: nil,
			}},
		}
		state.Deployments().AddOrUpdate(d)
		_, _, _ = flowkit.AddContract(
			ctx,
			srvAcc,
			resourceToContract(tests.ContractHelloString),
			false,
		)

		tx, err := flowkit.BuildTransaction(
			ctx,
			&TransactionAddressesRoles{
				signer,
				[]flow.Address{signer},
				signer,
			},
			srvAcc.Key().Index(),
			NewScript(tests.TransactionImports.Source, nil, tests.TransactionImports.Filename),
			flow.DefaultTransactionGasLimit,
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
		setupAccounts(state, flowkit)

		a, _ := state.Accounts().ByName("Alice")

		tx, err := flowkit.BuildTransaction(
			ctx,
			&TransactionAddressesRoles{
				a.Address(),
				nil,
				a.Address(),
			},
			0,
			NewScript(tests.TransactionSimple.Source, nil, tests.TransactionSimple.Filename),
			flow.DefaultTransactionGasLimit,
		)

		assert.Nil(t, err)
		assert.NotNil(t, tx)

		txSigned, err := flowkit.SignTransactionPayload(
			ctx,
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
		setupAccounts(state, flowkit)

		a, _ := state.Accounts().ByName("Alice")

		tx, err := flowkit.BuildTransaction(
			ctx,
			&TransactionAddressesRoles{
				a.Address(),
				[]flow.Address{a.Address()},
				a.Address(),
			},
			0,
			NewScript(tests.TransactionSingleAuth.Source, nil, tests.TransactionSingleAuth.Filename),
			flow.DefaultTransactionGasLimit,
		)

		assert.Nil(t, err)
		assert.NotNil(t, tx)

		txSigned, err := flowkit.SignTransactionPayload(
			ctx,
			a,
			[]byte(fmt.Sprintf("%x", tx.FlowTransaction().Encode())),
		)
		assert.Nil(t, err)
		assert.NotNil(t, txSigned)

		txSent, txResult, err := flowkit.SendSignedTransaction(ctx, txSigned)
		assert.Nil(t, err)
		assert.Equal(t, txResult.Status, flow.TransactionStatusSealed)
		assert.NotNil(t, txSent.ID())

	})

	t.Run("Fails signing transaction, wrong account", func(t *testing.T) {
		t.Parallel()
		state, flowkit := setupIntegration()
		setupAccounts(state, flowkit)

		a, _ := state.Accounts().ByName("Alice")

		tx, err := flowkit.BuildTransaction(
			ctx,
			&TransactionAddressesRoles{
				a.Address(),
				[]flow.Address{a.Address()},
				a.Address(),
			},
			0,
			NewScript(tests.TransactionSingleAuth.Source, nil, tests.TransactionSingleAuth.Filename),
			flow.DefaultTransactionGasLimit,
		)

		assert.Nil(t, err)
		assert.NotNil(t, tx)

		// sign with wrong account
		a, _ = state.Accounts().ByName("Bob")

		txSigned, err := flowkit.SignTransactionPayload(
			ctx,
			a,
			[]byte(fmt.Sprintf("%x", tx.FlowTransaction().Encode())),
		)
		assert.EqualError(t, err, "not a valid signer 179b6b1cb6755e31, proposer: 01cf0e2f2f715450, payer: 01cf0e2f2f715450, authorizers: [01cf0e2f2f715450]")
		assert.Nil(t, txSigned)
	})

	t.Run("Fails building, authorizers mismatch", func(t *testing.T) {
		t.Parallel()
		state, flowkit := setupIntegration()
		setupAccounts(state, flowkit)

		a, _ := state.Accounts().ByName("Alice")

		tx, err := flowkit.BuildTransaction(
			ctx,
			&TransactionAddressesRoles{
				proposer:    a.Address(),
				authorizers: []flow.Address{a.Address()},
				payer:       a.Address(),
			},
			0,
			NewScript(tests.TransactionTwoAuth.Source, nil, tests.TransactionTwoAuth.Filename),
			flow.DefaultTransactionGasLimit,
		)

		assert.EqualError(t, err, "provided authorizers length mismatch, required authorizers 2, but provided 1")
		assert.Nil(t, tx)
	})

	// todo(sideninja) we should convert different variations of sending transaction to table tests

	t.Run("Send Transaction No Auths", func(t *testing.T) {
		t.Parallel()
		state, flowkit := setupIntegration()
		setupAccounts(state, flowkit)

		a, _ := state.Accounts().ByName("Alice")

		tx, txr, err := flowkit.SendTransaction(
			ctx,
			&transactionAccountRoles{
				proposer: a,
				payer:    a,
			},
			NewScript(tests.TransactionSimple.Source, nil, tests.TransactionSimple.Filename),
			flow.DefaultTransactionGasLimit,
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
		setupAccounts(state, flowkit)

		a, _ := state.Accounts().ByName("Alice")

		tx, txr, err := flowkit.SendTransaction(
			ctx,
			NewSingleTransactionAccount(a),
			NewScript(tests.TransactionSingleAuth.Source, nil, tests.TransactionSingleAuth.Filename),
			flow.DefaultTransactionGasLimit,
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
		setupAccounts(state, flowkit)

		a, _ := state.Accounts().ByName("Alice")
		b, _ := state.Accounts().ByName("Bob")
		c, _ := state.Accounts().ByName("Charlie")

		roles, err := NewTransactionAccountRoles(a, b, []*Account{c})
		require.NoError(t, err)

		tx, txr, err := flowkit.SendTransaction(
			ctx,
			roles,
			NewScript(tests.TransactionSingleAuth.Source, nil, tests.TransactionSingleAuth.Filename),
			flow.DefaultTransactionGasLimit,
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
		setupAccounts(state, flowkit)

		a, _ := state.Accounts().ByName("Alice")
		b, _ := state.Accounts().ByName("Bob")

		roles, err := NewTransactionAccountRoles(a, b, []*Account{a})
		require.NoError(t, err)

		tx, txr, err := flowkit.SendTransaction(
			ctx,
			roles,
			NewScript(tests.TransactionSingleAuth.Source, nil, tests.TransactionSingleAuth.Filename),
			flow.DefaultTransactionGasLimit,
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
		setupAccounts(state, flowkit)

		a, _ := state.Accounts().ByName("Alice")

		tx, txr, err := flowkit.SendTransaction(
			ctx,
			NewSingleTransactionAccount(a),
			NewScript(
				tests.TransactionArgString.Source,
				[]cadence.Value{
					cadence.String("Bar"),
				},
				tests.TransactionArgString.Filename,
			),
			flow.DefaultTransactionGasLimit,
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
		setupAccounts(state, flowkit)

		a, _ := state.Accounts().ByName("Alice")

		tx, txr, err := flowkit.SendTransaction(
			ctx,
			NewSingleTransactionAccount(a),
			NewScript(
				tests.TransactionMultipleDeclarations.Source,
				nil,
				tests.TransactionMultipleDeclarations.Filename,
			),
			flow.DefaultTransactionGasLimit,
		)
		assert.NoError(t, err)
		assert.Equal(t, tx.Payer.String(), a.Address().String())
		assert.Equal(t, tx.ProposalKey.KeyIndex, a.Key().Index())
		assert.Nil(t, txr.Error)
		assert.Equal(t, txr.Status, flow.TransactionStatusSealed)
	})
}
