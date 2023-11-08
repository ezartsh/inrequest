package inrequest

import (
	"reflect"
	"sort"
	"strconv"
	"strings"
)

/*
transforming key value path into map
e.g. "path.to.this" : "value"
transform into :

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
Fixing value to the actual value type
e.g. RequestValue with key of numbers are transformed into slice of interface / []interface{}
*/
func fixValueToActualType(target *RequestValue) {
	t := *target
	for keyT, v := range t {
		if reflect.TypeOf(v).Kind() == reflect.Map {
			if vMap, ok := v.(RequestValue); ok {
				fixValueToActualType(&vMap)
				var arrMap []interface{}
				keys := make([]int, 0, len(vMap))
				for key, _ := range vMap {
					if intKey, err := strconv.Atoi(key); err == nil {
						keys = append(keys, intKey)
					}
				}
				sort.Ints(keys)
				for _, key := range keys {
					arrMap = append(arrMap, vMap[strconv.Itoa(key)])
				}
				if len(arrMap) > 0 {
					t[keyT] = arrMap
				}
			}
		}
	}
}

func replaceBracketKeyIntoDotKey(key string) string {
	replacer := strings.NewReplacer("]", "", "[", ".")
	return strings.Trim(replacer.Replace(key), ".")
}
