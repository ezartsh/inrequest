# Go Inrequest

#### Golang package for transform form request body into go map.

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

**Basic Usage**

## <a name="Simple-Form-Data">Simple Form Data</a>

```go
package main

import (
	"net/http"
	"log"
	"github.com/ezartsh/inrequest"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
            request := inrequest.GetFormRequest(r)
	})
	http.ListenAndServe(":3333", nil)
}
```

## Contributing

If you have a bug report or feature inrequest, you can [open an issue](https://github.com/ezartsh/inrequest/issues/new), and [pull requests](https://github.com/ezartsh/inrequest/pulls) are also welcome.

`inrequest` released under MIT license, refer [LICENSE](LICENSE) file.