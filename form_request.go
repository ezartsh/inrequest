package inrequest

import (
	"encoding/json"
	"errors"
	"mime/multipart"
	"reflect"
	"strings"
)

type formRequest struct {
	result RequestValue
}

func (r formRequest) ToMap() RequestValue {
	return r.result
}

func (r formRequest) ToBind(model interface{}) error {
	modelValue := reflect.ValueOf(model)
	if modelValue.Kind() != reflect.Ptr || modelValue.IsNil() {
		return errors.New("model must be a non-nil pointer to a struct")
	}

	modelElem := modelValue.Elem()
	if modelElem.Kind() != reflect.Struct {
		return errors.New("model must be a pointer to a struct")
	}

	for key, value := range r.result {
		field := modelElem.FieldByNameFunc(func(name string) bool {
			return strings.EqualFold(name, key)
		})

		if !field.IsValid() || !field.CanSet() {
			continue // Skip if the field does not exist or cannot be set
		}

		// Handle *multipart.FileHeader explicitly
		if fileHeader, ok := value.(*multipart.FileHeader); ok {
			if field.Type() == reflect.TypeOf(fileHeader) {
				field.Set(reflect.ValueOf(fileHeader))
			}
			continue
		}

		// Handle other types (e.g., strings, ints, etc.)
		fieldValue := reflect.ValueOf(value)
		if field.Type().AssignableTo(fieldValue.Type()) {
			field.Set(fieldValue)
		}
	}

	return nil
}

func (r formRequest) ToJsonByte() ([]byte, error) {
	jsonData, err := json.Marshal(r.result)
	if err != nil {
		return []byte{}, err
	}
	return jsonData, nil
}

func (r formRequest) ToJsonString() (string, error) {
	jsonData, err := json.Marshal(r.result)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
