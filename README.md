# gosession
gosession provides simple API that allows you to store visitor's data between HTTP requests.

## Quick start.
```go
package main

import (
	"fmt"
	"net/http"
	"github.com/kamiokk/gosession"
	"github.com/kamiokk/gosession/mem"
)

func login(w http.ResponseWriter,r *http.Request) {
	model := &mem.Model{}
	session,_ := gosession.Start(r,w,model)
	session.Set("user","Dark Flame Master")
	fmt.Fprint(w,"you are logined.")
}

func user(w http.ResponseWriter,r *http.Request) {
	model := &mem.Model{}
	session,_ := gosession.Start(r,w,model)
	if val,err := session.GetString("user");err == nil {
		fmt.Fprintf(w,"you are %s,your sessionID is %s", val,session.ID)
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

## API example
### Choose a store model and start a session with your own option. Although only memory model is available now. \_(:3 」∠)\_
```go
package main

import (
	"fmt"
	"net/http"
	"github.com/kamiokk/gosession"
	"github.com/kamiokk/gosession/mem"
)

func main() {
	http.HandleFunc("/",func(w http.ResponseWriter,r *http.Request) {
		model := &mem.Model{}
		option := &gosession.SessionOption {
			SessionName: "MYSESSIONID",
			Path: "/",
			Domain: "",
			MaxAge: 1200,
			HttpOnly: true,
			Secure: false,
		}
		session,err := gosession.Start(r,w,model,option)
		if err == nil {
			fmt.Fprintf(w,"your sessionID is %s", session.ID)
		} else {
			panic("failed to start a session.")
		}
	})
	http.ListenAndServe("localhost:8099",nil)
}
```
### Set and Get
```go
package main

import (
	"fmt"
	"net/http"
	"github.com/kamiokk/gosession"
	"github.com/kamiokk/gosession/mem"
)

func main() {
	http.HandleFunc("/set",func(w http.ResponseWriter,r *http.Request) {
		model := &mem.Model{}
		if session,err := gosession.Start(r,w,model);err == nil {
			session.Set("intVal",12345)
			session.Set("stringVal","some string")
			session.Set("whateverVal",map[string]string{"key":"val"})
			fmt.Fprintf(w,"All set.")
		}
	})
	http.HandleFunc("/get",func(w http.ResponseWriter,r *http.Request) {
		model := &mem.Model{}
		if session,err := gosession.Start(r,w,model);err == nil {
			i,_ := session.GetInt("intVal")
			s,_ := session.GetString("stringVal")
			m,_ := session.Get("whateverVal")
			fmt.Fprintf(w,"intval:%d stringVal:%s whateverVal:%v", i,s,m)
		}
	})
	http.ListenAndServe("localhost:8099",nil)
}
```
## That is it.