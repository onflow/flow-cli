module github.com/onflow/flow-cli

go 1.13

require (
	github.com/gosuri/uilive v0.0.4
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/onflow/cadence v0.12.5
	github.com/onflow/cadence/languageserver v0.12.5
	github.com/onflow/flow-emulator v0.14.1
	github.com/onflow/flow-go-sdk v0.14.2
	github.com/psiemens/sconfig v0.0.0-20190623041652-6e01eb1354fc
	github.com/sirupsen/logrus v1.5.0 // indirect
	github.com/spf13/cobra v0.0.7
	google.golang.org/grpc v1.32.0
)

replace github.com/fxamacker/cbor/v2 => github.com/turbolent/cbor/v2 v2.2.1-0.20200911003300-cac23af49154
