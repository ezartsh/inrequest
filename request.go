package inrequest

import (
	"fmt"
	"net/http"
	"strings"
)

func GetFormRequest(r *http.Request) RequestValue {
	r.ParseMultipartForm(0)
	r.ParseForm()
	var forms []GroupRequestProperty
	for key, value := range r.Form {
		forms = append(forms, GroupRequestProperty{Path: key, Value: value})
	}
	if r.MultipartForm != nil {
		for name := range r.MultipartForm.File {
			forms = append(forms, GroupRequestProperty{Path: name, Value: r.MultipartForm.File[name][0]})
		}
	}
	return mapValuesOf(forms)
}

func mapValuesOf(queries []GroupRequestProperty) RequestValue {
	maps := make(RequestValue)
	mapQuery := groupMapKey(queries)
	fmt.Println(mapQuery)
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
