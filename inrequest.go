package inrequest

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// Default max memory for multipart form parsing (32MB)
const DefaultMaxMemory = 32 << 20

// FormData parses multipart/form-data and application/x-www-form-urlencoded requests.
// It automatically handles nested fields, arrays, and file uploads.
//
// Basic usage:
//
//	form := inrequest.FormData(r)
//	data := form.ToMap()
//
// With struct binding:
//
//	var user User
//	inrequest.FormData(r).ToBind(&user)
//
// To clean up temporary files immediately (optional, for high-traffic servers):
//
//	form := inrequest.FormData(r)
//	defer form.Cleanup()
func FormData(r *http.Request) FormRequest {
	return FormDataWithOptions(r, DefaultMaxMemory)
}

// FormDataWithOptions parses form data with a custom max memory limit.
// maxMemory is the maximum bytes stored in memory (excess goes to temp files).
func FormDataWithOptions(r *http.Request, maxMemory int64) FormRequest {
	var forms []GroupRequestProperty

	contentType := r.Header.Get("Content-Type")

	// Handle multipart/form-data
	if strings.HasPrefix(contentType, "multipart/form-data") {
		err := r.ParseMultipartForm(maxMemory)
		if err == nil && r.MultipartForm != nil {
			forms = parseMultipartForm(r, forms)
		}
	} else {
		// Handle application/x-www-form-urlencoded
		err := r.ParseForm()
		if err == nil {
			forms = parseURLEncodedForm(r.PostForm, forms)
		}
	}

	return FormRequest{
		request: r,
		result:  mapValuesOf(forms),
	}
}

// parseMultipartForm extracts values and files from multipart form
func parseMultipartForm(r *http.Request, forms []GroupRequestProperty) []GroupRequestProperty {
	// Pre-allocate with estimated capacity
	if forms == nil {
		forms = make([]GroupRequestProperty, 0, len(r.MultipartForm.Value)+len(r.MultipartForm.File))
	}

	// Handle form values
	for name, val := range r.MultipartForm.Value {
		if strings.Contains(name, "[") || len(val) == 1 {
			forms = append(forms, GroupRequestProperty{Path: name, Value: val[0]})
		} else {
			for i, sVal := range val {
				forms = append(forms, GroupRequestProperty{Path: buildIndexedPath(name, i), Value: sVal})
			}
		}
	}

	// Handle files - support multiple files with same field name
	for name, files := range r.MultipartForm.File {
		if strings.Contains(name, "[") {
			// Already indexed (e.g., attachments[0][file])
			if len(files) == 1 {
				forms = append(forms, GroupRequestProperty{Path: name, Value: files[0]})
			} else {
				// Multiple files for indexed field - wrap as FileHeaders
				forms = append(forms, GroupRequestProperty{Path: name, Value: FileHeaders(files)})
			}
		} else {
			// Non-indexed field name
			if len(files) == 1 {
				forms = append(forms, GroupRequestProperty{Path: name, Value: files[0]})
			} else {
				// Multiple files - create indexed entries
				for i, file := range files {
					forms = append(forms, GroupRequestProperty{Path: buildIndexedPath(name, i), Value: file})
				}
			}
		}
	}

	return forms
}

// parseURLEncodedForm extracts values from URL-encoded form
func parseURLEncodedForm(values map[string][]string, forms []GroupRequestProperty) []GroupRequestProperty {
	if forms == nil {
		forms = make([]GroupRequestProperty, 0, len(values))
	}

	for name, val := range values {
		if strings.Contains(name, "[") || len(val) == 1 {
			forms = append(forms, GroupRequestProperty{Path: name, Value: val[0]})
		} else {
			for i, sVal := range val {
				forms = append(forms, GroupRequestProperty{Path: buildIndexedPath(name, i), Value: sVal})
			}
		}
	}

	return forms
}

// buildIndexedPath creates a path with an index suffix (e.g., "name[0]")
func buildIndexedPath(name string, index int) string {
	var sb strings.Builder
	sb.Grow(len(name) + 4) // name + "[" + digit(s) + "]"
	sb.WriteString(name)
	sb.WriteByte('[')
	sb.WriteString(strconv.Itoa(index))
	sb.WriteByte(']')
	return sb.String()
}

// Query parses URL query parameters.
//
// Usage:
//
//	query := inrequest.Query(r)
//	data := query.ToMap()
func Query(r *http.Request) QueryRequest {
	values := r.URL.Query()
	forms := make([]GroupRequestProperty, 0, len(values))

	for key, value := range values {
		if strings.Contains(key, "[") || len(value) == 1 {
			forms = append(forms, GroupRequestProperty{Path: key, Value: value[0]})
		} else {
			for i, sVal := range value {
				forms = append(forms, GroupRequestProperty{Path: buildIndexedPath(key, i), Value: sVal})
			}
		}
	}
	return QueryRequest{result: mapValuesOf(forms)}
}

// Json parses a JSON request body.
//
// Usage:
//
//	jsonReq, err := inrequest.Json(r)
//	if err != nil {
//	    // handle error
//	}
//	data := jsonReq.ToMap()
func Json(r *http.Request) (JsonRequest, error) {
	var result RequestValue
	err := json.NewDecoder(r.Body).Decode(&result)
	if result == nil {
		result = make(RequestValue)
	}
	return JsonRequest{result: result}, err
}

// Parse automatically detects the content type and parses the request.
//
// Usage:
//
//	req, err := inrequest.Parse(r)
//	if err != nil {
//	    // handle error (only for JSON)
//	}
//	data := req.ToMap()
func Parse(r *http.Request) (Request, error) {
	contentType := r.Header.Get("Content-Type")

	switch {
	case strings.HasPrefix(contentType, "application/json"):
		return Json(r)
	case strings.HasPrefix(contentType, "multipart/form-data"),
		strings.HasPrefix(contentType, "application/x-www-form-urlencoded"):
		return FormData(r), nil
	default:
		// Default to query params
		return Query(r), nil
	}
}
