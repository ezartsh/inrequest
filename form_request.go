package go_inrequest

import "encoding/json"

type formRequest struct {
	result RequestValue
}

func (r formRequest) ToMap() RequestValue {
	return r.result
}

func (r formRequest) ToBind(model interface{}) error {
	jsonData, err := json.Marshal(r.result)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(jsonData, &model); err != nil {
		return err
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
