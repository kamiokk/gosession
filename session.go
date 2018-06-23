package gosession

import (
	"io"
	"fmt"
	"time"
	"net/http"
	"crypto/md5"
	"math/rand"
)

// SessionOption defines session option like session key
type SessionOption struct {
	SessionName string // cookie name for session id
	Path string // see http.Cookie.Path
	Domain string // see http.Cookie.Domain
	MaxAge int64 // How long does a session last. Note: DO NOT SET IT TO A NEGATIVE NUMBER OR ZERO
	HttpOnly bool // see http.Cookie.HttpOnly
	Secure bool // see http.Cookie.MaxAgSecuree
}

// ISessionModel defines methods of session store model 
type ISessionModel interface {
	// Read session data by sessionID,return error if the session do not exists
	Read(ssid,key string) (interface{},error)
	// Write data into session,this method will create a session if not exist
	Write(ssid,key string,data interface{},expire int64) (error)
	// Refresh the session expire time and return the sessionID if the session exists
	Refresh(ssid string,expire int64) (string,bool)
}

type Session struct {
	ID string
	Request *http.Request
	Option SessionOption
	Model ISessionModel
}

func defaultOption() SessionOption {
	option := SessionOption{
		SessionName: "GOSESSID",
		Path: "/",
		Domain: "",
		MaxAge: 3600,
		HttpOnly: true,
		Secure: false,
	}
	return option
}

func createSessionID(r *http.Request) string {
	h := md5.New()
	s := fmt.Sprintf("%s_%d_%d",r.RemoteAddr,time.Now().UnixNano(),rand.Int63())
	io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Start session
func Start(r *http.Request,w http.ResponseWriter,model ISessionModel,options ...SessionOption) *Session {
	var s *Session
	var option SessionOption
	if options != nil {
		option = options[0]
	} else {
		option = defaultOption()
	}
	var sessionID string
	if sidCookie,err := r.Cookie(option.SessionName); err == nil {
		sessionID,_ = model.Refresh(sidCookie.Value,option.MaxAge)
	}
	if sessionID == "" {
		sessionID = createSessionID(r)
		model.Write(sessionID,"_sessid",sessionID,option.MaxAge)
	}
	s = &Session{
		ID: sessionID,
		Option: option,
		Model: model,
		Request: r,
	}
	http.SetCookie(w,&http.Cookie{
		Name:     s.Option.SessionName,
		Value:    s.ID,
		MaxAge:   int(s.Option.MaxAge),
		Path:     s.Option.Path,
		Domain:   s.Option.Domain,
		Secure:   s.Option.Secure,
		HttpOnly: s.Option.HttpOnly,
		Expires: time.Now().Add(time.Duration(s.Option.MaxAge)*time.Second),
	})
	return s
}

// Set to session
func (s *Session) Set(key string,value interface{}) (error) {
	return s.Model.Write(s.ID,key,value,s.Option.MaxAge)
}

// Get from session
func (s *Session) Get(key string) (interface{},error) {
	return s.Model.Read(s.ID,key)
}