package get

import (
	"fmt"
	"log"

	"github.com/onflow/flow-go-sdk"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"

	cli "github.com/dapperlabs/flow-cli/flow"
)

type Config struct {
	Host        string `flag:"host" info:"Flow Access API host address"`
	Latest      bool   `default:"false" flag:"latest" info:"Display latest block"`
	BlockID     string `default:"" flag:"id" info:"Display block by id"`
	BlockHeight uint64 `default:"0" flag:"height" info:"Display block by height"`
	Events      string `default:"" flag:"events" info:"List events of this type for the block"`
	Verbose     bool   `default:"false" flag:"verbose" info:"Display transactions in block"`
}

var conf Config

var Cmd = &cobra.Command{
	Use:   "get <block_id>",
	Short: "Get block info",
	Run: func(cmd *cobra.Command, args []string) {
		var block *flow.Block
		projectConf := new(cli.Config)
		if conf.Host == "" {
			projectConf = cli.LoadConfig()
		}
		host := projectConf.HostWithOverride(conf.Host)
		if conf.Latest {
			block = cli.GetLatestBlock(host)
		} else if len(conf.BlockID) > 0 {
			blockID := flow.HexToID(conf.BlockID)
			block = cli.GetBlockByID(host, blockID)
		} else if len(args) > 0 && len(args[0]) > 0 {
			blockID := flow.HexToID(args[0])
			block = cli.GetBlockByID(host, blockID)
		} else {
			block = cli.GetBlockByHeight(host, conf.BlockHeight)
		}
		printBlock(block, conf.Verbose)
		if conf.Events != "" {
			cli.GetBlockEvents(host, block.Height, conf.Events)
		}
	},
}

func init() {
	initConfig()
}

func initConfig() {
	err := sconfig.New(&conf).
		FromEnvironment(cli.EnvPrefix).
		BindFlags(Cmd.PersistentFlags()).
		Parse()
	if err != nil {
		log.Fatal(err)
	}
}

func printBlock(block *flow.Block, verbose bool) {
	fmt.Println()
	fmt.Println("Block ID: ", block.ID)
	fmt.Println("Parent ID: ", block.ParentID)
	fmt.Println("Height: ", block.Height)
	fmt.Println("Timestamp: ", block.Timestamp)
	fmt.Println("Total Collections: ", len(block.CollectionGuarantees))
	for i, guarantee := range block.CollectionGuarantees {
		fmt.Printf("  Collection %d: %s\n", i, guarantee.CollectionID)
		if verbose {
			collection := cli.GetCollectionByID(conf.Host, guarantee.CollectionID)
			for i, transaction := range collection.TransactionIDs {
				fmt.Printf("    Transaction %d: %s\n", i, transaction)
			}
		}
	}
	fmt.Println("Total Seals: ", len(block.Seals))
	fmt.Println()
}
