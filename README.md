# gosession
========

gosession provides simple API that allows you to store visitor's data between HTTP requests.

## Example
```go
package main

import (
	"fmt"
	"net/http"
	"github.com/kamiokk/gosession"
	memsess "github.com/kamiokk/gosession/mem"
)

func login(w http.ResponseWriter,r *http.Request) {
	model := memsess.Model{}
	session := gosession.Start(r,w,model)
	session.Set("user","kamiokk")
	fmt.Fprint(w,"you are logined.")
}

func user(w http.ResponseWriter,r *http.Request) {
	model := memsess.Model{}
	session := gosession.Start(r,w,model)
	if val,err := session.Get("user");err == nil {
		fmt.Fprintf(w,"you are %v", val)
	} else {
		fmt.Fprint(w,"not logined")
	}
}

func main() {
	http.HandleFunc("/login",login)
	http.HandleFunc("/user",user)
	http.ListenAndServe("localhost:8099",nil)
}
```

## features
- [x] memory store model
- [ ] file store model