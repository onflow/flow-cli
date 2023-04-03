module github.com/onflow/flow-cli

go 1.18

require (
	github.com/dukex/mixpanel v1.0.1
	github.com/getsentry/sentry-go v0.19.0
	github.com/go-git/go-git/v5 v5.6.1
	github.com/gosuri/uilive v0.0.4
	github.com/manifoldco/promptui v0.9.0
	github.com/onflow/cadence v0.37.0
	github.com/onflow/cadence-tools/languageserver v0.8.1-0.20230327102606-6be626d07eb8
	github.com/onflow/cadence-tools/test v0.6.0
	github.com/onflow/fcl-dev-wallet v0.6.0
	github.com/onflow/flow-cli/pkg/flowkit v0.0.0-20230327102447-8c34a92f8cbb
	github.com/onflow/flow-core-contracts/lib/go/templates v0.11.2-0.20221216161720-c1b31d5a4722
	github.com/onflow/flow-emulator v0.45.0
	github.com/onflow/flow-go-sdk v0.38.0
	github.com/onflowser/flowser/v2 v2.0.14-beta
	github.com/pkg/errors v0.9.1
	github.com/psiemens/sconfig v0.1.0
	github.com/radovskyb/watcher v1.0.7
	github.com/spf13/afero v1.9.5
	github.com/spf13/cobra v1.6.1
	github.com/spf13/viper v1.14.0
	github.com/stretchr/testify v1.8.2
	golang.org/x/exp v0.0.0-20221217163422-3c43f8badb15
	google.golang.org/grpc v1.54.0
)

require (

)

replace github.com/onflow/flow-cli/pkg/flowkit => ./pkg/flowkit

replace github.com/onflow/cadence-tools/languageserver => ../cadence-tools/languageserver

replace github.com/onflow/cadence-tools/lint => ../cadence-tools/lint
