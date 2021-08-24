module github.com/onflow/flow-cli

go 1.13

replace github.com/onflow/cadence/languageserver => github.com/bjartek/cadence/languageserver v0.9.5-0.20210804215405-0397187261d5

require (
	github.com/a8m/envsubst v1.2.0
	github.com/gosuri/uilive v0.0.4
	github.com/joho/godotenv v1.3.0
	github.com/manifoldco/promptui v0.8.0
	github.com/onflow/cadence v0.18.1-0.20210621144040-64e6b6fb2337
	github.com/onflow/cadence/languageserver v0.18.2
	github.com/onflow/flow-core-contracts/lib/go/templates v0.6.0
	github.com/onflow/flow-emulator v0.22.0
	github.com/onflow/flow-go v0.18.4
	github.com/onflow/flow-go-sdk v0.20.1-0.20210623043139-533a95abf071
	github.com/psiemens/sconfig v0.0.0-20190623041652-6e01eb1354fc
	github.com/spf13/afero v1.1.2
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/thoas/go-funk v0.7.0
	golang.org/x/tools v0.1.4 // indirect
	gonum.org/v1/gonum v0.6.1
	google.golang.org/grpc v1.37.0
)
