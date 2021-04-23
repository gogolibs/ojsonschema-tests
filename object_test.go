package ojsonschema_tests

import (
	"context"
	"encoding/json"
	"github.com/gogolibs/ojson"
	"github.com/gogolibs/ojsonschema"
	"github.com/qri-io/jsonschema"
	"github.com/stretchr/testify/require"
	"testing"
)

type validationCase struct {
	name     string
	expected []jsonschema.KeyError
	actual   ojson.Anything
}

var schemaCases = []struct {
	name            string
	schema          ojson.Anything
	validationCases []validationCase
}{
	{
		name:   "string: simple",
		schema: ojsonschema.String{},
		validationCases: []validationCase{
			{
				name:     "just a string, no errors",
				actual:   "hello",
				expected: []jsonschema.KeyError{},
			},
			{
				name:   "integer instead of string",
				actual: 42,
				expected: []jsonschema.KeyError{
					{PropertyPath: "/", InvalidValue: 42, Message: "type should be string, got integer"},
				},
			},
		},
	},
	{
		name:   "string: enum",
		schema: ojsonschema.String{Enum: ojson.Array{"one", "two", "three"}},
		validationCases: []validationCase{
			{
				name:     "valid value",
				actual:   "three",
				expected: []jsonschema.KeyError{},
			},
			{
				name:   "invalid value",
				actual: "four",
				expected: []jsonschema.KeyError{
					{
						PropertyPath: "/",
						InvalidValue: "four",
						Message:      `should be one of ["one", "two", "three"]`,
					},
				},
			},
		},
	},
	{
		name: "object: single required field, no additional properties",
		schema: ojsonschema.Object{
			AdditionalProperties: false,
			Properties: ojson.Object{
				"field": ojsonschema.String{},
			},
			Required: ojson.Array{"field"},
		},
		validationCases: []validationCase{
			{
				name:     "valid case",
				actual:   ojson.Object{"field": "hello"},
				expected: []jsonschema.KeyError{},
			},
			{
				name:   "missing required field and unknown field is present",
				actual: ojson.Object{"unknown-field": "hello"},
				expected: []jsonschema.KeyError{
					{
						PropertyPath: "/",
						InvalidValue: map[string]interface{}{"unknown-field": "hello"},
						Message:      `"field" value is required`,
					},
					{
						PropertyPath: "/",
						InvalidValue: map[string]interface{}{"unknown-field": "hello"},
						Message:      "additional properties are not allowed",
					},
				},
			},
		},
	},
	{
		name: "const",
		schema: ojsonschema.Const("hello"),
		validationCases: []validationCase{
			{
				name: "valid value",
				expected: []jsonschema.KeyError{},
				actual: "hello",
			},
			{
				name: "invalid value",
				expected: []jsonschema.KeyError{
					{
						PropertyPath: "/",
						InvalidValue: "sup",
						Message:      `must equal "hello"`,
					},
				},
				actual: "sup",
			},
		},
	},
}

func TestSchemaCases(t *testing.T) {
	for _, schemaCase := range schemaCases {
		t.Run(schemaCase.name, func(t *testing.T) {
			schemaData := ojson.MustMarshal(schemaCase.schema)
			schema := new(jsonschema.Schema)
			err := json.Unmarshal(schemaData, schema)
			require.NoError(t, err)
			for _, validationCase := range schemaCase.validationCases {
				t.Run(validationCase.name, func(t *testing.T) {
					state := schema.Validate(context.Background(), validationCase.actual)
					require.Equal(t, validationCase.expected, *state.Errs)
				})
			}
		})
	}
}
