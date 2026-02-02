package mcp

import (
	"reflect"
	"strings"
)

// GenerateSchema takes a Go struct and returns a JSON Schema map suitable for MCP tool inputSchema.
// It uses reflection to inspect struct fields and looks for `json` and `jsonschema` tags.
//
// Example tags:
//
//	type MyInput struct {
//	    Name string `json:"name" jsonschema:"description=The user name,required=true"`
//	    Age  int    `json:"age" jsonschema:"description=The user age"`
//	}
func GenerateSchema(v any) map[string]any {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return generateTypeSchema(t)
}

func generateTypeSchema(t reflect.Type) map[string]any {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Struct:
		properties := make(map[string]any)
		var required []string

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath != "" { // Skip unexported fields
				continue
			}

			name := field.Tag.Get("json")
			if name == "" {
				name = field.Name
			}
			name = strings.Split(name, ",")[0]
			if name == "-" {
				continue
			}

			fieldSchema := generateTypeSchema(field.Type)

			// Parse jsonschema tags
			jsTag := field.Tag.Get("jsonschema")
			if jsTag != "" {
				parts := strings.SplitSeq(jsTag, ",")
				for part := range parts {
					kv := strings.SplitN(part, "=", 2)
					if len(kv) == 2 {
						key, val := kv[0], kv[1]
						switch key {
						case "description":
							fieldSchema["description"] = val
						case "required":
							if val == "true" {
								required = append(required, name)
							}
						case "enum":
							enums := strings.Split(val, ";")
							fieldSchema["enum"] = enums
						case "default":
							fieldSchema["default"] = val
						}
					}
				}
			}

			properties[name] = fieldSchema
		}

		schema := map[string]any{
			"type":       "object",
			"properties": properties,
		}
		if len(required) > 0 {
			schema["required"] = required
		}
		return schema

	case reflect.Slice, reflect.Array:
		return map[string]any{
			"type":  "array",
			"items": generateTypeSchema(t.Elem()),
		}

	case reflect.String:
		return map[string]any{"type": "string"}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{"type": "integer"}

	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}

	case reflect.Bool:
		return map[string]any{"type": "boolean"}

	case reflect.Map:
		return map[string]any{
			"type":                 "object",
			"additionalProperties": generateTypeSchema(t.Elem()),
		}

	case reflect.Interface:
		return map[string]any{} // Generic object/any

	default:
		return map[string]any{"type": "string"} // Fallback
	}
}
