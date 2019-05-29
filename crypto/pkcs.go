// The difference between the PKCS#5 and PKCS#7 padding mechanisms is the block size; PKCS#5 padding is defined for 8-byte block sizes, PKCS#7 padding would work for any block size from 1 to 255 bytes.
// This is the definition of PKCS#5 padding (6.2) as defined in the RFC:
// The padding string PS shall consist of 8 - (||M|| mod 8) octets all having value 8 - (||M|| mod 8).
// The RFC that contains the PKCS#7 standard is the same except that it allows block sizes up to 255 bytes in size (10.3 note 2):
// For such algorithms, the method shall be to pad the input at the trailing end with k - (l mod k) octets all having value k - (l mod k), where l is the length of the input.
// So fundamentally PKCS#5 padding is a subset of PKCS#7 padding for 8 byte block sizes. Hence, PKCS#5 padding can not be used for AES. PKCS#5 padding was only defined with (triple) DES operation in mind.
// Many cryptographic libraries use an identifier indicating PKCS#5 or PKCS#7 to define the same padding mechanism. The identifier should indicate PKCS#7 if block sizes other than 8 are used within the calculation. Some cryptographic libraries such as the SUN provider in Java indicate PKCS#5 where PKCS#7 should be used - "PKCS5Padding" should have been "PKCS7Padding". This is a legacy from the time that only 8 byte block ciphers such as (triple) DES symmetric cipher were available.
// Note that neither PKCS#5 nor PKCS#7 is a standard created to describe a padding mechanism. The padding part is only a small subset of the defined functionality. PKCS#5 is a standard for Password Based Encryption or PBE, and PKCS#7 defines the Cryptographic Message Syntax or CMS.
package crypto

import "bytes"

func PKCS7Padding(data []byte, blockSize int) []byte {
	n := blockSize - len(data)%blockSize
	padded := bytes.Repeat([]byte{byte(n)}, n)
	return append(data, padded...)
}

func PKCS7UnPadding(data []byte) []byte {
	n := int(data[len(data)-1])
	if n == 0 || n > len(data) {
		panic("pkcs7 unpadding: invalid padding number")
	}
	return data[:len(data)-n]
}

/*
func PKCS5Padding(data []byte) []byte {
	return PKCS7Padding(data, 8)
}

func PKCS5UnPadding(data []byte) []byte {
	n := int(data[len(data)-1])
	if n == 0 || n > 8 || n > len(data) {
		panic("pkcs5 unpadding: invalid padding number")
	}
	return data[:len(data)-n]
}
*/
