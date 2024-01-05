# Go Inrequest

#### Golang package for transform http request body into go map.

<hr>

**Installation**

You first need [Go](https://go.dev/) installed (version 1.17+ is required), then you can use the below Go command to install req:

``` sh
go get github.com/ezartsh/inrequest
```

**Import**

Import inrequest to your code:

```go
import "github.com/ezartsh/inrequest"
```

### Basic Usage

**Example Form Data Request**

```html
<form action="some_url">
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

func main() {
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		// POST Method Request
		req := inrequest.FormData(r).ToMap()
		fmt.Println(req)
		// output : map[first_name:Jhon last_name:Smith]
	})
}
```

<br />

##### Request Body Type

- [ Form Data. ](#form-data)
- [ Query String. ](#query-string)
- [ Json Request. ](#json-request)

<a name="form-request"></a>
## 1. Form Data

- **Example 1**

```html
<form action="some_url" method="POST">
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
	FirstName string `json:"first_name"'`
	LastName string `json:"last_name"'`
}

func main() {
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		// POST Method Request
		req := inrequest.FormData(r)
		fmt.Println(req.ToMap())
		// output : map[first_name:Jhon last_name:Smith]
		breq := BodyRequest{}
		req.ToBind(&bodyRequest)
		fmt.Printf("My Name is %s %s:\n", breq.FirstName, breq.LastName)
		// output : My Name is Jhon Smith
		
		if jsonString, err := req.ToJsonString(); err != nil {
			fmt.Println(jsonString) 
			// output : {"first_name":"Jhon","last_name":"Smith"} 
		}
		
	})
}
```

- **Example 2**
```html
<form action="some_url" method="POST" enctype="multipart/form-data">
  <input type="text" value="John" name="names[0]"/>
  <input type="text" value="Smith" name="names[1]"/>
  <input type="text" value="Doe" name="names[2]"/>
  <input type="text" value="Title Attachment 1" name="attachments[0][title]"/>
  <textarea value="Description Attachment 1" name="attachments[0][description]"/>
  <input type="file" value="(binary file)" name="attachments[0][file]"/>
  <input type="text" value="Title Attachment 2" name="attachments[1][title]"/>
  <textarea value="Description Attachment 2" name="attachments[1][description]"/>
  <input type="file" value="(binary file)" name="attachments[1][file]"/>
</form>
```

```go
package main

import (
	"fmt"
	"net/http"
	"github.com/ezartsh/inrequest"
)

func main() {
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		// POST Method Request
		req := inrequest.FormData(r)
		
		if jsonString, err := req.ToJsonString(); err != nil {
			// output
			... 
		}
	})
}
```
#### Example 2 Output
```json
// example output 
{
	"names": [
		"Jhon",
		"Smith"
		"Doe"
	],
	"attachments": [
		{
			"title": "Title Attachment 1"
			"description": "Description Attachment 1",
			"file": {
				"Filename": "dummy.pdf",
				"Header": {
					"Content-Disposition": [
						"form-data; name=\"attachments[0][file]\"; filename=\"dummy.pdf\""
					],
					"Content-Type": [
						"application/pdf"
					]
				},
				"Size": 13264
			},
		},
		{
			"title": "Title Attachment 2"
			"description": "Description Attachment 2",
			"file": {
				"Filename": "dummy2.pdf",
				"Header": {
					"Content-Disposition": [
						"form-data; name=\"attachments[0][file]\"; filename=\"dummy2.pdf\""
					],
					"Content-Type": [
						"application/pdf"
					]
				},
				"Size": 13264
			},
		}
	]
}
```


<a name="query-string"></a>
## 2. Query String

- **Example Request**

```http request
http://localhost:8080/test?utm_source=1&callback_url=http://localhost:3000&status=active
```

```go
package main

import (
	"fmt"
	"net/http"
	"github.com/ezartsh/inrequest"
)

func main() {
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		// GET Method Request
		req := inrequest.Query(r)
		fmt.Println(req.ToMap())
		// output : map[callback_url:http://localhost:3000 status:active utm_source:1]
		if jsonString, err := req.ToJsonString(); err != nil {
			fmt.Println(jsonString) 
			// output : {"callback_url":"http://localhost:3000","status":"active","utm_source":"1"} 
		}
		
	})
}
```

<a name="json-request"></a>
## 3. Json Request

- **Example Request**

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

func main() {
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		// POST Method Request
		req := inrequest.Json(r)
		fmt.Println(req.ToMap())
		// output : map[age:31 first_name:John last_name:Smith]
		if jsonString, err := req.ToString(); err != nil {
			fmt.Println(jsonString) 
			// output : {"age":31,"first_name":"John","last_name":"Smith"} 
		}
		
	})
}
```

## Contributing

If you have a bug report or feature inrequest, you can [open an issue](https://github.com/ezartsh/inrequest/issues/new), and [pull requests](https://github.com/ezartsh/inrequest/pulls) are also welcome.

`inrequest` released under MIT license, refer [LICENSE](LICENSE) file.