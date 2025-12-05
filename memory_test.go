package inrequest

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"testing"
)

// =============================================================================
// Memory Leak Tests
// =============================================================================

func TestMemoryLeakMultipartWithCleanup(t *testing.T) {
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	for i := 0; i < 1000; i++ {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		writer.WriteField("name", "John Doe")
		part, _ := writer.CreateFormFile("file", "test.txt")
		part.Write([]byte("file content here for testing memory"))
		writer.Close()

		req, _ := http.NewRequest("POST", "/test", &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		form := FormData(req)
		_ = form.ToMap()
		form.Cleanup()
	}

	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	growth := int64(m2.Alloc) - int64(m1.Alloc)
	t.Logf("Memory growth WITH Cleanup (1k multipart): %d bytes (%d KB)", growth, growth/1024)

	if growth > 500*1024 {
		t.Errorf("Potential memory leak: grew %d KB", growth/1024)
	}
}

func TestMemoryLeakMultipartWithoutCleanup(t *testing.T) {
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	for i := 0; i < 1000; i++ {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		writer.WriteField("name", "John Doe")
		part, _ := writer.CreateFormFile("file", "test.txt")
		part.Write([]byte("file content here for testing memory"))
		writer.Close()

		req, _ := http.NewRequest("POST", "/test", &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		form := FormData(req)
		_ = form.ToMap()
		// NO Cleanup - temp files may accumulate
	}

	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	growth := int64(m2.Alloc) - int64(m1.Alloc)
	t.Logf("Memory growth WITHOUT Cleanup (1k multipart): %d bytes (%d KB)", growth, growth/1024)
}

func TestMemoryLeakURLEncoded(t *testing.T) {
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	for i := 0; i < 10000; i++ {
		formData := url.Values{}
		formData.Set("name", "John Doe")
		formData.Set("email", "john@example.com")
		formData.Set("items[0]", "item1")
		formData.Set("items[1]", "item2")

		req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		form := FormData(req)
		_ = form.ToMap()
	}

	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	growth := int64(m2.Alloc) - int64(m1.Alloc)
	t.Logf("Memory growth URL-encoded (10k requests): %d bytes (%d KB)", growth, growth/1024)

	if growth > 1024*1024 {
		t.Errorf("Potential memory leak: grew %d KB", growth/1024)
	}
}

func TestMemoryLeakJSON(t *testing.T) {
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	for i := 0; i < 10000; i++ {
		body := `{"name":"John","email":"john@example.com","age":30,"items":["a","b","c"]}`
		req, _ := http.NewRequest("POST", "/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		jsonReq, _ := Json(req)
		_ = jsonReq.ToMap()
	}

	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	growth := int64(m2.Alloc) - int64(m1.Alloc)
	t.Logf("Memory growth JSON (10k requests): %d bytes (%d KB)", growth, growth/1024)

	if growth > 1024*1024 {
		t.Errorf("Potential memory leak: grew %d KB", growth/1024)
	}
}

func TestMemoryLeakQuery(t *testing.T) {
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	for i := 0; i < 10000; i++ {
		req, _ := http.NewRequest("GET", "/test?name=John&email=john@example.com&items[0]=a&items[1]=b", nil)
		query := Query(req)
		_ = query.ToMap()
	}

	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	growth := int64(m2.Alloc) - int64(m1.Alloc)
	t.Logf("Memory growth Query (10k requests): %d bytes (%d KB)", growth, growth/1024)

	if growth > 1024*1024 {
		t.Errorf("Potential memory leak: grew %d KB", growth/1024)
	}
}

// =============================================================================
// CPU Profiling Tests
// =============================================================================

func BenchmarkCPUIntensiveNestedParsing(b *testing.B) {
	// Deeply nested structure - tests CPU usage on transformation
	formData := url.Values{}
	for i := 0; i < 10; i++ {
		for j := 0; j < 5; j++ {
			key := "data[" + string(rune('0'+i)) + "][items][" + string(rune('0'+j)) + "][name]"
			formData.Set(key, "value")
		}
	}
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

func BenchmarkCPUIntensiveTypeConversion(b *testing.B) {
	// Many fields requiring type conversion
	formData := url.Values{}
	for i := 0; i < 50; i++ {
		formData.Set("int_"+string(rune('a'+i%26)), "12345")
		formData.Set("float_"+string(rune('a'+i%26)), "123.456")
		formData.Set("bool_"+string(rune('a'+i%26)), "true")
	}
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

func BenchmarkCPUIntensiveLargeArrays(b *testing.B) {
	// Large array with many elements
	formData := url.Values{}
	for i := 0; i < 100; i++ {
		key := "items[" + string(rune('0'+i/10)) + string(rune('0'+i%10)) + "]"
		formData.Set(key, "item_value_"+string(rune('0'+i%10)))
	}
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
