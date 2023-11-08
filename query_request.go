package inrequest

import "encoding/json"

type queryRequest struct {
	result RequestValue
}

func (r queryRequest) ToMap() RequestValue {
	return r.result
}

func (r queryRequest) ToBind(model interface{}) error {
	jsonData, err := json.Marshal(r.result)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(jsonData, &model); err != nil {
		return err
	}
	return nil
}

func (r queryRequest) ToJsonByte() ([]byte, error) {
	jsonData, err := json.Marshal(r.result)
	if err != nil {
		return []byte{}, err
	}
	return jsonData, nil
}

func (r queryRequest) ToJsonString() (string, error) {
	jsonData, err := json.Marshal(r.result)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
