package inrequest

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// =============================================================================
// FormData Benchmarks
// =============================================================================

func BenchmarkFormDataMultipart(b *testing.B) {
	body, contentType := createMultipartForm(map[string]string{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   "30",
	}, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/test", bytes.NewReader(body))
		req.Header.Set("Content-Type", contentType)

		form := FormData(req)
		_ = form.ToMap()
		form.Cleanup()
	}
}

func BenchmarkFormDataURLEncoded(b *testing.B) {
	formData := url.Values{}
	formData.Set("name", "John Doe")
	formData.Set("email", "john@example.com")
	formData.Set("age", "30")
	formData.Set("active", "true")
	encoded := formData.Encode()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/test", strings.NewReader(encoded))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		form := FormData(req)
		_ = form.ToMap()
	}
}

func BenchmarkFormDataWithNestedFields(b *testing.B) {
	formData := url.Values{}
	formData.Set("user[name]", "John Doe")
	formData.Set("user[email]", "john@example.com")
	formData.Set("user[address][city]", "New York")
	formData.Set("user[address][zip]", "10001")
	formData.Set("user[tags][0]", "developer")
	formData.Set("user[tags][1]", "golang")
	formData.Set("user[tags][2]", "backend")
	encoded := formData.Encode()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/test", strings.NewReader(encoded))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		form := FormData(req)
		_ = form.ToMap()
	}
}

func BenchmarkFormDataWithFileUpload(b *testing.B) {
	body, contentType := createMultipartForm(
		map[string]string{"title": "My Upload"},
		map[string]string{"file": "This is the file content for benchmarking purposes."},
	)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/test", bytes.NewReader(body))
		req.Header.Set("Content-Type", contentType)

		form := FormData(req)
		_ = form.ToMap()
		form.Cleanup()
	}
}

func BenchmarkFormDataWithMultipleFiles(b *testing.B) {
	body, contentType := createMultipartFormWithMultipleFiles(
		map[string]string{"title": "Multiple Files"},
		"files",
		[]string{"file1 content", "file2 content", "file3 content", "file4 content", "file5 content"},
	)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/test", bytes.NewReader(body))
		req.Header.Set("Content-Type", contentType)

		form := FormData(req)
		_ = form.ToMap()
		form.Cleanup()
	}
}

func BenchmarkFormDataToBind(b *testing.B) {
	type User struct {
		Name   string `json:"name"`
		Email  string `json:"email"`
		Age    int    `json:"age"`
		Active bool   `json:"active"`
	}

	formData := url.Values{}
	formData.Set("name", "John Doe")
	formData.Set("email", "john@example.com")
	formData.Set("age", "30")
	formData.Set("active", "true")
	encoded := formData.Encode()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/test", strings.NewReader(encoded))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		form := FormData(req)
		var user User
		form.ToBind(&user)
	}
}

func BenchmarkFormDataToJsonByte(b *testing.B) {
	formData := url.Values{}
	formData.Set("name", "John Doe")
	formData.Set("email", "john@example.com")
	formData.Set("age", "30")
	encoded := formData.Encode()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/test", strings.NewReader(encoded))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		form := FormData(req)
		_, _ = form.ToJsonByte()
	}
}

// =============================================================================
// Query Benchmarks
// =============================================================================

func BenchmarkQuerySimple(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test?name=John&age=30&active=true", nil)
		query := Query(req)
		_ = query.ToMap()
	}
}

func BenchmarkQueryWithNestedParams(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test?user[name]=John&user[email]=john@example.com&user[address][city]=NYC&items[0]=a&items[1]=b&items[2]=c", nil)
		query := Query(req)
		_ = query.ToMap()
	}
}

func BenchmarkQueryWithManyParams(b *testing.B) {
	// Build URL with many parameters
	params := url.Values{}
	for i := 0; i < 50; i++ {
		params.Set("param"+string(rune('a'+i%26))+string(rune('0'+i/26)), "value")
	}
	urlStr := "/test?" + params.Encode()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", urlStr, nil)
		query := Query(req)
		_ = query.ToMap()
	}
}

func BenchmarkQueryToBind(b *testing.B) {
	type Params struct {
		Name   string `json:"name"`
		Age    int    `json:"age"`
		Active bool   `json:"active"`
	}

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test?name=John&age=30&active=true", nil)
		query := Query(req)
		var params Params
		query.ToBind(&params)
	}
}

func BenchmarkQueryToJsonByte(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test?name=John&age=30&active=true", nil)
		query := Query(req)
		_, _ = query.ToJsonByte()
	}
}

// =============================================================================
// JSON Benchmarks
// =============================================================================

func BenchmarkJsonSimple(b *testing.B) {
	body := `{"name":"John Doe","email":"john@example.com","age":30,"active":true}`

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		jsonReq, _ := Json(req)
		_ = jsonReq.ToMap()
	}
}

func BenchmarkJsonNested(b *testing.B) {
	body := `{
		"user": {
			"name": "John Doe",
			"email": "john@example.com",
			"address": {
				"city": "New York",
				"zip": "10001",
				"country": "USA"
			}
		},
		"items": [
			{"id": 1, "name": "Item 1"},
			{"id": 2, "name": "Item 2"},
			{"id": 3, "name": "Item 3"}
		],
		"metadata": {
			"version": "1.0",
			"timestamp": "2025-01-01T00:00:00Z"
		}
	}`

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		jsonReq, _ := Json(req)
		_ = jsonReq.ToMap()
	}
}

func BenchmarkJsonToBind(b *testing.B) {
	type User struct {
		Name   string `json:"name"`
		Email  string `json:"email"`
		Age    int    `json:"age"`
		Active bool   `json:"active"`
	}

	body := `{"name":"John Doe","email":"john@example.com","age":30,"active":true}`

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		jsonReq, _ := Json(req)
		var user User
		jsonReq.ToBind(&user)
	}
}

func BenchmarkJsonToJsonByte(b *testing.B) {
	body := `{"name":"John Doe","email":"john@example.com","age":30}`

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		jsonReq, _ := Json(req)
		_, _ = jsonReq.ToJsonByte()
	}
}

// =============================================================================
// Parse (Auto-detection) Benchmarks
// =============================================================================

func BenchmarkParseJSON(b *testing.B) {
	body := `{"name":"John","age":30}`

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		result, _ := Parse(req)
		_ = result.ToMap()
	}
}

func BenchmarkParseFormURLEncoded(b *testing.B) {
	formData := url.Values{}
	formData.Set("name", "John")
	formData.Set("age", "30")
	encoded := formData.Encode()

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/test", strings.NewReader(encoded))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		result, _ := Parse(req)
		_ = result.ToMap()
	}
}

func BenchmarkParseQuery(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test?name=John&age=30", nil)

		result, _ := Parse(req)
		_ = result.ToMap()
	}
}

// =============================================================================
// Utility Function Benchmarks
// =============================================================================

func BenchmarkTypeConversion(b *testing.B) {
	inputs := []string{"123", "45.67", "true", "false", "hello", "0", "007", "9999999999999"}

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, input := range inputs {
			_ = convertStringToActualType(input)
		}
	}
}

func BenchmarkBracketToDotConversion(b *testing.B) {
	inputs := []string{
		"name",
		"items[0]",
		"user[address][city]",
		"data[0][items][1][name]",
	}

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, input := range inputs {
			_ = replaceBracketKeyIntoDotKey(input)
		}
	}
}

func BenchmarkMapValuesOf(b *testing.B) {
	source := []GroupRequestProperty{
		{Path: "user[name]", Value: "John"},
		{Path: "user[email]", Value: "john@example.com"},
		{Path: "user[address][city]", Value: "NYC"},
		{Path: "user[address][zip]", Value: "10001"},
		{Path: "tags[0]", Value: "go"},
		{Path: "tags[1]", Value: "http"},
		{Path: "tags[2]", Value: "api"},
	}

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = mapValuesOf(source)
	}
}

func BenchmarkSparseArrayConversion(b *testing.B) {
	source := []GroupRequestProperty{
		{Path: "items[0]", Value: "first"},
		{Path: "items[5]", Value: "sixth"},
		{Path: "items[10]", Value: "eleventh"},
		{Path: "items[20]", Value: "twenty-first"},
	}

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = mapValuesOf(source)
	}
}

// =============================================================================
// Comparison Benchmarks (different sizes)
// =============================================================================

func BenchmarkFormDataSmall(b *testing.B) {
	formData := url.Values{}
	formData.Set("name", "John")
	encoded := formData.Encode()

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/test", strings.NewReader(encoded))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		form := FormData(req)
		_ = form.ToMap()
	}
}

func BenchmarkFormDataMedium(b *testing.B) {
	formData := url.Values{}
	for i := 0; i < 20; i++ {
		formData.Set("field"+string(rune('a'+i)), "value"+string(rune('0'+i)))
	}
	encoded := formData.Encode()

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/test", strings.NewReader(encoded))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		form := FormData(req)
		_ = form.ToMap()
	}
}

func BenchmarkFormDataLarge(b *testing.B) {
	formData := url.Values{}
	for i := 0; i < 100; i++ {
		formData.Set("field"+string(rune('a'+i%26))+string(rune('0'+i/26)), "value_"+strings.Repeat("x", 100))
	}
	encoded := formData.Encode()

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/test", strings.NewReader(encoded))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		form := FormData(req)
		_ = form.ToMap()
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func createMultipartForm(fields map[string]string, files map[string]string) ([]byte, string) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for key, value := range fields {
		writer.WriteField(key, value)
	}

	for filename, content := range files {
		part, _ := writer.CreateFormFile(filename, filename+".txt")
		io.Copy(part, strings.NewReader(content))
	}

	writer.Close()
	return body.Bytes(), writer.FormDataContentType()
}

func createMultipartFormWithMultipleFiles(fields map[string]string, fileField string, contents []string) ([]byte, string) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for key, value := range fields {
		writer.WriteField(key, value)
	}

	for i, content := range contents {
		part, _ := writer.CreateFormFile(fileField, "file"+string(rune('0'+i))+".txt")
		io.Copy(part, strings.NewReader(content))
	}

	writer.Close()
	return body.Bytes(), writer.FormDataContentType()
}

// =============================================================================
// Utility Function Benchmarks
// =============================================================================

func BenchmarkReplaceBracketKeyIntoDotKey_Simple(b *testing.B) {
	key := "name"
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = replaceBracketKeyIntoDotKey(key)
	}
}

func BenchmarkReplaceBracketKeyIntoDotKey_WithBrackets(b *testing.B) {
	key := "items[0][name]"
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = replaceBracketKeyIntoDotKey(key)
	}
}

func BenchmarkReplaceBracketKeyIntoDotKey_DeepNested(b *testing.B) {
	key := "data[users][0][profile][settings][notifications][email]"
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = replaceBracketKeyIntoDotKey(key)
	}
}

func BenchmarkConvertStringToActualType(b *testing.B) {
	testCases := []string{"123", "3.14", "true", "false", "null", "hello"}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			_ = convertStringToActualType(tc)
		}
	}
}
