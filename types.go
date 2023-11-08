package go_inrequest

type RequestValue = map[string]interface{}

type (
	GroupRequest         map[string][]GroupRequestProperty
	GroupRequestProperty struct {
		Path  string
		Value interface{}
	}
)
