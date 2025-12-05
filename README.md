# Go Inrequest

[![Go Version](https://img.shields.io/badge/Go-1.17+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/ezartsh/inrequest)](https://goreportcard.com/report/github.com/ezartsh/inrequest)

**Golang package for transforming HTTP request body into Go maps and structs.**

Supports multiple content types:
- `multipart/form-data` (with file uploads)
- `application/x-www-form-urlencoded`
- `application/json`
- Query string parameters

---

## Installation

Requires [Go](https://go.dev/) 1.17+

```sh
go get github.com/ezartsh/inrequest
```

## Import

```go
import "github.com/ezartsh/inrequest"
```

---

## Quick Start

```go
package main

import (
    "fmt"
    "net/http"
    "github.com/ezartsh/inrequest"
)

func main() {
    http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
        // Parse form data
        req := inrequest.FormData(r)
        defer req.Cleanup() // Optional: clean up temp files
        
        fmt.Println(req.ToMap())
        // output: map[first_name:John last_name:Smith]
    })
    
    http.ListenAndServe(":8080", nil)
}
```

---

## API Reference

### Entry Functions

| Function | Description |
|----------|-------------|
| `FormData(r *http.Request)` | Parse multipart/form-data or url-encoded form |
| `FormDataWithOptions(r *http.Request, opts Options)` | Parse form with custom options |
| `Query(r *http.Request)` | Parse query string parameters |
| `Json(r *http.Request)` | Parse JSON request body |
| `Parse(r *http.Request)` | Auto-detect and parse based on Content-Type |

### Options

```go
type Options struct {
    MaxMemory int64 // Max memory for multipart parsing (default: 32MB)
}
```

### Request Interface

All request types implement the `Request` interface:

```go
type Request interface {
    ToMap() RequestValue                    // Get as map[string]interface{}
    ToBind(model interface{}) error         // Bind to struct
    ToJsonByte() ([]byte, error)            // Get as JSON bytes
    ToJsonString() (string, error)          // Get as JSON string
}
```

### FormRequest Additional Methods

```go
// Cleanup removes temporary files created during multipart form parsing.
// Optional but recommended for high-traffic servers.
func (r FormRequest) Cleanup()
```

### Types

```go
// RequestValue is the parsed request data
type RequestValue = map[string]interface{}

// FileHeaders represents multiple uploaded files for a single field
type FileHeaders []*multipart.FileHeader
```

---

## Request Body Types

- [Form Data](#1-form-data)
- [Query String](#2-query-string)
- [JSON Request](#3-json-request)
- [Auto Parse](#4-auto-parse)

---

## 1. Form Data

### Basic Example

```html
<form action="/submit" method="POST">
    <input type="text" value="John" name="first_name"/>
    <input type="text" value="Smith" name="last_name"/>
</form>
```

```go
package main

import (
    "fmt"
    "net/http"
    "github.com/ezartsh/inrequest"
)

type BodyRequest struct {
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
}

func main() {
    http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
        req := inrequest.FormData(r)
        defer req.Cleanup()
        
        // Get as map
        fmt.Println(req.ToMap())
        // output: map[first_name:John last_name:Smith]
        
        // Bind to struct
        var body BodyRequest
        req.ToBind(&body)
        fmt.Printf("My name is %s %s\n", body.FirstName, body.LastName)
        // output: My name is John Smith
        
        // Get as JSON string
        if jsonStr, err := req.ToJsonString(); err == nil {
            fmt.Println(jsonStr)
            // output: {"first_name":"John","last_name":"Smith"}
        }
    })
    
    http.ListenAndServe(":8080", nil)
}
```

### Nested Arrays and Objects

```html
<form action="/submit" method="POST" enctype="multipart/form-data">
    <input type="text" value="John" name="names[0]"/>
    <input type="text" value="Smith" name="names[1]"/>
    <input type="text" value="Doe" name="names[2]"/>
    
    <input type="text" value="Title 1" name="attachments[0][title]"/>
    <textarea name="attachments[0][description]">Description 1</textarea>
    <input type="file" name="attachments[0][file]"/>
    
    <input type="text" value="Title 2" name="attachments[1][title]"/>
    <textarea name="attachments[1][description]">Description 2</textarea>
    <input type="file" name="attachments[1][file]"/>
</form>
```

```go
req := inrequest.FormData(r)
defer req.Cleanup()

jsonStr, _ := req.ToJsonString()
fmt.Println(jsonStr)
```

**Output:**
```json
{
    "names": ["John", "Smith", "Doe"],
    "attachments": [
        {
            "title": "Title 1",
            "description": "Description 1",
            "file": {
                "Filename": "document.pdf",
                "Header": {
                    "Content-Type": ["application/pdf"]
                },
                "Size": 13264
            }
        },
        {
            "title": "Title 2",
            "description": "Description 2",
            "file": {
                "Filename": "image.png",
                "Header": {
                    "Content-Type": ["image/png"]
                },
                "Size": 8192
            }
        }
    ]
}
```

### Multiple File Uploads

For multiple files with the same field name, use `FileHeaders`:

```html
<input type="file" name="documents[]" multiple/>
```

```go
type UploadRequest struct {
    Documents inrequest.FileHeaders `json:"documents"`
}

req := inrequest.FormData(r)
defer req.Cleanup()

var upload UploadRequest
req.ToBind(&upload)

for _, file := range upload.Documents {
    fmt.Printf("File: %s, Size: %d\n", file.Filename, file.Size)
}
```

### Custom Memory Limit

```go
req := inrequest.FormDataWithOptions(r, inrequest.Options{
    MaxMemory: 10 << 20, // 10 MB
})
defer req.Cleanup()
```

---

## 2. Query String

```
GET /search?utm_source=google&callback_url=http://localhost:3000&status=active
```

```go
package main

import (
    "fmt"
    "net/http"
    "github.com/ezartsh/inrequest"
)

func main() {
    http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
        req := inrequest.Query(r)
        
        fmt.Println(req.ToMap())
        // output: map[callback_url:http://localhost:3000 status:active utm_source:google]
        
        if jsonStr, err := req.ToJsonString(); err == nil {
            fmt.Println(jsonStr)
            // output: {"callback_url":"http://localhost:3000","status":"active","utm_source":"google"}
        }
    })
    
    http.ListenAndServe(":8080", nil)
}
```

### Nested Query Parameters

```
GET /filter?filters[status]=active&filters[category]=tech&items[0]=a&items[1]=b
```

```go
req := inrequest.Query(r)
fmt.Println(req.ToMap())
// output: map[filters:map[status:active category:tech] items:[a b]]
```

---

## 3. JSON Request

```json
{
    "first_name": "John",
    "last_name": "Smith",
    "age": 31
}
```

```go
package main

import (
    "fmt"
    "net/http"
    "github.com/ezartsh/inrequest"
)

type Person struct {
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    Age       int    `json:"age"`
}

func main() {
    http.HandleFunc("/api/person", func(w http.ResponseWriter, r *http.Request) {
        req := inrequest.Json(r)
        
        fmt.Println(req.ToMap())
        // output: map[age:31 first_name:John last_name:Smith]
        
        var person Person
        req.ToBind(&person)
        fmt.Printf("%s is %d years old\n", person.FirstName, person.Age)
        // output: John is 31 years old
        
        if jsonStr, err := req.ToJsonString(); err == nil {
            fmt.Println(jsonStr)
            // output: {"age":31,"first_name":"John","last_name":"Smith"}
        }
    })
    
    http.ListenAndServe(":8080", nil)
}
```

---

## 4. Auto Parse

Use `Parse()` to automatically detect and parse based on `Content-Type`:

```go
http.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
    req := inrequest.Parse(r)
    
    // Works with any content type:
    // - application/json
    // - multipart/form-data
    // - application/x-www-form-urlencoded
    // - Falls back to query string
    
    fmt.Println(req.ToMap())
})
```

---

## Type Conversion

The library automatically converts string values to appropriate Go types:

| Input | Output Type |
|-------|-------------|
| `"123"` | `int` |
| `"3.14"` | `float64` |
| `"true"`, `"false"` | `bool` |
| `"null"` | `nil` |
| Other strings | `string` |

---

## Binding to Structs

Use `ToBind()` to bind request data to a struct. The struct fields should have `json` tags:

```go
type CreateUserRequest struct {
    Name     string                  `json:"name"`
    Email    string                  `json:"email"`
    Age      int                     `json:"age"`
    Active   bool                    `json:"active"`
    Avatar   *multipart.FileHeader   `json:"avatar"`      // Single file
    Photos   inrequest.FileHeaders   `json:"photos"`      // Multiple files
}

req := inrequest.FormData(r)
defer req.Cleanup()

var user CreateUserRequest
if err := req.ToBind(&user); err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
}

// Access file
if user.Avatar != nil {
    fmt.Printf("Avatar: %s\n", user.Avatar.Filename)
}

// Access multiple files
for _, photo := range user.Photos {
    fmt.Printf("Photo: %s\n", photo.Filename)
}
```

---

## Performance

The library is optimized for performance:

| Operation | Time |
|-----------|------|
| Simple form (5 fields) | ~5 µs |
| Complex nested form | ~50 µs |
| Large form (100 fields) | ~200 µs |
| File upload | ~500 µs |

Memory allocation is minimal, and temporary files are cleaned up with `Cleanup()`.

---

## Contributing

If you have a bug report or feature request, you can [open an issue](https://github.com/ezartsh/inrequest/issues/new), and [pull requests](https://github.com/ezartsh/inrequest/pulls) are also welcome.

## License

`inrequest` is released under the MIT license. See [LICENSE](LICENSE) for details.
