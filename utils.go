package inrequest

import (
	"reflect"
	"sort"
	"strconv"
	"strings"
)

func mapValuesOf(queries []GroupRequestProperty) RequestValue {
	maps := make(RequestValue)
	mapQuery := groupMapKey(queries)
	for key, query := range mapQuery {
		if _, ok := maps[key]; !ok {
			maps[key] = make(RequestValue)
		}
		for _, v := range query {
			transformDotPathToMap(&maps, v.Path, v.Value)
		}
	}
	fixValueToActualType(&maps)
	return maps
}

func groupMapKey(data []GroupRequestProperty) GroupRequest {
	mapQuery := make(GroupRequest)
	for _, p := range data {
		dotKey := replaceBracketKeyIntoDotKey(p.Path)
		paths := strings.Split(dotKey, ".")
		mapQuery[paths[0]] = append(mapQuery[paths[0]], GroupRequestProperty{
			Path:  dotKey,
			Value: p.Value,
		})
	}
	return mapQuery
}

/*
transformDotPathToMap transforms a key value path into a nested map
e.g. "path.to.this" : "value"
transform into:

["path"] : {
["to"] : {
["this"] : "value"
}
}
*/
func transformDotPathToMap(target *RequestValue, dotPath string, value interface{}) {
	paths := strings.Split(dotPath, ".")
	t := *target
	if len(paths) == 1 {
		t[dotPath] = value
		return
	}
	if _, ok := t[paths[0]]; !ok {
		t[paths[0]] = make(RequestValue)
	}
	delete(t, dotPath)
	if query, ok := t[paths[0]].(RequestValue); ok {
		transformDotPathToMap(&query, strings.Join(paths[1:], "."), value)
	}
}

/*
fixValueToActualType fixes values to their actual types:
- RequestValue with numeric keys are transformed into []interface{}
- String values are converted to appropriate types (int64, float64, bool)
- Handles sparse arrays by filling gaps with nil
*/
func fixValueToActualType(target *RequestValue) {
	t := *target
	for keyT, v := range t {
		if v == nil {
			continue
		}

		if reflect.TypeOf(v).Kind() == reflect.Map {
			if vMap, ok := v.(RequestValue); ok {
				fixValueToActualType(&vMap)
				t[keyT] = convertMapToSliceIfNumericKeys(vMap)
			}
		} else if value, ok := v.(string); ok {
			t[keyT] = convertStringToActualType(value)
		}
	}
}

// convertMapToSliceIfNumericKeys converts a map with numeric keys to a slice
// Handles sparse arrays by filling gaps with nil
func convertMapToSliceIfNumericKeys(vMap RequestValue) interface{} {
	keys := make([]int, 0, len(vMap))
	allNumeric := true

	for key := range vMap {
		if intKey, err := strconv.Atoi(key); err == nil {
			keys = append(keys, intKey)
		} else {
			allNumeric = false
			break
		}
	}

	// If not all keys are numeric, or no numeric keys found, return the map as-is
	if !allNumeric || len(keys) == 0 {
		return vMap
	}

	sort.Ints(keys)

	// Determine the size of the array (max index + 1)
	maxIndex := keys[len(keys)-1]
	arrMap := make([]interface{}, maxIndex+1)

	// Fill in the values, gaps will be nil
	for _, key := range keys {
		arrMap[key] = vMap[strconv.Itoa(key)]
	}

	return arrMap
}

// convertStringToActualType converts a string to its appropriate type
func convertStringToActualType(value string) interface{} {
	// Handle empty strings
	if value == "" {
		return value
	}

	// Handle boolean strings
	if value == "true" {
		return true
	}
	if value == "false" {
		return false
	}

	// Handle numbers with leading zeros (except "0" itself)
	// Keep them as strings to preserve the leading zero
	if len(value) > 1 && value[0] == '0' && value[1] != '.' {
		return value
	}

	// Try to parse as float (includes integers with decimal points)
	if strings.Contains(value, ".") {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	} else {
		// Try to parse as int64 for large number support
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			// Return as int if it fits, otherwise int64
			if intValue >= int64(-1<<31) && intValue <= int64(1<<31-1) {
				return int(intValue)
			}
			return intValue
		}
	}

	return value
}

// bracketReplacer is a package-level reusable replacer for bracket-to-dot conversion.
// This avoids creating a new Replacer for each call, improving performance.
var bracketReplacer = strings.NewReplacer("]", "", "[", ".")

func replaceBracketKeyIntoDotKey(key string) string {
	// Fast path: if no brackets, return as-is
	if !strings.ContainsAny(key, "[]") {
		return key
	}
	return strings.Trim(bracketReplacer.Replace(key), ".")
}
