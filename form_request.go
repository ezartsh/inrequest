package inrequest

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"reflect"
	"strings"
)

type FormRequest struct {
	request *http.Request
	result  RequestValue
}

// Ensure FormRequest implements Request interface
var _ Request = FormRequest{}

// Cleanup removes temporary files created during multipart form parsing.
// This is optional - Go's garbage collector will eventually clean up,
// but calling this explicitly is recommended for high-traffic servers.
func (r FormRequest) Cleanup() {
	if r.request != nil && r.request.MultipartForm != nil {
		_ = r.request.MultipartForm.RemoveAll()
	}
}

func (r FormRequest) ToMap() RequestValue {
	return r.result
}

// ToBind binds the form data to a struct.
// It handles both regular fields via JSON marshaling and special types like *multipart.FileHeader.
func (r FormRequest) ToBind(model interface{}) error {
	modelValue := reflect.ValueOf(model)
	if modelValue.Kind() != reflect.Ptr || modelValue.IsNil() {
		return NewBindError("", "model must be a non-nil pointer to a struct", nil)
	}

	modelElem := modelValue.Elem()
	if modelElem.Kind() != reflect.Struct {
		return NewBindError("", "model must be a pointer to a struct", nil)
	}

	// First, handle file headers separately as they can't be JSON marshaled
	fileFields := make(map[string]interface{})
	nonFileResult := make(RequestValue)

	for key, value := range r.result {
		switch v := value.(type) {
		case *multipart.FileHeader:
			fileFields[key] = v
		case FileHeaders:
			fileFields[key] = v
		default:
			// Check if it's a slice containing file headers
			if slice, ok := value.([]interface{}); ok {
				hasFiles := false
				for _, item := range slice {
					if _, isFile := item.(*multipart.FileHeader); isFile {
						hasFiles = true
						break
					}
					if _, isFiles := item.(FileHeaders); isFiles {
						hasFiles = true
						break
					}
				}
				if hasFiles {
					fileFields[key] = value
				} else {
					nonFileResult[key] = value
				}
			} else {
				nonFileResult[key] = value
			}
		}
	}

	// Bind non-file fields via JSON
	if len(nonFileResult) > 0 {
		jsonData, err := json.Marshal(nonFileResult)
		if err != nil {
			return NewBindError("", "failed to marshal data", err)
		}
		if err = json.Unmarshal(jsonData, model); err != nil {
			return NewBindError("", "failed to unmarshal data", err)
		}
	}

	// Bind file fields directly via reflection
	for key, value := range fileFields {
		field := modelElem.FieldByNameFunc(func(name string) bool {
			return strings.EqualFold(name, key)
		})

		if !field.IsValid() || !field.CanSet() {
			continue
		}

		fieldValue := reflect.ValueOf(value)
		switch {
		case field.Type().AssignableTo(fieldValue.Type()):
			field.Set(fieldValue)
		case field.Type() == reflect.TypeOf((*multipart.FileHeader)(nil)):
			// Handle single file
			if fh, ok := value.(*multipart.FileHeader); ok {
				field.Set(reflect.ValueOf(fh))
			}
		case field.Type() == reflect.TypeOf(FileHeaders(nil)):
			// Handle multiple files
			if fhs, ok := value.(FileHeaders); ok {
				field.Set(reflect.ValueOf(fhs))
			}
		}
	}

	return nil
}

func (r FormRequest) ToJsonByte() ([]byte, error) {
	jsonData, err := json.Marshal(r.result)
	if err != nil {
		return []byte{}, err
	}
	return jsonData, nil
}

func (r FormRequest) ToJsonString() (string, error) {
	jsonData, err := json.Marshal(r.result)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
