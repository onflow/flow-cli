module github.com/dapperlabs/flow-cli

go 1.13

require (
	github.com/dapperlabs/cadence v0.0.0-20200327205214-136b868762e2
	github.com/dapperlabs/flow-emulator v0.0.0-00010101000000-000000000000
	github.com/dapperlabs/flow-go-sdk v0.5.0
	github.com/dapperlabs/flow-go/crypto v0.3.2-0.20200312195452-df4550a863b7
	github.com/psiemens/sconfig v0.0.0-20190623041652-6e01eb1354fc
	github.com/sirupsen/logrus v1.5.0
	github.com/spf13/cobra v0.0.7
)

replace github.com/dapperlabs/flow-go-sdk => ../flow-go-sdk

replace github.com/dapperlabs/flow-emulator => ../flow-emulator

replace github.com/dapperlabs/cadence => ../cadence
