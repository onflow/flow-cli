package config

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func Test_ParseAddress(t *testing.T) {
	tests := []string{
		"0x0000000000000002",
		"0x4c41cf317eec148e", // mainnet
		"0x78b84cd3c394708c", // testnet
		"0xf8d6e0586b0a20c7", // emulator
	}

	for _, test := range tests {
		addr, err := StringToAddress(test)
		require.NoError(t, err)
		assert.Equal(t, strings.TrimPrefix(test, "0x"), addr.Hex())
	}

}
