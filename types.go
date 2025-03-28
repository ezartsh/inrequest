package inrequest

type RequestValue = map[string]interface{}

type (
	GroupRequest         map[string][]GroupRequestProperty
	GroupRequestProperty struct {
		Path  string
		Value interface{}
	}
)

type CleanFormAction = func()
