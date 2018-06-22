package gosession

import (
	"net/http"
)

// SessionOption defines session option like session key
type SessionOption struct {
	SessionKey string // cookie key for session id
	Path string // see http.Cookie.Path
	Domain string // see http.Cookie.Domain
	MaxAge int // see http.Cookie.MaxAge
	HttpOnly bool // see http.Cookie.HttpOnly
	Secure bool // see http.Cookie.MaxAgSecuree
}

// ISessionModel defines methods of session store model 
type ISessionModel interface {
	//Read session data by ssid
	Read(ssid string) (map[string]interface{},error)
	//Write session data
	Write(ssid string,data map[string]interface{}) (error)
}

type session struct {
	ID string
	Request *http.Request
	Option SessionOption
	Model ISessionModel
}

var s *session

func defaultOption() SessionOption {
	option := SessionOption{
		SessionKey: "GOSESSID",
		Path: "/",
		Domain: "",
		MaxAge: 3600,
		HttpOnly: true,
		Secure: false,
	}
	return option
}

func defaultModel() ISessionModel {
	return ISessionModel{}
}

func createSessionID() string {
	return "12345"
}

// Start session
func Start(r *http.Request,w http.ResponseWriter,params ...interface{}) {
	var option SessionOption
	var model ISessionModel
	if len(params) >= 1 {
		if option,ok := params[0].(SessionOption); !ok {
			option = defaultOption()
		}
	} else {
		option = defaultOption()
	}
	if len(params) >= 2 {
		if model,ok := params[0].(ISessionModel); !ok {
			model = defaultModel()
		}
	} else {
		model = defaultModel()
	}
	var sessionID string
	if sidCookie,err := r.Cookie(s.ID); err != nil {
		sessionID = sidCookie.Value
	} else {
		sessionID = createSessionID()
	}
	s = &session{
		ID: sessionID,
		Option: option,
		Model: model,
		Request: r,
	}
	http.Setcookie(w,&http.Cookie{
		Name:     s.Option.SessionKey,
		Value:    s.ID,
		MaxAge:   s.Option.MaxAge,
		Path:     s.Option.Path,
		Domain:   s.Option.Domain,
		Secure:   s.Option.Secure,
		HttpOnly: s.Option.HttpOnly,
	})
}

// Set to session
func Set(key string,value interface{})  {
	
}

// Get from session
func Get(key string) interface{} {
	return interface{}
}