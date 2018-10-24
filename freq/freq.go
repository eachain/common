package freq

import (
	"sync"
	"time"
)

// Option 是规则条件集。
// 每Dur时间内，允许通过Num个请求。
// 如果超过，则进入Punish罚时；
// 进入罚时后，每Tick时间允许通过一个。
type Option struct {
	Dur    time.Duration
	Num    uint64
	Punish time.Duration
	Tick   time.Duration
}

// Cache 是redis命令的一个子集
type Cache interface {
	Len() uint64
	PushEX(ex time.Duration) uint64
	PushXX() uint64
	Expire(ex time.Duration)
	SetEX(ex time.Duration) bool
}

type Freq struct {
	opt   *Option
	cache Cache
}

func NewFreq(opt *Option, cache Cache) *Freq {
	o := *opt
	freq := &Freq{opt: &o, cache: cache}
	return freq
}

func (freq *Freq) Pass() bool {
	before := freq.cache.Len()
	var after uint64
	if before == 0 {
		after = freq.cache.PushEX(freq.opt.Dur)
	} else if before <= freq.opt.Num {
		after = freq.cache.PushXX()
	}
	if before < freq.opt.Num && after <= freq.opt.Num {
		return true
	}
	if before == freq.opt.Num || after > freq.opt.Num {
		if freq.opt.Punish > 0 {
			freq.cache.Expire(freq.opt.Punish)
		}
	}
	if freq.opt.Tick > 0 {
		return freq.cache.SetEX(freq.opt.Tick)
	} else {
		return false
	}
}

type MultiFreq struct {
	opt  *Option
	new  func(string) Cache
	mut  sync.Mutex
	freq map[string]*Freq
}

func NewMultiFreq(common *Option, newCache func(string) Cache) *MultiFreq {
	opt := *common
	return &MultiFreq{
		opt:  &opt,
		new:  newCache,
		freq: make(map[string]*Freq),
	}
}

func (mf *MultiFreq) Pass(key string) bool {
	mf.mut.Lock()
	freq := mf.freq[key]
	if freq == nil {
		freq = NewFreq(mf.opt, mf.new(key))
		mf.freq[key] = freq
	}
	mf.mut.Unlock()
	return freq.Pass()
}
