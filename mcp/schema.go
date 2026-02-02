package mcp

import (
	"reflect"
	"slices"
	"strconv"
	"strings"
)

// GenerateSchema takes a Go struct and returns a JSON Schema map suitable for MCP tool inputSchema.
// It uses reflection to inspect struct fields and looks for `json` and `jsonschema` tags.
//
// Pointers are treated as optional fields (not required) unless explicitly tagged.
// The `json:",omitempty"` tag is also used to determine if a field is optional.
func GenerateSchema(v any) map[string]any {
	t := reflect.TypeOf(v)
	if t == nil {
		return map[string]any{}
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return generateTypeSchema(t)
}

func generateTypeSchema(t reflect.Type) map[string]any {
	// Handle pointers by getting the underlying type
	isPointer := t.Kind() == reflect.Ptr
	if isPointer {
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

			jsonTag := field.Tag.Get("json")
			if jsonTag == "-" {
				continue
			}

			// Parse name and options from json tag
			tagParts := strings.Split(jsonTag, ",")
			name := tagParts[0]
			if name == "" {
				name = field.Name
			}

			omitempty := slices.Contains(tagParts[1:], "omitempty")

			fieldSchema := generateTypeSchema(field.Type)

			// Field is required if:
			// 1. It's not a pointer
			// 2. It doesn't have omitempty
			// 3. It's not explicitly marked as required=false in jsonschema tag
			isRequired := !isPointer && !omitempty && field.Type.Kind() != reflect.Ptr

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
							isRequired = (val == "true")
						case "enum":
							enums := strings.Split(val, ";")
							fieldSchema["enum"] = castSliceToType(enums, field.Type)
						case "default":
							fieldSchema["default"] = castToType(val, field.Type)
						}
					}
				}
			}

			if isRequired {
				required = append(required, name)
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

// castToType attempts to convert a string value from a tag into the actual Go type
// required by the field, ensuring the JSON Schema remains type-correct.
func castToType(val string, t reflect.Type) any {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if u, err := strconv.ParseUint(val, 10, 64); err == nil {
			return u
		}
	case reflect.Float32, reflect.Float64:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	case reflect.Bool:
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return val
}

// castSliceToType converts a slice of strings (from enum tag) into a slice of correctly typed values.
func castSliceToType(vals []string, t reflect.Type) []any {
	res := make([]any, len(vals))
	for i, v := range vals {
		res[i] = castToType(v, t)
	}
	return res
}
