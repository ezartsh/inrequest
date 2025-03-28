package inrequest

import "encoding/json"

type JsonRequest struct {
	result RequestValue
}

func (r JsonRequest) ToMap() RequestValue {
	return r.result
}

func (r JsonRequest) ToBind(model interface{}) error {
	jsonData, err := json.Marshal(r.result)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(jsonData, &model); err != nil {
		return err
	}
	return nil
}

func (r JsonRequest) ToByte() ([]byte, error) {
	jsonData, err := json.Marshal(r.result)
	if err != nil {
		return []byte{}, err
	}
	return jsonData, nil
}

func (r JsonRequest) ToString() (string, error) {
	jsonData, err := json.Marshal(r.result)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
