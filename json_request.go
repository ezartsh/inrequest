package inrequest

import "encoding/json"

type JsonRequest struct {
	result RequestValue
}

// Ensure JsonRequest implements Request interface
var _ Request = JsonRequest{}

func (r JsonRequest) ToMap() RequestValue {
	return r.result
}

func (r JsonRequest) ToBind(model interface{}) error {
	jsonData, err := json.Marshal(r.result)
	if err != nil {
		return NewBindError("", "failed to marshal JSON data", err)
	}
	if err = json.Unmarshal(jsonData, model); err != nil {
		return NewBindError("", "failed to unmarshal JSON data", err)
	}
	return nil
}

func (r JsonRequest) ToJsonByte() ([]byte, error) {
	jsonData, err := json.Marshal(r.result)
	if err != nil {
		return []byte{}, err
	}
	return jsonData, nil
}

func (r JsonRequest) ToJsonString() (string, error) {
	jsonData, err := json.Marshal(r.result)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// ToByte is deprecated, use ToJsonByte instead.
// Deprecated: Use ToJsonByte for consistent API.
func (r JsonRequest) ToByte() ([]byte, error) {
	return r.ToJsonByte()
}

// ToString is deprecated, use ToJsonString instead.
// Deprecated: Use ToJsonString for consistent API.
func (r JsonRequest) ToString() (string, error) {
	return r.ToJsonString()
}
