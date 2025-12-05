package inrequest

import "encoding/json"

type QueryRequest struct {
	result RequestValue
}

// Ensure QueryRequest implements Request interface
var _ Request = QueryRequest{}

func (r QueryRequest) ToMap() RequestValue {
	return r.result
}

func (r QueryRequest) ToBind(model interface{}) error {
	jsonData, err := json.Marshal(r.result)
	if err != nil {
		return NewBindError("", "failed to marshal query data", err)
	}
	if err = json.Unmarshal(jsonData, model); err != nil {
		return NewBindError("", "failed to unmarshal query data", err)
	}
	return nil
}

func (r QueryRequest) ToJsonByte() ([]byte, error) {
	jsonData, err := json.Marshal(r.result)
	if err != nil {
		return []byte{}, err
	}
	return jsonData, nil
}

func (r QueryRequest) ToJsonString() (string, error) {
	jsonData, err := json.Marshal(r.result)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
