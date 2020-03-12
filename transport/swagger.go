package transport

import (
	"github.com/go-openapi/spec"
	"github.com/severgroup-tt/gopkg-app/types"
	"github.com/utrack/clay/v2/transport/swagger"
)

const int64Type = "int64"

// SetIntegerTypeForInt64 replace type="string" to type="integer" for format="int64"
func SetIntegerTypeForInt64() swagger.Option {
	return func(swagger *spec.Swagger) {
		for _, definition := range swagger.Definitions {
			for propName, property := range definition.Properties {
				if property.Format == int64Type {
					definition.Properties[propName] = *spec.Int64Property()
				}
				if property.Items != nil && property.Items.Schema != nil && property.Items.Schema.Format == int64Type {
					definition.Properties[propName].Items.Schema = spec.Int64Property()
				}
			}
		}
	}
}

// Convert CamelCase names to snake_case
func SetNameSnakeCase() swagger.Option {
	return func(swagger *spec.Swagger) {
		for _, definition := range swagger.Definitions {
			for propName, property := range definition.Properties {
				propNameSnakeCase := types.CamelToSnakeCase(propName)
				if propNameSnakeCase != propName {
					definition.Properties[propNameSnakeCase] = property
					delete(definition.Properties, propName)
				}
			}
		}
	}
}
