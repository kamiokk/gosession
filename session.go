package gosession

import (
	"reflect"
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
	// New make a new session,return an error if the sessionID exist 
	New(ssid string,expire int64) (error)
	// Read session data by sessionID,return an error if the session do not exists
	Read(key string) (interface{},error)
	// Write data into session,return an error if the session do not exists
	Write(key string,data interface{}) (error)
	// Unset specified key,return an error if the session do not exists
	Unset(key string) (error)
	// Refresh checks if the sessionID exists, it will refresh the expire time and return true while the sessionID exists
	Refresh(ssid string,expire int64) (string,bool)
}

type Session struct {
	ID string
	Request *http.Request
	Option *SessionOption
	Model ISessionModel
}

func defaultOption() *SessionOption {
	option := &SessionOption{
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
	return fmt.Sprintf("%x%d", h.Sum(nil),time.Now().Unix())
}

// Start session
func Start(r *http.Request,w http.ResponseWriter,model ISessionModel,options ...*SessionOption) (*Session,error) {
	var s *Session
	var option *SessionOption
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
		if err := model.New(sessionID,option.MaxAge); err != nil {
			return nil,err
		}
	}
	s = &Session {
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
	return s,nil
}

// Set to session
func (s *Session) Set(key string,value interface{}) (error) {
	return s.Model.Write(key,value)
}

// Unset a specified key
func (s *Session) Unset(key string) (error) {
	return s.Model.Unset(key)
}

// Get a value from session
func (s *Session) Get(key string) (interface{},error) {
	return s.Model.Read(key)
}

func typeError(wanted string,actual interface{}) error {
	return fmt.Errorf("type missmatch,wanted: %s ,actual: %s",wanted,reflect.TypeOf(actual).String())
}

// Get a string value from session
func (s *Session) GetString(key string) (string,error) {
	if val,err := s.Model.Read(key);err != nil {
		return "",err
	} else {
		if s,ok := val.(string);ok {
			return s,nil
		} else {
			return "",typeError("string",val)
		}
	}
}

// Get an int value from session
func (s *Session) GetInt(key string) (int,error) {
	if val,err := s.Model.Read(key);err != nil {
		return 0,err
	} else {
		if i,ok := val.(int);ok {
			return i,nil
		} else {
			return 0,typeError("int",val)
		}
	}
}

// Get an uint value from session
func (s *Session) GetUInt(key string) (uint,error) {
	if val,err := s.Model.Read(key);err != nil {
		return 0,err
	} else {
		if i,ok := val.(uint);ok {
			return i,nil
		} else {
			return 0,typeError("uint",val)
		}
	}
}

// Get a float32 value from session
func (s *Session) GetFloat32(key string) (float32,error) {
	if val,err := s.Model.Read(key);err != nil {
		return 0.0,err
	} else {
		if i,ok := val.(float32);ok {
			return i,nil
		} else {
			return 0.0,typeError("float32",val)
		}
	}
}

// Get a float64 value from session
func (s *Session) GetFloat64(key string) (float64,error) {
	if val,err := s.Model.Read(key);err != nil {
		return 0.0,err
	} else {
		if i,ok := val.(float64);ok {
			return i,nil
		} else {
			return 0.0,typeError("float64",val)
		}
	}
}