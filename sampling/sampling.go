package sampling

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Iterator 遍历一个容器中所有元素
type Iterator interface {
	Range(func(interface{}))
}

// Writer 将元素写入容器，
// 如果index等于容器size，则为append操作；
// 如果index小于容器size，为替换操作；
// 确保index不会大于容器size。
type Writer interface {
	Store(int, interface{})
}

// Sampling 蓄水池算法，
// 参数n控制最后留下n个元素；
// 如果Iterator不足n个，则将Iterator全部元素拷贝到Writer。
func Sampling(w Writer, it Iterator, n int) {
	if n <= 0 {
		return
	}

	var i int
	it.Range(func(v interface{}) {
		i++
		if i <= n {
			w.Store(i-1, v)
		} else {
			k := rand.Intn(i)
			if k < n {
				w.Store(k, v)
			}
		}
	})
}
