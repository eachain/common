package utils

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type lockedSource struct {
	lk  sync.Mutex
	src rand.Source
}

type Random struct {
	index uint32
	srcs  []lockedSource
}

func NewRandom() *Random {
	const Size = 32
	r := &Random{
		srcs: make([]lockedSource, Size),
	}
	now := time.Now().UnixNano()
	for i := 0; i < len(r.srcs); i++ {
		r.srcs[i].src = rand.NewSource(now + int64(i))
	}
	return r
}

func (r *Random) Intn(n int) int {
	if n <= 0 {
		panic("invalid argument to Intn")
	}

	ls := &r.srcs[atomic.AddUint32(&r.index, 1)%uint32(len(r.srcs))]

	var x int64
	ls.lk.Lock()
	if n&(n-1) == 0 { // n is power of two, can mask
		x = ls.src.Int63() & int64(n-1)
	} else {
		max := int64((1 << 63) - 1 - (1<<63)%uint64(n))
		v := ls.src.Int63()
		for v > max {
			v = ls.src.Int63()
		}
		x = v % int64(n)
	}
	ls.lk.Unlock()
	return int(x)
}
