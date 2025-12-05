// Package inrequest provides utilities for transforming HTTP request bodies into Go maps and structs.
//
// It supports multiple content types including multipart/form-data, application/x-www-form-urlencoded,
// application/json, and query string parameters.
//
// # Basic Usage
//
// Parse form data from an HTTP request:
//
//	req := inrequest.FormData(r)
//	defer req.Cleanup() // Optional: clean up temp files
//
//	// Get as map
//	data := req.ToMap()
//
//	// Bind to struct
//	var user User
//	req.ToBind(&user)
//
//	// Get as JSON
//	jsonStr, _ := req.ToJsonString()
//
// # Request Types
//
// The package provides four main entry functions:
//
//   - [FormData]: Parse multipart/form-data or application/x-www-form-urlencoded
//   - [FormDataWithOptions]: Parse form data with custom options (e.g., max memory)
//   - [Query]: Parse query string parameters
//   - [Json]: Parse JSON request body
//   - [Parse]: Auto-detect content type and parse accordingly
//
// # Nested Data
//
// The package supports bracket notation for nested structures:
//
//	// Input: items[0][name]=foo&items[0][price]=10&items[1][name]=bar
//	// Output: map[items:[]map[name:foo price:10] map[name:bar]]
//
// # File Uploads
//
// For file uploads, use [FormData] with multipart/form-data:
//
//	type Upload struct {
//	    Title    string                  `json:"title"`
//	    Document *multipart.FileHeader   `json:"document"`  // Single file
//	    Photos   inrequest.FileHeaders   `json:"photos"`    // Multiple files
//	}
//
//	req := inrequest.FormData(r)
//	defer req.Cleanup()
//
//	var upload Upload
//	req.ToBind(&upload)
//
// # Type Conversion
//
// String values are automatically converted to appropriate Go types:
//
//   - "123" → int
//   - "3.14" → float64
//   - "true", "false" → bool
//   - "null" → nil
//   - Leading zeros preserved: "007" → "007" (string)
//
// # Error Handling
//
// The package provides custom error types for better error handling:
//
//	req := inrequest.FormData(r)
//	var user User
//	if err := req.ToBind(&user); err != nil {
//	    if inrequest.IsBindError(err) {
//	        // Handle binding error
//	    }
//	}
//
// # Performance
//
// The library is optimized for performance with minimal allocations:
//
//   - Simple form (5 fields): ~5 µs
//   - Complex nested form: ~50 µs
//   - Large form (100 fields): ~200 µs
//
// For high-traffic servers, call [FormRequest.Cleanup] to immediately release
// temporary files created during multipart parsing.
package inrequest
