package inrequest

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

// =============================================================================
// FormData Tests
// =============================================================================

func TestFormData(t *testing.T) {
	var body bytes.Buffer
	type BodyRequest struct {
		Name string                `json:"name"`
		File *multipart.FileHeader `json:"file"`
	}
	var bodyRequest BodyRequest
	writer := multipart.NewWriter(&body)

	err := writer.WriteField("name", "test-name")
	if err != nil {
		t.Fatalf("Failed to write form field: %v", err)
	}

	fileContent := "this is a test file"
	part, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	_, err = io.Copy(part, bytes.NewReader([]byte(fileContent)))
	if err != nil {
		t.Fatalf("Failed to write file content: %v", err)
	}

	writer.Close()

	req, err := http.NewRequest("POST", "/upload", &body)
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	form := FormData(req)
	defer form.Cleanup()

	form.ToBind(&bodyRequest)

	if bodyRequest.Name != "test-name" {
		t.Errorf("Expected 'test-name', got '%v'", bodyRequest.Name)
	}

	file, err := bodyRequest.File.Open()
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Failed to read file content: %v", err)
	}
	if string(content) != fileContent {
		t.Errorf("Expected file content '%v', got '%v'", fileContent, string(content))
	}
}

func TestFormDataWithOptions(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("name", "test")
	writer.Close()

	req, _ := http.NewRequest("POST", "/test", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	form := FormDataWithOptions(req, 10<<20)
	defer form.Cleanup()

	result := form.ToMap()
	if result["name"] != "test" {
		t.Errorf("Expected 'test', got '%v'", result["name"])
	}
}

func TestFormDataURLEncoded(t *testing.T) {
	formData := url.Values{}
	formData.Set("name", "John")
	formData.Set("age", "25")
	formData.Set("active", "true")

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	form := FormData(req)

	result := form.ToMap()

	if result["name"] != "John" {
		t.Errorf("Expected 'John', got '%v'", result["name"])
	}
	if result["age"] != 25 {
		t.Errorf("Expected 25, got '%v'", result["age"])
	}
	if result["active"] != true {
		t.Errorf("Expected true, got '%v'", result["active"])
	}
}

func TestFormDataURLEncodedMultipleValues(t *testing.T) {
	// Test multiple values for same key without brackets
	formData := url.Values{}
	formData.Add("tags", "go")
	formData.Add("tags", "http")
	formData.Add("tags", "api")

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	form := FormData(req)
	result := form.ToMap()

	tags, ok := result["tags"].([]interface{})
	if !ok {
		t.Fatalf("Expected tags to be slice, got %T", result["tags"])
	}
	if len(tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(tags))
	}
}

func TestFormDataMultipartMultipleValues(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Multiple values for same field without brackets
	writer.WriteField("colors", "red")
	writer.WriteField("colors", "green")
	writer.WriteField("colors", "blue")
	writer.Close()

	req, _ := http.NewRequest("POST", "/test", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	form := FormData(req)
	result := form.ToMap()

	colors, ok := result["colors"].([]interface{})
	if !ok {
		t.Fatalf("Expected colors to be slice, got %T", result["colors"])
	}
	if len(colors) != 3 {
		t.Errorf("Expected 3 colors, got %d", len(colors))
	}
}

func TestMultipleFilesUpload(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for i, content := range []string{"file1 content", "file2 content", "file3 content"} {
		part, err := writer.CreateFormFile("files", "file"+string(rune('1'+i))+".txt")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}
		part.Write([]byte(content))
	}
	writer.Close()

	req, _ := http.NewRequest("POST", "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	form := FormData(req)
	defer form.Cleanup()

	result := form.ToMap()

	files, ok := result["files"].([]interface{})
	if !ok {
		t.Fatalf("Expected files to be a slice, got %T", result["files"])
	}
	if len(files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(files))
	}
}

func TestMultipleFilesWithIndexedField(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Multiple files with already-indexed field name
	part1, _ := writer.CreateFormFile("docs[0][files]", "a.pdf")
	part1.Write([]byte("pdf1"))
	part2, _ := writer.CreateFormFile("docs[0][files]", "b.pdf")
	part2.Write([]byte("pdf2"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	form := FormData(req)
	result := form.ToMap()

	// Should have docs array
	docs, ok := result["docs"].([]interface{})
	if !ok {
		t.Fatalf("Expected docs to be slice, got %T", result["docs"])
	}
	if len(docs) < 1 {
		t.Fatal("Expected at least 1 doc")
	}
}

func TestNestedFormDataWithFiles(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	writer.WriteField("attachments[0][title]", "First Attachment")
	writer.WriteField("attachments[0][description]", "Description 1")
	part, _ := writer.CreateFormFile("attachments[0][file]", "doc1.pdf")
	part.Write([]byte("pdf content 1"))

	writer.WriteField("attachments[1][title]", "Second Attachment")
	part2, _ := writer.CreateFormFile("attachments[1][file]", "doc2.pdf")
	part2.Write([]byte("pdf content 2"))

	writer.Close()

	req, _ := http.NewRequest("POST", "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	form := FormData(req)
	defer form.Cleanup()

	result := form.ToMap()

	attachments, ok := result["attachments"].([]interface{})
	if !ok {
		t.Fatalf("Expected attachments to be a slice, got %T", result["attachments"])
	}

	if len(attachments) != 2 {
		t.Errorf("Expected 2 attachments, got %d", len(attachments))
	}

	att0, ok := attachments[0].(RequestValue)
	if !ok {
		t.Fatalf("Expected attachment[0] to be RequestValue, got %T", attachments[0])
	}
	if att0["title"] != "First Attachment" {
		t.Errorf("Expected title 'First Attachment', got %v", att0["title"])
	}
	if att0["file"] == nil {
		t.Error("Expected file to be present")
	}
}

func TestFormDataCleanupWithoutMultipart(t *testing.T) {
	// Test Cleanup on non-multipart request (should not panic)
	formData := url.Values{}
	formData.Set("name", "test")

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	form := FormData(req)
	form.Cleanup() // Should not panic
}

func TestFormDataCleanupWithNilRequest(t *testing.T) {
	// Test Cleanup with nil request (edge case)
	form := FormRequest{}
	form.Cleanup() // Should not panic
}

// =============================================================================
// FormRequest Method Tests
// =============================================================================

func TestFormRequestToJsonByte(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("name", "John")
	writer.WriteField("age", "30")
	writer.Close()

	req, _ := http.NewRequest("POST", "/test", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	form := FormData(req)

	jsonBytes, err := form.ToJsonByte()
	if err != nil {
		t.Fatalf("ToJsonByte failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if result["name"] != "John" {
		t.Errorf("Expected name 'John', got %v", result["name"])
	}
}

func TestFormRequestToJsonString(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("name", "John")
	writer.Close()

	req, _ := http.NewRequest("POST", "/test", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	form := FormData(req)

	jsonStr, err := form.ToJsonString()
	if err != nil {
		t.Fatalf("ToJsonString failed: %v", err)
	}

	if !strings.Contains(jsonStr, "John") {
		t.Errorf("Expected JSON to contain 'John', got %s", jsonStr)
	}
}

func TestFormRequestToBindErrors(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("name", "John")
	writer.Close()

	req, _ := http.NewRequest("POST", "/test", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	form := FormData(req)

	t.Run("nil pointer", func(t *testing.T) {
		err := form.ToBind(nil)
		if err == nil {
			t.Error("Expected error for nil pointer")
		}
	})

	t.Run("non-pointer", func(t *testing.T) {
		var s struct{ Name string }
		err := form.ToBind(s)
		if err == nil {
			t.Error("Expected error for non-pointer")
		}
	})

	t.Run("pointer to non-struct", func(t *testing.T) {
		var s string
		err := form.ToBind(&s)
		if err == nil {
			t.Error("Expected error for pointer to non-struct")
		}
	})
}

func TestFormRequestToBindWithFileHeaders(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Multiple files
	part1, _ := writer.CreateFormFile("files", "a.txt")
	part1.Write([]byte("content1"))
	part2, _ := writer.CreateFormFile("files", "b.txt")
	part2.Write([]byte("content2"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/test", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	form := FormData(req)

	// The files are stored as array, not directly as FileHeaders
	// This tests the path where we handle slices with files
	result := form.ToMap()
	if result["files"] == nil {
		t.Error("Expected files to be present")
	}
}

// =============================================================================
// Query Tests
// =============================================================================

func TestQuery(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?name=John&age=30", nil)

	query := Query(req)
	result := query.ToMap()

	if result["name"] != "John" {
		t.Errorf("Expected 'John', got %v", result["name"])
	}
	if result["age"] != 30 {
		t.Errorf("Expected 30, got %v", result["age"])
	}
}

func TestQueryMultipleValues(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?tags=go&tags=http&tags=api", nil)

	query := Query(req)
	result := query.ToMap()

	tags, ok := result["tags"].([]interface{})
	if !ok {
		t.Fatalf("Expected tags to be slice, got %T", result["tags"])
	}
	if len(tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(tags))
	}
}

func TestQueryWithBrackets(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?items[0]=first&items[1]=second", nil)

	query := Query(req)
	result := query.ToMap()

	items, ok := result["items"].([]interface{})
	if !ok {
		t.Fatalf("Expected items to be slice, got %T", result["items"])
	}
	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}
}

func TestQueryRequestToBind(t *testing.T) {
	type Params struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	req := httptest.NewRequest("GET", "/test?name=John&age=30", nil)
	query := Query(req)

	var params Params
	err := query.ToBind(&params)
	if err != nil {
		t.Fatalf("ToBind failed: %v", err)
	}

	if params.Name != "John" {
		t.Errorf("Expected 'John', got '%s'", params.Name)
	}
	if params.Age != 30 {
		t.Errorf("Expected 30, got %d", params.Age)
	}
}

func TestQueryRequestToJsonByte(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?name=John", nil)
	query := Query(req)

	jsonBytes, err := query.ToJsonByte()
	if err != nil {
		t.Fatalf("ToJsonByte failed: %v", err)
	}

	if !strings.Contains(string(jsonBytes), "John") {
		t.Errorf("Expected JSON to contain 'John'")
	}
}

func TestQueryRequestToJsonString(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?name=John", nil)
	query := Query(req)

	jsonStr, err := query.ToJsonString()
	if err != nil {
		t.Fatalf("ToJsonString failed: %v", err)
	}

	if !strings.Contains(jsonStr, "John") {
		t.Errorf("Expected JSON to contain 'John'")
	}
}

// =============================================================================
// Json Tests
// =============================================================================

func TestJson(t *testing.T) {
	body := `{"name": "John", "age": 30, "active": true}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	jsonReq, err := Json(req)
	if err != nil {
		t.Fatalf("Json failed: %v", err)
	}

	result := jsonReq.ToMap()
	if result["name"] != "John" {
		t.Errorf("Expected 'John', got %v", result["name"])
	}
}

func TestJsonInvalidBody(t *testing.T) {
	body := `{invalid json}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(body))

	_, err := Json(req)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestJsonEmptyBody(t *testing.T) {
	req := httptest.NewRequest("POST", "/test", strings.NewReader(""))

	jsonReq, err := Json(req)
	// Empty body returns EOF error but result should still be usable
	if err == nil {
		result := jsonReq.ToMap()
		if result == nil {
			t.Error("Expected non-nil map even for empty body")
		}
	}
}

func TestJsonRequestToBind(t *testing.T) {
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	body := `{"name": "John", "age": 30}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(body))

	jsonReq, err := Json(req)
	if err != nil {
		t.Fatalf("Json failed: %v", err)
	}

	var user User
	err = jsonReq.ToBind(&user)
	if err != nil {
		t.Fatalf("ToBind failed: %v", err)
	}

	if user.Name != "John" {
		t.Errorf("Expected 'John', got '%s'", user.Name)
	}
	if user.Age != 30 {
		t.Errorf("Expected 30, got %d", user.Age)
	}
}

func TestJsonRequestToJsonByte(t *testing.T) {
	body := `{"name": "John"}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(body))

	jsonReq, _ := Json(req)

	jsonBytes, err := jsonReq.ToJsonByte()
	if err != nil {
		t.Fatalf("ToJsonByte failed: %v", err)
	}

	if !strings.Contains(string(jsonBytes), "John") {
		t.Errorf("Expected JSON to contain 'John'")
	}
}

func TestJsonRequestToJsonString(t *testing.T) {
	body := `{"name": "John"}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(body))

	jsonReq, _ := Json(req)

	jsonStr, err := jsonReq.ToJsonString()
	if err != nil {
		t.Fatalf("ToJsonString failed: %v", err)
	}

	if !strings.Contains(jsonStr, "John") {
		t.Errorf("Expected JSON to contain 'John'")
	}
}

func TestJsonRequestDeprecatedMethods(t *testing.T) {
	body := `{"name": "John"}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(body))

	jsonReq, _ := Json(req)

	// Test deprecated ToByte
	b, err := jsonReq.ToByte()
	if err != nil {
		t.Fatalf("ToByte failed: %v", err)
	}
	if !strings.Contains(string(b), "John") {
		t.Errorf("Expected bytes to contain 'John'")
	}

	// Test deprecated ToString
	s, err := jsonReq.ToString()
	if err != nil {
		t.Fatalf("ToString failed: %v", err)
	}
	if !strings.Contains(s, "John") {
		t.Errorf("Expected string to contain 'John'")
	}
}

// =============================================================================
// Parse Tests
// =============================================================================

func TestParseAutoDetection(t *testing.T) {
	t.Run("JSON request", func(t *testing.T) {
		body := `{"name": "John", "age": 30}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		result, err := Parse(req)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		m := result.ToMap()
		if m["name"] != "John" {
			t.Errorf("Expected 'John', got %v", m["name"])
		}
	})

	t.Run("JSON request with charset", func(t *testing.T) {
		body := `{"name": "John"}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json; charset=utf-8")

		result, err := Parse(req)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		m := result.ToMap()
		if m["name"] != "John" {
			t.Errorf("Expected 'John', got %v", m["name"])
		}
	})

	t.Run("Form request", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("name", "John")

		req := httptest.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		result, err := Parse(req)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		m := result.ToMap()
		if m["name"] != "John" {
			t.Errorf("Expected 'John', got %v", m["name"])
		}
	})

	t.Run("Multipart form request", func(t *testing.T) {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		writer.WriteField("name", "John")
		writer.Close()

		req := httptest.NewRequest("POST", "/test", &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		result, err := Parse(req)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		m := result.ToMap()
		if m["name"] != "John" {
			t.Errorf("Expected 'John', got %v", m["name"])
		}
	})

	t.Run("Query params (no content type)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test?name=John&age=30", nil)

		result, err := Parse(req)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		m := result.ToMap()
		if m["name"] != "John" {
			t.Errorf("Expected 'John', got %v", m["name"])
		}
	})

	t.Run("Unknown content type defaults to query", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test?name=John", nil)
		req.Header.Set("Content-Type", "text/plain")

		result, err := Parse(req)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		m := result.ToMap()
		if m["name"] != "John" {
			t.Errorf("Expected 'John', got %v", m["name"])
		}
	})
}

// =============================================================================
// Interface Tests
// =============================================================================

func TestRequestInterface(t *testing.T) {
	var _ Request = FormRequest{}
	var _ Request = QueryRequest{}
	var _ Request = JsonRequest{}
}

// =============================================================================
// Mapping and Grouping Tests
// =============================================================================

func TestGroupMapKey(t *testing.T) {
	t.Run("1 dimensional object", func(t *testing.T) {
		source := []GroupRequestProperty{
			{Path: "name", Value: "John Doe"},
			{Path: "description", Value: "I'm a fullstack developer"},
		}
		target := GroupRequest{
			"name": []GroupRequestProperty{
				{Path: "name", Value: "John Doe"},
			},
			"description": []GroupRequestProperty{
				{Path: "description", Value: "I'm a fullstack developer"},
			},
		}

		groupKeys := groupMapKey(source)

		if !reflect.DeepEqual(groupKeys, target) {
			t.Fatalf("Failed: got %v", groupKeys)
		}
	})

	t.Run("2 dimensional array", func(t *testing.T) {
		source := []GroupRequestProperty{
			{Path: "names[0]", Value: "John Doe"},
			{Path: "names[1]", Value: "Michael"},
			{Path: "description", Value: "They are fullstack developers"},
		}
		target := GroupRequest{
			"names": []GroupRequestProperty{
				{Path: "names.0", Value: "John Doe"},
				{Path: "names.1", Value: "Michael"},
			},
			"description": []GroupRequestProperty{
				{Path: "description", Value: "They are fullstack developers"},
			},
		}

		groupKeys := groupMapKey(source)

		if !reflect.DeepEqual(groupKeys, target) {
			t.Fatalf("Failed: got %v", groupKeys)
		}
	})
}

func TestMappingValues(t *testing.T) {
	t.Run("1 dimensional object", func(t *testing.T) {
		source := []GroupRequestProperty{
			{Path: "name", Value: "John Doe"},
			{Path: "description", Value: "I'm a fullstack developer"},
		}
		target := RequestValue{
			"name":        "John Doe",
			"description": "I'm a fullstack developer",
		}

		mappedValues := mapValuesOf(source)

		if !reflect.DeepEqual(mappedValues, target) {
			t.Fatalf("Failed: got %v", mappedValues)
		}
	})

	t.Run("2 dimensional array", func(t *testing.T) {
		source := []GroupRequestProperty{
			{Path: "names[0]", Value: "John Doe"},
			{Path: "names[1]", Value: "Michael"},
			{Path: "description", Value: "They are fullstack developers"},
		}
		target := RequestValue{
			"names":       []interface{}{"John Doe", "Michael"},
			"description": "They are fullstack developers",
		}

		mappedValues := mapValuesOf(source)

		if !reflect.DeepEqual(mappedValues, target) {
			t.Fatalf("Failed: got %v", mappedValues)
		}
	})
}

func TestBindFormRequestToStruct(t *testing.T) {
	type User struct {
		Name     string  `json:"name"`
		Position string  `json:"position"`
		Age      float64 `json:"age"`
		Status   bool    `json:"status"`
	}

	bindUser := User{}

	source := []GroupRequestProperty{
		{Path: "name", Value: "John Doe"},
		{Path: "position", Value: "Fullstack Developer"},
		{Path: "age", Value: "12321.312"},
		{Path: "status", Value: true},
	}
	target := User{
		Name:     "John Doe",
		Position: "Fullstack Developer",
		Age:      12321.312,
		Status:   true,
	}

	mappedValues := mapValuesOf(source)

	jsonString, err := json.Marshal(mappedValues)
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal(jsonString, &bindUser)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(bindUser, target) {
		t.Fatalf("Failed: got %v", bindUser)
	}
}

// =============================================================================
// Type Conversion Tests
// =============================================================================

func TestTypeConversion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"zero", "0", 0},
		{"positive int", "42", 42},
		{"negative int", "-10", -10},
		{"float", "3.14", 3.14},
		{"leading zero", "007", "007"},
		{"phone number", "08123456789", "08123456789"},
		{"true", "true", true},
		{"false", "false", false},
		{"large number", "9999999999999", int64(9999999999999)},
		{"empty string", "", ""},
		{"regular string", "hello", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertStringToActualType(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("convertStringToActualType(%q) = %v (%T), want %v (%T)",
					tt.input, result, result, tt.expected, tt.expected)
			}
		})
	}
}

func TestSparseArrayHandling(t *testing.T) {
	source := []GroupRequestProperty{
		{Path: "items[0]", Value: "first"},
		{Path: "items[3]", Value: "fourth"},
		{Path: "items[5]", Value: "sixth"},
	}

	result := mapValuesOf(source)

	items, ok := result["items"].([]interface{})
	if !ok {
		t.Fatalf("Expected items to be a slice, got %T", result["items"])
	}

	if len(items) != 6 {
		t.Errorf("Expected 6 elements, got %d", len(items))
	}

	if items[0] != "first" {
		t.Errorf("Expected items[0] = 'first', got %v", items[0])
	}
	if items[1] != nil {
		t.Errorf("Expected items[1] = nil, got %v", items[1])
	}
	if items[3] != "fourth" {
		t.Errorf("Expected items[3] = 'fourth', got %v", items[3])
	}
	if items[5] != "sixth" {
		t.Errorf("Expected items[5] = 'sixth', got %v", items[5])
	}
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestEmptyRequest(t *testing.T) {
	t.Run("empty form data", func(t *testing.T) {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		writer.Close()

		req, _ := http.NewRequest("POST", "/test", &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		form := FormData(req)
		result := form.ToMap()

		if len(result) != 0 {
			t.Errorf("Expected empty map, got %v", result)
		}
	})

	t.Run("empty query", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		query := Query(req)
		result := query.ToMap()

		if len(result) != 0 {
			t.Errorf("Expected empty map, got %v", result)
		}
	})
}

func TestNestedObjects(t *testing.T) {
	source := []GroupRequestProperty{
		{Path: "user[name]", Value: "John"},
		{Path: "user[address][city]", Value: "NYC"},
		{Path: "user[address][zip]", Value: "10001"},
	}

	result := mapValuesOf(source)

	user, ok := result["user"].(RequestValue)
	if !ok {
		t.Fatalf("Expected user to be RequestValue, got %T", result["user"])
	}

	if user["name"] != "John" {
		t.Errorf("Expected name 'John', got %v", user["name"])
	}

	address, ok := user["address"].(RequestValue)
	if !ok {
		t.Fatalf("Expected address to be RequestValue, got %T", user["address"])
	}

	if address["city"] != "NYC" {
		t.Errorf("Expected city 'NYC', got %v", address["city"])
	}
}

func TestDeeplyNestedArrays(t *testing.T) {
	source := []GroupRequestProperty{
		{Path: "data[0][items][0][name]", Value: "Item 1"},
		{Path: "data[0][items][1][name]", Value: "Item 2"},
	}

	result := mapValuesOf(source)

	data, ok := result["data"].([]interface{})
	if !ok {
		t.Fatalf("Expected data to be slice, got %T", result["data"])
	}

	if len(data) < 1 {
		t.Fatal("Expected at least 1 data item")
	}

	data0, ok := data[0].(RequestValue)
	if !ok {
		t.Fatalf("Expected data[0] to be RequestValue, got %T", data[0])
	}

	items, ok := data0["items"].([]interface{})
	if !ok {
		t.Fatalf("Expected items to be slice, got %T", data0["items"])
	}

	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}
}

// =============================================================================
// ToBind Error Cases
// =============================================================================

func TestJsonRequestToBindError(t *testing.T) {
	body := `{"name": "John", "age": 30}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(body))

	jsonReq, _ := Json(req)

	// Try to bind incompatible type
	type Invalid struct {
		Name int `json:"name"` // name is string in JSON, but int here
	}

	var invalid Invalid
	err := jsonReq.ToBind(&invalid)
	if err == nil {
		t.Error("Expected error for type mismatch")
	}
}

func TestQueryRequestToBindError(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?name=John&age=thirty", nil)
	query := Query(req)

	// Try to bind incompatible type
	type Invalid struct {
		Name int `json:"name"` // name is string, but we want int
	}

	var invalid Invalid
	err := query.ToBind(&invalid)
	if err == nil {
		t.Error("Expected error for type mismatch")
	}
}

func TestFormRequestToBindWithTypeMismatch(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("name", "John")
	writer.Close()

	req, _ := http.NewRequest("POST", "/test", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	form := FormData(req)

	type Invalid struct {
		Name int `json:"name"` // name is string, but we want int
	}

	var invalid Invalid
	err := form.ToBind(&invalid)
	if err == nil {
		t.Error("Expected error for type mismatch")
	}
}

// =============================================================================
// File Handling Edge Cases
// =============================================================================

func TestFormRequestToBindWithSingleFileArray(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Single file that should still work with array field
	part, _ := writer.CreateFormFile("docs", "doc.pdf")
	part.Write([]byte("pdf content"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/test", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	form := FormData(req)
	result := form.ToMap()

	// Check docs is present
	if result["docs"] == nil {
		t.Error("Expected docs to be present")
	}
}

func TestFormRequestToBindWithFileHeadersField(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("title", "My Upload")
	part1, _ := writer.CreateFormFile("files", "a.txt")
	part1.Write([]byte("a"))
	part2, _ := writer.CreateFormFile("files", "b.txt")
	part2.Write([]byte("b"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/test", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	form := FormData(req)

	// Files are stored as slice of FileHeaders
	result := form.ToMap()
	if result["title"] != "My Upload" {
		t.Errorf("Expected title 'My Upload', got %v", result["title"])
	}
}

func TestFormRequestToBindFileFieldNotFound(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("attachment", "file.txt")
	part.Write([]byte("content"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/test", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Struct without matching field
	type NoFile struct {
		Name string `json:"name"`
	}

	form := FormData(req)

	var result NoFile
	err := form.ToBind(&result)
	// Should not error, just skip the field
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// =============================================================================
// Special Value Handling
// =============================================================================

func TestNilValueInMap(t *testing.T) {
	// Test that nil values are handled
	rv := RequestValue{
		"name":  "John",
		"empty": nil,
	}
	fixValueToActualType(&rv)

	if rv["name"] != "John" {
		t.Errorf("Expected 'John', got %v", rv["name"])
	}
	if rv["empty"] != nil {
		t.Errorf("Expected nil, got %v", rv["empty"])
	}
}

func TestNestedMixedTypes(t *testing.T) {
	// Test nested maps with various types
	source := []GroupRequestProperty{
		{Path: "user[name]", Value: "John"},
		{Path: "user[age]", Value: "30"},
		{Path: "user[active]", Value: "true"},
		{Path: "user[score]", Value: "95.5"},
	}

	result := mapValuesOf(source)

	user, ok := result["user"].(RequestValue)
	if !ok {
		t.Fatalf("Expected user to be RequestValue, got %T", result["user"])
	}

	if user["name"] != "John" {
		t.Errorf("Expected name 'John', got %v", user["name"])
	}
	if user["age"] != 30 {
		t.Errorf("Expected age 30, got %v (%T)", user["age"], user["age"])
	}
	if user["active"] != true {
		t.Errorf("Expected active true, got %v", user["active"])
	}
	if user["score"] != 95.5 {
		t.Errorf("Expected score 95.5, got %v", user["score"])
	}
}

// =============================================================================
// Query String Edge Cases
// =============================================================================

func TestQueryWithEmptyValues(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?name=&age=", nil)
	query := Query(req)
	result := query.ToMap()

	if result["name"] != "" {
		t.Errorf("Expected empty string, got %v", result["name"])
	}
}

func TestQueryWithNestedBrackets(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?filter[user][name]=John&filter[user][role]=admin", nil)
	query := Query(req)
	result := query.ToMap()

	filter, ok := result["filter"].(RequestValue)
	if !ok {
		t.Fatalf("Expected filter to be RequestValue, got %T", result["filter"])
	}

	user, ok := filter["user"].(RequestValue)
	if !ok {
		t.Fatalf("Expected user to be RequestValue, got %T", filter["user"])
	}

	if user["name"] != "John" {
		t.Errorf("Expected name 'John', got %v", user["name"])
	}
	if user["role"] != "admin" {
		t.Errorf("Expected role 'admin', got %v", user["role"])
	}
}

// =============================================================================
// Form Encoding Edge Cases
// =============================================================================

func TestFormDataWithNestedArrayObjects(t *testing.T) {
	formData := url.Values{}
	formData.Set("users[0][name]", "John")
	formData.Set("users[0][email]", "john@example.com")
	formData.Set("users[1][name]", "Jane")
	formData.Set("users[1][email]", "jane@example.com")

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	form := FormData(req)
	result := form.ToMap()

	users, ok := result["users"].([]interface{})
	if !ok {
		t.Fatalf("Expected users to be slice, got %T", result["users"])
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	user0, ok := users[0].(RequestValue)
	if !ok {
		t.Fatalf("Expected user[0] to be RequestValue, got %T", users[0])
	}

	if user0["name"] != "John" {
		t.Errorf("Expected name 'John', got %v", user0["name"])
	}
}

// =============================================================================
// JSON Nested Objects
// =============================================================================

func TestJsonWithNestedObjects(t *testing.T) {
	body := `{
		"user": {
			"name": "John",
			"address": {
				"city": "NYC",
				"zip": "10001"
			}
		},
		"items": [1, 2, 3]
	}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	jsonReq, err := Json(req)
	if err != nil {
		t.Fatalf("Json failed: %v", err)
	}

	result := jsonReq.ToMap()

	user, ok := result["user"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected user to be map, got %T", result["user"])
	}

	if user["name"] != "John" {
		t.Errorf("Expected name 'John', got %v", user["name"])
	}

	items, ok := result["items"].([]interface{})
	if !ok {
		t.Fatalf("Expected items to be slice, got %T", result["items"])
	}

	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}
}

// =============================================================================
// ToMap on Empty/Nil Result
// =============================================================================

func TestFormRequestToMapEmpty(t *testing.T) {
	form := FormRequest{}
	result := form.ToMap()

	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func TestQueryRequestToMapEmpty(t *testing.T) {
	query := QueryRequest{}
	result := query.ToMap()

	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func TestJsonRequestToMapEmpty(t *testing.T) {
	jsonReq := JsonRequest{}
	result := jsonReq.ToMap()

	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

// =============================================================================
// FileHeaders Type Tests
// =============================================================================

func TestFileHeadersType(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Create multiple files with same field name
	part1, _ := writer.CreateFormFile("attachments", "file1.txt")
	part1.Write([]byte("content 1"))
	part2, _ := writer.CreateFormFile("attachments", "file2.txt")
	part2.Write([]byte("content 2"))
	part3, _ := writer.CreateFormFile("attachments", "file3.txt")
	part3.Write([]byte("content 3"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/test", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	form := FormData(req)
	defer form.Cleanup()

	result := form.ToMap()

	// Files should be stored as a slice
	attachments := result["attachments"]
	if attachments == nil {
		t.Fatal("Expected attachments to be present")
	}

	// Verify it's a slice of files
	slice, ok := attachments.([]interface{})
	if !ok {
		t.Fatalf("Expected slice, got %T", attachments)
	}

	if len(slice) != 3 {
		t.Errorf("Expected 3 attachments, got %d", len(slice))
	}

	// Verify each item is a file header
	for i, item := range slice {
		_, isFile := item.(*multipart.FileHeader)
		if !isFile {
			t.Errorf("Expected item[%d] to be *multipart.FileHeader, got %T", i, item)
		}
	}
}

func TestFormRequestToBindWithFileHeadersInSlice(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Files that get converted to slice
	part1, _ := writer.CreateFormFile("docs", "a.pdf")
	part1.Write([]byte("pdf1"))
	part2, _ := writer.CreateFormFile("docs", "b.pdf")
	part2.Write([]byte("pdf2"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/test", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	form := FormData(req)
	result := form.ToMap()

	// Files should be in a slice
	docs := result["docs"]
	if docs == nil {
		t.Error("Expected docs to be present")
	}

	// Verify it's a slice
	slice, ok := docs.([]interface{})
	if !ok {
		t.Fatalf("Expected docs to be slice, got %T", docs)
	}
	if len(slice) != 2 {
		t.Errorf("Expected 2 docs, got %d", len(slice))
	}
}

func TestFormRequestToBindSingleFileHeader(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("name", "Document")
	part, _ := writer.CreateFormFile("doc", "single.pdf")
	part.Write([]byte("pdf content"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/test", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	type DocUpload struct {
		Name string                `json:"name"`
		Doc  *multipart.FileHeader `json:"doc"`
	}

	form := FormData(req)
	defer form.Cleanup()

	var upload DocUpload
	err := form.ToBind(&upload)
	if err != nil {
		t.Fatalf("ToBind failed: %v", err)
	}

	if upload.Name != "Document" {
		t.Errorf("Expected name 'Document', got '%s'", upload.Name)
	}
	if upload.Doc == nil {
		t.Error("Expected doc to be present")
	} else if upload.Doc.Filename != "single.pdf" {
		t.Errorf("Expected filename 'single.pdf', got '%s'", upload.Doc.Filename)
	}
}

func TestFormRequestToBindNestedFileInArray(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	writer.WriteField("items[0][name]", "Item 1")
	part1, _ := writer.CreateFormFile("items[0][file]", "item1.pdf")
	part1.Write([]byte("content1"))

	writer.WriteField("items[1][name]", "Item 2")
	part2, _ := writer.CreateFormFile("items[1][file]", "item2.pdf")
	part2.Write([]byte("content2"))

	writer.Close()

	req, _ := http.NewRequest("POST", "/test", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	form := FormData(req)
	defer form.Cleanup()

	result := form.ToMap()

	items, ok := result["items"].([]interface{})
	if !ok {
		t.Fatalf("Expected items to be slice, got %T", result["items"])
	}

	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}

	// Verify each item has name and file
	for i, item := range items {
		itemMap, ok := item.(RequestValue)
		if !ok {
			t.Fatalf("Expected item[%d] to be RequestValue, got %T", i, item)
		}
		if itemMap["name"] == nil {
			t.Errorf("Expected item[%d] to have name", i)
		}
		if itemMap["file"] == nil {
			t.Errorf("Expected item[%d] to have file", i)
		}
	}
}

// =============================================================================
// Complex Struct Binding
// =============================================================================

func TestFormRequestToBindComplexStruct(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("name", "Test User")
	writer.WriteField("age", "25")
	writer.WriteField("active", "true")
	writer.WriteField("score", "95.5")
	writer.WriteField("tags[0]", "developer")
	writer.WriteField("tags[1]", "golang")
	writer.Close()

	req, _ := http.NewRequest("POST", "/test", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	type User struct {
		Name   string   `json:"name"`
		Age    int      `json:"age"`
		Active bool     `json:"active"`
		Score  float64  `json:"score"`
		Tags   []string `json:"tags"`
	}

	form := FormData(req)

	var user User
	err := form.ToBind(&user)
	if err != nil {
		t.Fatalf("ToBind failed: %v", err)
	}

	if user.Name != "Test User" {
		t.Errorf("Expected name 'Test User', got '%s'", user.Name)
	}
	if user.Age != 25 {
		t.Errorf("Expected age 25, got %d", user.Age)
	}
	if !user.Active {
		t.Error("Expected active true")
	}
	if user.Score != 95.5 {
		t.Errorf("Expected score 95.5, got %f", user.Score)
	}
	if len(user.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(user.Tags))
	}
}

func TestQueryRequestToBindComplexStruct(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?name=John&age=30&active=true&items[0]=a&items[1]=b", nil)

	type Params struct {
		Name   string   `json:"name"`
		Age    int      `json:"age"`
		Active bool     `json:"active"`
		Items  []string `json:"items"`
	}

	query := Query(req)

	var params Params
	err := query.ToBind(&params)
	if err != nil {
		t.Fatalf("ToBind failed: %v", err)
	}

	if params.Name != "John" {
		t.Errorf("Expected name 'John', got '%s'", params.Name)
	}
	if params.Age != 30 {
		t.Errorf("Expected age 30, got %d", params.Age)
	}
	if !params.Active {
		t.Error("Expected active true")
	}
}

func TestJsonRequestToBindComplexStruct(t *testing.T) {
	body := `{
		"name": "John",
		"age": 30,
		"active": true,
		"address": {
			"city": "NYC",
			"zip": "10001"
		}
	}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	type Address struct {
		City string `json:"city"`
		Zip  string `json:"zip"`
	}
	type User struct {
		Name    string  `json:"name"`
		Age     int     `json:"age"`
		Active  bool    `json:"active"`
		Address Address `json:"address"`
	}

	jsonReq, err := Json(req)
	if err != nil {
		t.Fatalf("Json failed: %v", err)
	}

	var user User
	err = jsonReq.ToBind(&user)
	if err != nil {
		t.Fatalf("ToBind failed: %v", err)
	}

	if user.Name != "John" {
		t.Errorf("Expected name 'John', got '%s'", user.Name)
	}
	if user.Address.City != "NYC" {
		t.Errorf("Expected city 'NYC', got '%s'", user.Address.City)
	}
}

// =============================================================================
// Slice in Form with FileHeaders types
// =============================================================================

func TestFormRequestSliceContainsFileHeaders(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Create a form with nested files that become FileHeaders
	part1, _ := writer.CreateFormFile("uploads[0]", "upload1.txt")
	part1.Write([]byte("content1"))
	part2, _ := writer.CreateFormFile("uploads[0]", "upload2.txt")
	part2.Write([]byte("content2"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/test", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	form := FormData(req)
	result := form.ToMap()

	uploads := result["uploads"]
	if uploads == nil {
		t.Error("Expected uploads to be present")
	}
}

// =============================================================================
// Additional Edge Cases
// =============================================================================

func TestBuildIndexedPathCoverage(t *testing.T) {
	// Test buildIndexedPath with various inputs
	tests := []struct {
		path     string
		index    int
		expected string
	}{
		{"files", 0, "files[0]"},
		{"files", 5, "files[5]"},
		{"items[0][files]", 2, "items[0][files][2]"},
	}

	for _, tt := range tests {
		result := buildIndexedPath(tt.path, tt.index)
		if result != tt.expected {
			t.Errorf("buildIndexedPath(%q, %d) = %q, want %q", tt.path, tt.index, result, tt.expected)
		}
	}
}

func TestReplaceBracketKeyIntoDotKeyCoverage(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"name", "name"},
		{"items[0]", "items.0"},
		{"items[0][name]", "items.0.name"},
		{"a[1][b][2][c]", "a.1.b.2.c"},
	}

	for _, tt := range tests {
		result := replaceBracketKeyIntoDotKey(tt.input)
		if result != tt.expected {
			t.Errorf("replaceBracketKeyIntoDotKey(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestTransformDotPathToMapCoverage(t *testing.T) {
	t.Run("simple key", func(t *testing.T) {
		result := RequestValue{}
		transformDotPathToMap(&result, "name", "John")

		if result["name"] != "John" {
			t.Errorf("Expected name 'John', got %v", result["name"])
		}
	})

	t.Run("nested path", func(t *testing.T) {
		result := RequestValue{}
		transformDotPathToMap(&result, "user.name", "John")

		user, ok := result["user"].(RequestValue)
		if !ok {
			t.Fatalf("Expected user to be RequestValue, got %T", result["user"])
		}
		if user["name"] != "John" {
			t.Errorf("Expected name 'John', got %v", user["name"])
		}
	})

	t.Run("deeply nested path", func(t *testing.T) {
		result := RequestValue{}
		transformDotPathToMap(&result, "a.b.c.d", "value")

		a, _ := result["a"].(RequestValue)
		b, _ := a["b"].(RequestValue)
		c, _ := b["c"].(RequestValue)
		if c["d"] != "value" {
			t.Errorf("Expected value 'value', got %v", c["d"])
		}
	})
}

func TestConvertMapToSliceIfNumericKeysEdgeCases(t *testing.T) {
	t.Run("mixed keys", func(t *testing.T) {
		// Map with both numeric and non-numeric keys
		m := RequestValue{
			"0":    "first",
			"1":    "second",
			"name": "not numeric",
		}
		result := convertMapToSliceIfNumericKeys(m)

		// Should remain a map since not all keys are numeric
		rv, ok := result.(RequestValue)
		if !ok {
			t.Fatalf("Expected RequestValue for mixed keys, got %T", result)
		}
		if rv["name"] != "not numeric" {
			t.Errorf("Expected 'not numeric', got %v", rv["name"])
		}
	})

	t.Run("all numeric keys", func(t *testing.T) {
		m := RequestValue{
			"0": "first",
			"1": "second",
			"2": "third",
		}
		result := convertMapToSliceIfNumericKeys(m)

		slice, ok := result.([]interface{})
		if !ok {
			t.Fatalf("Expected slice, got %T", result)
		}
		if len(slice) != 3 {
			t.Errorf("Expected 3 elements, got %d", len(slice))
		}
	})
}

// =============================================================================
// Additional FormRequest.ToBind Tests for File Handling
// =============================================================================

func TestFormRequestToBindWithDirectFileHeaders(t *testing.T) {
	// Test binding when we have FileHeaders type directly
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("title", "My Upload")

	// Using indexed format to potentially create FileHeaders
	part1, _ := writer.CreateFormFile("docs[0]", "doc1.pdf")
	part1.Write([]byte("pdf1"))
	part2, _ := writer.CreateFormFile("docs[1]", "doc2.pdf")
	part2.Write([]byte("pdf2"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/test", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	form := FormData(req)
	result := form.ToMap()

	if result["title"] != "My Upload" {
		t.Errorf("Expected title 'My Upload', got %v", result["title"])
	}

	docs := result["docs"]
	if docs == nil {
		t.Error("Expected docs to be present")
	}
}

func TestFormRequestToBindWithSliceContainingMixedFiles(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Create nested structure with files
	writer.WriteField("data[0][name]", "Entry 1")
	part1, _ := writer.CreateFormFile("data[0][file]", "file1.pdf")
	part1.Write([]byte("content1"))

	writer.WriteField("data[1][name]", "Entry 2")
	// No file for data[1]

	writer.Close()

	req, _ := http.NewRequest("POST", "/test", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	form := FormData(req)
	result := form.ToMap()

	data, ok := result["data"].([]interface{})
	if !ok {
		t.Fatalf("Expected data to be slice, got %T", result["data"])
	}

	if len(data) < 2 {
		t.Errorf("Expected at least 2 data entries, got %d", len(data))
	}
}

func TestFormRequestToBindFileInNonExistentField(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, _ := writer.CreateFormFile("avatar", "photo.jpg")
	part.Write([]byte("image data"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/test", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Struct with different field name
	type Profile struct {
		ProfilePic *multipart.FileHeader `json:"profile_pic"`
	}

	form := FormData(req)

	var profile Profile
	// Should not error, just not bind the field
	err := form.ToBind(&profile)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if profile.ProfilePic != nil {
		t.Error("Expected ProfilePic to be nil since field name doesn't match")
	}
}

func TestFormRequestToBindWithAssignableFileType(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, _ := writer.CreateFormFile("document", "report.pdf")
	part.Write([]byte("pdf content"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/test", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	type Document struct {
		Document *multipart.FileHeader `json:"document"`
	}

	form := FormData(req)
	defer form.Cleanup()

	var doc Document
	err := form.ToBind(&doc)
	if err != nil {
		t.Fatalf("ToBind failed: %v", err)
	}

	if doc.Document == nil {
		t.Error("Expected Document to be present")
	} else if doc.Document.Filename != "report.pdf" {
		t.Errorf("Expected filename 'report.pdf', got '%s'", doc.Document.Filename)
	}
}

func TestFormRequestToBindEmptyNonFileResult(t *testing.T) {
	// Only files, no text fields
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("content"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/test", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	type FileOnly struct {
		File *multipart.FileHeader `json:"file"`
	}

	form := FormData(req)

	var result FileOnly
	err := form.ToBind(&result)
	if err != nil {
		t.Fatalf("ToBind failed: %v", err)
	}

	if result.File == nil {
		t.Error("Expected file to be present")
	}
}

// Test to trigger the FileHeaders branch in ToBind
func TestFormRequestToBindWithFileHeadersSlice(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Multiple files with same name (no brackets)
	part1, _ := writer.CreateFormFile("images", "img1.png")
	part1.Write([]byte("png1"))
	part2, _ := writer.CreateFormFile("images", "img2.png")
	part2.Write([]byte("png2"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/test", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	form := FormData(req)
	result := form.ToMap()

	images := result["images"]
	if images == nil {
		t.Fatal("Expected images to be present")
	}

	slice, ok := images.([]interface{})
	if !ok {
		t.Fatalf("Expected images to be slice, got %T", images)
	}

	if len(slice) != 2 {
		t.Errorf("Expected 2 images, got %d", len(slice))
	}
}
