package dict

import (
	"sync"
)

// NewMutexMap返回一个带锁的map
func NewMutexMap(m Map) Map {
	return &mutexMap{m: m}
}

type mutexMap struct {
	m Map
	l sync.RWMutex
}

func (mm *mutexMap) Load(key interface{}) (value interface{}, ok bool) {
	mm.l.RLock()
	value, ok = mm.m.Load(key)
	mm.l.RUnlock()
	return
}

func (mm *mutexMap) LoadOrStore(key interface{}, new func() interface{}) (actual interface{}, loaded bool) {
	mm.l.Lock()
	actual, loaded = mm.m.LoadOrStore(key, new)
	mm.l.Unlock()
	return
}

func (mm *mutexMap) Store(key, value interface{}) {
	mm.l.Lock()
	mm.m.Store(key, value)
	mm.l.Unlock()
}

func (mm *mutexMap) Delete(key interface{}) {
	mm.l.Lock()
	mm.m.Delete(key)
	mm.l.Unlock()
}

func (mm *mutexMap) Range(f func(key, value interface{}) bool) {
	mm.l.RLock()
	mm.m.Range(f)
	mm.l.RUnlock()
}

func (mm *mutexMap) MarshalJSON() ([]byte, error) {
	return marshalJSON(mm)
}

func (mm *mutexMap) UnmarshalJSON(p []byte) error {
	return unmarshalJSON(p, mm)
}
