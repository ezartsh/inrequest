package inrequest

import "mime/multipart"

// RequestValue is the map type used to store parsed request data
type RequestValue = map[string]interface{}

type (
	GroupRequest         map[string][]GroupRequestProperty
	GroupRequestProperty struct {
		Path  string
		Value interface{}
	}
)

// Request is the common interface for all request types
type Request interface {
	ToMap() RequestValue
	ToBind(model interface{}) error
	ToJsonByte() ([]byte, error)
	ToJsonString() (string, error)
}

// FileHeaders represents multiple uploaded files for a single field
type FileHeaders []*multipart.FileHeader
