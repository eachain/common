package dict

import (
	"container/list"
)

// EvictCallback is used to get a callback when a cache entry is evicted
type EvictCallback func(key interface{}, value interface{})

// NewLRUMap返回lru map,
// size必须大于0,
// onEvict可以为nil
func NewLRUMap(size int, onEvict EvictCallback) Map {
	if size <= 0 {
		panic("Must provide a positive size")
	}
	return &lruMap{
		m: make(map[interface{}]*list.Element),
		l: list.New(),
		n: size,
		c: onEvict,
	}
}

type lruMap struct {
	m map[interface{}]*list.Element
	l *list.List
	n int
	c EvictCallback
}

func (lm *lruMap) Load(key interface{}) (value interface{}, ok bool) {
	e, ok := lm.m[key]
	if !ok {
		return
	}
	lm.l.MoveToFront(e)
	return e.Value.(*_linkedPair).v, true
}

func (lm *lruMap) LoadOrStore(key interface{}, new func() interface{}) (actual interface{}, loaded bool) {
	e, loaded := lm.m[key]
	if loaded {
		lm.l.MoveToFront(e)
		return e.Value.(_linkedPair).v, true
	}

	actual = new()
	lm.m[key] = lm.l.PushFront(&_linkedPair{key, actual})
	if lm.l.Len() > lm.n {
		lm.removeOldest()
	}
	return
}

func (lm *lruMap) Store(key, value interface{}) {
	e, ok := lm.m[key]
	if ok {
		e.Value.(*_linkedPair).v = value
		lm.l.MoveToFront(e)
	} else {
		lm.m[key] = lm.l.PushFront(&_linkedPair{key, value})
		if lm.l.Len() > lm.n {
			lm.removeOldest()
		}
	}
}

func (lm *lruMap) Delete(key interface{}) {
	e, ok := lm.m[key]
	if ok {
		lm.removeElement(e)
	}
}

func (lm *lruMap) Range(f func(key, value interface{}) bool) {
	for e := lm.l.Front(); e != nil; e = e.Next() {
		pair := e.Value.(*_linkedPair)
		if !f(pair.k, pair.v) {
			return
		}
	}
}

func (lm *lruMap) removeOldest() {
	e := lm.l.Back()
	if e != nil {
		lm.removeElement(e)
	}
}

func (lm *lruMap) removeElement(e *list.Element) {
	lm.l.Remove(e)
	pair := e.Value.(*_linkedPair)
	delete(lm.m, pair.k)
	if lm.c != nil {
		lm.c(pair.k, pair.v)
	}
}

func (lm *lruMap) MarshalJSON() ([]byte, error) {
	return marshalJSON(lm)
}

func (lm *lruMap) UnmarshalJSON(p []byte) error {
	err := unmarshalJSON(p, lm)
	if err != nil {
		return err
	}

	// reverse
	e := lm.l.Front()
	if e == nil {
		return nil
	}
	var next *list.Element
	for e = e.Next(); e != nil; e = next {
		next = e.Next()
		lm.l.MoveToFront(e)
	}
	return nil
}
