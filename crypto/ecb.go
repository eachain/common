package crypto

import "crypto/cipher"

type ecb struct {
	b         cipher.Block
	blockSize int
	f         func(dst, src []byte) // encrypt or decrypt function
}

func newECB(b cipher.Block, f func(dst, src []byte)) *ecb {
	return &ecb{
		b:         b,
		blockSize: b.BlockSize(),
		f:         f,
	}
}

func (x *ecb) BlockSize() int { return x.blockSize }

func (x *ecb) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		panic("crypto: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("crypto: output smaller than input")
	}
	if InexactOverlap(dst[:len(src)], src) {
		panic("crypto/cipher: invalid buffer overlap")
	}

	for len(src) > 0 {
		x.f(dst[:x.blockSize], src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

type ecbEncAble interface {
	NewECBEncrypter() cipher.BlockMode
}

func NewECBEncrypter(b cipher.Block) cipher.BlockMode {
	if ecb, ok := b.(ecbEncAble); ok {
		return ecb.NewECBEncrypter()
	}
	return newECB(b, b.Encrypt)
}

type ecbDecAble interface {
	NewECBDecrypter() cipher.BlockMode
}

func NewECBDecrypter(b cipher.Block) cipher.BlockMode {
	if ecb, ok := b.(ecbDecAble); ok {
		return ecb.NewECBDecrypter()
	}
	return newECB(b, b.Decrypt)
}
