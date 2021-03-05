package services

import (
	emulator "github.com/onflow/flow-emulator"
	"github.com/onflow/flow-go-sdk"
	"strings"
)

type EmulatorService struct {
	emulator *emulator.Blockchain
}

func NewEmulatorService() *EmulatorService {
	return &EmulatorService{
		emulator: newEmulator(),
	}
}

func newEmulator() *emulator.Blockchain {
	b, err := emulator.NewBlockchain()
	if err != nil {
		panic(err)
	}
	return b
}

func (s *EmulatorService) GetAccount(address string) (*flow.Account, error) {
	flowAddress := flow.HexToAddress(
		strings.ReplaceAll(address, "0x", ""),
	)

	return s.emulator.GetAccount(flowAddress)
}
