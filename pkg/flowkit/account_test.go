package flowkit

import (
	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Accounts(t *testing.T) {

	t.Run("List accounts", func(t *testing.T) {
		accs := Accounts{
			Account{Name: "alice"},
			Account{Name: "bob"},
			Account{Name: "charlie"},
			Account{Name: "dave"},
		}

		assert.Equal(t, "alice,bob,charlie,dave", accs.String())
		assert.Equal(t, []string{"alice", "bob", "charlie", "dave"}, accs.Names())
	})

	t.Run("Get by name", func(t *testing.T) {
		accs := Accounts{
			Account{Name: "alice"},
			Account{Name: "bob"},
		}

		a, err := accs.ByName("alice")
		assert.NoError(t, err)
		assert.Equal(t, "alice", a.Name)
		assert.Equal(t, "0000000000000000", a.Address.String())
	})

	t.Run("Change address", func(t *testing.T) {
		accs := Accounts{
			Account{Name: "alice"},
			Account{Name: "bob"},
		}

		a, err := accs.ByName("alice")
		assert.NoError(t, err)
		newAddr := flow.HexToAddress("0x02")
		a.Address = newAddr
		// change gets reflected in the collection
		a2, _ := accs.ByName("alice")
		assert.Equal(t, "0000000000000002", a2.Address.String())
	})

	t.Run("Get by address", func(t *testing.T) {
		accs := Accounts{
			Account{Name: "alice"},
			Account{Name: "bob"},
		}

		a, _ := accs.ByName("alice")
		newAddr := flow.HexToAddress("0x02")
		a.Address = newAddr

		a2, err := accs.ByAddress(newAddr)
		assert.NoError(t, err)
		assert.Equal(t, "alice", a2.Name)
	})

	t.Run("Update", func(t *testing.T) {
		accs := Accounts{
			Account{Name: "alice"},
			Account{Name: "bob"},
		}

		accs.AddOrUpdate(&Account{Name: "mike"})
		assert.Equal(t, "alice,bob,mike", accs.String())

		m1, err := accs.ByName("mike")
		assert.NoError(t, err)
		assert.Equal(t, "0000000000000000", m1.Address.String())

		m1.Address = flow.HexToAddress("0x02")
		accs.AddOrUpdate(m1)
		m2, err := accs.ByName("mike")
		assert.NoError(t, err)
		assert.Equal(t, "0000000000000002", m2.Address.String())
	})

	t.Run("Remove", func(t *testing.T) {
		accs := Accounts{
			Account{Name: "alice"},
			Account{Name: "bob"},
			Account{Name: "mike"},
		}

		err := accs.Remove("mike")
		assert.NoError(t, err)
		assert.Equal(t, "alice,bob", accs.String())
	})

	t.Run("Fail not found", func(t *testing.T) {
		accs := Accounts{
			Account{Name: "alice"},
		}

		_, err := accs.ByName("bob")
		assert.EqualError(t, err, "could not find account with name bob in the configuration")

		_, err = accs.ByAddress(flow.HexToAddress("0x01"))
		assert.EqualError(t, err, "could not find account with address 0000000000000001 in the configuration")
	})

}
