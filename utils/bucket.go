package utils

import (
	"hash/crc32"
	"sync"
)

type Rand interface {
	Intn(int) int
}

type bucket struct {
	m map[string]interface{}
	l sync.RWMutex
}

type BucketMap struct {
	bkts []bucket
	pool *sync.Pool
	rand Rand
}

func NewBucketMap(rand Rand, cap ...int) *BucketMap {
	const Size = 31
	var n int = Size
	if len(cap) > 0 && cap[0] > 0 {
		n = cap[0]
	}
	bm := &BucketMap{
		bkts: make([]bucket, n),
		pool: &sync.Pool{
			New: func() interface{} {
				return make([]bool, n)
			},
		},
		rand: rand,
	}
	for i := 0; i < len(bm.bkts); i++ {
		bm.bkts[i].m = make(map[string]interface{})
	}
	return bm
}

func (bm *BucketMap) index(k string) uint32 {
	return crc32.ChecksumIEEE([]byte(k)) % uint32(len(bm.bkts))
}

func (bm *BucketMap) Store(k string, v interface{}) {
	bkt := &bm.bkts[bm.index(k)]
	bkt.l.Lock()
	if v == nil {
		delete(bkt.m, k)
	} else {
		bkt.m[k] = v
	}
	bkt.l.Unlock()
}

func (bm *BucketMap) Load(k string) interface{} {
	bkt := &bm.bkts[bm.index(k)]
	bkt.l.RLock()
	v := bkt.m[k]
	bkt.l.RUnlock()
	return v
}

func (bm *BucketMap) LoadOrStore(k string, new func() interface{}) interface{} {
	bkt := &bm.bkts[bm.index(k)]
	bkt.l.Lock()
	v := bkt.m[k]
	if v == nil {
		v = new()
		if v != nil {
			bkt.m[k] = v
		}
	}
	bkt.l.Unlock()
	return v
}

var BucketDeleteFunc = func(interface{}) interface{} { return nil }

func (bm *BucketMap) Replace(k string, f func(old interface{}) interface{}) {
	bkt := &bm.bkts[bm.index(k)]
	bkt.l.Lock()
	v := f(bkt.m[k])
	if v == nil {
		delete(bkt.m, k)
	} else {
		bkt.m[k] = v
	}
	bkt.l.Unlock()
}

func (bm *BucketMap) Range(f func(key string, value interface{}) bool) {
	scanned := bm.pool.Get().([]bool)
	for n := 0; n < len(scanned); n++ {
		scanned[n] = false
	}

	for n := 0; n < len(scanned); n++ {
		var i int
		for {
			i = bm.rand.Intn(len(scanned))
			if !scanned[i] {
				scanned[i] = true
				break
			}
		}

		exit := false
		bkt := &bm.bkts[i]
		bkt.l.RLock()
		for k, v := range bkt.m {
			if !f(k, v) {
				exit = true
				break
			}
		}
		bkt.l.RUnlock()
		if exit {
			break
		}
	}

	bm.pool.Put(scanned)
}

func (bm *BucketMap) RangeReplace(f func(key string, value interface{}) interface{}) {
	scanned := bm.pool.Get().([]bool)
	for n := 0; n < len(scanned); n++ {
		scanned[n] = false
	}

	for n := 0; n < len(scanned); n++ {
		var i int
		for {
			i = bm.rand.Intn(len(scanned))
			if !scanned[i] {
				scanned[i] = true
				break
			}
		}

		bkt := &bm.bkts[i]
		bkt.l.Lock()
		for k, v := range bkt.m {
			new := f(k, v)
			if new == nil {
				delete(bkt.m, k)
			} else {
				bkt.m[k] = new
			}
		}
		bkt.l.Unlock()
	}
	bm.pool.Put(scanned)
}
