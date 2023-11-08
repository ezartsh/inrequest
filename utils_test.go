package go_inrequest

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
		t.Fatalf("Failed to tranform dot path to map interface %v, %v", source, target)
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
