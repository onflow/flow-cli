package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDependencies_ByName(t *testing.T) {
	dependencies := Dependencies{
		Dependency{Name: "mydep"},
	}

	dep := dependencies.ByName("mydep")
	assert.NotNil(t, dep)
}

func TestDependencies_AddOrUpdate(t *testing.T) {
	dependencies := Dependencies{}
	dependencies.AddOrUpdate(Dependency{Name: "mydep"})

	assert.Len(t, dependencies, 1)

	dep := dependencies.ByName("mydep")
	assert.NotNil(t, dep)
}
