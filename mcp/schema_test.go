package mcp

import (
	"encoding/json"
	"testing"
)

type NestedStruct struct {
	Active bool `json:"active" jsonschema:"description=Is it active?"`
}

type TestInput struct {
	Name        string       `json:"name" jsonschema:"description=The name of the item,required=true"`
	Age         *int         `json:"age,omitempty" jsonschema:"description=Optional age"`
	Tags        []string     `json:"tags" jsonschema:"description=List of tags"`
	Status      string       `json:"status" jsonschema:"enum=TODO;PROG;DONE,default=TODO"`
	Count       int          `json:"count" jsonschema:"default=10"`
	Metadata    NestedStruct `json:"metadata"`
	Probability float64      `json:"probability" jsonschema:"default=0.5"`
	Internal    string       `json:"-"`
	unexported  string
}

func TestGenerateSchema(t *testing.T) {
	schema := GenerateSchema(TestInput{})

	// Verify top-level object
	if schema["type"] != "object" {
		t.Errorf("expected type object, got %v", schema["type"])
	}

	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("properties not found or not a map")
	}

	// Test required fields
	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("required fields not found")
	}

	isRequired := func(name string) bool {
		for _, r := range required {
			if r == name {
				return true
			}
		}
		return false
	}

	if !isRequired("name") {
		t.Error("field 'name' should be required")
	}
	if isRequired("age") {
		t.Error("field 'age' should not be required (pointer/omitempty)")
	}

	// Test nested struct
	metadata, ok := properties["metadata"].(map[string]any)
	if !ok {
		t.Fatal("nested metadata property not found")
	}
	if metadata["type"] != "object" {
		t.Errorf("nested metadata should be type object, got %v", metadata["type"])
	}

	// Test enum casting
	status, ok := properties["status"].(map[string]any)
	if !ok {
		t.Fatal("status property not found")
	}
	enums, ok := status["enum"].([]any)
	if !ok || len(enums) != 3 {
		t.Errorf("expected 3 enum values, got %v", status["enum"])
	}

	// Test default value casting
	count, ok := properties["count"].(map[string]any)
	if !ok {
		t.Fatal("count property not found")
	}
	if count["default"] != int64(10) {
		t.Errorf("expected default 10 (int64), got %T %v", count["default"], count["default"])
	}

	// Test float default casting
	prob, ok := properties["probability"].(map[string]any)
	if !ok {
		t.Fatal("probability property not found")
	}
	if prob["default"] != 0.5 {
		t.Errorf("expected default 0.5 (float64), got %T %v", prob["default"], prob["default"])
	}

	// Verify unexported and ignored fields are missing
	if _, ok := properties["Internal"]; ok {
		t.Error("field 'Internal' should be ignored via json tag '-'")
	}
	if _, ok := properties["unexported"]; ok {
		t.Error("unexported field should be ignored")
	}
}

func TestPointerHandling(t *testing.T) {
	type PointerInput struct {
		Val *string `json:"val"`
	}

	schema := GenerateSchema(&PointerInput{})
	properties := schema["properties"].(map[string]any)
	val := properties["val"].(map[string]any)

	if val["type"] != "string" {
		t.Errorf("expected pointer to string to have type string, got %v", val["type"])
	}

	if _, ok := schema["required"]; ok {
		t.Error("pointer field should not result in a required list")
	}
}

func TestManualRequiredOverride(t *testing.T) {
	type OverrideInput struct {
		Name string `json:"name" jsonschema:"required=false"`
	}

	schema := GenerateSchema(OverrideInput{})
	if _, ok := schema["required"]; ok {
		t.Error("field with required=false should not be in required list")
	}
}

func TestJsonOutput(t *testing.T) {
	// Simple visual check to ensure valid JSON is produced
	schema := GenerateSchema(TestInput{})
	_, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal schema to JSON: %v", err)
	}
}
