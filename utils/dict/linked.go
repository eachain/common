package dict

import (
	"container/list"
)

// NewLinkedMap返回有序的 map,
// 按key的store顺序存储,
// 在json.Marshal和json.Unmarshal时会保持顺序
func NewLinkedMap() Map {
	return &linkedMap{
		m: make(map[interface{}]*list.Element),
		l: list.New(),
	}
}

type _linkedPair struct {
	k, v interface{}
}

type linkedMap struct {
	m map[interface{}]*list.Element
	l *list.List
}

func (lm *linkedMap) Load(key interface{}) (value interface{}, ok bool) {
	e, ok := lm.m[key]
	if !ok {
		return
	}
	return e.Value.(*_linkedPair).v, true
}

func (lm *linkedMap) LoadOrStore(key interface{}, new func() interface{}) (actual interface{}, loaded bool) {
	e, loaded := lm.m[key]
	if loaded {
		return e.Value.(_linkedPair).v, true
	}
	actual = new()
	lm.m[key] = lm.l.PushBack(&_linkedPair{key, actual})
	return
}

func (lm *linkedMap) Store(key, value interface{}) {
	e, ok := lm.m[key]
	if ok {
		e.Value.(*_linkedPair).v = value
	} else {
		lm.m[key] = lm.l.PushBack(&_linkedPair{key, value})
	}
}

func (lm *linkedMap) Delete(key interface{}) {
	e, ok := lm.m[key]
	if ok {
		delete(lm.m, key)
		lm.l.Remove(e)
	}
}

func (lm *linkedMap) Range(f func(key, value interface{}) bool) {
	for e := lm.l.Front(); e != nil; e = e.Next() {
		pair := e.Value.(*_linkedPair)
		if !f(pair.k, pair.v) {
			return
		}
	}
}

func (lm *linkedMap) MarshalJSON() ([]byte, error) {
	return marshalJSON(lm)
}

func (lm *linkedMap) UnmarshalJSON(p []byte) error {
	return unmarshalJSON(p, lm)
}
