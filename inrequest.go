package go_inrequest

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func FormData(r *http.Request) formRequest {
	r.ParseMultipartForm(0)
	var forms []GroupRequestProperty

	if r.MultipartForm != nil {
		for name, val := range r.MultipartForm.Value {
			if strings.Contains(name, "[") || len(val) == 1 {
				forms = append(forms, GroupRequestProperty{Path: name, Value: val[0]})
			} else {
				for i, sVal := range val {
					forms = append(forms, GroupRequestProperty{Path: name + "[" + strconv.Itoa(i) + "]", Value: sVal})
				}
			}
		}
		for name := range r.MultipartForm.File {
			forms = append(forms, GroupRequestProperty{Path: name, Value: r.MultipartForm.File[name][0]})
		}
	}
	return formRequest{result: mapValuesOf(forms)}
}

func Query(r *http.Request) queryRequest {
	var forms []GroupRequestProperty
	var values = r.URL.Query()
	for key, value := range values {
		if strings.Contains(key, "[") || len(value) == 1 {
			forms = append(forms, GroupRequestProperty{Path: key, Value: value[0]})
		} else {
			for i, sVal := range value {
				forms = append(forms, GroupRequestProperty{Path: key + "[" + strconv.Itoa(i) + "]", Value: sVal})
			}
		}
	}
	return queryRequest{result: mapValuesOf(forms)}
}

func Json(r *http.Request) (jsonRequest, error) {
	var result RequestValue
	err := json.NewDecoder(r.Body).Decode(&result)

	return jsonRequest{result: result}, err
}

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
		var paths = strings.Split(dotKey, ".")
		mapQuery[paths[0]] = append(mapQuery[paths[0]], GroupRequestProperty{
			Path:  dotKey,
			Value: p.Value,
		})
	}
	return mapQuery
}
