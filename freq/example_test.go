package freq_test

import (
	"fmt"
	"sync"
	"time"

	"github.com/eachain/common/freq"
)

type inmemList struct {
	u uint64
	t time.Time
	m sync.Mutex
}

func (l *inmemList) clean(now time.Time) {
	if l.t.Before(now) {
		l.u = 0
	}
}

func (l *inmemList) Len() uint64 {
	now := time.Now()
	l.m.Lock()
	l.clean(now)
	length := l.u
	l.m.Unlock()
	return length
}

func (l *inmemList) PushEX(expire time.Duration) uint64 {
	now := time.Now()
	l.m.Lock()
	l.clean(now)
	l.t = now.Add(expire)
	l.u++
	length := l.u
	l.m.Unlock()
	return length
}

func (l *inmemList) PushXX() uint64 {
	now := time.Now()
	var length uint64
	l.m.Lock()
	l.clean(now)
	if l.u > 0 {
		l.u++
		length = l.u
	}
	l.m.Unlock()
	return length
}

func (l *inmemList) Expire(expire time.Duration) {
	now := time.Now()
	l.m.Lock()
	l.clean(now)
	l.t = now.Add(expire)
	l.m.Unlock()
}

type expiration struct {
	t time.Time
	m sync.Mutex
}

func (e *expiration) SetEX(expire time.Duration) bool {
	ok := false
	now := time.Now()
	e.m.Lock()
	if e.t.Before(now) {
		e.t = now.Add(expire)
		ok = true
	}
	e.m.Unlock()
	return ok
}

type inmemCache struct {
	*inmemList
	*expiration
}

func NewCache() freq.Cache {
	return &inmemCache{&inmemList{}, &expiration{}}
}

func NewMultiCache() func(key string) freq.Cache {
	var mut sync.Mutex
	var ic = make(map[string]freq.Cache)
	return func(key string) freq.Cache {
		mut.Lock()
		c := ic[key]
		if c == nil {
			c = &inmemCache{&inmemList{}, &expiration{}}
			ic[key] = c
		}
		mut.Unlock()
		return c
	}
}

func ExamplePass() {
	opt := &freq.Option{
		Dur:    1 * time.Second,
		Num:    1,
		Punish: 10 * time.Second,
		Tick:   time.Second,
	}
	freq := freq.NewMultiFreq(opt, NewMultiCache())

	for i := 0; i < 3; i++ {
		fmt.Println(freq.Pass("test"))
	}

	// Output:
	// true
	// true
	// false
}
