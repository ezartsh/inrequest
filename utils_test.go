package inrequest

import (
	"reflect"
	"testing"
)

// TestTransformDotPathToMap calls inrequest.transformDotPathToMap
func TestTransformDotPathToMap(t *testing.T) {
	source := RequestValue{
		"index.0":    0,
		"index.1":    1,
		"my.name.is": "elon",
	}
	target := RequestValue{
		"index": []interface{}{0, 1},
		"my": RequestValue{
			"name": RequestValue{
				"is": "elon",
			},
		},
	}

	for key, value := range source {
		transformDotPathToMap(&source, key, value)
	}

	fixValueToActualType(&source)

	if !reflect.DeepEqual(source, target) {
		t.Fatalf("Failed to transform dot path to map interface %v, %v", source, target)
	}
}

// TestReplaceBracketKeyIntoDotKey calls inrequest.replaceBracketKeyIntoDotKey
func TestReplaceBracketKeyIntoDotKey(t *testing.T) {
	caseValues := map[string]string{
		"path[to][this]": "path.to.this",
		"path[0][to]":    "path.0.to",
		"[0][to]":        "0.to",
	}

	for key, target := range caseValues {
		replacedKey := replaceBracketKeyIntoDotKey(key)
		if replacedKey != target {
			t.Fatalf(`Key %s failed to transform to %s, got %s`, key, target, replacedKey)
		}
	}
}

func TestConvertStringToActualType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"empty string", "", ""},
		{"regular string", "hello world", "hello world"},
		{"zero", "0", 0},
		{"positive integer", "123", 123},
		{"negative integer", "-456", -456},
		{"float", "3.14159", 3.14159},
		{"negative float", "-2.5", -2.5},
		{"true boolean", "true", true},
		{"false boolean", "false", false},
		{"leading zero preserved", "007", "007"},
		{"phone number preserved", "08123456789", "08123456789"},
		{"zip code preserved", "00123", "00123"},
		{"large int64", "9223372036854775807", int64(9223372036854775807)},
		{"float starting with 0.", "0.5", 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertStringToActualType(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("convertStringToActualType(%q) = %v (%T), want %v (%T)",
					tt.input, result, result, tt.expected, tt.expected)
			}
		})
	}
}

func TestConvertMapToSliceIfNumericKeys(t *testing.T) {
	t.Run("all numeric keys", func(t *testing.T) {
		input := RequestValue{
			"0": "first",
			"1": "second",
			"2": "third",
		}
		result := convertMapToSliceIfNumericKeys(input)
		expected := []interface{}{"first", "second", "third"}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("mixed keys returns map", func(t *testing.T) {
		input := RequestValue{
			"0":    "first",
			"name": "John",
		}
		result := convertMapToSliceIfNumericKeys(input)

		// Should return the original map since not all keys are numeric
		if _, ok := result.(RequestValue); !ok {
			t.Errorf("Expected RequestValue, got %T", result)
		}
	})

	t.Run("sparse array with gaps", func(t *testing.T) {
		input := RequestValue{
			"0": "first",
			"2": "third",
			"5": "sixth",
		}
		result := convertMapToSliceIfNumericKeys(input)
		slice, ok := result.([]interface{})
		if !ok {
			t.Fatalf("Expected slice, got %T", result)
		}

		if len(slice) != 6 {
			t.Errorf("Expected length 6, got %d", len(slice))
		}
		if slice[0] != "first" {
			t.Errorf("Expected slice[0] = 'first', got %v", slice[0])
		}
		if slice[1] != nil {
			t.Errorf("Expected slice[1] = nil, got %v", slice[1])
		}
		if slice[2] != "third" {
			t.Errorf("Expected slice[2] = 'third', got %v", slice[2])
		}
		if slice[5] != "sixth" {
			t.Errorf("Expected slice[5] = 'sixth', got %v", slice[5])
		}
	})

	t.Run("empty map", func(t *testing.T) {
		input := RequestValue{}
		result := convertMapToSliceIfNumericKeys(input)

		// Should return the original map
		if _, ok := result.(RequestValue); !ok {
			t.Errorf("Expected RequestValue for empty map, got %T", result)
		}
	})
}

func TestBuildIndexedPath(t *testing.T) {
	tests := []struct {
		name     string
		index    int
		expected string
	}{
		{"names", 0, "names[0]"},
		{"items", 10, "items[10]"},
		{"data", 999, "data[999]"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := buildIndexedPath(tt.name, tt.index)
			if result != tt.expected {
				t.Errorf("buildIndexedPath(%q, %d) = %q, want %q",
					tt.name, tt.index, result, tt.expected)
			}
		})
	}
}
