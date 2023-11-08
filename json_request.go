package go_inrequest

import "encoding/json"

type jsonRequest struct {
	result RequestValue
}

func (r jsonRequest) ToMap() RequestValue {
	return r.result
}

func (r jsonRequest) ToBind(model interface{}) error {
	jsonData, err := json.Marshal(r.result)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(jsonData, &model); err != nil {
		return err
	}
	return nil
}

func (r jsonRequest) ToByte() ([]byte, error) {
	jsonData, err := json.Marshal(r.result)
	if err != nil {
		return []byte{}, err
	}
	return jsonData, nil
}

func (r jsonRequest) ToString() (string, error) {
	jsonData, err := json.Marshal(r.result)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
