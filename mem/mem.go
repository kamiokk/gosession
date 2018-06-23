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

type Model struct {}

func (m Model) Read(ssid,key string) (interface{},error) {
	index := indexFor(ssid)
	data[index].mu.RLock()
	s,ok := data[index].table[ssid]
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

func (m Model) Write(ssid,key string,value interface{},expire int64) (error) {
	index := indexFor(ssid)
	expire = time.Now().Unix() + expire + EXPIRE_BUFFER_TIME
	data[index].mu.Lock()
	if _,ok := data[index].table[ssid];ok {
		data[index].table[ssid].values[key] = value
		data[index].table[ssid].expireAt = expire
	} else {
		data[index].table[ssid] = &session{expireAt:expire,values:make(map[string]interface{})}
		data[index].table[ssid].values[key] = value
	}
	data[index].mu.Unlock()
	return nil
}

func (m Model) Refresh(ssid string,expire int64) (string,bool) {
	ok := false
	index := indexFor(ssid)
	data[index].mu.Lock()
	if _,ok = data[index].table[ssid];ok {
		data[index].table[ssid].expireAt = time.Now().Unix() + expire + EXPIRE_BUFFER_TIME
	}
	data[index].mu.Unlock()
	if ok {
		return ssid,true
	} else {
		return "",false
	}
}