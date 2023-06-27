package json

import (
	"github.com/invopop/jsonschema"
)

func GenerateSchema() *jsonschema.Schema {
	schema := jsonschema.Reflect(jsonConfig{})

	// Recursively move all definitions to the root of the schema
	// This is necessary because the jsonschema library does not support
	// definitions in nested schemas and is a workaround
	var moveDefinitions func(*jsonschema.Schema)
	moveDefinitions = func (s *jsonschema.Schema) {
		for k, v := range s.Definitions {
			schema.Definitions[k] = v
			moveDefinitions(v)
		}
		if (s != schema) {
			s.Definitions = nil
		}
	}
	moveDefinitions(schema)

	return schema
}
