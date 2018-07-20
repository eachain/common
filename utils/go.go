package utils

import (
	"context"
	"sync"
)

type routineGroup struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

var (
	gomutex    sync.Mutex
	goroutines = make(map[string]*routineGroup)
)

/*
Go 用于协程管理，可以用于后台协连接池管理等。

	Go("biz_name", func(ctx context.Context) {
		for !CtxDone(ctx) {
			...
		}
	})

参考StopAndWait组合用法。
*/
func Go(biz string, fn func(context.Context)) {
	gomutex.Lock()
	defer gomutex.Unlock()

	rg := goroutines[biz]
	if rg == nil {
		ctx, cancel := context.WithCancel(context.Background())
		rg = &routineGroup{
			ctx:    ctx,
			cancel: cancel,
			wg:     &sync.WaitGroup{},
		}
		goroutines[biz] = rg
	}

	rg.wg.Add(1)
	go func() {
		defer rg.wg.Done()
		fn(rg.ctx)
	}()
}

/*
StopAndWait 用于停掉所有biz的协程，并等待退出。biz必须和Go(biz, fn)里面的biz一致。
Go和StopAndWait经常被用于各种自治组件，比如连接池、任务池等管理。
例如：

	type Pool struct {
		biz string
		... // contains other fields
	}

	func NewPool(...) *Pool {
		pool := &Pool{biz: "unique_biz_name", ...}
		Go(pool.biz, pool.run)
		return pool
	}

	func (pool *Pool) run(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(3 * time.Minute):
				pool.refresh(ctx)
			}
		}
	}

	func (pool *Pool) refresh(ctx context.Context) {
		...
	}

	func (pool *Pool) Stop() {
		StopAndWait(pool.biz)
		// clear the pool
		...
	}
*/
func StopAndWait(biz string) {
	gomutex.Lock()
	rg := goroutines[biz]
	gomutex.Unlock()

	if rg == nil {
		return
	}

	rg.cancel()
	rg.wg.Wait()
}
