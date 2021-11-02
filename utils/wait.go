package utils

import (
	"context"
	"sync"
)

type WaitContext struct {
	once   sync.Once
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func (wc *WaitContext) init() {
	wc.once.Do(func() {
		wc.ctx, wc.cancel = context.WithCancel(context.Background())
	})
}

func (wc *WaitContext) Add(n int) {
	wc.wg.Add(n)
}

func (wc *WaitContext) Done() {
	wc.wg.Done()
}

func (wc *WaitContext) Wait() {
	wc.wg.Wait()
}

func (wc *WaitContext) Cancel() {
	wc.init()
	wc.cancel()
}

func (wc *WaitContext) Canceled() bool {
	wc.init()
	select {
	case <-wc.ctx.Done():
		return true
	default:
		return false
	}
}

func (wc *WaitContext) Signal() <-chan struct{} {
	wc.init()
	return wc.ctx.Done()
}
