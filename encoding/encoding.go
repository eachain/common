package encoding

import (
	"fmt"
	"math"
	"math/big"
)

const (
	base62Std = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	base58Std = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
)

var (
	Base62 = NewEncoding(base62Std)
	Base58 = NewEncoding(base58Std)
)

type Encoding struct {
	tbl    [256]byte
	encode []byte
	radix  *big.Int
	zero   *big.Int
	multi  int
}

func NewEncoding(encode string) *Encoding {
	enc := &Encoding{
		encode: make([]byte, len(encode)),
		radix:  big.NewInt(int64(len(encode))),
		zero:   big.NewInt(0),
		multi:  int(8/math.Log2(float64(len(encode)))*100) + 1,
	}
	copy(enc.encode, encode)
	for i := 0; i < len(enc.tbl); i++ {
		enc.tbl[i] = 0xFF
	}
	for i, b := range enc.encode {
		enc.tbl[b] = byte(i)
	}
	return enc
}

func (enc *Encoding) Decode(b string) ([]byte, error) {
	answer := big.NewInt(0)
	j := big.NewInt(1)

	scratch := new(big.Int)
	for i := len(b) - 1; i >= 0; i-- {
		tmp := enc.tbl[b[i]]
		if tmp == 0xFF {
			return nil, fmt.Errorf("base%v: invalid character '%v'(%v)",
				len(enc.encode), b[i], i)
		}
		scratch.SetInt64(int64(tmp))

		//scratch = j*scratch
		scratch.Mul(j, scratch)

		answer.Add(answer, scratch)
		j.Mul(j, enc.radix)
	}

	tmpval := answer.Bytes()

	var numZeros int
	for numZeros = 0; numZeros < len(b); numZeros++ {
		if b[numZeros] != enc.encode[0] {
			break
		}
	}
	//得到原来数字的长度
	flen := numZeros + len(tmpval)
	//构造一个新地存放结果的空间
	val := make([]byte, flen, flen)
	copy(val[numZeros:], tmpval)

	return val, nil
}

// Encode encodes a byte slice to a modified base58 string.
func (enc *Encoding) EncodeToString(b []byte) string {
	x := new(big.Int)
	x.SetBytes(b)

	answer := make([]byte, 0, len(b)*enc.multi/100)

	mod := new(big.Int)
	for x.Cmp(enc.zero) > 0 {
		x.DivMod(x, enc.radix, mod)
		answer = append(answer, enc.encode[mod.Int64()])
	}

	for _, i := range b {
		if i != 0 {
			break
		}
		answer = append(answer, enc.encode[0])
	}

	alen := len(answer)
	for i := 0; i < alen/2; i++ {
		answer[i], answer[alen-1-i] = answer[alen-1-i], answer[i]
	}

	return string(answer)
}
