package jcon

// 解析json注释
// 允许json中出现单行注释"//..."
// 和多行注释"/*...*/"

import (
	"bytes"
	"io"
	"io/ioutil"
)

type reader struct {
	r io.Reader

	prefix          byte
	inString        bool
	inSingleComment bool
	inMultiComment  bool
}

func NewReader(r io.Reader) io.Reader {
	return &reader{r: r}
}

func (r *reader) Read(b []byte) (int, error) {
	n, err := r.r.Read(b)
	l := 0
	for i := 0; i < n; i++ {
		if r.inSingleComment {
			if b[i] == '\n' || b[i] == '\r' {
				r.inSingleComment = false
				r.prefix = 0
			} else {
				continue
			}
		}

		if r.inMultiComment {
			if r.prefix == '*' && b[i] == '/' {
				r.inMultiComment = false
				r.prefix = 0
			} else {
				r.prefix = b[i]
			}
			continue
		}

		if b[i] == '"' {
			r.inString = !r.inString
		}
		if !r.inString {
			if r.prefix == '/' {
				if b[i] == '/' {
					r.inSingleComment = true
					r.prefix = 0
					l--
					continue
				}
				if b[i] == '*' {
					r.inMultiComment = true
					r.prefix = 0
					l--
					continue
				}
			}
		}

		r.prefix = b[i]
		if l < i {
			b[l] = b[i]
		}
		l++
	}
	return l, err
}

func TrimComment(p []byte) []byte {
	b, _ := ioutil.ReadAll(NewReader(bytes.NewReader(p)))
	return b
}
