package utils

import (
	"context"
	"time"
)

/*
CtxDone 将<-ctx.Done()的判断以bool返回，
用于在for循环里面判断退出条件。

	for !CtxDone(ctx) {
		...
	}

也可用作函数里面:

	if CtxDone(ctx) {
		return ...
	}
*/
func CtxDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

/*
Retry用于出错重试，等待时间从50ms一直倍增，最大到10s。
比如：

	var ctx = context.Background()
	var conn net.Conn
	var err error
	err = Retry(ctx, func() error {
		conn, err = (&net.Dialer{}).DialContext(ctx, "tcp", "www.baidu.com:443")
		return err
	}, 5)
	if err != nil {
		...
	}
	_ = conn
*/
func Retry(ctx context.Context, fn func() error, n int) (err error) {
	var wait = 50 * time.Millisecond
	const maxWait = 10 * time.Second

	for i := 0; i < n; i++ {
		err = fn()
		if err == nil {
			return nil
		}

		if i < n-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
				wait *= 2
				if wait > maxWait {
					wait = maxWait
				}
			}
		}
	}
	return
}
