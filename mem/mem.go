// Store model for gosession.
// This model stores data into memory.So this model IS NOT PERSISTENT.
package mem

import (
	"sync"
	"time"
	"errors"
)

// split the table to reduce concurrent racing.
// you can modify it for better performance.
const PARTITION int32 = 128
// prevents server-side session expires before client-side.
const EXPIRE_BUFFER_TIME int64 = 30
const RECYCLE_INTERVAL time.Duration = time.Minute

type session struct {
	expireAt int64
	values map[string]interface{}
}

type sessionTable struct {
	mu sync.RWMutex
	table map[string]*session
}

var data [PARTITION]sessionTable

func init() {
	for i := range data {
		data[i] = sessionTable{table: make(map[string]*session)}
	}
	go func() {
		for {
			time.Sleep(RECYCLE_INTERVAL)
			recycle()
		}
	}()
}

func hashCode(key string) int32 {
	var h int32 = 0
	for _,c := range key {
		h = 31 * h + c
	}
	return h
}

func indexFor(key string) int32 {
	h := hashCode(key)
	return h & (PARTITION - 1)
}

func recycle() {
	timestamp := time.Now().Unix()
	for i := range data {
		data[i].mu.Lock()
		for k,v := range data[i].table {
			if v.expireAt < timestamp  {
				delete(data[i].table,k)
			}
		}
		data[i].mu.Unlock()
	}
}

type Model struct {
	sessionID string
	expire int64
}

// New make a new session,return an error if the sessionID exist 
func (m *Model) New(ssid string,expire int64) error {
	var err error
	index := indexFor(ssid)
	expireAt := time.Now().Unix() + expire + EXPIRE_BUFFER_TIME
	data[index].mu.Lock()
	_,ok := data[index].table[ssid]
	if ok {
		err = errors.New("sessionID:" + ssid + " already exists")
	} else {
		data[index].table[ssid] = &session{expireAt:expireAt,values:make(map[string]interface{})}
	}
	data[index].mu.Unlock()
	if err == nil {
		m.sessionID = ssid
		m.expire = expire
	}
	return err
}

func (m *Model) Read(key string) (interface{},error) {
	if m.sessionID == "" {
		return nil,errors.New("not init yet")
	}
	index := indexFor(m.sessionID)
	data[index].mu.RLock()
	s,ok := data[index].table[m.sessionID]
	data[index].mu.RUnlock()
	var ret interface{}
	var err error
	if ok && s.expireAt >= time.Now().Unix() {
		if ret,ok = s.values[key];!ok {
			err = errors.New("session key not found")
		}
	} else {
		err = errors.New("session not exists or expired")
	}
	return ret,err
}

func (m *Model) Write(key string,value interface{}) (error) {
	if m.sessionID == "" {
		return errors.New("not init yet")
	}
	index := indexFor(m.sessionID)
	var err error
	data[index].mu.Lock()
	if v,ok := data[index].table[m.sessionID];ok && v.expireAt >= time.Now().Unix() {
		data[index].table[m.sessionID].values[key] = value
	} else {
		err = errors.New("session not exists or expired")
	}
	data[index].mu.Unlock()
	return err
}

func (m *Model) Unset(key string) (error) {
	if m.sessionID == "" {
		return errors.New("not init yet")
	}
	index := indexFor(m.sessionID)
	var err error
	data[index].mu.Lock()
	if v,ok := data[index].table[m.sessionID];ok && v.expireAt >= time.Now().Unix() {
		if _,ok = data[index].table[m.sessionID].values[key]; ok {
			delete(data[index].table[m.sessionID].values,key)
		}
	} else {
		err = errors.New("session not exists or expired")
	}
	data[index].mu.Unlock()
	return err
}

// Refresh checks if the sessionID exists, it will refresh the expire time and return true while the sessionID exists
func (m *Model) Refresh(ssid string,expire int64) (string,bool) {
	ok := false
	index := indexFor(ssid)
	data[index].mu.Lock()
	if _,ok = data[index].table[ssid];ok {
		data[index].table[ssid].expireAt = time.Now().Unix() + expire + EXPIRE_BUFFER_TIME
	}
	data[index].mu.Unlock()
	if ok {
		m.sessionID = ssid
		m.expire = expire
		return ssid,true
	} else {
		return "",false
	}
}