package inrequest

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestGroupMapKey(t *testing.T) {
	t.Run("should get grouped keys from 1 dimensional object", func(t *testing.T) {
		var source = []GroupRequestProperty{
			{Path: "name", Value: "John Doe"},
			{Path: "description", Value: "I'm a fullstack developer"},
		}
		target := GroupRequest{
			"name": []GroupRequestProperty{
				{
					Path:  "name",
					Value: "John Doe",
				},
			},
			"description": []GroupRequestProperty{
				{
					Path:  "description",
					Value: "I'm a fullstack developer",
				},
			},
		}

		groupKeys := groupMapKey(source)

		if !reflect.DeepEqual(groupKeys, target) {
			t.Fatalf("Failed group keys %v, %v, got %v", source, target, groupKeys)
		}
	})
	t.Run("should get grouped keys from 1 dimension object and 2 dimensional array", func(t *testing.T) {
		var source = []GroupRequestProperty{
			{Path: "names[0]", Value: "John Doe"},
			{Path: "names[1]", Value: "Michael"},
			{Path: "description", Value: "They are fullstack developers"},
		}
		target := GroupRequest{
			"names": []GroupRequestProperty{
				{
					Path:  "names.0",
					Value: "John Doe",
				},
				{
					Path:  "names.1",
					Value: "Michael",
				},
			},
			"description": []GroupRequestProperty{
				{
					Path:  "description",
					Value: "They are fullstack developers",
				},
			},
		}

		groupKeys := groupMapKey(source)

		if !reflect.DeepEqual(groupKeys, target) {
			t.Fatalf("Failed group keys %v, %v, got %v", source, target, groupKeys)
		}
	})
}

func TestMappingValues(t *testing.T) {
	t.Run("should get 1 dimensional object", func(t *testing.T) {
		var source []GroupRequestProperty = []GroupRequestProperty{
			{Path: "name", Value: "John Doe"},
			{Path: "description", Value: "I'm a fullstack developer"},
		}
		target := RequestValue{
			"name":        "John Doe",
			"description": "I'm a fullstack developer",
		}

		mappedValues := mapValuesOf(source)

		if !reflect.DeepEqual(mappedValues, target) {
			t.Fatalf("Failed mapping values %v, %v, got %v", source, target, mappedValues)
		}
	})
	t.Run("should get 1 dimension object and 2 dimensional array", func(t *testing.T) {
		var source = []GroupRequestProperty{
			{Path: "names[0]", Value: "John Doe"},
			{Path: "names[1]", Value: "Michael"},
			{Path: "description", Value: "They are fullstack developers"},
		}
		target := RequestValue{
			"names": []interface{}{
				"John Doe",
				"Michael",
			},
			"description": "They are fullstack developers",
		}

		mappedValues := mapValuesOf(source)

		if !reflect.DeepEqual(mappedValues, target) {
			t.Fatalf("Failed mapping values %v, %v, got %v", source, target, mappedValues)
		}
	})
}

func TestBindFormRequestToStruct(t *testing.T) {
	type User struct {
		Name     string  `json:"name"`
		Position string  `json:"position"`
		Age      float64 `json:"age"`
		Status   bool    `json:"status"`
	}

	bindUser := User{}

	var source = []GroupRequestProperty{
		{Path: "name", Value: "John Doe"},
		{Path: "position", Value: "Fullstack Developer"},
		{Path: "age", Value: "12321.312"},
		{Path: "status", Value: true},
	}
	target := User{
		Name:     "John Doe",
		Position: "Fullstack Developer",
		Age:      12321.312,
		Status:   true,
	}

	mappedValues := mapValuesOf(source)

	jsonString, err := json.Marshal(mappedValues)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(jsonString))

	err = json.Unmarshal(jsonString, &bindUser)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("result", bindUser)

	if !reflect.DeepEqual(bindUser, target) {
		t.Fatalf("Failed bind values to struct %v, %v, got %v", source, target, bindUser)
	}
}
