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

```http request
+ first_name: "John"
+ last_name: "Smith"
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

- **Example Request**

```http request
+ first_name: "John"
+ last_name: "Smith"
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

- **Advanced Example Request**

```http request
+ names: John
+ names: Mike
+ numbers[0]: 1
+ numbers[1]: 2
+ numbers[3]: 3
+ attachments[0][title]: Title Attachment 1
+ attachments[0][description]: Description Attachment 1
+ attachments[0][file]: (binary)
+ attachments[1][title]: Title Attachment 2
+ attachments[1][description]: Description Attachment 2
+ attachments[1][file]: (binary)
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
			fmt.Println(jsonString) 
			// output : 
			/* {
			    "names": [
			        "Jhon",
			        "Mike"
			    ],
			    "numbers": [
			        "1",
			        "2",
			        "3"
			    ],
			    "attachments": [
			        {
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
			            "title": "Title Attachment 1"
			        },
			        {
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
			            "title": "Title Attachment 2"
			        }
			    ]
			} */
		}
	})
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