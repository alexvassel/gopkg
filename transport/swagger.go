package transport

import (
	"github.com/go-openapi/spec"
	"github.com/utrack/clay/v2/transport/swagger"
	"regexp"
	"strings"
)

const int64Type = "int64"

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

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
				propNameSnakeCase := toSnakeCase(propName)
				if propNameSnakeCase != propName {
					definition.Properties[propNameSnakeCase] = property
					delete(definition.Properties, propName)
				}
			}
		}
	}
}

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
