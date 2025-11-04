# httperr

[![Go Reference](https://pkg.go.dev/badge/github.com/jorenkoyen/httperr.svg)](https://pkg.go.dev/github.com/jorenkoyen/httperr)

`httperr` is a minimal Go library that extends the standard [`net/http`](https://pkg.go.dev/net/http) package with *
*explicit error handling** for HTTP handlers.

It lets you write HTTP handlers that **return errors directly**, while providing a simple mechanism to control how those
errors are rendered to the client.

---

## ✨ Features

- Drop-in compatible with `net/http`
- Handlers can return `error` instead of writing directly
- Customizable error writers (plain text or JSON)
- Optional HTTP status embedding with `WithStatus`
- Lightweight wrapper types with no external dependencies

---

## 🚀 Installation

```bash
go get github.com/jorenkoyen/httperr
```

--- 

## 🧩 Basic Example

```go
package main

import (
	"net/http"

	"github.com/jorenkoyen/httperr"
)

func handler(w http.ResponseWriter, r *http.Request) error {
	name := r.URL.Query().Get("name")
	if name == "" {
		return httperr.New("missing 'name' parameter", http.StatusBadRequest)
	}
	_, _ = w.Write([]byte("Hello, " + name + "!"))
	return nil
}

func main() {
	http.HandleFunc("/", httperr.StdHandler(handler))
	http.ListenAndServe(":8080", nil)
}
```

If the request is missing a name query parameter, the client receives:

```
HTTP/1.1 400 Bad Request
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff

missing 'name' parameter
```

---

## 💡 Custom Error Writers

You can provide your own ErrorWriter to control how errors are written to the response.

### JSON Example

```go
package main

import (
	"net/http"

	"github.com/jorenkoyen/httperr"
)

func handler(w http.ResponseWriter, r *http.Request) error {
	return httperr.New("invalid parameters", http.StatusBadRequest)
}

func main() {
	http.HandleFunc("/hello", httperr.StdHandlerWithError(handler, httperr.JsonErrorWriter))
}
```

Produces:

```json
{
  "error": "missing 'name' parameter",
  "status": 400
}
```

---

## 🧱 Using ErrorServeMux

`ErrorServeMux` behaves like a regular `http.ServeMux` but supports `ErrorHandlerFunc`.

```go
package main

import (
	"net/http"

	"github.com/jorenkoyen/httperr"
)

func main() {
	mux := httperr.NewErrorServeMux(httperr.JsonErrorWriter)

	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) error {
		return httperr.New("something went wrong", http.StatusInternalServerError)
	})

	http.ListenAndServe(":8080", mux)
}

```