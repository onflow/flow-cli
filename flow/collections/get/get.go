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
	Host string `default:"127.0.0.1:3569" flag:"host" info:"Flow Observation API host address"`
}

var conf Config

var Cmd = &cobra.Command{
	Use:   "get <collection_id>",
	Short: "Get collection info",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		collectionID := flow.HexToID(args[0])
		collection := cli.GetCollectionByID(conf.Host, collectionID)
		printCollection(collection)
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

func printCollection(collection *flow.Collection) {
	fmt.Println()
	fmt.Println("Collection ID: ", collection.ID())
	for i, transaction := range collection.TransactionIDs {
		fmt.Printf("  Transaction %d: %s\n", i, transaction)
	}
	fmt.Println()
}
